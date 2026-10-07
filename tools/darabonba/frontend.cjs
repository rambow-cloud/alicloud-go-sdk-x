"use strict";

// Lower recognized product SDK functions; imported runtime programs are not emitted.
const fs = require("node:fs");
const path = require("node:path");
const crypto = require("node:crypto");
const parser = require("@darabonba/parser");
const repository = path.resolve(__dirname, "../..");
const sourceRoot = path.join(repository, "sources/darabonba");
const sha = (data) => crypto.createHash("sha256").update(data).digest("hex");
const lex = (token) => token?.lexeme;
const attr = (field, name) =>
  field.attrs.find((a) => lex(a.attrName) === name)?.attrValue;
const wire = (field) => attr(field, "name")?.string || lex(field.fieldName);
function requireProfile(ok, message) {
  if (!ok) throw new Error("darabonba: unsupported SDK pattern: " + message);
}
function safePath(root, relative) {
  requireProfile(
    typeof relative === "string" &&
      relative !== "" &&
      !path.isAbsolute(relative) &&
      !relative.includes("\\") &&
      !relative.split("/").some((p) => !p || p === ".." || p === "."),
    "source path",
  );
  requireProfile(!fs.lstatSync(root).isSymbolicLink(), "source root");
  const target = path.join(root, relative);
  let current = root;
  for (const part of relative.split("/")) {
    current = path.join(current, part);
    if (fs.lstatSync(current).isSymbolicLink())
      throw new Error("darabonba: symlink source");
  }
  return target;
}
function verifySources(root = sourceRoot) {
  const raw = fs.readFileSync(path.join(root, "manifest.json"));
  const manifest = JSON.parse(raw);
  requireProfile(
    manifest.schemaVersion === 1 &&
      manifest.parserVersion ===
        require("@darabonba/parser/package.json").version,
    "parser/manifest version",
  );
  requireProfile(
    manifest.repository === "https://github.com/aliyun/alibabacloud-sdk" &&
      /^[a-f0-9]{40}$/.test(manifest.revision) &&
      manifest.license === "Apache-2.0",
    "source provenance",
  );
  const names = new Set();
  for (const file of manifest.files) {
    requireProfile(!names.has(file.file), "duplicate source");
    names.add(file.file);
    if (sha(fs.readFileSync(safePath(root, file.file))) !== file.sha256)
      throw new Error("darabonba: source checksum mismatch: " + file.file);
  }
  function inventory(relative) {
    for (const entry of fs.readdirSync(path.join(root, relative), {
      withFileTypes: true,
    })) {
      const name = relative + "/" + entry.name;
      requireProfile(!entry.isSymbolicLink(), "source symlink");
      if (entry.isDirectory()) inventory(name);
      else requireProfile(names.has(name), "unlisted source " + name);
    }
  }
  inventory("products");
  inventory("modules");
  inventory("licenses");
  for (const module of manifest.modules) {
    const directory = safePath(root, module.directory);
    const meta = JSON.parse(
      fs.readFileSync(
        path.join(
          directory,
          fs.existsSync(path.join(directory, "Darafile"))
            ? "Darafile"
            : "Teafile",
        ),
        "utf8",
      ),
    );
    requireProfile(
      module.directory ===
        "modules/" + meta.scope + "_" + meta.name + "_" + meta.version &&
        module.scope === meta.scope &&
        module.name === meta.name &&
        module.version === meta.version,
      "resolved module version",
    );
    requireProfile(
      names.has(path.posix.join(module.directory, meta.main)),
      "module entrypoint",
    );
  }
  for (const pkg of Object.keys(manifest.products)) {
    const directory = "products/" + pkg;
    for (const name of ["main.tea", "Teafile", ".libraries.json"])
      requireProfile(names.has(directory + "/" + name), "product source");
    const libraries = JSON.parse(
      fs.readFileSync(path.join(root, directory, ".libraries.json"), "utf8"),
    );
    for (const [spec, target] of Object.entries(libraries))
      requireProfile(
        manifest.modules.some(
          (m) => m.spec === spec && target === "../../" + m.directory,
        ),
        "import resolution",
      );
  }
  return { manifest, hash: sha(raw) };
}
function variable(expr, name) {
  return expr?.type === "variable" && lex(expr.id) === name;
}
function requestProperty(expr) {
  return expr?.type === "property_access" &&
    lex(expr.id) === "request" &&
    expr.propertyPath.length === 1
    ? lex(expr.propertyPath[0])
    : undefined;
}
function staticCall(expr, alias, name, arg) {
  return (
    expr?.type === "call" &&
    expr.left.type === "static_call" &&
    lex(expr.left.id) === alias &&
    expr.left.propertyPath.length === 1 &&
    lex(expr.left.propertyPath[0]) === name &&
    expr.args.length === 1 &&
    arg(expr.args[0])
  );
}
function construct(expr, model) {
  return (
    expr?.type === "construct_model" &&
    lex(expr.aliasId) === "OpenApi" &&
    expr.propertyPath.length === 1 &&
    lex(expr.propertyPath[0]) === model
  );
}

