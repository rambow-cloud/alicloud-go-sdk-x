"use strict";

const { test } = require("node:test");
const assert = require("node:assert/strict");
const fs = require("node:fs");
const path = require("node:path");
const os = require("node:os");
const crypto = require("node:crypto");
const parser = require("@darabonba/parser");
const { reuseFixture } = require("./reuse-fixture.cjs");
const {
  buildProduct,
  project,
  run,
  modelGraph,
  discoverCandidates,
} = require("./discovery.cjs");
const root = path.resolve(__dirname, "../..");
const file = "products/sts/main.tea";
const main = path.join(root, "sources/darabonba", file);
const ast = parser.parse(fs.readFileSync(main, "utf8"), main);
const info = JSON.parse(
  fs.readFileSync(
    path.join(root, "sources/darabonba/products/sts/api-info.json"),
  ),
);
const context = {
  pkg: "sts",
  identifier: "sts-20150401",
  info,
  file,
  provenance: { fixture: true },
};
const roleFunction = (a) =>
  a.moduleBody.nodes.find(
    (n) => n.functionName?.lexeme === "assumeRoleWithOptions",
  );
const roleRequest = (a) =>
  a.moduleBody.nodes.find((n) => n.modelName?.lexeme === "AssumeRoleRequest");
const params = (fn) =>
  fn.functionBody.stmts.stmts.find((n) => n.id?.lexeme === "params").expr.object
    .fields;

test("renamed actions use the official parser and shared discovery without STS names", () => {
  const { ir, coverage } = reuseFixture(root);
  assert.equal(ir.product, "authfixture");
  assert.equal(ir.models.length, 19);
  assert.equal(coverage.counts.lowered, 4);
  assert.deepEqual(ir.operations.map((op) => op.name).sort(), [
    "ExchangeAssertion",
    "ExchangeIdentity",
    "InspectIdentity",
    "ObtainRole",
  ]);
  for (const op of ir.operations) {
    assert.equal(op.status, "lowered");
    assert.equal(op.protocol.action, op.name);
    assert.equal(op.origin, "dsl-only");
    assert.equal(
      op.protocol.authType,
      op.name.startsWith("Exchange") ? "Anonymous" : "AK",
    );
  }
  const identity = ir.operations.find((op) => op.name === "InspectIdentity");
  assert.equal(identity.roots.request.kind, "empty");
  assert.deepEqual(identity.bindings, []);
  const role = ir.operations.find((op) => op.name === "ObtainRole");
  assert.ok(role.bindings.some((binding) => binding.wire === "RoleArn"));
  const anonymous = ir.operations.find((op) => op.name === "ExchangeIdentity");
  assert.equal(anonymous.handoff, "doRPCRequest");
  assert.ok(anonymous.bindings.some((binding) => binding.wire === "OIDCToken"));
});

test("complete real corpus is deterministic and accounts for every SDK operation and declared model", () => {
  const first = project(root),
    second = project(root);
  assert.deepEqual(first.files, second.files);
  for (const [relative, data] of Object.entries(first.files))
    assert.deepEqual(
      data,
      fs.readFileSync(path.join(root, relative)),
      relative,
    );
  for (const [pkg, product] of Object.entries(first.products)) {
    const { ir, coverage } = product;
    assert.equal(ir.schemaVersion, 4);
    const sourceFile = path.join(
      root,
      "sources/darabonba/products",
      pkg,
      "main.tea",
    );
    const parsed = parser.parse(
      fs.readFileSync(sourceFile, "utf8"),
      sourceFile,
    );
    const functions = parsed.moduleBody.nodes.filter((n) =>
      n.functionName?.lexeme?.endsWith("WithOptions"),
    );
    assert.equal(coverage.counts.dslOperations, functions.length);
    assert.equal(
      coverage.counts.discovered,
      coverage.counts.lowered + coverage.counts.unsupported,
    );
    assert.equal(
      coverage.counts.declaredModels,
      coverage.counts.reachableNamedModels + coverage.counts.unusedNamedModels,
    );
    assert.equal(coverage.acceptance.goEmission, "not-assessed");
    assert.equal(coverage.acceptance.compilation, "not-assessed");
    assert.equal(coverage.acceptance.live, "not-assessed");
    for (const op of ir.operations) {
      assert.ok(op.source.line > 0, op.name);
      assert.ok(
        op.reachableModels.every((id) =>
          ir.models.some((model) => model.id === id),
        ),
      );
      assert.equal(op.status === "lowered", op.reasons.length === 0);
      assert.ok(
        op.reasons.every(
          (reason) => reason.code && reason.source.file && reason.message,
        ),
      );
    }
  }
  const lock = JSON.parse(first.files["models/manifest.json"]);
  assert.equal(lock.schemaVersion, 4);
  for (const artifact of lock.files)
    assert.equal(
      crypto
        .createHash("sha256")
        .update(first.files[artifact.file])
        .digest("hex"),
      artifact.sha256,
    );
});

