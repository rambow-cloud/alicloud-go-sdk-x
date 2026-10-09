"use strict";

const fs = require("node:fs"), path = require("node:path"), crypto = require("node:crypto");
const { verifyCanonical, requestShape, responseShape, indexedLeaf, compatibleShape } = require("./normalization.cjs");
const repository = path.resolve(__dirname, "../..");
const sha = b => crypto.createHash("sha256").update(b).digest("hex");
const escapePointer = s => s.replace(/~/g, "~0").replace(/\//g, "~1");
const english = s => typeof s === "string" && /[A-Za-z]/.test(s) && !/[\u3400-\u9fff]/.test(s);
function annotate(value, pointer = "") {
  if (!value || typeof value !== "object") return;
  for (const [key, child] of Object.entries(value)) annotate(child, pointer + "/" + escapePointer(key));
  if (!Array.isArray(value) && (typeof value.type === "string" || typeof value.description_en === "string" || typeof value.help_en === "string")) value.__prosePointer = pointer;
}
function nativeShape(type, models, depth = 0) {
  if (!type || depth > 32) return { type: "unsupported" };
  if (type.kind === "scalar") return { type: type.wireType };
  if (type.kind === "array") return { type: "array", items: nativeShape(type.items, models, depth + 1) };
  if (type.kind !== "model") return { type: "unsupported" };
  const model = models.get(type.ref); if (!model) return { type: "unsupported" };
  const properties = {};
  for (const f of [...(model.inheritedFields || []), ...model.fields]) {
    properties[f.wireName] = { ...nativeShape(f.type, models, depth + 1), nativeKey: model.id + "#" + f.dslName, nativeSource: f.source, upstreamEnglish: (f.documentation || []).some(d => d.attribute !== "example" && /[A-Za-z]/.test(d.text)) };
  }
  return { type: "object", properties };
}
function matchProse(metadata, native, context, output, reasons, location = "") {
  if (!metadata || !native || metadata.type !== native.type || native.type === "unsupported") { reasons.push({ ...context, path: location, reason: "wire-type-or-path-differs" }); return; }
  const a = metadata.attributes || {}, text = a.description_en || a.help_en;
  if (native.nativeKey && !native.upstreamEnglish && english(text)) {
    try {
      compatibleShape(metadata, native);
      const attribute = a.description_en ? "description_en" : "help_en";
      const entry = { ...context, fieldSource: native.nativeSource, pointer: a.__prosePointer + "/" + attribute, english: text, textSHA256: sha(text) };
      const existing = output.get(native.nativeKey);
      if (!existing) output.set(native.nativeKey, entry);
      else if (existing.english !== text) { existing.conflict = true; reasons.push({ ...context, path: location, reason: "conflicting-field-prose", field: native.nativeKey }); }
    } catch { reasons.push({ ...context, path: location, reason: "container-shape-differs" }); }
  }
  if (native.type === "array") matchProse(metadata.items, native.items, context, output, reasons, location + "[]");
  if (native.type === "object") for (const key of Object.keys(metadata.properties || {}).sort()) matchProse(metadata.properties[key], native.properties?.[key], context, output, reasons, location ? location + "." + key : key);
}
function project(root = repository) {
  const corpus = path.join(root, "sources/openapi-meta/prose");
  if (!fs.existsSync(corpus)) return {};
  const verified = verifyCanonical(corpus), pins = new Map(verified.manifest.files.map(f => [f.file, f]));
  const files = {};
  for (const product of fs.readdirSync(path.join(root, "models")).sort()) {
    const irpath = path.join(root, "models", product, "ir.json"); if (!fs.existsSync(irpath)) continue;
    const irbytes = fs.readFileSync(irpath), ir = JSON.parse(irbytes), models = new Map(ir.models.map(m => [m.id, m])), output = new Map(), reasons = [];
    let matchedOperations = 0;
    for (const op of ir.operations) {
      const file = "canonical/" + product + "/" + ir.version + "/" + op.name + ".json", pin = pins.get(file);
      if (!pin) { reasons.push({ operation: op.name, reason: "metadata-absent-at-pinned-revision" }); continue; }
      const doc = JSON.parse(fs.readFileSync(path.join(corpus, file))); annotate(doc);
      const context = { operation: op.name, file, sourceSHA256: pin.sha256 };
      if (doc.operation?.action !== op.name || doc.operation?.api_version !== ir.version) throw Error("prose: operation/version differs");
      const request = nativeShape(op.roots.request, models), response = nativeShape(op.roots.body, models);
      const pending = new Map();
      try {
        for (const p of doc.parameters || []) {
          if (p.location !== "query") throw Error("unsupported metadata location");
          let leaf; try { leaf = indexedLeaf(request, p.raw_name); } catch { reasons.push({ ...context, path: p.raw_name, reason: "request-path-differs" }); continue; }
          matchProse(requestShape(p), leaf, context, pending, reasons, p.raw_name);
        }
        matchProse(responseShape(doc.responses?.["200"]?.schema), response, context, pending, reasons);
      } catch { reasons.push({ ...context, reason: "unsupported-metadata-representation" }); continue; }
      for (const [key, entry] of pending) {
        const previous = output.get(key); if (previous && previous.english !== entry.english) { previous.conflict = true; reasons.push({ ...context, field: key, reason: "conflicting-field-prose" }); } else if (!previous) output.set(key, entry);
      }
      matchedOperations++;
    }
    const fields = Object.fromEntries([...output].filter(([,d]) => !d.conflict).sort(([a],[b]) => a.localeCompare(b)));
    files["metadata/prose/" + product + ".json"] = Buffer.from(JSON.stringify({ generator: "darabonba prose", schemaVersion: 1, product, version: ir.version, irSHA256: sha(irbytes), corpusManifestSHA256: verified.hash, repository: verified.manifest.repository, revision: verified.manifest.revision, license: verified.manifest.license, matchedOperations, fields, reasons }, null, 2) + "\n");
  }
  return files;
}
function run(root, check) {
  const files = project(root);
  for (const [file, bytes] of Object.entries(files)) {
    const target = path.join(root, file);
    if (check) { if (!fs.existsSync(target) || !fs.readFileSync(target).equals(bytes)) throw Error("prose: projection differs: " + file); }
    else { fs.mkdirSync(path.dirname(target), { recursive: true }); fs.writeFileSync(target, bytes); }
  }
  return files;
}
if (require.main === module) {
  try { const command = process.argv[2]; if (!["generate", "check"].includes(command)) throw Error("prose: use generate|check"); const files=run(repository,command==="check"); console.log("Optional canonical prose " + command + " passed for " + Object.keys(files).length + " products."); } catch(e) { console.error(e.message); process.exit(1); }
}
module.exports = { annotate, nativeShape, matchProse, project, run };
