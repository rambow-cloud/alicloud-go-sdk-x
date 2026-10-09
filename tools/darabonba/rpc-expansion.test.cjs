"use strict";
const { test } = require("node:test");
const assert = require("node:assert/strict");
const fs = require("node:fs");
const path = require("node:path");
const parser = require("@darabonba/parser");
const { lowerOperation } = require("./frontend.cjs");
const { buildProduct } = require("./discovery.cjs");
const main = path.resolve(
  __dirname,
  "../../sources/darabonba/products/ecs/main.tea",
);
const ast = parser.parse(fs.readFileSync(main, "utf8"), main);
const fn = (a, name) =>
  a.moduleBody.nodes.find(
    (n) =>
      n.functionName?.lexeme ===
      name[0].toLowerCase() + name.slice(1) + "WithOptions",
  );
const stmts = (a, name) => fn(a, name).functionBody.stmts.stmts;

test("real ECS corpus lowers every action with complete typed public roots", () => {
  const { ir, coverage } = buildProduct(ast, {
    pkg: "ecs",
    identifier: "ecs-20140526",
    info: JSON.parse(
      fs.readFileSync(path.join(path.dirname(main), "api-info.json")),
    ),
    file: "products/ecs/main.tea",
    provenance: { fixture: true },
  });
  assert.equal(coverage.counts.discovered, 380);
  assert.equal(coverage.counts.lowered, 380);
  const invoke = ir.operations.find((o) => o.name === "InvokeCommand");
  assert.equal(invoke.roots.request.ref, "InvokeCommandRequest");
  assert.deepEqual(
    invoke.bindings.find((b) => b.wire === "Parameters").encoding,
    "json",
  );
  assert.equal(
    invoke.bindings.find((b) => b.wire === "Parameters").field,
    "parameters",
  );
  const interfaces = ir.models.find(
    (m) => m.id === "DescribeNetworkInterfacesRequest",
  );
  assert.equal(
    interfaces.fields.find((f) => f.dslName === "pageNumber").attributes
      .deprecated,
    true,
  );
});

test("identical duplicate guarded bindings normalize without changing source AST", () => {
  const before = JSON.stringify(fn(ast, "RunInstances"));
  const op = lowerOperation(ast, "RunInstances");
  assert.equal(op.inputs.filter((i) => i.wire === "RegionId").length, 1);
  assert.equal(JSON.stringify(fn(ast, "RunInstances")), before);
  const component = lowerOperation(ast, "CreateImageComponent");
  assert.equal(
    component.inputs.filter((i) => i.wire === "ResourceGroupId").length,
    1,
  );
});

test("duplicate query aliases and conflicting guard/source assignments stay rejected", () => {
  for (const mutate of [
    (guard) => {
      guard.stmts.stmts[0].expr.propertyPath[0].lexeme = "ownerAccount";
    },
    (guard) => {
      guard.condition.expr.args[0].propertyPath[0].lexeme = "ownerAccount";
      guard.stmts.stmts[0].expr.propertyPath[0].lexeme = "ownerAccount";
    },
    (guard) => {
      guard.elseStmts = { type: "stmts", stmts: [] };
    },
  ]) {
    const a = structuredClone(ast);
    const guards = stmts(a, "AcceptInquiredSystemEvent").filter(
      (s) =>
        s.type === "if" &&
        s.condition.expr?.args?.[0]?.propertyPath?.[0]?.lexeme === "regionId",
    );
    mutate(guards[1]);
    assert.throws(
      () => lowerOperation(a, "AcceptInquiredSystemEvent"),
      /query (guard|alias|assignment)|direct query assignment/,
    );
  }
});

