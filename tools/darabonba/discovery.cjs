"use strict";

// Complete product discovery is independent of legacy snapshots, overlays and decisions.
const fs = require("node:fs");
const path = require("node:path");
const crypto = require("node:crypto");
const parser = require("@darabonba/parser");
const { verifySources, lowerOperation } = require("./frontend.cjs");
const repository = path.resolve(__dirname, "../..");
const lex = (token) => token?.lexeme;
const order = (a, b) => (a < b ? -1 : a > b ? 1 : 0);
const sha = (data) => crypto.createHash("sha256").update(data).digest("hex");
function requireInventory(ok, message) {
  if (!ok) throw new Error("discovery: " + message);
}
function source(file, token) {
  const loc = token?.loc;
  return {
    file,
    ...(loc
      ? {
          line: loc.start.line,
          column: loc.start.column,
          endLine: loc.end.line,
          endColumn: loc.end.column,
        }
      : {}),
  };
}
// Consume the official parser's annotation token; prose never affects lowering.
function annotationDocuments(node, file) {
  if (!node.annotation) return [];
  const lines = node.annotation.value
    .replace(/^\/\*\*/, "")
    .replace(/\*\/$/, "")
    .split(/\r?\n/)
    .map((line) => line.replace(/^\s*\* ?/, ""));
  const docs = [];
  let current;
  for (const line of lines) {
    const tag = line.match(/^\s*@([A-Za-z]+)\b\s*(.*)$/);
    if (tag) {
      current = ["summary", "description"].includes(tag[1])
        ? {
            attribute: tag[1],
            text: tag[2],
            source: source(file, node.annotation),
          }
        : null;
      if (current) docs.push(current);
    } else if (current) current.text += "\n" + line;
  }
  return docs.map((doc) => ({ ...doc, text: doc.text.trim() }));
}
function walk(node, visit, seen = new WeakSet()) {
  if (!node || typeof node !== "object" || seen.has(node)) return;
  seen.add(node);
  visit(node);
  for (const value of Object.values(node)) walk(value, visit, seen);
}
function protocolEvidence(fn, file) {
  const constructors = [];
  walk(fn.functionBody || fn.apiBody, (node) => {
    if (
      node.type !== "construct_model" ||
      lex(node.aliasId) !== "OpenApi" ||
      node.propertyPath?.length !== 1 ||
      lex(node.propertyPath[0]) !== "Params"
    )
      return;
    constructors.push({
      source: source(file, node.aliasId),
      fields: (node.object?.fields || []).map((field) => ({
        name: lex(field.fieldName),
        source: source(file, field.fieldName),
        expressionKind: field.expr?.type || "missing",
        ...(field.expr?.type === "string"
          ? { value: field.expr.value.string }
          : {}),
      })),
    });
  });
  return constructors;
}
function bindingEvidence(fn, file) {
  const bindings = [];
  walk(fn.functionBody || fn.apiBody, (node) => {
    const left = node.left;
    if (
      node.type !== "assign" ||
      left?.type !== "map_access" ||
      !["query", "body", "headers", "path"].includes(lex(left.id))
    )
      return;
    const expression = node.expr;
    bindings.push({
      target: lex(left.id),
      source: source(file, left.id),
      ...(left.accessKey?.type === "string"
        ? { wire: left.accessKey.value.string }
        : { dynamicKey: true }),
      expressionKind: expression?.type || "missing",
      ...(expression?.type === "property_access" &&
      lex(expression.id) === "request"
        ? { requestPath: expression.propertyPath.map(lex) }
        : {}),
      ...(expression?.left?.type === "static_call"
        ? {
            helper: [
              lex(expression.left.id),
              ...(expression.left.propertyPath || []).map(lex),
            ].join("."),
          }
        : {}),
    });
  });
  return bindings;
}
function discoverCandidates(ast, info, file) {
  const candidates = new Map();
  function candidate(name) {
    requireInventory(
      typeof name === "string" && /^[A-Za-z][A-Za-z0-9]*$/.test(name),
      "invalid operation identity",
    );
    if (!candidates.has(name))
      candidates.set(name, { name, declarations: [], catalog: false });
    return candidates.get(name);
  }
  for (const node of ast.moduleBody.nodes) {
    if (!["function", "api"].includes(node.type)) continue;
    const sdkName = lex(node.functionName || node.apiName);
    const evidence = protocolEvidence(node, file);
    const withOptions = sdkName?.endsWith("WithOptions");
    if (!withOptions && node.type !== "api" && evidence.length === 0) continue;
    const actions = [
      ...new Set(
        evidence.flatMap((c) =>
          c.fields
            .filter((f) => f.name === "action" && Object.hasOwn(f, "value"))
            .map((f) => f.value),
        ),
      ),
    ].sort(order);
    const base = withOptions
      ? sdkName.slice(0, -"WithOptions".length)
      : sdkName;
    const fallback = base ? base[0].toUpperCase() + base.slice(1) : "";
    // Ambiguous actions remain one unsupported declaration, never several claimed APIs.
    const name = actions.length === 1 ? actions[0] : fallback;
    candidate(name).declarations.push({
      node,
      sdkName,
      evidence,
      actions,
      bindings: bindingEvidence(node, file),
      source: source(file, node.functionName || node.apiName),
    });
  }
  if (info) {
    requireInventory(
      Array.isArray(info.apiDoc?.hasDoc) && Array.isArray(info.apiDoc?.noDoc),
      "malformed upstream API catalog",
    );
    const listed = [...info.apiDoc.hasDoc, ...info.apiDoc.noDoc];
    requireInventory(
      new Set(listed).size === listed.length,
      "duplicate upstream catalog entry",
    );
    for (const name of listed) candidate(name).catalog = true;
  }
  return [...candidates.values()].sort((a, b) => order(a.name, b.name));
}

