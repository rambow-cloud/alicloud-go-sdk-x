"use strict";

// Test-only fixture: rename in memory; never change pinned source artifacts.
const fs = require("node:fs");
const path = require("node:path");
const parser = require("@darabonba/parser");
const { buildProduct } = require("./discovery.cjs");

function reuseFixture(root) {
  const main = path.join(root, "sources/darabonba/products/sts/main.tea");
  let source = fs.readFileSync(main, "utf8");
  for (const [before, after] of [
    ["AssumeRoleWithOIDC", "ExchangeIdentity"],
    ["AssumeRoleWithSAML", "ExchangeAssertion"],
    ["GetCallerIdentity", "InspectIdentity"],
    ["AssumeRole", "ObtainRole"],
  ]) {
    source = source
      .replaceAll(before, after)
      .replaceAll(
        before[0].toLowerCase() + before.slice(1),
        after[0].toLowerCase() + after.slice(1),
      );
  }
  return buildProduct(parser.parse(source, main), {
    pkg: "authfixture",
    identifier: "authfixture-20150401",
    file: "products/authfixture/main.tea",
    provenance: {
      repository: "synthetic-fixture",
      revision: "in-memory-test",
      license: "Apache-2.0",
      parserVersion: "2.2.1",
      sourceManifestSHA256: "synthetic-not-for-publication",
    },
  });
}

module.exports = { reuseFixture };
if (require.main === module) {
  process.stdout.write(JSON.stringify(reuseFixture(process.argv[2]).ir));
}
