"use strict";

const scalarTypes = { string: "string", boolean: "bool", long: "int64", int32: "int32", integer: "int", number: "float64", double: "float64", float: "float32" };

// Native helper names are lookup keys, never generated public model identities.
// Compare wire shape recursively; preserve newer DSL members as discrepancies.
function compareShape(graph, dsl, native, models, path, findings = [], stack = new Set(), coordinates = {}) {
  const difference = (code, detail) => findings.push({ code, path, ...detail, ...coordinates });
  if (!dsl || !native || dsl.kind !== native.kind) {
    difference("XML_KIND", { dslKind: dsl?.kind ?? "missing", nativeKind: native?.kind ?? "missing" });
    return findings;
  }
  if (dsl.kind === "scalar") {
    if (!scalarTypes[dsl.dslType] || scalarTypes[dsl.dslType] !== native.name) difference("XML_SCALAR_TYPE", { dslType: dsl.dslType, nativeType: native.name });
    return findings;
  }
  if (dsl.kind === "array") return compareShape(graph, dsl.items, native.items, models, path + "[]", findings, stack, coordinates);
  if (dsl.kind !== "model") { difference("XML_TYPE_PROFILE", { dslKind: dsl.kind }); return findings; }
  const identity = dsl.ref + "/" + native.name;
  if (stack.size >= 64) { difference("XML_MODEL_DEPTH", {}); return findings; }
  if (stack.has(identity)) { difference("XML_RECURSION", {}); return findings; }
  const body = graph.models.get(dsl.ref), nativeBody = models.get(native.name);
  if (!body || !nativeBody || body.extends || body.inheritedFields?.length) { difference("XML_MODEL_PROFILE", {}); return findings; }
  const seen = new Set(stack); seen.add(identity);
  const nativeFields = new Map();
  for (const field of nativeBody.fields) {
    if (!field.jsonName || nativeFields.has(field.jsonName)) { difference("XML_NATIVE_FIELD_PROFILE", { nativeSource: field.source }); return findings; }
    nativeFields.set(field.jsonName, field);
  }
  for (const field of body.fields) {
    const nativeField = nativeFields.get(field.wireName), fieldPath = path + "." + field.wireName;
    const locations = { dslSource: field.source, nativeSource: nativeField?.source ?? nativeBody.fields[0]?.source ?? coordinates.nativeSource };
    if (!nativeField) findings.push({ code: "XML_NATIVE_FIELD_ABSENT", path: fieldPath, ...locations });
    else if (nativeField.xmlName !== field.wireName || (nativeField.xmlOptions || []).some(option => option !== "omitempty")) findings.push({ code: "XML_NATIVE_WIRE_TAG", path: fieldPath, ...locations });
    else compareShape(graph, field.type, nativeField.type, models, fieldPath, findings, seen, locations);
  }
  return findings;
}

function xmlReviewer(graph, inventory, file) {
  if (inventory.schemaVersion !== 1 || !Array.isArray(inventory.roots) || !Array.isArray(inventory.models)) throw Error("OSS native inventory schema");
  const roots = new Map(), models = new Map();
  for (const root of inventory.roots) { if (roots.has(root.action)) throw Error("Duplicate native XML action"); roots.set(root.action, root); }
  for (const model of inventory.models) { if (models.has(model.name)) throw Error("Duplicate native XML model"); models.set(model.name, model); }
  return (action, returnType) => {
    const fail = (reason, findings = []) => { const error = Error("darabonba: unsupported SDK pattern: OSS XML " + reason); error.findings = findings; throw error; };
    const responseType = graph.type(returnType, action + ".$response", returnType);
    const response = graph.models.get(responseType.ref), bodyType = response?.fields.find(f => f.wireName === "body")?.type;
    const body = graph.models.get(bodyType?.ref), root = roots.get(action);
    if (!body) fail("response body target missing");
    const headers = response.fields.find(f => f.wireName === "headers")?.type, status = response.fields.find(f => f.wireName === "statusCode")?.type;
    if (response.fields.length !== 3 || headers?.kind !== "map" || headers.keys?.dslType !== "string" || headers.values?.dslType !== "string" || status?.kind !== "scalar" || status.dslType !== "int32") fail("response facade profile");
    if (!root || !["structured", "scalar"].includes(root.kind)) fail("root missing or unsupported");
    if (!/^[A-Za-z_][A-Za-z0-9_.-]*$/.test(root.name) || !["local", "exact"].includes(root.namespaceMatch) || root.namespaceMatch === "local" && root.namespace) fail("root namespace profile");
    if (root.field.xmlName !== root.name || (root.field.xmlOptions || []).some(option => option !== "omitempty")) fail("root wire tag profile");
    const wrapped = root.kind === "structured" && body.fields.length === 1 && body.fields[0].wireName === root.field.jsonName;
    if (wrapped) fail("structured wrapper requires a separate binding");
    const findings = root.kind === "scalar" ? compareShape(graph, bodyType, { kind: "model", name: root.model }, models, action) : compareShape(graph, bodyType, root.field.type, models, action);
    if (findings.length) fail("shape differs", findings);
    if (root.kind === "scalar" && (body.fields.length !== 1 || body.fields[0].wireName !== root.field.jsonName || root.field.xmlName !== root.field.jsonName)) fail("scalar root binding");
    return { name: root.name, namespaceMatch: root.namespaceMatch, ...(root.namespace ? { namespace: root.namespace } : {}), normalization: root.kind === "scalar" ? "scalar-field" : "unwrap-structured-root", ...(root.kind === "scalar" ? { scalarField: root.field.jsonName } : {}), dslSource: body.source, nativeSource: root.source, nativeFieldSource: root.field.source };
  };
}

module.exports = { compareShape, xmlReviewer };