function shape(value, models, stack = []) {
  if (value.type === "modelBody") {
    requireProfile(!value.extendFileds?.length, "inherited model fields");
    const properties = {};
    for (const field of value.nodes) {
      requireProfile(
        field.type === "modelField" && !properties[wire(field)],
        "model member",
      );
      requireProfile(
        field.attrs.every((a) =>
          ["name", "description", "example", "nullable"].includes(
            lex(a.attrName),
          ),
        ),
        "unreviewed model attribute",
      );
      properties[wire(field)] = {
        ...shape(field.fieldValue, models, stack),
        required: field.required,
      };
    }
    return { type: "object", properties };
  }
  let kind = value.fieldType ?? lex(value);
  if (kind && typeof kind === "object") kind = lex(kind);
  if (kind === "array")
    return { type: "array", items: shape(value.fieldItemType, models, stack) };
  if (kind === "string") return { type: "string" };
  if (kind === "boolean") return { type: "boolean" };
  if (
    [
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
    ].includes(kind)
  )
    return { type: "integer" };
  if (["float", "double"].includes(kind)) return { type: "number" };
  if (models.has(kind)) {
    requireProfile(
      stack.length < 32 && !stack.includes(kind),
      "recursive model",
    );
    const model = models.get(kind);
    requireProfile(!model.extendOn, "model inheritance");
    return shape(model.modelBody, models, [...stack, kind]);
  }
  // Preserve unsupported source types. Selected unsupported fields fail cross-checks.
  return { type: "unsupported" };
}

