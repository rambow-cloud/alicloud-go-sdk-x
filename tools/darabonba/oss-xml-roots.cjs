"use strict";

const fs = require("node:fs"), path = require("node:path"), crypto = require("node:crypto");
const { execFileSync } = require("node:child_process");
const parser = require("@darabonba/parser");
const { verifySources } = require("./frontend.cjs");
const { discoverCandidates, modelGraph } = require("./discovery.cjs");
const projectRoot = path.resolve(__dirname, "../..");

// Compare explicit root-level wire names only. Nested shapes and runtime policy
// remain outside this report; no missing root/name is inferred.
function compareBody(body, root, nativeModels) {
  if (!root) return { status: "unsupported", reason: "native registry action missing" };
  if (root.kind === "unsupported") return { status: "unsupported", reason: root.reason };
  if (!body) return { status: "unsupported", reason: "DSL response body is not a concrete model" };
  const wrapped = body.fields.length === 1 && body.fields[0].wireName === root.field.jsonName;
  let fields;
  if (root.kind === "scalar" || wrapped) fields = [root.field];
  else fields = nativeModels.get(root.field.type.name)?.fields;
  if (!fields) return { status: "unsupported", reason: "native root child model missing" };
  const names = fields.map(f => f.jsonName);
  if (names.some(n => !n || n === "-") || new Set(names).size !== names.length || fields.some(f => f.xmlName !== f.jsonName || (f.xmlOptions || []).some(o => o !== "omitempty"))) {
    return { status: "unsupported", reason: "native child wire tags require additional normalization" };
  }
  const dsl = body.fields.map(f => f.wireName), missing = dsl.filter(n => !names.includes(n)).sort(), extra = names.filter(n => !dsl.includes(n)).sort();
  return {
    status: missing.length ? "divergent" : "root-fields-match",
    normalization: root.kind === "scalar" ? "scalar-field" : wrapped ? "preserve-wrapper" : "unwrap-structured-root",
    missingNativeFields: missing, extraNativeFields: extra,
  };
}

function compareProduct(ast, file, inventory) {
  if (inventory.schemaVersion !== 1) throw Error("Unsupported native inventory schema");
  const graph = modelGraph(ast, file), roots = new Map(), nativeModels = new Map();
  for (const root of inventory.roots) {
    if (roots.has(root.action)) throw Error("Duplicate native registry action");
    roots.set(root.action, root);
  }
  for (const model of inventory.models) {
    if (nativeModels.has(model.name)) throw Error("Duplicate native model");
    nativeModels.set(model.name, model);
  }
  const candidates = discoverCandidates(ast, null, file), operations = [], actions = new Set();
  for (const candidate of candidates) {
    for (const declaration of candidate.declarations) {
      const evidence = declaration.evidence.flatMap(e => e.fields);
      const modes = evidence.filter(f => f.name === "bodyType");
      if (!modes.some(f => f.value === "xml")) continue;
      if (modes.length !== 1 || modes[0].expressionKind !== "string") throw Error("Ambiguous XML response declaration");
      const actionFields = evidence.filter(f => f.name === "action");
      if (actionFields.length !== 1 || actionFields[0].expressionKind !== "string" || !actionFields[0].value) throw Error("XML action must have one explicit constant");
      const action = actionFields[0].value;
      if (actions.has(action)) throw Error("Duplicate DSL XML action");
      actions.add(action);
      const responseType = graph.type(declaration.node.returnType, action + ".$response", declaration.node.functionName);
      const response = responseType.kind === "model" ? graph.models.get(responseType.ref) : null;
      const bodyFields = response?.fields.filter(f => f.wireName === "body") || [];
      const bodyType = bodyFields.length === 1 ? bodyFields[0].type : null;
      const body = bodyType?.kind === "model" ? graph.models.get(bodyType.ref) : null;
      const root = roots.get(action);
      operations.push({
        operation: candidate.name, action, source: declaration.source,
        ...(response ? { responseModel: response.id, responseSource: response.source } : {}),
        ...(body ? { bodyModel: body.id, bodySource: body.source } : {}),
        ...(root ? { nativeRoot: root } : {}),
        ...(response && bodyFields.length === 0 ? { status: "no-body-target", reason: "DSL response facade declares no body field" } : compareBody(body, root, nativeModels)),
      });
    }
  }
  operations.sort((a,b) => a.action < b.action ? -1 : a.action > b.action ? 1 : 0);
  const counts = { discovered: candidates.length, xmlResponses: operations.length, nativeModels: inventory.models.length, nativeRoots: inventory.roots.length, rootFieldsMatch: 0, noBodyTarget: 0, divergent: 0, unsupported: 0 };
  for (const op of operations) counts[op.status === "root-fields-match" ? "rootFieldsMatch" : op.status === "no-body-target" ? "noBodyTarget" : op.status]++;
  return { counts, operations, unusedNativeRoots: inventory.roots.filter(r => !actions.has(r.action)).map(r => r.action) };
}

function run(sourceRoot, registry, models, pins = path.join(projectRoot, "metadata/native-helper-pins/oss.json")) {
  const verified = verifySources(sourceRoot), product = "oss", file = "products/oss/main.tea";
  if (verified.manifest.products[product] !== "oss-20190517") throw Error("Complete pinned OSS source is required");
  const source = fs.readFileSync(path.join(sourceRoot, file));
  // The Go reader validates original file hashes before returning any facts.
  const inventory = JSON.parse(execFileSync("go", ["run", "./internal/cmd/xmltraits", "-pins", path.resolve(pins), "-registry", path.resolve(registry), "-models", path.resolve(models)], { cwd: projectRoot, encoding: "utf8", maxBuffer: 8 << 20, stdio: ["ignore", "pipe", "pipe"] }));
  const ast = parser.parse(source.toString("utf8"), path.join(sourceRoot, file));
  const result = compareProduct(ast, file, inventory);
  return {
    schemaVersion: 1,
    scope: "XML root and root-field discovery only; no nested compatibility, lowering, emission, compilation or live acceptance",
    sourceRevision: verified.manifest.revision,
    sourceSHA256: crypto.createHash("sha256").update(source).digest("hex"),
    sourceManifestSHA256: verified.hash,
    parserVersion: require("@darabonba/parser/package.json").version,
    nativeRegistrySHA256: inventory.registrySHA256, nativeModelsSHA256: inventory.modelsSHA256,
    nativePins: JSON.parse(fs.readFileSync(pins,"utf8")),
    ...result,
  };
}

if (require.main === module) {
  try {
    const args = process.argv.slice(2), options = {};
    for (let i=0;i<args.length;i+=2) {
      if (!["--source-root","--registry","--models"].includes(args[i]) || !args[i+1] || options[args[i]]) throw Error("Provide unique --source-root, --registry and --models paths");
      options[args[i]] = args[i+1];
    }
    if (Object.keys(options).length !== 3) throw Error("Provide --source-root, --registry and --models paths");
    process.stdout.write(JSON.stringify(run(path.resolve(options["--source-root"]), options["--registry"], options["--models"]), null, 2) + "\n");
  } catch (error) {
    // Native compiler/file errors must not echo complete source/paths into reports.
    console.error(error.status !== undefined ? "Native XML source extraction failed" : error.message);
    process.exitCode = 1;
  }
}

module.exports = { compareBody, compareProduct, run };
