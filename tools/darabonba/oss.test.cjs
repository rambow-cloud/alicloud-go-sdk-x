"use strict";

const test = require("node:test"), assert = require("node:assert/strict");
const fs = require("node:fs"), path = require("node:path"), os = require("node:os");
const parser = require("@darabonba/parser");
const { buildProduct, modelGraph } = require("./discovery.cjs");
const { lowerOperation } = require("./frontend.cjs");
const { xmlReviewer, compareShape } = require("./oss-xml-shape.cjs");
const { project, run } = require("./oss-discover.cjs");
const root = path.resolve(__dirname, "../.."), file = "products/oss/main.tea";
const main = path.join(root, "sources/darabonba", file), source = fs.readFileSync(main, "utf8");
const ast = parser.parse(source, main);
const inventory = JSON.parse(fs.readFileSync(path.join(root, "metadata/native-helper-pins/oss-inventory.json")));
const operation = (tree, action = "GetBucketAcl") => tree.moduleBody.nodes.find(n => n.functionName?.lexeme === action[0].toLowerCase() + action.slice(1) + "WithOptions");
const statements = (tree, action) => operation(tree, action).functionBody.stmts.stmts;
const field = (list, name) => list.find(f => f.fieldName?.lexeme === name);
const protocol = (tree, action) => statements(tree, action).find(n => n.id?.lexeme === "params").expr.object.fields;
const lower = (tree, action = "GetBucketAcl", facts = inventory) => lowerOperation(tree, action, { reviewXML: xmlReviewer(modelGraph(tree, file), facts, file) });
function temporaryRoot(t) {
  const temp = fs.mkdtempSync(path.join(os.tmpdir(), "oss-semantic-"));
  t.after(() => {
    assert.equal(path.dirname(path.resolve(temp)), path.resolve(os.tmpdir()));
    assert.ok(path.basename(temp).startsWith("oss-semantic-"));
    fs.rmSync(temp, { recursive: true, force: true });
  });
  return temp;
}
function copyInputs(temp) {
  fs.cpSync(path.join(root, "sources/darabonba"), path.join(temp, "sources/darabonba"), { recursive: true });
  fs.cpSync(path.join(root, "metadata/native-helper-pins"), path.join(temp, "metadata/native-helper-pins"), { recursive: true });
  fs.copyFileSync(path.join(root, "metadata/oss-semantic-pins.json"), path.join(temp, "metadata/oss-semantic-pins.json"));
}

test("complete pinned OSS discovery separates frontend acceptance from public coverage", () => {
  const result = project();
  assert.equal(result.coverage.counts.discovered, 90);
  assert.equal(result.coverage.counts.declaredModels, 303);
  assert.equal(result.coverage.counts.lowered, 16);
  assert.equal(result.coverage.counts.unsupported, 74);
  assert.equal(result.coverage.acceptance.goEmission, "not-assessed");
  assert.equal(result.coverage.acceptance.compilation, "not-assessed");
  assert.equal(result.coverage.acceptance.live, "not-assessed");
  assert.deepEqual(result, project());
  assert.equal(result.ir.endpoints.profile, "gateway-explicit-origin-v1");
  const acl = result.ir.operations.find(o => o.name === "GetBucketAcl");
  assert.deepEqual(acl.oss.query, [{ name: "acl", value: "" }]);
  assert.equal(acl.protocol.pathname, "/");
  assert.equal(acl.protocol.reqBodyType, "xml");
  assert.equal(acl.oss.response.name, "AccessControlPolicy");
  assert.equal(acl.oss.response.namespaceMatch, "local");
  assert.equal(acl.bindings.find(b => b.location === "bucket").guard, "required");
  const location = result.ir.operations.find(o => o.name === "GetBucketLocation");
  assert.equal(location.oss.response.scalarField, "LocationConstraint");
  const list = result.ir.operations.find(o => o.name === "ListObjectsV2");
  assert.deepEqual(list.oss.query, [{ name: "list-type", value: "2" }]);
  assert.ok(list.bindings.some(b => b.wire === "continuation-token" && b.location === "query"));
});

