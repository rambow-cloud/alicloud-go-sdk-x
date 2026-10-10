"use strict";

// Offline source-only maintenance rehearsal. Production sources are read-only.
const fs = require("node:fs");
const path = require("node:path");
const crypto = require("node:crypto");
const { project, run } = require("./discovery.cjs");
const parser = require("@darabonba/parser");
const repository = path.resolve(__dirname, "../..");
const sha = (b) => crypto.createHash("sha256").update(b).digest("hex");

function revisions(root = repository) {
  const base = path.join(root, "tools/darabonba/fixtures/sts-revisions");
  const manifest = JSON.parse(
    fs.readFileSync(path.join(base, "manifest.json")),
  );
  if (
    manifest.schemaVersion !== 1 ||
    manifest.repository !== "https://github.com/aliyun/alibabacloud-sdk" ||
    manifest.license !== "Apache-2.0" ||
    ![manifest.baselineRevision, manifest.candidateRevision].every((r) =>
      /^[a-f0-9]{40}$/.test(r),
    )
  )
    throw Error("invalid rehearsal provenance");
  const files = new Map();
  for (const record of manifest.files) {
    if (
      ![manifest.baselineRevision, manifest.candidateRevision].includes(
        record.revision,
      ) ||
      !["main.tea", "Teafile", "api-info.json", "LICENSE"].some(
        (n) => record.file === record.revision + "/" + n,
      ) ||
      files.has(record.file)
    )
      throw Error("unsafe/duplicate revision fixture");
    const bytes = fs.readFileSync(path.join(base, record.file));
    if (sha(bytes) !== record.sha256)
      throw Error("revision fixture checksum mismatch");
    files.set(record.file, bytes);
  }
  if (files.size !== 8) throw Error("incomplete revision fixture");
  return { manifest, files };
}

function withoutCoordinates(value) {
  if (Array.isArray(value)) return value.map(withoutCoordinates);
  if (value && typeof value === "object")
    return Object.fromEntries(
      Object.entries(value)
        .filter(
          ([k]) =>
            ![
              "source",
              "loc",
              "tokenRange",
              "inferred",
              "expectedType",
              "index",
              "needCast",
              "needValidate",
            ].includes(k),
        )
        .map(([k, v]) => [k, withoutCoordinates(v)]),
    );
  return value;
}

function sourceFacts(ast) {
  const init = ast.moduleBody.nodes.find((n) => n.type === "init");
  const assignments = init.initBody.stmts.filter(
    (s) => s.type === "assign" && s.left.type === "virtualVariable",
  );
  const signature = assignments.find(
    (s) => s.left.vid.lexeme === "@signatureAlgorithm",
  );
  const endpoints = assignments.find(
    (s) => s.left.vid.lexeme === "@endpointMap",
  );
  const handoffs = ast.moduleBody.nodes
    .filter((n) => n.functionName?.lexeme?.endsWith("WithOptions"))
    .map((n) => {
      const params = n.functionBody.stmts.stmts.find(
        (s) => s.id?.lexeme === "params",
      ).expr.object.fields;
      return {
        name: n.functionName.lexeme,
        authType: params.find((f) => f.fieldName.lexeme === "authType").expr
          .value.string,
        handoff: n.functionBody.stmts.stmts.at(-1).expr.left.id.lexeme,
      };
    })
    .sort((a, b) => a.name.localeCompare(b.name));
  return {
    signatureAlgorithm: signature?.expr.value.string || "inherited-default",
    endpointMap: withoutCoordinates(endpoints.expr),
    handoffs,
  };
}

function fingerprints(ir) {
  const documentation = {
    operations: ir.operations.map((o) => ({
      name: o.name,
      documentation: o.documentation,
    })),
    models: ir.models.map((m) => ({
      id: m.id,
      documentation: m.documentation,
      fields: m.fields.map((f) => ({
        name: f.wireName,
        documentation: f.documentation,
      })),
    })),
  };
  return {
    operations: ir.operations.map((o) => ({
      name: o.name,
      status: o.status,
      protocol: o.protocol,
      handoff: o.handoff || "",
      bindings: o.bindings,
    })),
    fields: ir.models
      .flatMap((m) =>
        m.fields.map((f) => ({
          key: m.id + "." + f.wireName,
          type: f.type,
          required: f.required,
        })),
      )
      .sort((a, b) => a.key.localeCompare(b.key)),
    documentationCoordinates: sha(Buffer.from(JSON.stringify(documentation))),
    documentationText: sha(
      Buffer.from(JSON.stringify(withoutCoordinates(documentation))),
    ),
  };
}