function lowerOperation(ast, operation) {
  const nodes = ast.moduleBody.nodes;
  const fn = nodes.find(
    (n) =>
      lex(n.functionName) ===
      operation[0].toLowerCase() + operation.slice(1) + "WithOptions",
  );
  requireProfile(
    fn?.isAsync && !fn.isStatic && !fn.notes?.length,
    "operation function " + operation,
  );
  const params = fn.params.params;
  requireProfile(
    params.length === 2 &&
      lex(params[0].paramName) === "request" &&
      lex(params[1].paramName) === "runtime",
    "operation signature",
  );
  requireProfile(
    params[1].paramType.type === "moduleModel" &&
      lex(params[1].paramType.path[0]) === "Util" &&
      lex(params[1].paramType.path[1]) === "RuntimeOptions",
    "runtime signature",
  );
  const models = new Map(
    nodes.filter((n) => n.type === "model").map((n) => [lex(n.modelName), n]),
  );
  const requestModel = models.get(lex(params[0].paramType));
  requireProfile(requestModel && !requestModel.extendOn, "request model");
  const inputFields = new Map(
    requestModel.modelBody.nodes.map((f) => [lex(f.fieldName), f]),
  );
  requireProfile(
    requestModel.modelBody.nodes.every((f) =>
      f.attrs.every((a) =>
        ["name", "description", "example", "nullable"].includes(
          lex(a.attrName),
        ),
      ),
    ),
    "unreviewed input attribute",
  );
  const statements = fn.functionBody.stmts.stmts;
  let index = 0;
  requireProfile(
    staticCall(statements[index++], "Util", "validateModel", (e) =>
      variable(e, "request"),
    ),
    "model validation",
  );
  const query = statements[index++];
  requireProfile(
    query?.type === "declare" &&
      lex(query.id) === "query" &&
      query.expr.type === "object" &&
      query.expr.fields.length === 0,
    "query initialization",
  );
  const bindings = [];
  while (statements[index]?.type === "if") {
    const guard = statements[index++];
    requireProfile(
      guard.condition.type === "not" &&
        staticCall(
          guard.condition.expr,
          "Util",
          "isUnset",
          (e) => !!requestProperty(e),
        ) &&
        !guard.elseIfs.length &&
        !guard.elseStmts,
      "query guard",
    );
    const fieldName = requestProperty(guard.condition.expr.args[0]);
    const body = guard.stmts.stmts;
    requireProfile(body.length === 1, "query guard body");
    const assign = body[0];
    requireProfile(
      assign.type === "assign" &&
        assign.left.type === "map_access" &&
        lex(assign.left.id) === "query" &&
        assign.left.accessKey.type === "string" &&
        requestProperty(assign.expr) === fieldName,
      "direct query assignment",
    );
    const field = inputFields.get(fieldName);
    const fieldWire = assign.left.accessKey.value.string;
    requireProfile(
      field &&
        wire(field) === fieldWire &&
        !bindings.some((b) => b.wire === fieldWire),
      "query alias/duplicate",
    );
    bindings.push({
      wire: fieldWire,
      location: "query",
      guard: "isUnset",
      schema: { ...shape(field.fieldValue, models), required: field.required },
    });
  }
  requireProfile(bindings.length === inputFields.size, "unbound model input");
  const request = statements[index++];
  requireProfile(
    request?.type === "declare" &&
      lex(request.id) === "req" &&
      construct(request.expr, "OpenApiRequest"),
    "request construction",
  );
  const requestFields = request.expr.object.fields;
  requireProfile(
    requestFields.length === 1 &&
      lex(requestFields[0].fieldName) === "query" &&
      staticCall(requestFields[0].expr, "OpenApiUtil", "query", (e) =>
        variable(e, "query"),
      ),
    "query encoding helper",
  );
  const protocol = statements[index++];
  requireProfile(
    protocol?.type === "declare" &&
      lex(protocol.id) === "params" &&
      construct(protocol.expr, "Params"),
    "protocol construction",
  );
  const expected = [
    "action",
    "version",
    "protocol",
    "pathname",
    "method",
    "authType",
    "style",
    "reqBodyType",
    "bodyType",
  ];
  const facts = {};
  for (const field of protocol.expr.object.fields) {
    const name = lex(field.fieldName);
    requireProfile(
      field.type === "objectField" &&
        expected.includes(name) &&
        !Object.hasOwn(facts, name) &&
        field.expr.type === "string",
      "protocol constants",
    );
    facts[name] = field.expr.value.string;
  }
  requireProfile(
    Object.keys(facts).length === expected.length &&
      facts.action === operation &&
      facts.protocol === "HTTPS" &&
      facts.pathname === "/" &&
      facts.method === "POST" &&
      facts.authType === "AK" &&
      facts.style === "RPC" &&
      facts.reqBodyType === "formData" &&
      facts.bodyType === "json",
    "RPC profile",
  );
  const result = statements[index++];
  requireProfile(
    result?.type === "return" &&
      result.expr.type === "call" &&
      result.expr.left.type === "method_call" &&
      lex(result.expr.left.id) === "callApi" &&
      result.expr.args.length === 3 &&
      result.expr.args.every((e, i) =>
        variable(e, ["params", "req", "runtime"][i]),
      ) &&
      index === statements.length,
    "runtime handoff and trailing statements",
  );
  const responseModel = models.get(lex(fn.returnType));
  requireProfile(responseModel, "response model");
  const body = responseModel.modelBody.nodes.find((f) => wire(f) === "body");
  requireProfile(body, "response body");
  return {
    name: operation,
    protocol: facts,
    inputs: bindings.sort((a, b) => a.wire.localeCompare(b.wire)),
    response: shape(body.fieldValue, models),
  };
}