test("all ten official ECS shrink flows retain structured inputs and exact JSON query keys", () => {
  for (const name of [
    "CreateAutoProvisioningGroup",
    "CreateDiagnosticReport",
    "CreatePlanMaintenanceWindow",
    "DescribePlanMaintenanceWindows",
    "InvokeCommand",
    "ModifyCloudAssistantSettings",
    "ModifyInvocationAttribute",
    "ModifyPlanMaintenanceWindow",
    "RunCommand",
    "StartTerminalSession",
  ]) {
    const op = lowerOperation(ast, name);
    const encoded = op.inputs.filter((i) => i.encoding === "json");
    assert.ok(encoded.length > 0, name);
    assert.ok(
      encoded.every((i) => i.field && !i.field.endsWith("Shrink")),
      name,
    );
  }
});

for (const [label, mutate] of [
  ["conversion direction", (ss) => ss[2].args.reverse()],
  ["conversion helper", (ss) => (ss[2].left.propertyPath[0].lexeme = "other")],
  ["shrink model", (ss) => (ss[1].expr.aliasId.lexeme = "OtherShrinkRequest")],
  [
    "JSON style",
    (ss) => (ss[3].stmts.stmts[0].expr.args[2].value.string = "repeatList"),
  ],
  [
    "JSON prefix",
    (ss) => (ss[3].stmts.stmts[0].expr.args[1].value.string = "Other"),
  ],
  [
    "JSON source",
    (ss) =>
      (ss[3].stmts.stmts[0].expr.args[0].propertyPath[0].lexeme = "commandId"),
  ],
  [
    "JSON target",
    (ss) => (ss[3].stmts.stmts[0].left.propertyPath[0].lexeme = "commandId"),
  ],
  [
    "transform guard",
    (ss) => (ss[3].condition.expr.args[0].propertyPath[0].lexeme = "commandId"),
  ],
  ["missing transform", (ss) => ss.splice(3, 1)],
  ["trailing execution", (ss) => ss.push({ type: "declare" })],
])
  test("shrink drift rejects " + label, () => {
    const a = structuredClone(ast);
    mutate(stmts(a, "InvokeCommand"));
    assert.throws(
      () => lowerOperation(a, "InvokeCommand"),
      /unsupported SDK pattern/,
    );
  });

test("deprecated requires a boolean token and unknown attributes remain rejected", () => {
  for (const attr of [
    {
      attrName: { lexeme: "deprecated" },
      attrValue: { tag: 1, string: "true" },
    },
    { attrName: { lexeme: "future" }, attrValue: { tag: 13, lexeme: "true" } },
  ]) {
    const a = structuredClone(ast);
    a.moduleBody.nodes
      .find((n) => n.modelName?.lexeme === "InvokeCommandRequest")
      .modelBody.nodes[0].attrs.push(attr);
    assert.throws(
      () => lowerOperation(a, "InvokeCommand"),
      /unreviewed input attribute/,
    );
  }
});

test("renamed shrink action lowers through the same semantic parser and IR", () => {
  const changed = structuredClone(ast);
  const operation = fn(changed, "InvokeCommand");
  operation.functionName.lexeme = "executeTaskWithOptions";
  const params = operation.functionBody.stmts.stmts.find(
    (s) => s.id?.lexeme === "params",
  );
  params.expr.object.fields.find(
    (f) => f.fieldName.lexeme === "action",
  ).expr.value.string = "ExecuteTask";
  const lower = lowerOperation(changed, "ExecuteTask");
  assert.equal(lower.protocol.action, "ExecuteTask");
  assert.equal(
    lower.inputs.find((i) => i.wire === "Parameters").encoding,
    "json",
  );
  const { ir } = buildProduct(changed, {
    pkg: "fixture",
    identifier: "fixture-20140526",
    info: null,
    file: "products/ecs/main.tea",
    provenance: { fixture: true },
  });
  const op = ir.operations.find((o) => o.name === "ExecuteTask");
  assert.equal(op.status, "lowered");
  assert.equal(op.roots.request.ref, "InvokeCommandRequest");
  assert.equal(
    op.bindings.find((b) => b.wire === "Parameters").field,
    "parameters",
  );
});