function modelGraph(ast, file) {
  const declared = new Map();
  for (const node of ast.moduleBody.nodes.filter((n) => n.type === "model")) {
    const name = lex(node.modelName);
    requireInventory(
      !declared.has(name),
      "duplicate model declaration: " + name,
    );
    declared.set(name, node);
  }
  const models = new Map();
  function register(id, body, token, named = false, extendsType = null) {
    if (models.has(id)) return;
    const entry = { id, named, source: source(file, token), fields: [] };
    models.set(id, entry);
    if (extendsType) entry.extends = type(extendsType, id + ".$base", token);
    const wireNames = new Set();
    for (const field of body.nodes || []) {
      requireInventory(
        field.type === "modelField",
        "unknown model member: " + id,
      );
      const attributes = {},
        documentation = [];
      for (const attr of field.attrs || []) {
        const name = lex(attr.attrName);
        requireInventory(
          !Object.hasOwn(attributes, name) &&
            !documentation.some((a) => a.attribute === name),
          "duplicate model attribute: " + id,
        );
        if (["description", "example"].includes(name)) {
          documentation.push({
            attribute: name,
            source: source(file, attr.attrValue),
            ...(name === "description" ? { text: attr.attrValue.string } : {}),
          });
        } else {
          const literal = attr.attrValue;
          attributes[name] =
            name === "deprecated" && ["true", "false"].includes(lex(literal))
              ? lex(literal) === "true"
              : (literal.string ?? literal.value ?? lex(literal) ?? null);
        }
      }
      const dslName = lex(field.fieldName),
        wireName = attributes.name ?? dslName;
      requireInventory(
        typeof wireName === "string" &&
          wireName.length > 0 &&
          !wireNames.has(wireName),
        "duplicate/invalid wire model member: " + id,
      );
      wireNames.add(wireName);
      entry.fields.push({
        dslName,
        wireName,
        required: !!field.required,
        source: source(file, field.fieldName),
        attributes,
        ...(documentation.length ? { documentation } : {}),
        type: type(field.fieldValue, id + "." + dslName, field.fieldName),
      });
    }
    if (body.extendFileds?.length) entry.inheritedFields = true;
  }
  function type(value, id, token) {
    if (!value) return { kind: "unsupported", dslType: "missing" };
    if (value.type === "modelBody") {
      register(id, value, token);
      return { kind: "model", ref: id };
    }
    if (["moduleModel", "subModel"].includes(value.type)) {
      return {
        kind: "external",
        dslType: (value.path || []).map(lex).join(".") || "unknown",
        source: source(file, value),
      };
    }
    let kind = value.fieldType ?? lex(value);
    if (kind && typeof kind === "object") {
      if (kind.type === "moduleModel") return type(kind, id, token);
      kind = lex(kind);
    }
    if (kind === "array")
      return {
        kind: "array",
        items: type(value.fieldItemType, id + "[]", token),
      };
    if (kind === "map")
      return {
        kind: "map",
        keys: type(value.keyType, id + ".key", token),
        values: type(value.valueType, id + ".value", token),
      };
    if (declared.has(kind)) {
      const model = declared.get(kind);
      register(kind, model.modelBody, model.modelName, true, model.extendOn);
      return { kind: "model", ref: kind };
    }
    const integer = [
      "integer",
      "number",
      "int8",
      "int16",
      "int32",
      "int64",
      "long",
      "ulong",
      "uint8",
      "uint16",
      "uint32",
      "uint64",
    ];
    if (kind === "any") return { kind: "json", dslType: "any" };
    const wireType = integer.includes(kind)
      ? "integer"
      : ["float", "double"].includes(kind)
        ? "number"
        : ["string", "boolean"].includes(kind)
          ? kind
          : null;
    return wireType
      ? { kind: "scalar", dslType: kind, wireType }
      : { kind: "unsupported", dslType: kind || value.type || "unknown" };
  }
  function reachable(roots) {
    const refs = new Set();
    function visit(t) {
      if (!t) return;
      if (t.kind === "model" && !refs.has(t.ref)) {
        refs.add(t.ref);
        const model = models.get(t.ref);
        if (model) {
          visit(model.extends);
          for (const field of model.fields) visit(field.type);
        }
      }
      if (t.kind === "array") visit(t.items);
      if (t.kind === "map") {
        visit(t.keys);
        visit(t.values);
      }
    }
    for (const root of roots) visit(root);
    return [...refs].sort(order);
  }
  return { declared, models, type, reachable };
}

