"use strict";

// Explicit reviewed corpus extension; generation never resolves imports online.
const fs = require("node:fs");
const path = require("node:path");
const crypto = require("node:crypto");
const tar = require("tar");
const { verifySources } = require("./frontend.cjs");
const sourceRoot = path.resolve(__dirname, "../../sources/darabonba");
const hash = (bytes, algorithm = "sha256") =>
  crypto.createHash(algorithm).update(bytes).digest("hex");
const fail = (condition, message) => {
  if (!condition) throw Error("import-modules: " + message);
};
const order = (a, b) => (a < b ? -1 : a > b ? 1 : 0);

function validatePin(pin, manifest) {
  fail(pin && typeof pin.scope === "string" && typeof pin.name === "string" && typeof pin.version === "string" && /^[a-z][a-z0-9]*$/.test(pin.scope) && /^[A-Za-z][A-Za-z0-9_]*$/.test(pin.name) && /^\d+\.\d+\.\d+$/.test(pin.version), "invalid module identity");
  fail(pin.spec === `${pin.scope}:${pin.name}:*` && pin.directory === `modules/${pin.scope}_${pin.name}_${pin.version}`, "module spec/directory mismatch");
  fail(!manifest.modules.some(m => m.spec === pin.spec || m.directory === pin.directory), "cannot replace a pinned module");
  fail(/^[a-f0-9]{40}$/.test(pin.archiveSHA1) && /^[a-f0-9]{64}$/.test(pin.archiveSHA256), "archive checksums are required");
  fail(pin.url === `https://darabonba-module-prod.oss-cn-zhangjiakou.aliyuncs.com/${pin.scope}/${pin.name}-${pin.version}.tar.gz`, "unreviewed archive origin");
  const license = pin.license;
  fail(license && ["Apache-2.0", "MIT"].includes(license.spdx), "unreviewed license");
  if (license.evidence === "archive-declaration") {
    fail(license.file === pin.directory + "/README.md", "archive license declaration path");
  } else {
    fail(license.evidence === "module-source-repository" && manifest.modules.some(m => m.license?.file === license.file && m.license?.repository === license.repository && m.license?.spdx === license.spdx && m.license?.gitBlobSHA === license.gitBlobSHA), "source license evidence is not already pinned");
  }
}