test("reachable graph preserves numeric source types, response envelopes, inline models and document coordinates", () => {
  const { ir } = buildProduct(ast, context);
  const request = ir.models.find((m) => m.id === "AssumeRoleRequest");
  const seconds = request.fields.find((f) => f.wireName === "DurationSeconds");
  assert.deepEqual(seconds.type, {
    kind: "scalar",
    dslType: "long",
    wireType: "integer",
  });
  assert.equal(seconds.required, false);
  assert.equal(
    seconds.documentation.find((d) => d.attribute === "description").source
      .line,
    66,
  );
  const envelope = ir.models.find((m) => m.id === "AssumeRoleResponse");
  assert.equal(
    envelope.fields.find((f) => f.wireName === "headers").type.kind,
    "map",
  );
  const role = ir.operations.find((o) => o.name === "AssumeRole");
  assert.equal(role.roots.body.ref, "AssumeRoleResponseBody");
  assert.ok(
    role.reachableModels.includes("AssumeRoleResponseBody.credentials"),
  );
  assert.equal(
    role.parameters.find((p) => p.name === "runtime").type.dslType,
    "Util.RuntimeOptions",
  );
  assert.ok(
    role.bindings.some((b) => b.wire === "RoleArn" && b.field === "roleArn"),
  );
});

test("catalog is an optional cross-check rather than a discovery whitelist", () => {
  const changed = structuredClone(info);
  changed.apiDoc.hasDoc = ["AssumeRole", "CatalogOnly"];
  const { ir, coverage } = buildProduct(ast, { ...context, info: changed });
  assert.equal(ir.operations.length, 5);
  assert.equal(
    ir.operations.find((o) => o.name === "GetCallerIdentity").origin,
    "dsl-only",
  );
  assert.equal(
    ir.operations.find((o) => o.name === "CatalogOnly").reasons[0].code,
    "DSL_MISSING_OPERATION",
  );
  assert.equal(coverage.counts.catalogOnlyOperations, 1);
  const noCatalog = buildProduct(ast, { ...context, info: null });
  assert.equal(noCatalog.ir.operations.length, 4);
  assert.equal(noCatalog.coverage.counts.dslOnlyOperations, 4);
});

test("new DSL operations are lowered automatically without catalog or per-operation configuration", () => {
  const changed = structuredClone(ast),
    added = structuredClone(roleFunction(changed));
  added.functionName.lexeme = "futureOperationWithOptions";
  params(added).find((f) => f.fieldName.lexeme === "action").expr.value.string =
    "FutureOperation";
  changed.moduleBody.nodes.push(added);
  const product = buildProduct(changed, context);
  const operation = product.ir.operations.find(
    (o) => o.name === "FutureOperation",
  );
  assert.equal(operation.origin, "dsl-only");
  assert.equal(operation.status, "lowered");
  assert.equal(operation.protocol.action, "FutureOperation");
  assert.ok(operation.reachableModels.includes("AssumeRoleRequest"));
  assert.equal(product.coverage.counts.discovered, 5);
});

test("ambiguous and dynamic actions remain visible and unsupported", () => {
  const changed = structuredClone(ast),
    fn = roleFunction(changed);
  params(fn).find((f) => f.fieldName.lexeme === "action").expr = {
    type: "variable",
    id: { lexeme: "request" },
  };
  let result = buildProduct(changed, context).ir.operations.find(
    (o) => o.name === "AssumeRole",
  );
  assert.equal(result.reasons[0].code, "DSL_PROTOCOL_CONSTANTS");
  assert.equal(
    result.declarations[0].protocolEvidence[0].fields.find(
      (f) => f.name === "action",
    ).expressionKind,
    "variable",
  );
  const extra = structuredClone(
    fn.functionBody.stmts.stmts.find((n) => n.id?.lexeme === "params"),
  );
  extra.expr.object.fields.find((f) => f.fieldName.lexeme === "action").expr = {
    type: "string",
    value: { string: "OtherAction" },
  };
  fn.functionBody.stmts.stmts.push(extra);
  result = buildProduct(changed, context).ir.operations.find(
    (o) => o.name === "OtherAction",
  );
  assert.equal(result.status, "unsupported");
  assert.equal(result.reasons[0].code, "DSL_OPERATION_FUNCTION");
  params(fn).find((f) => f.fieldName.lexeme === "action").expr = {
    type: "string",
    value: { string: "AssumeRole" },
  };
  result = buildProduct(changed, context).ir.operations.find(
    (o) => o.name === "AssumeRole",
  );
  assert.equal(result.reasons[0].code, "DSL_AMBIGUOUS_ACTION");
});