test("renamed complete product reuses the parser, Gateway lowerer and explicit native roots", () => {
  const renamedSource = source.replaceAll("GetBucketAcl", "ReadAccessPolicy").replaceAll("getBucketAcl", "readAccessPolicy");
  const renamed = parser.parse(renamedSource, main), native = structuredClone(inventory);
  native.roots.find(r => r.action === "GetBucketAcl").action = "ReadAccessPolicy";
  const result = buildProduct(renamed, { pkg: "storagefixture", identifier: "storagefixture-20190517", file: "products/storagefixture/main.tea", provenance: { fixture: true }, nativeXML: native });
  const read = result.ir.operations.find(o => o.name === "ReadAccessPolicy");
  assert.equal(read.status, "lowered");
  assert.equal(read.protocol.action, "ReadAccessPolicy");
  assert.equal(read.oss.response.name, "AccessControlPolicy");
  assert.equal(result.coverage.counts.lowered, 16);
  assert.ok(result.ir.models.some(m => m.id === "ReadAccessPolicy.$input"));
  assert.ok(!result.ir.operations.some(o => o.name === "GetBucketAcl"));
});

test("complete semantic program rejects initialization, addressing and handoff mutations", () => {
  const mutations = [
    ["initializer", tree => tree.moduleBody.nodes.find(n => n.type === "init").initBody.stmts.push({ type: "return" })],
    ["host key", tree => statements(tree)[1].left.accessKey.value.string = "userId"],
    ["host source", tree => statements(tree)[1].expr.id.lexeme = "headers"],
    ["host handoff", tree => field(statements(tree)[2].expr.object.fields, "hostMap").expr.id.lexeme = "headers"],
    ["request body", tree => statements(tree)[2].expr.object.fields.push({ type: "objectField", fieldName: { lexeme: "body" }, expr: { type: "object", fields: [] } })],
    ["method", tree => field(protocol(tree), "method").expr.value.string = "PUT"],
    ["duplicate constant", tree => protocol(tree).push(structuredClone(field(protocol(tree), "bodyType")))],
    ["duplicate subresource", tree => field(protocol(tree), "pathname").expr.elements[0].value.string = "/?acl&acl"],
    ["execute", tree => statements(tree).at(-1).expr.left.id.lexeme = "callApi"],
    ["execute arguments", tree => statements(tree).at(-1).expr.args[1].id.lexeme = "headers"],
    ["trailing statement", tree => statements(tree).push({ type: "return", expr: { type: "object", fields: [] } })],
    ["hidden statement", tree => statements(tree).splice(2, 0, { type: "return", expr: { type: "object", fields: [] } })],
  ];
  for (const [name, mutate] of mutations) {
    const tree = structuredClone(ast); mutate(tree);
    assert.throws(() => lower(tree), /unsupported SDK pattern/, name);
  }
  assert.throws(() => lowerOperation(ast, "GetBucketAcl"), /native XML facts missing/);
});

test("recursive XML differences retain exact DSL and native coordinates without rewriting fields", () => {
  const { ir } = project();
  const cors = ir.operations.find(o => o.name === "GetBucketCors");
  const difference = cors.reasons[0].differences.find(d => d.path.endsWith("AllowedHeader"));
  assert.equal(difference.dslKind, "scalar"); assert.equal(difference.nativeKind, "array");
  assert.equal(difference.dslSource.line, 150); assert.equal(difference.nativeSource.line, 837);
  assert.ok(!cors.reasons[0].differences.some(d => d.path.endsWith("ExposeHeader")));
  const buckets = ir.operations.find(o => o.name === "ListBuckets");
  assert.equal(buckets.reasons[0].code, "DSL_OSS_XML_SHAPE");
  assert.ok(buckets.reasons[0].differences.some(d => d.path.endsWith("buckets")));
  assert.ok(ir.models.find(m => m.id === "Bucket").fields.some(f => f.wireName === "ResourceGroupId"));
  for (const difference of ir.operations.flatMap(o => o.reasons.flatMap(r => r.differences || []))) {
    assert.ok(difference.dslSource?.line > 0); assert.ok(difference.nativeSource?.line > 0);
  }
});

