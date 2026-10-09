"use strict";

// Lower recognized product SDK functions; imported runtime programs are not emitted.
const fs = require("node:fs");
const path = require("node:path");
const crypto = require("node:crypto");
const parser = require("@darabonba/parser");
const { Tag } = require("@darabonba/parser/lib/tag");
const { roaStyle, lowerROA } = require("./roa.cjs");
const {
  bindInputs,
  verifyCanonical,
  normalizeCanonical,
} = require("./normalization.cjs");
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
function requestProperty(expr, owner = "request") {
  return expr?.type === "property_access" &&
    lex(expr.id) === owner &&
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

function reviewedAttributes(field, message) {
  requireProfile(
    field.attrs.every(
      (a) =>
        ["name", "description", "example", "nullable", "deprecated"].includes(
          lex(a.attrName),
        ) &&
        (lex(a.attrName) !== "deprecated" ||
          (a.attrValue.tag === Tag.BOOL &&
            ["true", "false"].includes(lex(a.attrValue)))),
    ),
    message,
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
      reviewedAttributes(field, "unreviewed model attribute");
      properties[wire(field)] = {
        ...shape(field.fieldValue, models, stack),
        required: field.required,
      };
    }
    return { type: "object", properties };
  }
  let kind = value.fieldType ?? lex(value) ?? (["array","map"].includes(value.type) ? value.type : undefined);
  if (kind && typeof kind === "object") kind = lex(kind);
  if (kind === "array")
    return { type: "array", items: shape(value.fieldItemType || value.subType, models, stack) };
  if (kind === "map") {
    requireProfile(lex(value.keyType) === "string", "map key type");
    return { type: "map", values: shape(value.valueType, models, stack) };
  }
  if (kind === "any") return { type: "json" };
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
  if (roaStyle(fn)) {
    const inspect=node=>{if(!node||typeof node!=="object")return;if(node.type==="assign"&&node.left?.type==="virtualVariable"&&lex(node.left.vid)==="@signatureAlgorithm")requireProfile(node.expr?.type==="string"&&node.expr.value.string==="ACS3-HMAC-SHA256","signed product signature initializer");for(const value of Object.values(node)){if(Array.isArray(value))value.forEach(inspect);else if(value&&typeof value==="object")inspect(value);}};
    nodes.filter(n=>n.type==="init").forEach(inspect);
    return lowerROA(ast, fn, operation, {shape, reviewedAttributes, requireProfile});
  }
  const params = fn.params.params;
  const requestless =
    params.length === 1 && lex(params[0].paramName) === "runtime";
  requireProfile(
    requestless ||
      (params.length === 2 &&
        ["request", "tmpReq"].includes(lex(params[0].paramName)) &&
        lex(params[1].paramName) === "runtime"),
    "operation signature",
  );
  const runtime = params.at(-1);
  requireProfile(
    runtime.paramType.type === "moduleModel" &&
      runtime.paramType.path.length === 2 &&
      lex(runtime.paramType.path[0]) === "Util" &&
      lex(runtime.paramType.path[1]) === "RuntimeOptions",
    "runtime signature",
  );
  const models = new Map(
    nodes.filter((n) => n.type === "model").map((n) => [lex(n.modelName), n]),
  );
  const requestModel = requestless
    ? null
    : models.get(lex(params[0].paramType));
  requireProfile(
    requestless || (requestModel && !requestModel.extendOn),
    "request model",
  );
  const inputFields = new Map(
    (requestModel?.modelBody.nodes || []).map((f) => [lex(f.fieldName), f]),
  );
  for (const field of inputFields.values())
    reviewedAttributes(field, "unreviewed input attribute");
  const statements = fn.functionBody.stmts.stmts;
  const inputName = requestless ? null : lex(params[0].paramName);
  const transforms = new Map();
  const transformEncodings = new Map();
  let wholeQuery = false;
  let queryFields = inputFields;
  let index = 0;
  if (!requestless) {
    requireProfile(
      staticCall(statements[index++], "Util", "validateModel", (e) =>
        variable(e, inputName),
      ),
      "model validation",
    );
    if (inputName === "tmpReq") {
      const declaration = statements[index++];
      requireProfile(
        declaration?.type === "declare" &&
          lex(declaration.id) === "request" &&
          declaration.expr.type === "construct_model" &&
          lex(declaration.expr.aliasId) ===
            lex(params[0].paramType).replace(/Request$/, "ShrinkRequest") &&
          declaration.expr.propertyPath.length === 0 &&
          declaration.expr.object.fields.length === 0,
        "shrink request construction",
      );
      const shrinkModel = models.get(lex(declaration.expr.aliasId));
      requireProfile(
        shrinkModel && !shrinkModel.extendOn,
        "shrink request model",
      );
      queryFields = new Map(
        shrinkModel.modelBody.nodes.map((f) => [lex(f.fieldName), f]),
      );
      for (const field of queryFields.values())
        reviewedAttributes(field, "unreviewed input attribute");
      const convert = statements[index++];
      requireProfile(
        convert?.type === "call" &&
          convert.left.type === "static_call" &&
          lex(convert.left.id) === "OpenApiUtil" &&
          convert.left.propertyPath.length === 1 &&
          lex(convert.left.propertyPath[0]) === "convert" &&
          convert.args.length === 2 &&
          variable(convert.args[0], "tmpReq") &&
          variable(convert.args[1], "request"),
        "shrink conversion",
      );
      while (statements[index]?.type === "if") {
        const guard = statements[index++];
        requireProfile(
          guard.condition.type === "not" &&
            staticCall(
              guard.condition.expr,
              "Util",
              "isUnset",
              (e) => !!requestProperty(e, "tmpReq"),
            ) &&
            !guard.elseIfs.length &&
            !guard.elseStmts &&
            guard.stmts.stmts.length === 1,
          "shrink transform guard",
        );
        const original = requestProperty(
          guard.condition.expr.args[0],
          "tmpReq",
        );
        const assignment = guard.stmts.stmts[0],
          target =
            assignment.left?.type === "property"
              ? requestProperty({ ...assignment.left, type: "property_access" })
              : undefined,
          call = assignment.expr;
        const field = inputFields.get(original),
          shrink = queryFields.get(target);
        requireProfile(
          assignment.type === "assign" &&
            field &&
            shrink &&
            target === original + "Shrink" &&
            shrink.fieldValue.fieldType === "string" &&
            wire(shrink) === wire(field) &&
            !transforms.has(target) &&
            call?.type === "call" &&
            call.left.type === "static_call" &&
            lex(call.left.id) === "OpenApiUtil" &&
            call.left.propertyPath.length === 1 &&
            lex(call.left.propertyPath[0]) ===
              "arrayToStringWithSpecifiedStyle" &&
            call.args.length === 3 &&
            requestProperty(call.args[0], "tmpReq") === original &&
            call.args[1].type === "string" &&
            call.args[1].value.string === wire(field) &&
            call.args[2].type === "string" &&
            ["json", "simple"].includes(call.args[2].value.string),
          "shrink JSON transform",
        );
        if (call.args[2].value.string === "simple") {
          const schema = shape(field.fieldValue, models);
          requireProfile(schema.type === "array" && schema.items.type === "string", "shrink simple string array");
        }
        transforms.set(target, original);
        transformEncodings.set(target, call.args[2].value.string);
      }
      requireProfile(
        [...queryFields].every(([name, field]) => {
          const original = inputFields.get(transforms.get(name) || name);
          return (
            original &&
            wire(original) === wire(field) &&
            (transforms.has(name) ||
              JSON.stringify(shape(original.fieldValue, models)) ===
                JSON.stringify(shape(field.fieldValue, models)))
          );
        }) && queryFields.size === inputFields.size,
        "shrink model correspondence",
      );
    }
    const query = statements[index++];
    wholeQuery = inputName === "request" && staticCall(query?.expr, "OpenApiUtil", "query", (e) =>
      staticCall(e, "Util", "toMap", (value) => variable(value, "request")),
    );
    requireProfile(
      query?.type === "declare" &&
        lex(query.id) === "query" &&
        (wholeQuery || (query.expr.type === "object" && query.expr.fields.length === 0)),
      "query initialization",
    );
  }
  const bindings = [];
  const boundSources = new Map();
  const boundFields = new Set();
  function addBinding(fieldName, fieldWire, location) {
    const field = queryFields.get(fieldName);
    const originalName = transforms.get(fieldName) || fieldName;
    const key = location + ":" + fieldWire;
    requireProfile(field && wire(field) === fieldWire &&
      (!boundSources.has(key) || boundSources.get(key) === originalName), "query alias/duplicate");
    if (boundSources.has(key)) return;
    requireProfile(!boundFields.has(originalName), "input location conflict");
    boundFields.add(originalName);
    boundSources.set(key, originalName);
    const originalField = inputFields.get(originalName);
    bindings.push({
      ...(inputName === "tmpReq" ? { field: originalName } : {}),
      ...(transforms.has(fieldName) ? { encoding: transformEncodings.get(fieldName) } : {}),
      wire: fieldWire, location, guard: "isUnset",
      schema: { ...shape(originalField.fieldValue, models), required: originalField.required },
    });
  }
  function readBindings(target, location) {
  while (!requestless && statements[index]?.type === "if") {
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
        lex(assign.left.id) === target &&
        assign.left.accessKey.type === "string" &&
        requestProperty(assign.expr) === fieldName,
      "direct query assignment",
    );
    const fieldWire = assign.left.accessKey.value.string;
    addBinding(fieldName, fieldWire, location);
  }
  }
  if (wholeQuery) {
    for (const [name, field] of inputFields) addBinding(name, wire(field), "query");
  } else readBindings("query", "query");
  let formBody = false;
  if (!requestless && statements[index]?.type === "declare" && lex(statements[index].id) === "body") {
    requireProfile(!wholeQuery && inputName === "request", "form body input");
    for (const name of ["body", "bodyFlat"]) {
      const declaration = statements[index++];
      requireProfile(declaration?.type === "declare" && lex(declaration.id) === name &&
        declaration.expr.type === "object" && declaration.expr.fields.length === 0 &&
        (!declaration.expectedType || (declaration.expectedType.type === "map" &&
          lex(declaration.expectedType.keyType) === "string" && lex(declaration.expectedType.valueType) === "any")), "form body initialization");
    }
    const previous = bindings.length;
    readBindings("bodyFlat", "form");
    requireProfile(bindings.length > previous, "empty form binding program");
    const merge = statements[index++];
    const fields = merge?.expr?.fields;
    requireProfile(merge?.type === "assign" && variable(merge.left, "body") &&
      merge.expr.type === "object" && fields.length === 2 && fields.every((f) => f.type === "expandField") &&
      variable(fields[0].expr, "body") && staticCall(fields[1].expr, "OpenApiUtil", "query", (e) => variable(e, "bodyFlat")), "form body merge");
    formBody = true;
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
    requestless
      ? requestFields.length === 0
      : requestFields.length === (formBody ? 2 : 1) &&
          lex(requestFields[0].fieldName) === "query" &&
          staticCall(requestFields[0].expr, "OpenApiUtil", "query", (e) =>
            variable(e, "query"),
          ) && (!formBody || (lex(requestFields[1].fieldName) === "body" &&
            staticCall(requestFields[1].expr, "OpenApiUtil", "parseToMap", (e) => variable(e, "body")))),
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
      ["POST", "GET"].includes(facts.method) &&
      (!formBody || facts.method === "POST") &&
      ["AK", "Anonymous"].includes(facts.authType) &&
      facts.style === "RPC" &&
      facts.reqBodyType === "formData" &&
      facts.bodyType === "json",
    "RPC profile",
  );
  const result = statements[index++];
  if (facts.authType === "AK") {
    function checkSignature(node) {
      if (!node || typeof node !== "object") return;
      if (
        node.type === "assign" &&
        node.left?.type === "virtualVariable" &&
        lex(node.left.vid) === "@signatureAlgorithm"
      ) {
        requireProfile(
          node.expr?.type === "string" &&
            node.expr.value.string === "ACS3-HMAC-SHA256",
          "signed product signature initializer",
        );
      }
      for (const value of Object.values(node)) {
        if (Array.isArray(value)) value.forEach(checkSignature);
        else if (value && typeof value === "object") checkSignature(value);
      }
    }
    nodes.filter((n) => n.type === "init").forEach(checkSignature);
  }
  if (facts.authType === "Anonymous") {
    const reserved = new Set([
      "action",
      "version",
      "format",
      "timestamp",
      "signaturenonce",
      "accesskeyid",
      "accesskeysecret",
      "securitytoken",
      "signature",
      "signaturemethod",
      "signatureversion",
      "signaturetype",
      "bearertoken",
    ]);
    requireProfile(
      bindings.every((b) => !reserved.has(b.wire.toLowerCase())),
      "anonymous reserved query member",
    );
  }
  const call = result?.expr;
  const namedCall = (name) =>
    call?.type === "call" &&
    call.left.type === "method_call" &&
    lex(call.left.id) === name;
  const callApi =
    namedCall("callApi") &&
    call.args.length === 3 &&
    call.args.every((e, i) => variable(e, ["params", "req", "runtime"][i]));
  const rpcFields = [
    "action",
    "version",
    "protocol",
    "method",
    "authType",
    "bodyType",
  ];
  const doRPC =
    namedCall("doRPCRequest") &&
    call.args.length === 8 &&
    call.args.every((e, i) =>
      i < 6
        ? e.type === "property_access" &&
          lex(e.id) === "params" &&
          e.propertyPath.length === 1 &&
          lex(e.propertyPath[0]) === rpcFields[i]
        : variable(e, ["req", "runtime"][i - 6]),
    );
  requireProfile(
    result?.type === "return" &&
      ((facts.authType === "AK" && callApi) ||
        (facts.authType === "Anonymous" && doRPC)) &&
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
    ...(facts.authType === "Anonymous" ? { handoff: "doRPCRequest" } : {}),
    inputs: bindings.sort((a, b) => a.wire.localeCompare(b.wire)),
    response: shape(body.fieldValue, models),
  };
}

