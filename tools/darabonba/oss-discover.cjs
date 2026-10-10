"use strict";

const fs = require("node:fs"), path = require("node:path"), crypto = require("node:crypto");
const parser = require("@darabonba/parser");
const { verifySources } = require("./frontend.cjs");
const { buildProduct } = require("./discovery.cjs");
const repository = path.resolve(__dirname, "../..");
const sha = bytes => crypto.createHash("sha256").update(bytes).digest("hex");

function project(root = repository, selected = []) {
  const sourceRoot = path.join(root, "sources/darabonba"), verified = verifySources(sourceRoot);
  if (verified.manifest.stagedProducts?.oss !== "oss-20190517" || verified.manifest.products.oss) throw Error("OSS must be staged, not production registered");
  const pins = JSON.parse(fs.readFileSync(path.join(root, "metadata/oss-semantic-pins.json")));
  if (pins.schemaVersion !== 1 || pins.inventoryFile !== "metadata/native-helper-pins/oss-inventory.json" || pins.pinsFile !== "metadata/native-helper-pins/oss.json" || pins.gatewayFile !== "modules/alibabacloud_GatewayOSS_0.0.42/main.tea") throw Error("Unsupported OSS semantic pins");
  const inventoryBytes = fs.readFileSync(path.join(root, pins.inventoryFile)), inventory = JSON.parse(inventoryBytes);
  const nativePins = JSON.parse(fs.readFileSync(path.join(root, pins.pinsFile)));
  if (sha(inventoryBytes) !== pins.inventorySHA256 || nativePins.registry.sha256 !== inventory.registrySHA256 || nativePins.models.sha256 !== inventory.modelsSHA256) throw Error("OSS native inventory checksum differs");
  const file = "products/oss/main.tea", main = path.join(sourceRoot, file), bytes = fs.readFileSync(main);
  if (sha(bytes) !== pins.sourceSHA256 || sha(fs.readFileSync(path.join(sourceRoot, pins.gatewayFile))) !== pins.gatewaySHA256) throw Error("OSS semantic source binding differs");
  const libraries = JSON.parse(fs.readFileSync(path.join(sourceRoot, "products/oss/Teafile"))).libraries;
  const gateway = verified.manifest.modules.find(m => m.directory + "/main.tea" === pins.gatewayFile);
  if (!gateway || libraries.GatewayClient !== gateway.spec) throw Error("OSS Gateway import binding differs");
  const provenance = { repository: verified.manifest.repository, revision: verified.manifest.revision, license: verified.manifest.license, parserVersion: verified.manifest.parserVersion, sourceManifestSHA256: verified.hash, sourceSHA256: sha(bytes), nativeInventorySHA256: pins.inventorySHA256, nativeRegistrySHA256: inventory.registrySHA256, nativeModelsSHA256: inventory.modelsSHA256, gatewaySHA256: pins.gatewaySHA256 };
  const ast = parser.parse(bytes.toString("utf8"), main);
  const result = buildProduct(ast, { pkg: "oss", identifier: "oss-20190517", file, provenance, nativeXML: inventory });
  for (const name of selected) {
    if (!/^[A-Za-z][A-Za-z0-9]*$/.test(name) || result.ir.operations.find(o => o.name === name)?.status !== "lowered") throw Error("Selected OSS operation is unknown/unsupported: " + name);
  }
  return result;
}

const targets = [["docs/research/oss-semantic-ir.json", "ir"], ["docs/research/oss-semantic-coverage.json", "coverage"]];
function run(mode, { root = repository, selected = [] } = {}) {
  if (!["generate", "check", "report"].includes(mode)) throw Error("Use generate, check or report");
  const result = project(root, selected);
  if (mode === "report") return JSON.stringify(result.coverage, null, 2) + "\n";
  const files = targets.map(([relative, key]) => ({ relative, bytes: Buffer.from(JSON.stringify(result[key], null, 2) + "\n") }));
  // Selection and all output-path preflight finish before either file is written.
  for (const { relative, bytes } of files) {
    let current = root;
    for (const part of ["", ...relative.split("/")]) {
      current = path.join(current, part);
      let entry;
      try { entry = fs.lstatSync(current); } catch (error) { if (error.code !== "ENOENT") throw error; }
      if (entry?.isSymbolicLink()) throw Error("OSS semantic output symlink");
    }
    const target = path.join(root, relative);
    if (fs.existsSync(target) && (!fs.statSync(target).isFile() || JSON.parse(fs.readFileSync(target)).product !== "oss")) throw Error("OSS semantic output is not owned");
    if (mode === "check" && (!fs.existsSync(target) || !fs.readFileSync(target).equals(bytes))) throw Error("OSS semantic artifact drift: " + relative);
  }
  if (mode === "generate") for (const { relative, bytes } of files) { fs.mkdirSync(path.dirname(path.join(root, relative)), { recursive: true }); fs.writeFileSync(path.join(root, relative), bytes); }
  return result;
}

if (require.main === module) {
  try { const output = run(process.argv[2], { selected: process.argv.slice(3) }); if (typeof output === "string") process.stdout.write(output); }
  catch (error) { console.error(error.message); process.exitCode = 1; }
}
module.exports = { project, run };