test("non-WithOptions protocol functions and API declarations cannot silently disappear", () => {
  const changed = structuredClone(ast);
  const fn = roleFunction(changed);
  fn.functionName.lexeme = "customRole";
  const result = discoverCandidates(changed, null, file);
  assert.ok(result.some((o) => o.name === "AssumeRole"));
  const api = structuredClone(fn);
  api.type = "api";
  api.apiName = { ...api.functionName, lexeme: "futureApi" };
  delete api.functionName;
  api.apiBody = api.functionBody;
  delete api.functionBody;
  changed.moduleBody.nodes.push(api);
  const product = buildProduct(changed, { ...context, info: null });
  assert.equal(
    product.ir.operations.find((o) => o.name === "AssumeRole").reasons[0].code,
    "DSL_AMBIGUOUS_OPERATION",
  );
});

for (const [description, change, code] of [
  [
    "ROA profile",
    (a) => {
      params(roleFunction(a)).find(
        (f) => f.fieldName.lexeme === "style",
      ).expr.value.string = "ROA";
    },
    "DSL_PROTOCOL_PROFILE",
  ],
  [
    "helper transform",
    (a) => {
      roleFunction(a).functionBody.stmts.stmts[2].stmts.stmts[0].expr = {
        type: "call",
        left: {
          type: "static_call",
          id: { lexeme: "Util" },
          propertyPath: [{ lexeme: "transform" }],
        },
        args: [],
      };
    },
    "DSL_QUERY_TRANSFORM",
  ],
  [
    "unknown wire type",
    (a) => {
      roleRequest(a).modelBody.nodes[0].fieldValue.fieldType = "futureWireType";
    },
    "DSL_WIRE_TYPE",
  ],
  [
    "recursive wire model",
    (a) => {
      roleRequest(a).modelBody.nodes[0].fieldValue.fieldType = {
        lexeme: "AssumeRoleRequest",
      };
    },
    "DSL_RECURSIVE_MODEL",
  ],
  [
    "unreviewed attributes",
    (a) => {
      roleRequest(a).modelBody.nodes[0].attrs.push({
        attrName: { lexeme: "futureAttribute" },
        attrValue: { string: "value" },
      });
    },
    "DSL_INPUT_ATTRIBUTE",
  ],
])
  test(
    "unsupported " +
      description +
      " retains a reason, source and reachable inventory",
    () => {
      const changed = structuredClone(ast);
      change(changed);
      const { ir } = buildProduct(changed, context);
      const role = ir.operations.find((o) => o.name === "AssumeRole");
      assert.equal(role.status, "unsupported");
      assert.equal(role.reasons[0].code, code);
      assert.equal(role.reasons[0].source.line, 147);
      assert.ok(role.reachableModels.includes("AssumeRoleRequest"));
      if (description === "helper transform")
        assert.ok(
          role.declarations[0].bindingEvidence.some(
            (b) => b.helper === "Util.transform",
          ),
        );
    },
  );

test("duplicate wire fields and malformed catalogs fail rather than becoming supported IR", () => {
  const changed = structuredClone(ast);
  roleRequest(changed).modelBody.nodes.push(
    structuredClone(roleRequest(changed).modelBody.nodes[0]),
  );
  assert.throws(
    () => buildProduct(changed, context),
    /duplicate\/invalid wire model member/,
  );
  const catalog = structuredClone(info);
  catalog.apiDoc.noDoc.push("AssumeRole");
  assert.throws(
    () => buildProduct(ast, { ...context, info: catalog }),
    /duplicate upstream catalog/,
  );
});

test("cycle-safe graph stores named references instead of duplicating recursive models", () => {
  const changed = structuredClone(ast);
  roleRequest(changed).modelBody.nodes[0].fieldValue.fieldType = {
    lexeme: "AssumeRoleRequest",
  };
  const graph = modelGraph(changed, file);
  graph.type(
    { lexeme: "AssumeRoleRequest" },
    "unused",
    roleRequest(changed).modelName,
  );
  assert.deepEqual(
    graph.reachable([{ kind: "model", ref: "AssumeRoleRequest" }]),
    ["AssumeRoleRequest"],
  );
  assert.equal(graph.models.size, 1);
});

