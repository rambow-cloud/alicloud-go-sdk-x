"use strict";
const { test } = require("node:test");
const assert = require("node:assert/strict");
const fs = require("node:fs");
const path = require("node:path");
const parser = require("@darabonba/parser");
const { annotationDocuments, modelGraph, project } = require("./discovery.cjs");
const root = path.resolve(__dirname, "../..");
const file = "products/sts/main.tea";
const main = path.join(root, "sources/darabonba", file);
const ast = parser.parse(fs.readFileSync(main, "utf8"), main);

test("official annotations retain summary/description coordinates and exclude parameter/example tags", () => {
  const fn = ast.moduleBody.nodes.find(
    (n) => n.functionName?.lexeme === "assumeRoleWithOptions",
  );
  const docs = annotationDocuments(fn, file);
  assert.deepEqual(
    docs.map((d) => d.attribute),
    ["summary", "description"],
  );
  assert.match(docs[0].text, /Security Token Service/);
  assert.match(docs[1].text, /Prerequisites/);
  assert.equal(docs[0].source.line, fn.annotation.loc.start.line);
  assert.equal(docs[0].source.file, file);
  assert.ok(!docs.some((d) => /^@(param|return|example)/m.test(d.text)));
  assert.deepEqual(annotationDocuments({}, file), []);
  const empty = structuredClone(fn);
  empty.annotation.value =
    "/**\n * @summary\n * @example real-account-value\n */";
  assert.equal(annotationDocuments(empty, file)[0].text, "");
  assert.equal(annotationDocuments(empty, file).length, 1);
});

test("field descriptions use parser values and source example values stay absent from IR", () => {
  const graph = modelGraph(ast, file);
  const fn = ast.moduleBody.nodes.find(
    (n) => n.functionName?.lexeme === "assumeRoleWithOptions",
  );
  graph.type(
    fn.params.params[0].paramType,
    "AssumeRole.$request",
    fn.params.params[0].paramName,
  );
  const model = graph.models.get("AssumeRoleRequest");
  const field = model.fields.find((f) => f.dslName === "durationSeconds");
  const docs = field.documentation;
  assert.match(
    docs.find((d) => d.attribute === "description").text,
    /Minimum value: 900/,
  );
  const example = docs.find((d) => d.attribute === "example");
  assert.equal(example.text, undefined);
  assert.ok(example.source.line > 0);
});

test("documentation projection preserves complete coverage and changes no wire protocol", () => {
  const projected = project(root);
  for (const [pkg, product] of Object.entries(projected.products)) {
    const expected = { ecs: 380, vpc: 403, sts: 4, fc:72 }[pkg];
    assert.equal(product.coverage.counts.lowered, expected);
    assert.ok(
      product.ir.operations.some((o) =>
        o.documentation.some((d) => d.attribute === "summary"),
      ),
    );
    for (const model of product.ir.models)
      for (const field of model.fields)
        for (const doc of field.documentation || []) {
          assert.ok(doc.source.file.endsWith("main.tea"));
          if (doc.attribute === "example") assert.equal(doc.text, undefined);
        }
  }
});