function prepare(output, root = repository) {
  output = path.resolve(output);
  if (fs.existsSync(output))
    throw Error(
      "rehearsal output already exists; choose a new owned directory",
    );
  const pinned = revisions(root);
  const accepted = path.join(root, "sources/darabonba");
  const sourceManifest = JSON.parse(
    fs.readFileSync(path.join(accepted, "manifest.json")),
  );
  // The current candidate is a real newer repository revision with byte-identical
  // STS inputs. Do not silently approve a future source or import change.
  for (const name of ["main.tea", "Teafile", "api-info.json"])
    if (
      !pinned.files
        .get(pinned.manifest.candidateRevision + "/" + name)
        .equals(fs.readFileSync(path.join(accepted, "products/sts", name)))
    )
      throw Error(
        "candidate differs from accepted STS source; separate source review required",
      );
  for (const revision of [
    pinned.manifest.baselineRevision,
    pinned.manifest.candidateRevision,
  ])
    if (
      !pinned.files
        .get(revision + "/LICENSE")
        .equals(fs.readFileSync(path.join(accepted, "LICENSE.upstream")))
    )
      throw Error("source license drift requires review");
  const records = {};
  for (const [stage, revision] of [
    ["baseline", pinned.manifest.baselineRevision],
    ["candidate", pinned.manifest.candidateRevision],
  ]) {
    const workspace = path.join(output, stage),
      sources = path.join(workspace, "sources/darabonba");
    const manifest = structuredClone(sourceManifest);
    manifest.revision = revision;
    manifest.products = { sts: "sts-20150401" };
    // This rehearsal intentionally narrows the corpus to STS. Do not retain
    // registrations whose product files are excluded from the fixture.
    delete manifest.stagedProducts;
    manifest.files = manifest.files.filter(
      (f) =>
        !f.file.startsWith("products/") || f.file.startsWith("products/sts/"),
    );
    for (const record of manifest.files) {
      let bytes = fs.readFileSync(path.join(accepted, record.file));
      const name = record.file.startsWith("products/sts/")
        ? record.file.slice("products/sts/".length)
        : record.file === "LICENSE.upstream"
          ? "LICENSE"
          : null;
      if (name && pinned.files.has(revision + "/" + name)) {
        bytes = pinned.files.get(revision + "/" + name);
        record.url = pinned.manifest.files.find(
          (r) => r.file === revision + "/" + name,
        ).url;
      }
      record.sha256 = sha(bytes);
      const target = path.join(sources, record.file);
      fs.mkdirSync(path.dirname(target), { recursive: true });
      fs.writeFileSync(target, bytes);
    }
    const manifestBytes = Buffer.from(JSON.stringify(manifest, null, 2) + "\n");
    fs.writeFileSync(path.join(sources, "manifest.json"), manifestBytes);
    run("generate", { root: workspace });
    run("check", { root: workspace });
    const projected = project(
      workspace,
      stage === "candidate"
        ? [
            "sts/AssumeRole",
            "sts/GetCallerIdentity",
            "sts/AssumeRoleWithOIDC",
            "sts/AssumeRoleWithSAML",
          ]
        : [],
    );
    const second = project(workspace);
    if (
      Object.keys(projected.files).some(
        (f) => !projected.files[f].equals(second.files[f]),
      )
    )
      throw Error("non-deterministic source lowering");
    const main = path.join(sources, "products/sts/main.tea");
    records[stage] = {
      revision,
      sourceManifestSHA256: sha(manifestBytes),
      fingerprints: fingerprints(projected.products.sts.ir),
      facts: sourceFacts(parser.parse(fs.readFileSync(main, "utf8"), main)),
      counts: projected.products.sts.coverage.counts,
    };
  }
  const baseline = records.baseline.fingerprints,
    candidate = records.candidate.fingerprints;
  const oldFields = new Map(baseline.fields.map((f) => [f.key, f])),
    newFields = new Map(candidate.fields.map((f) => [f.key, f]));
  const report = {
    schemaVersion: 1,
    baseline: records.baseline.revision,
    candidate: records.candidate.revision,
    productionRevision: sourceManifest.revision,
    productionSTSSourceUnchanged: true,
    parserVersion: sourceManifest.parserVersion,
    counts: {
      baseline: records.baseline.counts,
      candidate: records.candidate.counts,
    },
    drift: {
      addedFields: candidate.fields.filter((f) => !oldFields.has(f.key)),
      removedFields: baseline.fields.filter((f) => !newFields.has(f.key)),
      changedFields: candidate.fields.filter(
        (f) =>
          oldFields.has(f.key) &&
          JSON.stringify(oldFields.get(f.key)) !== JSON.stringify(f),
      ),
      operationLoweringChanged:
        JSON.stringify(baseline.operations) !==
        JSON.stringify(candidate.operations),
      handoffs: {
        baseline: records.baseline.facts.handoffs,
        candidate: records.candidate.facts.handoffs,
      },
      signatureAlgorithm: {
        baseline: records.baseline.facts.signatureAlgorithm,
        candidate: records.candidate.facts.signatureAlgorithm,
      },
      endpointInitializationChanged:
        JSON.stringify(records.baseline.facts.endpointMap) !==
        JSON.stringify(records.candidate.facts.endpointMap),
      documentationTextChanged:
        baseline.documentationText !== candidate.documentationText,
      documentationCoordinatesChanged:
        baseline.documentationCoordinates !==
        candidate.documentationCoordinates,
      licenseChanged: false,
      importsRepinned: false,
    },
    compilation: "not-assessed",
    live: "not-assessed",
  };
  const target = path.join(output, "candidate");
  fs.mkdirSync(path.join(target, "policies"), { recursive: true });
  const policy = JSON.parse(
    fs.readFileSync(path.join(root, "policies/sts.json")),
  );
  fs.writeFileSync(
    path.join(target, "policies/sts.json"),
    JSON.stringify(policy, null, 2) + "\n",
  );
  policy.sourceManifestSHA256 = records.candidate.sourceManifestSHA256;
  fs.writeFileSync(
    path.join(output, "reviewed-sts-policy.json"),
    JSON.stringify(policy, null, 2) + "\n",
  );
  fs.writeFileSync(
    path.join(output, "report.json"),
    JSON.stringify(report, null, 2) + "\n",
  );
  return report;
}