test("XML registry facts require unique actions/models and supported explicit root bindings", () => {
  const graph = modelGraph(ast, file);
  assert.throws(() => xmlReviewer(graph, { schemaVersion: 1, roots: [inventory.roots[0], inventory.roots[0]], models: [] }), /Duplicate/);
  assert.throws(() => xmlReviewer(graph, { schemaVersion: 1, roots: [], models: [inventory.models[0], inventory.models[0]] }), /Duplicate/);
  for (const change of [facts => facts.roots = facts.roots.filter(r => r.action !== "GetBucketAcl"), facts => facts.roots.find(r => r.action === "GetBucketAcl").namespaceMatch = "guess", facts => facts.models.find(m => m.name === "Owner").fields[0].xmlOptions = ["attr"]]) {
    const facts = structuredClone(inventory); change(facts); assert.throws(() => lower(ast, "GetBucketAcl", facts), /unsupported SDK pattern/);
  }
  const recursive = { models: new Map([["Loop", { fields: [{ wireName: "Child", type: { kind: "model", ref: "Loop" } }] }]]) };
  const findings = compareShape(recursive, { kind: "model", ref: "Loop" }, { kind: "model", name: "Native" }, new Map([["Native", { fields: [{ jsonName: "Child", xmlName: "Child", type: { kind: "model", name: "Native" } }] }]]), "Loop");
  assert.equal(findings[0].code, "XML_RECURSION");
});

test("selected unsupported or unknown operations fail before writing either report", () => {
  const files = ["oss-semantic-ir.json", "oss-semantic-coverage.json"].map(name => path.join(root, "docs/research", name));
  const before = files.map(file => fs.readFileSync(file));
  for (const selected of [["GetBucketCors"], ["PutBucketCors"], ["MissingAction"]]) assert.throws(() => run("generate", { selected }), /unknown\/unsupported/);
  files.forEach((file, index) => assert.deepEqual(fs.readFileSync(file), before[index]));
});

test("offline projection rejects inventory and source hash drift before output", t => {
  const temp = temporaryRoot(t); copyInputs(temp);
  const nativeFile = path.join(temp, "metadata/native-helper-pins/oss-inventory.json"), bytes = fs.readFileSync(nativeFile);
  fs.writeFileSync(nativeFile, Buffer.concat([bytes, Buffer.from(" ")]));
  assert.throws(() => run("generate", { root: temp }), /checksum differs/);
  assert.equal(fs.existsSync(path.join(temp, "docs/research/oss-semantic-ir.json")), false);
  fs.writeFileSync(nativeFile, bytes);
  fs.appendFileSync(path.join(temp, "sources/darabonba/products/oss/main.tea"), "\n");
  assert.throws(() => run("generate", { root: temp }), /source checksum mismatch/);
  assert.equal(fs.existsSync(path.join(temp, "docs/research/oss-semantic-ir.json")), false);
});

test("dangling output symlinks cannot redirect either semantic report", { skip: process.platform === "win32" ? "Windows symlink privileges are not assumed; Linux CI verifies this contract" : false }, t => {
  const temp = temporaryRoot(t); copyInputs(temp);
  const output = path.join(temp, "docs/research"); fs.mkdirSync(output, { recursive: true });
  const redirected = path.join(temp, "redirected.json");
  fs.symlinkSync(redirected, path.join(output, "oss-semantic-coverage.json"));
  assert.throws(() => run("generate", { root: temp }), /output symlink/);
  assert.equal(fs.existsSync(redirected), false);
  assert.equal(fs.existsSync(path.join(output, "oss-semantic-ir.json")), false);
});