function shapeIssue(shape, location, allowJSON = false) {
  if (!shape || shape.type === "unsupported") return location;
  if (shape.type === "json") return allowJSON ? null : location;
  if (shape.type === "map")
    return shapeIssue(shape.values, location + ".*", allowJSON);
  if (shape.type === "array")
    return shapeIssue(shape.items, location + "[]", allowJSON);
  for (const [name, field] of Object.entries(shape.properties || {}).sort(
    ([a], [b]) => order(a, b),
  )) {
    const issue = shapeIssue(field, location + "." + name, allowJSON);
    if (issue) return issue;
  }
  return null;
}
const reasonPrefixes = {
  "operation function": "DSL_OPERATION_FUNCTION",
  "operation signature": "DSL_OPERATION_SIGNATURE",
  "runtime signature": "DSL_RUNTIME_SIGNATURE",
  "request model": "DSL_REQUEST_MODEL",
  "unreviewed input attribute": "DSL_INPUT_ATTRIBUTE",
  "model validation": "DSL_MODEL_VALIDATION",
  "query initialization": "DSL_QUERY_INITIALIZATION",
  "query guard": "DSL_QUERY_GUARD",
  "direct query assignment": "DSL_QUERY_TRANSFORM",
  "query alias/duplicate": "DSL_QUERY_BINDING",
  shrink: "DSL_SHRINK_TRANSFORM",
  "map key type": "DSL_WIRE_TYPE",
  "unbound model input": "DSL_UNBOUND_INPUT",
  "request construction": "DSL_REQUEST_CONSTRUCTION",
  "query encoding helper": "DSL_REQUEST_ENCODING",
  "protocol construction": "DSL_PROTOCOL_CONSTRUCTION",
  "protocol constants": "DSL_PROTOCOL_CONSTANTS",
  "RPC profile": "DSL_PROTOCOL_PROFILE",
  "runtime handoff and trailing statements": "DSL_RUNTIME_HANDOFF",
  "signed product signature initializer": "DSL_PRODUCT_AUTH_INITIALIZER",
  "response model": "DSL_RESPONSE_MODEL",
  "response body": "DSL_RESPONSE_BODY",
  "recursive model": "DSL_RECURSIVE_MODEL",
  "model inheritance": "DSL_MODEL_INHERITANCE",
  "inherited model fields": "DSL_MODEL_INHERITANCE",
  "model member": "DSL_MODEL_MEMBER",
  "unreviewed model attribute": "DSL_MODEL_ATTRIBUTE",
};
function profileIssue(error, location) {
  const prefix = "darabonba: unsupported SDK pattern: ";
  if (!error.message.startsWith(prefix)) throw error;
  const message = error.message.slice(prefix.length);
  const matched = Object.entries(reasonPrefixes).find(
    ([key]) => message === key || message.startsWith(key + " "),
  );
  return {
    code: matched?.[1] || "DSL_UNSUPPORTED_PATTERN",
    message,
    source: location,
  };
}
function buildProduct(ast, { pkg, identifier, info, file, provenance }) {
  const graph = modelGraph(ast, file);
  const candidates = discoverCandidates(ast, info, file);
  const operations = [];
  for (const candidate of candidates) {
    const declarations = candidate.declarations;
    const declaration = declarations[0];
    const location = declaration?.source || {
      file: file.replace("main.tea", "api-info.json"),
      pointer: "/apiDoc",
    };
    const record = {
      name: candidate.name,
      source: location,
      origin: declarations.length
        ? candidate.catalog
          ? "dsl-and-catalog"
          : "dsl-only"
        : "catalog-only",
      declarations: declarations.map((d) => ({
        sdkName: d.sdkName,
        source: d.source,
        protocolEvidence: d.evidence,
        bindingEvidence: d.bindings,
      })),
      documentation: declarations.flatMap((d) =>
        annotationDocuments(d.node, file),
      ),
      status: "unsupported",
      reasons: [],
      parameters: [],
      roots: {},
      reachableModels: [],
    };
    const allRoots = [];
    for (const d of declarations) {
      for (const param of d.node.params?.params || []) {
        record.parameters.push({
          name: lex(param.paramName),
          source: source(file, param.paramName),
          type: graph.type(
            param.paramType,
            candidate.name + ".$" + lex(param.paramName),
            param.paramName,
          ),
        });
      }
      const response = graph.type(
        d.node.returnType,
        candidate.name + ".$response",
        d.node.functionName || d.node.apiName,
      );
      const request = record.parameters.find((p) =>
        ["request", "tmpReq"].includes(p.name),
      )?.type;
      record.roots = {
        request: request || null,
        response,
        body:
          graph.models
            .get(response.ref)
            ?.fields.find((f) => f.wireName === "body")?.type || null,
      };
      allRoots.push(...Object.values(record.roots));
    }
    record.reachableModels = graph.reachable([
      ...record.parameters.map((p) => p.type),
      ...allRoots,
    ]);
    if (declarations.length !== 1) {
      record.reasons.push({
        code: declarations.length
          ? "DSL_AMBIGUOUS_OPERATION"
          : "DSL_MISSING_OPERATION",
        message: declarations.length
          ? "Multiple DSL declarations claim this operation"
          : "Catalog entry has no DSL operation declaration",
        source: location,
      });
    } else if (declaration.actions.length > 1) {
      record.reasons.push({
        code: "DSL_AMBIGUOUS_ACTION",
        message: "Multiple constant actions in one declaration",
        source: location,
      });
    } else {
      try {
        const lowered = lowerOperation(ast, candidate.name);
        const invalid =
          lowered.inputs
            .map((input) =>
              shapeIssue(
                input.schema,
                "request." + input.wire,
                input.encoding === "json",
              ),
            )
            .find(Boolean) || shapeIssue(lowered.response, "response.body");
        if (invalid)
          record.reasons.push({
            code: "DSL_WIRE_TYPE",
            message: "Unsupported reachable wire shape: " + invalid,
            source: location,
          });
        else {
          record.status = "lowered";
          record.protocol = lowered.protocol;
          if (lowered.handoff) record.handoff = lowered.handoff;
          // An explicit empty root preserves source absence without inventing a DSL model.
          if (!record.roots.request) record.roots.request = { kind: "empty" };
          const fields =
            graph.models.get(record.roots.request?.ref)?.fields || [];
          record.bindings = lowered.inputs.map((input) => ({
            wire: input.wire,
            location: input.location,
            guard: input.guard,
            field:
              input.field ||
              fields.find((f) => f.wireName === input.wire)?.dslName,
            ...(input.encoding ? { encoding: input.encoding } : {}),
            source:
              fields.find((f) => f.wireName === input.wire)?.source || location,
          }));
        }
      } catch (error) {
        record.reasons.push(profileIssue(error, location));
      }
    }
    operations.push(record);
  }
  const models = [...graph.models.values()].sort((a, b) => order(a.id, b.id));
  const reachableNamed = models.filter((m) => m.named).length;
  const unusedModels = [...graph.declared.keys()]
    .filter((name) => !graph.models.has(name))
    .sort(order);
  const counts = {
    discovered: operations.length,
    lowered: operations.filter((o) => o.status === "lowered").length,
    unsupported: operations.filter((o) => o.status === "unsupported").length,
    dslOperations: operations.filter((o) => o.origin !== "catalog-only").length,
    catalogOnlyOperations: operations.filter((o) => o.origin === "catalog-only")
      .length,
    dslOnlyOperations: operations.filter((o) => o.origin === "dsl-only").length,
    declaredModels: graph.declared.size,
    reachableNamedModels: reachableNamed,
    reachableAnonymousModels: models.length - reachableNamed,
    unusedNamedModels: unusedModels.length,
  };
  const version =
    info?.version ||
    identifier.replace(/^.*-(\d{4})(\d{2})(\d{2})$/, "$1-$2-$3");
  requireInventory(
    /^\d{4}-\d{2}-\d{2}$/.test(version),
    "missing product version",
  );
  for (const op of operations.filter((o) => o.status === "lowered"))
    requireInventory(
      op.protocol.version === version,
      "product protocol version differs: " + op.name,
    );
  const ir = {
    schemaVersion: 3,
    profile: "rpc-query-json-v1",
    product: pkg,
    identifier,
    version,
    provenance,
    sourceFile: file,
    operations,
    models,
    unusedModels,
  };
  const byReason = {};
  for (const op of operations)
    for (const reason of op.reasons)
      byReason[reason.code] = (byReason[reason.code] || 0) + 1;
  const coverage = {
    schemaVersion: 1,
    product: pkg,
    version,
    provenance,
    counts,
    acceptance: {
      lowering: "assessed",
      goEmission: "not-assessed",
      compilation: "not-assessed",
      live: "not-assessed",
    },
    reasonCounts: Object.fromEntries(
      Object.entries(byReason).sort(([a], [b]) => order(a, b)),
    ),
    operations: operations.map((o) => ({
      name: o.name,
      status: o.status,
      origin: o.origin,
      source: o.source,
      reasons: o.reasons,
    })),
  };
  return { ir, coverage };
}