function prepareRuntime(candidate, root = repository) {
  candidate = path.resolve(candidate);
  if (!fs.existsSync(path.join(candidate, "service/sts/operations.gen.go")))
    throw Error("generate reviewed candidate before preparing compilation");
  const files = new Map([
    [
      "go.mod",
      Buffer.from(
        "module github.com/rambow-cloud/alicloud-go-sdk-x\n\ngo 1.27.0\n",
      ),
    ],
  ]);
  for (const directory of [
    ".",
    "credentials",
    "endpoint",
    "middleware",
    "retry",
    "pagination",
    "waiter",
    "sdktest",
    "internal/signing",
    "internal/rpcmodel",
    "internal/checksum",
    "internal/xmlmodel",
  ]) {
    for (const entry of fs.readdirSync(path.join(root, directory), {
      withFileTypes: true,
    }))
      if (
        entry.isFile() &&
        entry.name.endsWith(".go") &&
        !entry.name.endsWith(".gen.go") &&
        !entry.name.endsWith("_test.go")
      )
        files.set(
          path.join(directory, entry.name),
          fs.readFileSync(path.join(root, directory, entry.name)),
        );
  }
  for (const name of [
    "identity_test.go",
    "federation_test.go",
    "policy_test.go",
  ])
    files.set(
      "service/sts/" + name,
      fs.readFileSync(path.join(root, "service/sts", name)),
    );
  for (const relative of files.keys())
    if (fs.existsSync(path.join(candidate, relative)))
      throw Error("candidate compilation fixture already exists");
  for (const [relative, bytes] of files) {
    const target = path.join(candidate, relative);
    fs.mkdirSync(path.dirname(target), { recursive: true });
    fs.writeFileSync(target, bytes);
  }
  return "Prepared standalone canonical runtime + generated STS + independent contracts/Examples; compilation not yet assessed.";
}
if (require.main === module) {
  try {
    if (process.argv.length === 4 && process.argv[2] === "--runtime")
      console.log(prepareRuntime(process.argv[3]));
    else if (process.argv.length === 3)
      console.log(JSON.stringify(prepare(process.argv[2]), null, 2));
    else
      throw Error(
        "usage: node tools/darabonba/rehearse-sts.cjs NEW_OUTPUT_DIRECTORY | --runtime CANDIDATE_DIRECTORY",
      );
  } catch (e) {
    console.error(e.message);
    process.exitCode = 1;
  }
}
module.exports = { revisions, fingerprints, prepare, prepareRuntime };
