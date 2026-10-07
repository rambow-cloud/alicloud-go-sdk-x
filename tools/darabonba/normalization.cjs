"use strict";

const fs = require("node:fs");
const path = require("node:path");
const crypto = require("node:crypto");
const sha = (data) => crypto.createHash("sha256").update(data).digest("hex");
function requireFact(ok, message) {
  if (!ok) throw new Error("normalization: " + message);
}
function verifyCanonical(
  root = path.resolve(__dirname, "../../sources/openapi-meta"),
) {
  requireFact(!fs.lstatSync(root).isSymbolicLink(), "symlink source root");
  requireFact(
    !fs.lstatSync(path.join(root, "manifest.json")).isSymbolicLink(),
    "symlink source manifest",
  );
  const raw = fs.readFileSync(path.join(root, "manifest.json"));
  const manifest = JSON.parse(raw);
  requireFact(
    manifest.schemaVersion === 1 &&
      manifest.adapterVersion === 1 &&
      manifest.repository === "https://github.com/aliyun/aliyun-openapi-meta" &&
      /^[a-f0-9]{40}$/.test(manifest.revision) &&
      manifest.license === "Apache-2.0" &&
      Array.isArray(manifest.files) &&
      manifest.files.length > 0,
    "unsupported source manifest",
  );
  const seen = new Set();
  for (const file of manifest.files) {
    requireFact(
      typeof file.file === "string" &&
        !file.file.includes("\\") &&
        !path.isAbsolute(file.file) &&
        file.file.split("/").every((p) => p && p !== "." && p !== "..") &&
        (file.file === "LICENSE" || file.file.startsWith("canonical/")) &&
        !seen.has(file.file),
      "unsafe/duplicate source path",
    );
    seen.add(file.file);
    let target = root;
    for (const part of file.file.split("/")) {
      target = path.join(target, part);
      requireFact(
        !fs.lstatSync(target).isSymbolicLink(),
        "symlink source path",
      );
    }
    requireFact(
      file.url ===
        `https://raw.githubusercontent.com/aliyun/aliyun-openapi-meta/${manifest.revision}/${file.file}`,
      "source URL differs from pinned revision",
    );
    requireFact(
      sha(fs.readFileSync(target)) === file.sha256,
      "source checksum mismatch: " + file.file,
    );
  }
  requireFact(seen.has("LICENSE"), "missing license artifact");
  function inventory(dir) {
    for (const entry of fs.readdirSync(path.join(root, dir), {
      withFileTypes: true,
    })) {
      const relative = dir + "/" + entry.name;
      requireFact(!entry.isSymbolicLink(), "symlink source inventory");
      if (entry.isDirectory()) inventory(relative);
      else requireFact(seen.has(relative), "unlisted source: " + relative);
    }
  }
  inventory("canonical");
  return { manifest, hash: sha(raw) };
}

// Source annotations never become codecs, validators, names or retry policies.
function annotations(source, structural) {
  return Object.fromEntries(
    Object.entries(source).filter(([key]) => !structural.includes(key)),
  );
}
function name(value) {
  requireFact(
    typeof value === "string" && value.length > 0 && !value.includes("."),
    "invalid wire member name",
  );
  return value;
}
function requestShape(source, depth = 0) {
  requireFact(source && depth < 32, "missing/deep request shape");
  const type = { bool: "boolean", int: "integer" }[source.type] || source.type;
  requireFact(
    ["string", "integer", "number", "boolean", "object", "array"].includes(
      type,
    ),
    "unsupported request type",
  );
  requireFact(
    source.required === undefined || typeof source.required === "boolean",
    "invalid requiredness",
  );
  const result = {
    type,
    required: !!source.required,
    attributes: annotations(source, ["type", "required", "fields", "element"]),
  };
  if (type === "array") result.items = requestShape(source.element, depth + 1);
  if (type === "object") {
    requireFact(Array.isArray(source.fields), "missing request fields");
    const entries = source.fields.map((field) => [
      name(field.raw_name),
      requestShape(field, depth + 1),
    ]);
    requireFact(
      new Set(entries.map(([key]) => key)).size === entries.length,
      "duplicate wire field",
    );
    result.properties = Object.fromEntries(entries);
  }
  return result;
}
function responseShape(source, depth = 0) {
  requireFact(source && depth < 32, "missing/deep response shape");
  requireFact(
    ["string", "integer", "number", "boolean", "object", "array"].includes(
      source.type,
    ),
    "unsupported response type",
  );
  const result = {
    type: source.type,
    attributes: annotations(source, [
      "type",
      "items",
      "properties",
      "itemName",
    ]),
  };
  requireFact(
    !Object.hasOwn(source, "itemName") || source.type === "array",
    "itemName on non-array",
  );
  if (source.type === "object") {
    requireFact(
      source.properties &&
        typeof source.properties === "object" &&
        !Array.isArray(source.properties),
      "missing response properties",
    );
    result.properties = Object.fromEntries(
      Object.entries(source.properties).map(([key, value]) => [
        name(key),
        responseShape(value, depth + 1),
      ]),
    );
  }
  if (source.type === "array") {
    result.items = responseShape(source.items, depth + 1);
    if (Object.hasOwn(source, "itemName")) {
      const member = name(source.itemName);
      return {
        type: "object",
        properties: { [member]: { type: "array", items: result.items } },
        attributes: { ...result.attributes, itemName: member },
      };
    }
  }
  return result;
}
function normalizeCanonical(document, provenance) {
  const operation = document.operation;
  requireFact(
    operation &&
      typeof operation.action === "string" &&
      typeof operation.api_version === "string" &&
      operation.api_style === "RPC" &&
      operation.method === "POST" &&
      operation.protocol === "HTTPS",
    "unsupported canonical profile",
  );
  requireFact(
    Array.isArray(document.parameters),
    "missing canonical parameters",
  );
  const inputs = document.parameters
    .map((parameter) => {
      requireFact(
        typeof parameter.raw_name === "string" &&
          parameter.raw_name.length > 0 &&
          parameter.location === "query",
        "invalid query binding",
      );
      requireFact(
        parameter.param_style === undefined ||
          parameter.param_style === "repeatList",
        "unsupported parameter style",
      );
      requireFact(
        parameter.param_style !== "repeatList" || parameter.type === "array",
        "repeatList on non-array",
      );
      return {
        wire: parameter.raw_name,
        location: parameter.location,
        style: parameter.param_style || "",
        schema: requestShape(parameter),
        cli: { name: parameter.name, options: parameter.options || [] },
      };
    })
    .sort((a, b) => (a.wire < b.wire ? -1 : a.wire > b.wire ? 1 : 0));
  requireFact(
    new Set(inputs.map((i) => i.wire)).size === inputs.length,
    "duplicate query binding",
  );
  return {
    schemaVersion: 1,
    adapterVersion: 1,
    provenance,
    protocol: {
      action: operation.action,
      version: operation.api_version,
      style: operation.api_style,
      method: operation.method,
      protocol: operation.protocol,
    },
    inputs,
    response: responseShape(document.responses?.["200"]?.schema),
    attributes: annotations(document, ["operation", "parameters", "responses"]),
    operationAttributes: annotations(operation, [
      "action",
      "api_version",
      "api_style",
      "method",
      "protocol",
    ]),
  };
}