function sourceFixture(t) {
  const temporary = fs.mkdtempSync(
    path.join(os.tmpdir(), "product-discovery-"),
  );
  // Cleanup is confined to the exact task-owned temporary directory created above.
  const temporaryBase = path.resolve(os.tmpdir());
  t.after(() => {
    assert.equal(path.dirname(path.resolve(temporary)), temporaryBase);
    fs.rmSync(temporary, { recursive: true, force: true });
  });
  fs.mkdirSync(path.join(temporary, "sources"));
  fs.cpSync(
    path.join(root, "sources/darabonba"),
    path.join(temporary, "sources/darabonba"),
    { recursive: true },
  );
  return temporary;
}
test("discovery runs without metadata, decisions, overlays or canonical fixtures; check is read-only", (t) => {
  const temporary = sourceFixture(t);
  assert.throws(() => run("check", { root: temporary }), /artifact drift/);
  assert.equal(fs.existsSync(path.join(temporary, "models")), false);
  run("generate", {
    root: temporary,
    selected: ["ecs/DescribeImages", "sts/AssumeRole"],
  });
  run("check", { root: temporary });
  const target = path.join(temporary, "models/ecs/ir.json"),
    before = fs.readFileSync(target);
  fs.writeFileSync(target, "modified");
  assert.throws(() => run("check", { root: temporary }), /artifact drift/);
  assert.equal(fs.readFileSync(target, "utf8"), "modified");
  assert.ok(before.length > 0);
});
test("strict unknown/unsupported selections fail before creating outputs", (t) => {
  const temporary = sourceFixture(t);
  for (const target of [
    "ecs/DoesNotExist",
    "vpc/DoesNotExist",
    "../AssumeRole",
  ]) {
    assert.throws(
      () => run("generate", { root: temporary, selected: [target] }),
      /selected operation/,
    );
    assert.equal(fs.existsSync(path.join(temporary, "models")), false);
  }
});
test("source checksum failures abort discovery before any outputs", (t) => {
  const temporary = sourceFixture(t);
  fs.appendFileSync(
    path.join(temporary, "sources/darabonba/products/sts/main.tea"),
    " ",
  );
  assert.throws(
    () => run("generate", { root: temporary }),
    /source checksum mismatch/,
  );
  assert.equal(fs.existsSync(path.join(temporary, "models")), false);
});
test("output directory symlinks fail during path preflight without writing targets", (t) => {
  const temporary = sourceFixture(t),
    other = path.join(temporary, "other");
  fs.mkdirSync(other);
  try {
    fs.symlinkSync(
      other,
      path.join(temporary, "models"),
      process.platform === "win32" ? "junction" : "dir",
    );
  } catch (error) {
    if (error.code === "EPERM") {
      t.skip("Directory symlinks require unavailable OS permission");
      return;
    }
    throw error;
  }
  assert.throws(
    () => run("generate", { root: temporary }),
    /symlink output path/,
  );
  assert.deepEqual(fs.readdirSync(other), []);
});
test("report distinguishes lowering from Go/live acceptance and prints reason locations", () => {
  const text = run("report", { product: "sts" });
  assert.match(text, /^sts: 4 discovered, 4 lowered, 0 unsupported/);
  assert.match(text, /Go emission\/compilation\/live: not assessed/);
  assert.match(
    run("report", { product: "ecs" }),
    /^ecs: 380 discovered, 380 lowered, 0 unsupported/,
  );
  assert.match(
    run("report", { product: "vpc" }),
    /^vpc: 403 discovered, 403 lowered, 0 unsupported/,
  );
});

test("requestless discovery retains absent source model without a synthetic DSL declaration", () => {
  const { ir } = buildProduct(ast, context);
  const identity = ir.operations.find((o) => o.name === "GetCallerIdentity");
  assert.equal(identity.status, "lowered");
  assert.deepEqual(identity.roots.request, { kind: "empty" });
  assert.deepEqual(identity.bindings, []);
  assert.deepEqual(
    identity.parameters.map((p) => p.name),
    ["runtime"],
  );
  assert.ok(identity.reachableModels.includes("GetCallerIdentityResponseBody"));
  assert.ok(!ir.models.some((m) => m.id === "GetCallerIdentityRequest"));
});