async function archiveFiles(bytes) {
  fail(Buffer.isBuffer(bytes) && bytes.length <= (8 << 20), "archive exceeds limit");
  const files = new Map(), names = new Set(), fileNames = new Set();
  let total = 0, entries = 0;
  return new Promise((resolve, reject) => {
    const reader = new tar.Parser({ strict: true, onReadEntry(entry) {
      try {
        let name = entry.path;
        // node-tar normalizes backslashes on Windows; inspect the raw header too.
        fail(!entry.header.path.includes("\\"), "unsafe archive path");
        if (entry.type === "Directory" && [".", "./"].includes(name)) {
          fail(++entries <= 128, "too many archive entries"); entry.resume(); return;
        }
        if (name.startsWith("./")) name = name.slice(2);
        if (entry.type === "Directory") name = name.replace(/\/$/, "");
        fail(++entries <= 128 && !path.isAbsolute(name) && !/[\x00-\x1f\\:*?"<>|]/.test(name) && name.split("/").every(p => p && p !== "." && p !== ".." && !/[. ]$/.test(p) && !/^(?:con|prn|aux|nul|com[1-9]|lpt[1-9])(?:\.|$)/i.test(p)), "unsafe archive path");
        fail(["File", "Directory", "OldFile"].includes(entry.type), "unsupported archive entry");
        const folded = name.toLowerCase(), parts = folded.split("/");
        fail(!names.has(folded), "duplicate archive path");
        fail(parts.slice(0,-1).every((_, i) => !fileNames.has(parts.slice(0,i+1).join("/"))), "archive file/directory conflict");
        if (entry.type !== "Directory") {
          fail([...names].every(existing => !existing.startsWith(folded + "/")), "archive file/directory conflict");
          fileNames.add(folded);
        }
        names.add(folded);
        if (entry.type === "Directory") { entry.resume(); return; }
        fail(entry.size <= (8 << 20) && (total += entry.size) <= (32 << 20), "expanded archive exceeds limit");
        const chunks = [];
        entry.on("data", chunk => chunks.push(chunk));
        entry.on("end", () => files.set(name, Buffer.concat(chunks)));
        entry.on("error", reject);
      } catch (error) { reader.abort(error); }
    }});
    reader.on("error", reject);
    reader.on("ignoredEntry", () => reader.abort(Error("import-modules: unsupported archive entry")));
    reader.on("end", () => resolve(files));
    reader.end(bytes);
  });
}

async function prepareModules(root, plan, loadArchive) {
  plan = structuredClone(plan);
  const verified = verifySources(root), manifest = structuredClone(verified.manifest);
  fail(plan?.schemaVersion === 1 && Array.isArray(plan.modules) && plan.modules.length > 0, "invalid extension plan");
  fail(plan.sourceManifestSHA256 === verified.hash, "extension plan source binding differs");
  const seen = new Set();
  for (const pin of plan.modules) {
    validatePin(pin, manifest);
    fail(!seen.has(pin.spec), "duplicate planned module"); seen.add(pin.spec);
    fail(!fs.existsSync(path.join(root, pin.directory)), "module directory already exists");
  }
  const results = await Promise.allSettled(plan.modules.map(pin => loadArchive(structuredClone(pin))));
  const failure = results.find(r => r.status === "rejected");
  if (failure) throw failure.reason;
  const files = new Map(), metadata = [];
  for (let i = 0; i < plan.modules.length; i++) {
    const pin = plan.modules[i], bytes = results[i].value;
    fail(Buffer.isBuffer(bytes) && hash(bytes, "sha1") === pin.archiveSHA1 && hash(bytes) === pin.archiveSHA256, "archive checksum mismatch");
    const contents = await archiveFiles(bytes);
    fail(contents.has("Teafile") !== contents.has("Darafile"), "module descriptor must be unique");
    const meta = JSON.parse(contents.get(contents.has("Darafile") ? "Darafile" : "Teafile"));
    fail(meta.scope === pin.scope && meta.name === pin.name && meta.version === pin.version && typeof meta.main === "string" && contents.has(meta.main.replace(/^\.\//, "")), "archive module identity/entrypoint mismatch");
    fail(meta.libraries === undefined || (meta.libraries && typeof meta.libraries === "object" && !Array.isArray(meta.libraries)), "invalid module imports");
    metadata.push(meta);
    if (pin.license.evidence === "archive-declaration")
      fail(contents.has("README.md") && contents.get("README.md").toString("utf8").includes(pin.license.spdx), "missing archive license declaration");
    for (const [name, data] of contents) files.set(pin.directory + "/" + name, data);
    manifest.modules.push(structuredClone(pin));
  }
  const pinned = new Set(manifest.modules.map(m => m.spec));
  // Check all current and new module imports, not just the new direct edges.
  for (const module of verified.manifest.modules) {
    const directory = path.join(root, module.directory);
    metadata.push(JSON.parse(fs.readFileSync(path.join(directory, fs.existsSync(path.join(directory, "Darafile")) ? "Darafile" : "Teafile"))));
  }
  for (const meta of metadata)
    for (const spec of Object.values(meta.libraries || {}))
      fail(typeof spec === "string" && pinned.has(spec), "unresolved module import: " + spec);
  for (const pin of plan.modules)
    for (const [file, data] of files)
      if (file.startsWith(pin.directory + "/"))
        manifest.files.push({ file, sha256: hash(data), url: pin.url + "#" + file.slice(pin.directory.length + 1) });
  manifest.modules.sort((a,b) => order(a.spec,b.spec));
  const libraries = Buffer.from(JSON.stringify(Object.fromEntries(manifest.modules.map(m => [m.spec, "../../" + m.directory])), null, 2) + "\n");
  for (const product of Object.keys(manifest.products)) {
    const file = `products/${product}/.libraries.json`;
    files.set(file, libraries);
    const record = manifest.files.find(f => f.file === file);
    fail(record, "missing product import projection");
    record.sha256 = hash(libraries); record.url = "local:pinned-import-projection";
  }
  manifest.files.sort((a,b) => order(a.file,b.file));
  return { manifest, files, expectedSourceHash: verified.hash };
}

function writeModules(root, prepared) {
  fail(verifySources(root).hash === prepared.expectedSourceHash, "source corpus changed after preflight");
  for (const [file, bytes] of [...prepared.files].sort(([a],[b]) => order(a,b))) {
    fs.mkdirSync(path.dirname(path.join(root,file)), { recursive: true });
    fs.writeFileSync(path.join(root,file), bytes);
  }
  fs.writeFileSync(path.join(root,"manifest.json"), JSON.stringify(prepared.manifest,null,2) + "\n");
}

async function main() {
  fail(process.argv.length === 3 || process.argv.length === 4, "use <reviewed-plan.json> [archive-directory]");
  const plan = JSON.parse(fs.readFileSync(process.argv[2]));
  const load = async pin => {
    if (process.argv[3]) return fs.readFileSync(path.join(process.argv[3],`${pin.scope}_${pin.name}_${pin.version}.tgz`));
    const response = await fetch(pin.url, { signal: AbortSignal.timeout(40000), redirect: "error" });
    fail(response.ok && response.url === pin.url, "archive download failed or redirected");
    fail(response.body, "empty archive download");
    const chunks = []; let size = 0;
    for await (const chunk of response.body) {
      fail((size += chunk.length) <= (8 << 20), "archive exceeds limit");
      chunks.push(chunk);
    }
    return Buffer.concat(chunks);
  };
  const prepared = await prepareModules(sourceRoot,plan,load);
  writeModules(sourceRoot,prepared);
  console.log(`Pinned ${plan.modules.length} additional modules; existing versions are unchanged. Review source hash bindings and regenerate before emission.`);
}
if (require.main === module) main().catch(error => { console.error(error.message); process.exitCode = 1; });
module.exports = { archiveFiles, prepareModules, writeModules };
