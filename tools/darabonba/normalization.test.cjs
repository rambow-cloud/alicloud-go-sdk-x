"use strict";

const { test } = require("node:test");
const assert = require("node:assert/strict");
const fs = require("node:fs");
const path = require("node:path");
const os = require("node:os");
const parser = require("@darabonba/parser");
const {
  normalizeCanonical,
  verifyCanonical,
  bindInputs,
  indexedLeaf,
  compatibleShape,
} = require("./normalization.cjs");
const { lowerOperation, reviewInputs } = require("./frontend.cjs");
const root = path.resolve(__dirname, "../..");
const canonicalRoot = path.join(root, "sources/openapi-meta");
const read = (name) =>
  JSON.parse(
    fs.readFileSync(
      path.join(canonicalRoot, `canonical/ecs/2014-05-26/${name}.json`),
    ),
  );
const normalized = (name) =>
  normalizeCanonical(read(name), { license: "Apache-2.0" });
const main = path.join(root, "sources/darabonba/products/ecs/main.tea");
const ast = parser.parse(fs.readFileSync(main, "utf8"), main);
const decisions = JSON.parse(
  fs.readFileSync(path.join(root, "metadata/darabonba-decisions.json")),
);
const snapshot = (name) =>
  JSON.parse(fs.readFileSync(path.join(root, `metadata/ecs/${name}.json`)));

test("canonical adapter preserves wire case, repeated models, CLI attributes and bilingual evidence", () => {
  const product = normalized("DescribeImages");
  const region = product.inputs.find((i) => i.wire === "RegionId");
  assert.equal(region.cli.name, "region_id");
  assert.deepEqual(region.cli.options, ["--biz-region-id"]);
  assert.equal(region.schema.required, true);
  assert.match(region.schema.attributes.help_en, /region/);
  assert.match(region.schema.attributes.help_zh, /地域/);
  const tag = product.inputs.find((i) => i.wire === "Tag");
  assert.equal(tag.style, "repeatList");
  assert.deepEqual(Object.keys(tag.schema.items.properties).sort(), [
    "Key",
    "Value",
    "key",
    "value",
  ]);
  assert.match(
    tag.schema.items.properties.key.attributes.help_en,
    /\[Deprecated\]/,
  );
  assert.equal(product.protocol.method, "POST");
  assert.equal(product.attributes.method, "GET|POST");
  assert.equal(product.protocol.protocol, "HTTPS");
  assert.equal(product.operationAttributes.operation_type, "read");
});

test("itemName restores real DSL response wrappers without applying backend transformations", () => {
  const response = normalized("DescribeImages").response;
  const dsl = lowerOperation(ast, "DescribeImages").response;
  const paths = [
    "Images.Image",
    "Images.Image.1.ImageId",
    "Images.Image.1.DiskDeviceMappings.DiskDeviceMapping",
    "Images.Image.1.Tags.Tag",
    "Images.Image.1.DetectionOptions.Items.Item",
  ];
  for (const location of paths) {
    const meta = indexedLeaf(response, location);
    const wire = indexedLeaf(dsl, location);
    assert.equal(meta.type, wire.type, location);
  }
  assert.equal(response.properties.Images.type, "object");
  assert.equal(response.properties.Images.attributes.backendName, "data.data");
  assert.equal(
    response.properties.Images.properties.Image.items.properties.Architecture
      .attributes.nullToEmpty,
    true,
  );
  assert.equal(response.properties.Images.properties["data.data"], undefined);
  assert.equal(
    response.properties.Images.properties.Image.items.properties.Architecture
      .properties,
    undefined,
  );
});

test("indexed Filter bindings match the DSL model before approved differences are computed", () => {
  const op = lowerOperation(ast, "DescribeInstances");
  reviewInputs("ecs", op, snapshot(op.name), decisions);
  const filter = op.inputs.find((i) => i.wire === "Filter");
  assert.equal(filter.metadataBindings.length, 8);
  assert.equal(filter.schema.type, "array");
  assert.equal(filter.schema.maxItems, undefined);
  const canonical = normalized(op.name);
  assert.deepEqual(
    bindInputs(
      op.inputs,
      canonical.inputs.map((i) => ({
        name: i.wire,
        in: i.location,
        schema: i.schema,
      })),
    ).get("Filter").aliases,
    filter.metadataBindings,
  );
  assert.equal(
    decisions.operations["ecs/DescribeInstances"].dslOnlyInputs.includes(
      "Filter",
    ),
    false,
  );
});