function reviewInputs(pkg, op, snapshot, decisions) {
  const parameters = new Map(snapshot.parameters.map((p) => [p.name, p]));
  const bindings = bindInputs(op.inputs, snapshot.parameters);
  for (const input of op.inputs) {
    const aliases = bindings.get(input.wire).aliases;
    if (aliases.length) input.metadataBindings = aliases;
    else delete input.metadataBindings;
    const metadata = parameters.get(input.wire);
    requireProfile(
      metadata || aliases.length || !input.schema.required,
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
      .filter(
        (i) => !parameters.has(i.wire) && !bindings.get(i.wire).aliases.length,
      )
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
    metadataOnlyFields: Object.fromEntries(
      Object.entries(bindings.metadataOnlyFields).sort(([a], [b]) =>
        a < b ? -1 : a > b ? 1 : 0,
      ),
    ),
  };
  const approved = decisions.operations[pkg + "/" + op.name];
  requireProfile(
    approved &&
      Object.keys(approved).every((key) => Object.hasOwn(actual, key)) &&
      Object.entries(actual).every(([key, value]) =>
        key === "metadataOnlyFields"
          ? JSON.stringify(value) ===
            JSON.stringify(
              Object.fromEntries(
                Object.entries(approved[key] || {}).sort(([a], [b]) =>
                  a < b ? -1 : a > b ? 1 : 0,
                ),
              ),
            )
          : JSON.stringify(value) === JSON.stringify(approved[key] || []),
      ),
    "unreviewed metadata/DSL difference " + op.name,
  );
}

function project(root = repository) {
  const sourceDir = path.join(root, "sources/darabonba");
  const verified = verifySources(sourceDir);
  // CLI metadata is optional enrichment, never an operation-discovery prerequisite.
  // These pinned fixtures exercise its versioned adapter before any bridge writes.
  const canonicalRoot = path.join(root, "sources/openapi-meta");
  const canonical = new Map();
  if (fs.existsSync(canonicalRoot)) {
    const lock = verifyCanonical(canonicalRoot);
    for (const file of lock.manifest.files.filter(
      (f) => f.file.endsWith(".json") && !f.file.endsWith("/version.json"),
    )) {
      const normalized = normalizeCanonical(
        JSON.parse(fs.readFileSync(path.join(canonicalRoot, file.file))),
        {
          repository: lock.manifest.repository,
          revision: lock.manifest.revision,
          file: file.file,
          sha256: file.sha256,
          license: lock.manifest.license,
        },
      );
      canonical.set("ecs/" + normalized.protocol.action, normalized);
    }
  }
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
  // This retained fixture checker covers only explicitly declared bridge decisions.
  // A new full-DSL product needs no per-operation metadata fixture directory.
  const fixtureProducts = [...new Set(Object.keys(decisions.operations).map(key => key.split("/")[0]))].sort();
  for (const pkg of fixtureProducts) {
    requireProfile(Object.hasOwn(verified.manifest.products,pkg),"fixture product source missing");
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
      const enrichment = canonical.get(pkg + "/" + op.name);
      if (enrichment) {
        requireProfile(
          Object.entries(enrichment.protocol).every(
            ([key, value]) => op.protocol[key] === value,
          ),
          "canonical protocol differs",
        );
        reviewInputs(
          pkg,
          structuredClone(op),
          {
            parameters: enrichment.inputs.map((i) => ({
              name: i.wire,
              in: i.location,
              schema: i.schema,
            })),
          },
          decisions,
        );
      }
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
