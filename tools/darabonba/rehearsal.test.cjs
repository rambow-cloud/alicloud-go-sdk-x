"use strict";
const { test } = require("node:test");
const assert = require("node:assert/strict");
const fs = require("node:fs");
const path = require("node:path");
const os = require("node:os");
const { revisions, prepare } = require("./rehearse-sts.cjs");
const { run } = require("./discovery.cjs");
const { lowerOperation } = require("./frontend.cjs");
const parser = require("@darabonba/parser");
const root = path.resolve(__dirname, "../..");
function temporary(t) {
  const dir = fs.mkdtempSync(path.join(os.tmpdir(), "sts-rehearsal-"));
  t.after(() => {
    assert.equal(path.dirname(path.resolve(dir)), path.resolve(os.tmpdir()));
    fs.rmSync(dir, { recursive: true, force: true });
  });
  return dir;
}

test("real source revision exposes initializer and handoff drift without promoting historical coverage", (t) => {
  const report = prepare(path.join(temporary(t), "fresh"));
  assert.equal(report.counts.baseline.discovered, 4);
  assert.equal(report.counts.baseline.lowered, 0);
  assert.equal(report.counts.candidate.lowered, 4);
  assert.equal(report.productionSTSSourceUnchanged, true);
  assert.equal(report.drift.signatureAlgorithm.baseline, "v2");
  assert.equal(report.drift.signatureAlgorithm.candidate, "inherited-default");
  assert.equal(report.drift.endpointInitializationChanged, true);
  assert.equal(report.drift.documentationTextChanged, false);
  assert.equal(report.drift.documentationCoordinatesChanged, true);
  assert.deepEqual(report.drift.addedFields, []);
  assert.deepEqual(report.drift.removedFields, []);
  assert.deepEqual(report.drift.changedFields, []);
  assert.equal(
    report.drift.handoffs.baseline.find(
      (o) => o.name === "assumeRoleWithOIDCWithOptions",
    ).handoff,
    "callApi",
  );
  assert.equal(
    report.drift.handoffs.candidate.find(
      (o) => o.name === "assumeRoleWithOIDCWithOptions",
    ).handoff,
    "doRPCRequest",
  );
  const destination = path.join(temporary(t), "selection");
  prepare(destination);
  assert.throws(
    () =>
      run("generate", {
        root: path.join(destination, "baseline"),
        selected: ["sts/AssumeRole"],
      }),
    /DSL_PRODUCT_AUTH_INITIALIZER/,
  );
  assert.throws(() => prepare(destination), /already exists/);
});

test("signed initializer rejects dynamic and unsupported overrides while explicit ACS3 remains supported", () => {
  const file = path.join(root, "sources/darabonba/products/sts/main.tea");
  const ast = parser.parse(fs.readFileSync(file, "utf8"), file);
  const initializer = ast.moduleBody.nodes.find((n) => n.type === "init");
  const template = initializer.initBody.stmts.find((n) => n.type === "assign");
  for (const value of ["v2", "FutureAlgorithm", null, "ACS3-HMAC-SHA256"]) {
    const changed = structuredClone(ast),
      statement = structuredClone(template);
    statement.left.vid.lexeme = "@signatureAlgorithm";
    statement.expr =
      value === null
        ? { type: "variable", id: { lexeme: "config" } }
        : { type: "string", value: { string: value } };
    changed.moduleBody.nodes
      .find((n) => n.type === "init")
      .initBody.stmts.push(statement);
    if (value === "ACS3-HMAC-SHA256")
      assert.equal(
        lowerOperation(changed, "AssumeRole").protocol.authType,
        "AK",
      );
    else
      assert.throws(
        () => lowerOperation(changed, "AssumeRole"),
        /signed product signature initializer/,
      );
  }
});

test("revision fixture checksum and traversal fail before outputs", (t) => {
  const dir = temporary(t),
    base = path.join(dir, "tools/darabonba/fixtures/sts-revisions");
  fs.cpSync(path.join(root, "tools/darabonba/fixtures/sts-revisions"), base, {
    recursive: true,
  });
  const manifestFile = path.join(base, "manifest.json");
  const manifest = JSON.parse(fs.readFileSync(manifestFile));
  const original = fs.readFileSync(path.join(base, manifest.files[0].file));
  fs.appendFileSync(path.join(base, manifest.files[0].file), " ");
  assert.throws(() => revisions(dir), /checksum mismatch/);
  fs.writeFileSync(path.join(base, manifest.files[0].file), original);
  manifest.files[0].file = "../outside";
  fs.writeFileSync(manifestFile, JSON.stringify(manifest));
  assert.throws(() => revisions(dir), /unsafe/);
});