test("explicit metadata-only lowercase tag exceptions do not approve other case drift", () => {
  const op = lowerOperation(ast, "DescribeInstances"),
    data = snapshot(op.name);
  const tag = data.parameters.find((p) => p.name === "Tag").schema.items;
  tag.properties.KEY = tag.properties.key;
  delete tag.properties.key;
  assert.throws(
    () => reviewInputs("ecs", op, data, decisions),
    /unreviewed metadata\/DSL difference/,
  );
  assert.throws(
    () =>
      compatibleShape(
        { type: "object", properties: { key: { type: "string" } } },
        { type: "object", properties: { Key: { type: "string" } } },
      ),
    /wire path\/case differs/,
  );
});

test("metadata-only scalar approvals retain types and cannot become required", () => {
  for (const change of [
    (s) => {
      s.type = "integer";
    },
    (s) => {
      s.required = true;
    },
  ]) {
    const data = snapshot("DescribeInstances");
    change(
      data.parameters.find((p) => p.name === "Tag").schema.items.properties.key,
    );
    assert.throws(
      () =>
        reviewInputs(
          "ecs",
          lowerOperation(ast, "DescribeInstances"),
          data,
          decisions,
        ),
      /unreviewed metadata\/DSL difference|wire path\/case differs/,
    );
  }
});

for (const [description, mutate, error] of [
  [
    "indexed wire type",
    (p) => {
      p.schema.type = "integer";
    },
    /wire type differs/,
  ],
  [
    "indexed wire case",
    (p) => {
      p.name = "Filter.1.key";
    },
    /wire path\/case differs/,
  ],
  [
    "zero index",
    (p) => {
      p.name = "Filter.0.Key";
    },
    /invalid array index/,
  ],
  [
    "leading zero index",
    (p) => {
      p.name = "Filter.01.Key";
    },
    /invalid array index/,
  ],
  [
    "indexed requiredness",
    (p) => {
      p.schema.required = true;
    },
    /indexed requiredness differs/,
  ],
])
  test(
    "rejects " + description + " before approving representation differences",
    () => {
      const data = snapshot("DescribeInstances");
      mutate(data.parameters.find((p) => p.name === "Filter.1.Key"));
      assert.throws(
        () =>
          reviewInputs(
            "ecs",
            lowerOperation(ast, "DescribeInstances"),
            data,
            decisions,
          ),
        error,
      );
    },
  );

test("direct and flattened bindings cannot compete; declared indexes do not impose a maximum", () => {
  const data = snapshot("DescribeInstances"),
    op = lowerOperation(ast, "DescribeInstances");
  data.parameters.push({
    in: "query",
    name: "Filter",
    schema: op.inputs.find((i) => i.wire === "Filter").schema,
  });
  assert.throws(
    () => bindInputs(op.inputs, data.parameters),
    /ambiguous direct\/indexed binding/,
  );
  data.parameters.pop();
  data.parameters.push({
    in: "query",
    name: "Filter.100.Key",
    schema: { type: "string", required: false },
  });
  const bindings = bindInputs(op.inputs, data.parameters);
  assert.equal(bindings.get("Filter").aliases.length, 9);
});

test("invalid canonical styles, names and response itemName fail instead of becoming guessed codecs", () => {
  for (const mutate of [
    (d) => {
      d.parameters[0].param_style = "unknown";
    },
    (d) => {
      d.parameters[1].raw_name = d.parameters[0].raw_name;
    },
    (d) => {
      d.responses["200"].schema.properties.Images.itemName = "";
    },
    (d) => {
      d.responses["200"].schema.properties.RequestId.itemName = "Item";
    },
  ]) {
    const data = read("DescribeImages");
    mutate(data);
    assert.throws(() => normalizeCanonical(data, {}), /normalization:/);
  }
});

function fixture(t) {
  const dir = fs.mkdtempSync(path.join(os.tmpdir(), "canonical-source-"));
  // Only delete the exact temporary directory created above.
  t.after(() => fs.rmSync(dir, { recursive: true, force: true }));
  fs.cpSync(canonicalRoot, dir, { recursive: true });
  return dir;
}
test("canonical source lock detects modified or unlisted fixtures", (t) => {
  assert.equal(verifyCanonical().manifest.adapterVersion, 1);
  const dir = fixture(t);
  fs.appendFileSync(
    path.join(dir, "canonical/ecs/2014-05-26/DescribeImages.json"),
    " ",
  );
  assert.throws(() => verifyCanonical(dir), /checksum mismatch/);
});
test("canonical source inventory rejects unlisted metadata", (t) => {
  const dir = fixture(t);
  fs.writeFileSync(path.join(dir, "canonical/extra.json"), "{}");
  assert.throws(() => verifyCanonical(dir), /unlisted source/);
});
test("canonical source lock rejects traversal before reading external files", (t) => {
  const dir = fixture(t),
    file = path.join(dir, "manifest.json");
  const manifest = JSON.parse(fs.readFileSync(file));
  manifest.files[0].file = "../outside";
  fs.writeFileSync(file, JSON.stringify(manifest));
  assert.throws(() => verifyCanonical(dir), /source path/);
});