function reviewInputs(pkg, op, snapshot, decisions) {
  const parameters = new Map(snapshot.parameters.map((p) => [p.name, p]));
  for (const input of op.inputs) {
    const metadata = parameters.get(input.wire);
    requireProfile(
      metadata || !input.schema.required,
      "unselected required DSL input " + input.wire,
    );
    if (metadata && !!input.schema.required !== !!metadata.schema.required) {
      requireProfile(
        metadata.schema.required && !input.schema.required,
        "unreviewed requiredness direction " + input.wire,
      );
    }
  }
  const actual = {
    dslOnlyInputs: op.inputs
      .filter((i) => !parameters.has(i.wire))
      .map((i) => i.wire)
      .sort(),
    requiredInputs: op.inputs
      .filter(
        (i) =>
          parameters.has(i.wire) &&
          !!i.schema.required !== !!parameters.get(i.wire).schema.required,
      )
      .map((i) => i.wire)
      .sort(),
  };
  const approved = decisions.operations[pkg + "/" + op.name];
  requireProfile(
    approved &&
      Object.keys(approved).length === 2 &&
      Object.entries(actual).every(
        ([key, value]) =>
          JSON.stringify(value) === JSON.stringify(approved[key]),
      ),
    "unreviewed metadata/DSL difference " + op.name,
  );
}

function project(root = repository) {
  const sourceDir = path.join(root, "sources/darabonba");
  const verified = verifySources(sourceDir);
  const files = {};
  const decisionData = fs.readFileSync(
    path.join(root, "metadata/darabonba-decisions.json"),
  );
  const decisions = JSON.parse(decisionData);
  requireProfile(
    decisions.schemaVersion === 1 &&
      decisions.revision === verified.manifest.revision,
    "decision revision",
  );
  for (const pkg of Object.keys(verified.manifest.products).sort()) {
    const metaDir = path.join(root, "metadata", pkg),
      metadata = JSON.parse(
        fs.readFileSync(path.join(metaDir, "manifest.json"), "utf8"),
      );
    const main = path.join(sourceDir, "products", pkg, "main.tea");
    const ast = parser.parse(fs.readFileSync(main, "utf8"), main);
    const projection = {
      schemaVersion: 1,
      parserVersion: verified.manifest.parserVersion,
      sourceManifestSHA256: verified.hash,
      revision: verified.manifest.revision,
      product: metadata.product,
      version: metadata.version,
      operations: metadata.sources
        .map((s) => lowerOperation(ast, s.operation))
        .sort((a, b) => a.name.localeCompare(b.name)),
    };
    requireProfile(
      projection.operations.every(
        (o) => o.protocol.version === metadata.version,
      ),
      "API version",
    );
    for (const op of projection.operations) {
      const snapshot = JSON.parse(
        fs.readFileSync(path.join(metaDir, op.name + ".json"), "utf8"),
      );
      reviewInputs(pkg, op, snapshot, decisions);
    }
    const data = Buffer.from(JSON.stringify(projection, null, 2) + "\n");
    files["metadata/" + pkg + "/dsl.json"] = data;
    metadata.dsl = {
      file: "dsl.json",
      sha256: sha(data),
      sourceManifestSHA256: verified.hash,
      decisionsSHA256: sha(decisionData),
    };
    files["metadata/" + pkg + "/manifest.json"] = Buffer.from(
      JSON.stringify(metadata, null, 2) + "\n",
    );
  }
  return files;
}
function run(mode) {
  requireProfile(["generate", "check"].includes(mode), "command");
  const files = project(); // All products preflight before writes.
  for (const [relative, data] of Object.entries(files)) {
    const target = path.join(repository, relative);
    if (mode === "check") {
      if (!fs.existsSync(target) || !fs.readFileSync(target).equals(data))
        throw new Error("darabonba: projection drift: " + relative);
    } else fs.writeFileSync(target, data);
  }
  console.log(
    "Official Darabonba semantic projection " +
      mode +
      " passed for " +
      Object.keys(files).length / 2 +
      " products (offline).",
  );
}
if (require.main === module) {
  try {
    run(process.argv[2]);
  } catch (error) {
    console.error(error.message);
    process.exitCode = 1;
  }
}
module.exports = {
  verifySources,
  lowerOperation,
  shape,
  reviewInputs,
  project,
  run,
};