// Match literal indexed wire paths against structural arrays; indices are not limits.
function indexedLeaf(root, suffix) {
  let shape = root;
  const parts = suffix.split(".");
  requireFact(parts.length <= 32, "deep indexed binding");
  for (const part of parts) {
    if (shape?.type === "array") {
      requireFact(/^[1-9][0-9]*$/.test(part), "invalid array index: " + suffix);
      shape = shape.items;
    } else {
      requireFact(
        shape?.type === "object" && Object.hasOwn(shape.properties || {}, part),
        "wire path/case differs: " + suffix,
      );
      shape = shape.properties[part];
    }
  }
  requireFact(shape, "missing indexed leaf");
  return shape;
}
function compatibleShape(
  metadata,
  wire,
  depth = 0,
  location = "",
  missing = null,
) {
  requireFact(
    metadata &&
      wire &&
      depth < 32 &&
      metadata.type === wire.type &&
      wire.type !== "unsupported",
    "wire type differs",
  );
  if (wire.type === "array")
    compatibleShape(
      metadata.items,
      wire.items,
      depth + 1,
      location + "[]",
      missing,
    );
  if (metadata.type === "object") {
    for (const [key, value] of Object.entries(metadata.properties || {})) {
      const member = location + "." + key;
      if (
        !Object.hasOwn(wire.properties || {}, key) &&
        missing &&
        !value.required
      ) {
        requireFact(
          ["string", "integer", "number", "boolean"].includes(value.type),
          "unsupported metadata-only member: " + member,
        );
        missing[member] = value.type;
        continue;
      }
      requireFact(
        Object.hasOwn(wire.properties || {}, key),
        "wire path/case differs: " + member,
      );
      compatibleShape(value, wire.properties[key], depth + 1, member, missing);
    }
  }
}
function bindInputs(inputs, parameters) {
  const seen = new Set();
  for (const p of parameters) {
    requireFact(
      typeof p.name === "string" &&
        !seen.has(p.name) &&
        p.in === "query" &&
        p.schema,
      "invalid/duplicate metadata binding",
    );
    seen.add(p.name);
  }
  const bindings = new Map();
  bindings.metadataOnlyFields = Object.create(null);
  for (const input of inputs) {
    const direct = parameters.find((p) => p.name === input.wire);
    const aliases = parameters.filter((p) =>
      p.name.startsWith(input.wire + "."),
    );
    requireFact(
      !direct || aliases.length === 0,
      "ambiguous direct/indexed binding: " + input.wire,
    );
    if (direct)
      compatibleShape(
        direct.schema,
        input.schema,
        0,
        input.wire,
        bindings.metadataOnlyFields,
      );
    for (const p of aliases) {
      const leaf = indexedLeaf(
        input.schema,
        p.name.slice(input.wire.length + 1),
      );
      compatibleShape(p.schema, leaf);
      requireFact(
        !!p.schema.required === !!leaf.required,
        "indexed requiredness differs: " + p.name,
      );
    }
    if (aliases.length)
      requireFact(
        !input.schema.required,
        "required indexed root lacks API requiredness: " + input.wire,
      );
    bindings.set(input.wire, {
      direct,
      aliases: aliases.map((p) => p.name).sort(),
    });
  }
  return bindings;
}
module.exports = {
  verifyCanonical,
  normalizeCanonical,
  requestShape,
  responseShape,
  indexedLeaf,
  compatibleShape,
  bindInputs,
};