// Each model/operation occupies one JSON line to keep large generated diffs bounded.
function encodeIR(ir) {
  const { operations, models, ...header } = ir;
  return Buffer.from(
    JSON.stringify(header, null, 2).slice(0, -2) +
      ',\n  "operations": [\n' +
      operations.map((o) => "    " + JSON.stringify(o)).join(",\n") +
      '\n  ],\n  "models": [\n' +
      models.map((m) => "    " + JSON.stringify(m)).join(",\n") +
      "\n  ]\n}\n",
  );
}
function project(root = repository, selected = []) {
  const sourceRoot = path.join(root, "sources/darabonba");
  const verified = verifySources(sourceRoot),
    files = {},
    products = {};
  for (const pkg of Object.keys(verified.manifest.products).sort(order)) {
    requireInventory(/^[a-z][a-z0-9]*$/.test(pkg), "unsafe product key");
    const file = "products/" + pkg + "/main.tea",
      main = path.join(sourceRoot, file);
    const infoFile = "products/" + pkg + "/api-info.json";
    const info = fs.existsSync(path.join(sourceRoot, infoFile))
      ? JSON.parse(fs.readFileSync(path.join(sourceRoot, infoFile)))
      : null;
    const provenance = {
      repository: verified.manifest.repository,
      revision: verified.manifest.revision,
      license: verified.manifest.license,
      parserVersion: verified.manifest.parserVersion,
      sourceManifestSHA256: verified.hash,
      sourceSHA256: sha(fs.readFileSync(main)),
      apiInfoSHA256: info
        ? sha(fs.readFileSync(path.join(sourceRoot, infoFile)))
        : null,
    };
    const ast = parser.parse(fs.readFileSync(main, "utf8"), main);
    const product = buildProduct(ast, {
      pkg,
      identifier: verified.manifest.products[pkg],
      info,
      file,
      provenance,
    });
    products[pkg] = product;
    files[`models/${pkg}/ir.json`] = encodeIR(product.ir);
    files[`models/${pkg}/coverage.json`] = Buffer.from(
      JSON.stringify(product.coverage, null, 2) + "\n",
    );
  }
  for (const target of selected) {
    requireInventory(
      /^[a-z][a-z0-9]*\/[A-Za-z][A-Za-z0-9]*$/.test(target),
      "invalid selected operation: " + target,
    );
    const [pkg, name] = target.split("/"),
      op = products[pkg]?.ir.operations.find((o) => o.name === name);
    requireInventory(
      op?.status === "lowered",
      "selected operation is unknown/unsupported: " +
        target +
        (op ? " (" + op.reasons.map((r) => r.code).join(",") + ")" : ""),
    );
  }
  const manifest = {
    schemaVersion: 3,
    profile: "rpc-query-json-v1",
    sourceManifestSHA256: verified.hash,
    files: Object.entries(files).map(([file, data]) => ({
      file,
      sha256: sha(data),
    })),
  };
  files["models/manifest.json"] = Buffer.from(
    JSON.stringify(manifest, null, 2) + "\n",
  );
  return { files, products };
}
function safeTarget(root, relative) {
  requireInventory(
    relative.startsWith("models/") &&
      relative.split("/").every((p) => p && p !== "." && p !== ".."),
    "unsafe output path",
  );
  let target = root;
  requireInventory(
    !fs.lstatSync(target).isSymbolicLink(),
    "symlink output root",
  );
  for (const part of relative.split("/")) {
    target = path.join(target, part);
    let entry;
    try {
      entry = fs.lstatSync(target);
    } catch (error) {
      if (error.code !== "ENOENT") throw error;
    }
    requireInventory(!entry?.isSymbolicLink(), "symlink output path");
  }
  return target;
}
function run(mode, { root = repository, selected = [], product } = {}) {
  requireInventory(
    ["generate", "check", "report"].includes(mode),
    "unknown command",
  );
  const result = project(root, selected);
  if (mode === "report") {
    requireInventory(result.products[product], "unknown report product");
    const coverage = result.products[product].coverage;
    const lines = [
      `${product}: ${coverage.counts.discovered} discovered, ${coverage.counts.lowered} lowered, ${coverage.counts.unsupported} unsupported. Go emission/compilation/live: not assessed.`,
    ];
    for (const [code, count] of Object.entries(coverage.reasonCounts))
      lines.push(`${code}: ${count}`);
    for (const op of coverage.operations.filter(
      (o) => o.status === "unsupported",
    ))
      for (const reason of op.reasons) {
        lines.push(
          `${op.name}: ${reason.code}: ${reason.message} (${reason.source.file}:${reason.source.line || reason.source.pointer})`,
        );
      }
    return lines.join("\n");
  }
  const targets = Object.entries(result.files).map(([relative, data]) => ({
    relative,
    data,
    target: safeTarget(root, relative),
  }));
  if (mode === "check") {
    for (const { relative, data, target } of targets)
      requireInventory(
        fs.existsSync(target) && fs.readFileSync(target).equals(data),
        "artifact drift: " + relative,
      );
  } else {
    // Complete source/product/selection/path preflight before any output writes.
    for (const { data, target } of targets) {
      fs.mkdirSync(path.dirname(target), { recursive: true });
      fs.writeFileSync(target, data);
    }
  }
  return `Product discovery ${mode} passed for ${Object.keys(result.products).length} products (offline).`;
}
function cli(args) {
  const mode = args.shift();
  const options = {};
  if (mode === "report") options.product = args.shift();
  while (args.length) {
    const flag = args.shift(),
      value = args.shift();
    requireInventory(
      value && ["--root", "--operations"].includes(flag),
      "usage: discovery generate|check|report PRODUCT [--root PATH] [--operations PRODUCT/Operation,...]",
    );
    if (flag === "--root") options.root = path.resolve(value);
    else options.selected = value.split(",");
  }
  return run(mode, options);
}
if (require.main === module) {
  try {
    console.log(cli(process.argv.slice(2)));
  } catch (error) {
    console.error(error.message);
    process.exitCode = 1;
  }
}
module.exports = {
  annotationDocuments,
  discoverCandidates,
  modelGraph,
  buildProduct,
  project,
  run,
  cli,
};
