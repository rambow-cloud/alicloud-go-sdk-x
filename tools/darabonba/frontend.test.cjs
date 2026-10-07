"use strict";

const { test } = require("node:test");
const assert = require("node:assert/strict");
const fs = require("node:fs");
const path = require("node:path");
const os = require("node:os");
const crypto = require("node:crypto");
const parser = require("@darabonba/parser");
const { project, verifySources, lowerOperation } = require("./frontend.cjs");
const root = path.resolve(__dirname, "../..");
const main = path.join(root, "sources/darabonba/products/sts/main.tea");
const source = fs.readFileSync(main, "utf8");
const ast = parser.parse(source, main);
const operation = (a) =>
  a.moduleBody.nodes.find(
    (n) => n.functionName?.lexeme === "assumeRoleWithOptions",
  );

test("all real product projections are deterministic and match committed artifacts", () => {
  const first = project(root),
    second = project(root);
  assert.deepEqual(first, second);
  for (const [relative, data] of Object.entries(first))
    assert.deepEqual(
      data,
      fs.readFileSync(path.join(root, relative)),
      relative,
    );
  const sts = JSON.parse(first["metadata/sts/dsl.json"]);
  assert.equal(sts.operations[0].protocol.action, "AssumeRole");
  assert.equal(
    sts.operations[0].response.properties.Credentials.type,
    "object",
  );
  assert.equal(
    sts.operations[0].inputs.find((i) => i.wire === "RoleArn").schema.required,
    false,
  );
});

test("official parser performs semantic checking, including imported declarations", () => {
  assert.throws(() =>
    parser.parse(
      source.replace(
        "return callApi(params, req, runtime);",
        "return missingApi(params, req, runtime);",
      ),
      main,
    ),
  );
  assert.throws(() =>
    parser.parse(source.replace("import Util;", "import Unknown;"), main),
  );
});

for (const [name, mutate] of Object.entries({
  "extra request behavior": (fn) =>
    fn.functionBody.stmts.stmts.push({ type: "declare" }),
  "else behavior": (fn) => {
    fn.functionBody.stmts.stmts[2].elseStmts = { stmts: [{ type: "declare" }] };
  },
  "dynamic protocol": (fn) => {
    fn.functionBody.stmts.stmts.find(
      (s) => s.id?.lexeme === "params",
    ).expr.object.fields[0].expr = {
      type: "variable",
      id: { lexeme: "request" },
    };
  },
  "unrecognized query transform": (fn) => {
    fn.functionBody.stmts.stmts[2].stmts.stmts[0].expr = { type: "call" };
  },
  "request headers": (fn) => {
    fn.functionBody.stmts.stmts
      .find((s) => s.id?.lexeme === "req")
      .expr.object.fields.push({
        type: "objectField",
        fieldName: { lexeme: "headers" },
      });
  },
  "language override": (fn) => {
    fn.notes = [{ type: "note" }];
  },
  "unbound input": (fn) => {
    fn.functionBody.stmts.stmts.splice(2, 1);
  },
}))
  test("rejects " + name + " instead of discarding it", () => {
    const changed = structuredClone(ast);
    mutate(operation(changed));
    assert.throws(
      () => lowerOperation(changed, "AssumeRole"),
      /unsupported SDK pattern/,
    );
  });

function fixture(t) {
  const dir = fs.mkdtempSync(path.join(os.tmpdir(), "darabonba-source-"));
  t.after(() => fs.rmSync(dir, { recursive: true, force: true }));
  for (const name of ["products", "modules", "licenses"])
    fs.mkdirSync(path.join(dir, name));
  const data = Buffer.from("pinned");
  fs.writeFileSync(path.join(dir, "products/fixture.tea"), data);
  const manifest = {
    schemaVersion: 1,
    repository: "https://github.com/aliyun/alibabacloud-sdk",
    revision: "ec489e5c3deae95496daae2b41503ac58b221adb",
    license: "Apache-2.0",
    parserVersion: "2.2.1",
    products: {},
    modules: [],
    files: [
      {
        file: "products/fixture.tea",
        sha256: crypto.createHash("sha256").update(data).digest("hex"),
      },
    ],
  };
  fs.writeFileSync(path.join(dir, "manifest.json"), JSON.stringify(manifest));
  return { dir, manifest };
}

test("detects source tampering before parsing", (t) => {
  const { dir } = fixture(t);
  fs.writeFileSync(path.join(dir, "products/fixture.tea"), "modified");
  assert.throws(() => verifySources(dir), /checksum mismatch/);
});
test("rejects unlisted files that could change import resolution", (t) => {
  const { dir } = fixture(t);
  fs.writeFileSync(path.join(dir, "products/main.dara"), "shadow");
  assert.throws(() => verifySources(dir), /unlisted source/);
});
test("rejects traversal in a source manifest", (t) => {
  const { dir, manifest } = fixture(t);
  manifest.files[0].file = "../outside";
  fs.writeFileSync(path.join(dir, "manifest.json"), JSON.stringify(manifest));
  assert.throws(() => verifySources(dir), /source path/);
});
