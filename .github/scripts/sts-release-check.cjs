"use strict";

// Read-only release gate. Agent acceptance never implies independent human UX.
const fs = require("node:fs");
const path = require("node:path");
const { execFileSync } = require("node:child_process");
const { requiredTests } = require("./sts-consumer-record.cjs");
const root = path.resolve(__dirname, "../..");
const taskIDs = [
  "identity",
  "provider-cache",
  "mock-errors",
  "anonymous",
  "comparison",
];

function validateHuman(human) {
  if (human.schemaVersion !== 1 || human.status !== "PASS")
    return "independent developer acceptance is not PASS";
  if (human.developer?.independent !== true || !human.developer.alias?.trim())
    return "independent developer identity/assertion is missing";
  if (
    !/^[a-f0-9]{40}$/.test(human.sdkCommit || "") ||
    !human.evidence?.trim() ||
    !human.os?.trim() ||
    human.officialSTSVersion !== "v2.1.0"
  )
    return "pinned independent acceptance evidence is incomplete";
  const version = /^go(\d+)\.(\d+)(?:\.|$)/.exec(human.goVersion || "");
  if (
    !version ||
    Number(version[1]) < 1 ||
    (Number(version[1]) === 1 && Number(version[2]) < 27)
  )
    return "independent acceptance requires Go 1.27+";
  if (
    !Array.isArray(human.tasks) ||
    human.tasks.length !== taskIDs.length ||
    taskIDs.some((id) => human.tasks.filter((t) => t.id === id).length !== 1)
  )
    return "independent task inventory is incomplete or duplicated";
  if (
    human.tasks.some(
      (t) =>
        t.status !== "PASS" ||
        typeof t.elapsedMinutes !== "number" ||
        !Number.isFinite(t.elapsedMinutes) ||
        t.elapsedMinutes < 0,
    )
  )
    return "required independent tasks/timings are incomplete";
  return null;
}

function validateAgent(record) {
  if (record.schemaVersion !== 1 || record.status !== "PASS")
    return "agent consumer acceptance is not PASS";
  if (
    record.reviewKind !== "implementation-agent" ||
    record.independentHuman !== false
  )
    return "agent reviewer classification is missing or misleading";
  const version = /^go(\d+)\.(\d+)(?:\.|$)/.exec(record.goVersion || "");
  if (
    !/^[a-f0-9]{40}$/.test(record.sdkCommit || "") ||
    !record.evidence?.trim() ||
    !record.os?.trim() ||
    record.officialSTSVersion !== "v2.1.0" ||
    !version ||
    Number(version[1]) < 1 ||
    (Number(version[1]) === 1 && Number(version[2]) < 27)
  )
    return "pinned agent acceptance evidence is incomplete";
  if (
    record.timingKind !== "automated-test-execution" ||
    !Array.isArray(record.tasks) ||
    record.tasks.length !== taskIDs.length ||
    taskIDs.some((id) => record.tasks.filter((t) => t.id === id).length !== 1)
  )
    return "agent task inventory or timing classification is incomplete";
  if (
    record.tasks.some(
      (t) =>
        t.status !== "PASS" ||
        !Array.isArray(t.tests) ||
        t.tests.length === 0 ||
        typeof t.elapsedSeconds !== "number" ||
        !Number.isFinite(t.elapsedSeconds) ||
        t.elapsedSeconds < 0,
    )
  )
    return "required agent consumer tasks are incomplete";
  if (
    record.tasks.some(
      (t) =>
        t.tests.length !== requiredTests[t.id].length ||
        requiredTests[t.id].some(
          (name) =>
            t.tests.filter(
              (test) =>
                test.name === name &&
                test.status === "PASS" &&
                typeof test.elapsedSeconds === "number" &&
                Number.isFinite(test.elapsedSeconds) &&
                test.elapsedSeconds >= 0,
            ).length !== 1,
        ),
    )
  )
    return "required agent consumer test evidence is incomplete";
  return null;
}

function validateProduct(record, product) {
  if (
    record.schemaVersion !== 1 ||
    record.product !== product ||
    record.status !== "PASS"
  )
    return product + " product acceptance is not PASS";
  const ids = [
    "generation",
    "consumer",
    "pagination",
    "retry-errors",
    "live",
    "docs",
  ];
  if (product === "ecs") ids.push("waiter");
  if (
    !/^[a-f0-9]{40}$/.test(record.sdkCommit || "") ||
    !record.evidence?.trim() ||
    !Array.isArray(record.requiredCases) ||
    record.requiredCases.length !== ids.length ||
    ids.some(
      (id) =>
        record.requiredCases.filter((c) => c.id === id && c.status === "PASS")
          .length !== 1,
    )
  )
    return product + " product acceptance evidence is incomplete";
  return null;
}

function check(repository = root) {
  function readEvidence(relative, label) {
    try {
      return JSON.parse(fs.readFileSync(path.join(repository, relative)));
    } catch {
      throw Error(label + " evidence cannot be read or decoded");
    }
  }
  const agent = readEvidence(
    "docs/acceptance/sts-agent-result.json",
    "Agent consumer",
  );
  const problem = validateAgent(agent);
  if (problem) throw Error(problem);
  const coverage = readEvidence(
    "docs/products/sts.coverage.json",
    "STS coverage",
  );
  const actions = [
    "AssumeRole",
    "GetCallerIdentity",
    "AssumeRoleWithOIDC",
    "AssumeRoleWithSAML",
  ];
  if (
    coverage.discovered !== 4 ||
    coverage.lowered !== 4 ||
    coverage.emitted !== 4 ||
    coverage.models !== 19 ||
    coverage.operations.length !== 4 ||
    actions.some(
      (name) =>
        coverage.operations.filter(
          (o) => o.name === name && o.status === "emitted",
        ).length !== 1,
    )
  )
    throw Error("pinned STS emission scope does not match release notes");
  const live = readEvidence(
    "docs/acceptance/sts-identity-live.json",
    "Live identity",
  );
  const source = readEvidence(
    "docs/acceptance/sts-source-rehearsal.json",
    "Source rehearsal",
  );
  const profile = readEvidence(
    "docs/acceptance/profile-oauth-live.json",
    "Native Profile/OAuth",
  );
  if (
    profile.Status !== "PASS" ||
    profile.NativeCLIConfig !== true ||
    profile.CLISubprocess !== false ||
    profile.SuccessfulNativeOAuthExchange !== "PASS"
  )
    throw Error("required native Profile/OAuth acceptance is incomplete");
  if (
    live.Status !== "PASS" ||
    live.IdentityFieldsCompared !== 6 ||
    source.verification?.candidateCompilationAndIndependentContracts !== "PASS"
  )
    throw Error("scoped live/source acceptance evidence is incomplete");
  const products = ["ecs", "vpc"].map((product) => {
    const record = readEvidence(
      "docs/acceptance/" + product + "-product-result.json",
      product + " product",
    );
    const problem = validateProduct(record, product);
    if (problem) throw Error(problem);
    return record;
  });
  if (
    execFileSync("git", ["branch", "--show-current"], {
      cwd: repository,
      encoding: "utf8",
    }).trim() !== "main"
  )
    throw Error("release candidate must be on main");
  if (
    execFileSync("git", ["status", "--porcelain"], {
      cwd: repository,
      encoding: "utf8",
    }).trim()
  )
    throw Error("release candidate working tree must be clean");
  requireUnchangedBehavior(repository, agent.sdkCommit);
  for (const product of products)
    requireUnchangedBehavior(repository, product.sdkCommit);
  return "PASS local STS/ECS/VPC acceptance guard; final-main CI, closed #60/#74/#75, remote-tag absence, publication and browser indexing must still be verified separately. No mutation performed.";
}

function requireUnchangedBehavior(repository, revision) {
  try {
    execFileSync(
      "git",
      [
        "diff",
        "--quiet",
        revision,
        "HEAD",
        "--",
        ":(glob)**/*.go",
        "go.mod",
        "go.sum",
        "sources",
        "models",
        "metadata",
        "policies",
        "tools/darabonba",
        "examples/stsacceptance",
        "examples/productacceptance",
        "docs/products",
        "docs/credentials.md",
        "docs/default-configuration.md",
        "docs/sts-credentials.md",
        "docs/sts-anonymous-rpc.md",
        "docs/sts-consumer-acceptance.md",
        ".github/workflows/ci.yml",
        ".github/scripts/sts-release-check.cjs",
        ".github/scripts/sts-consumer-record.cjs",
      ],
      { cwd: repository, stdio: "pipe" },
    );
  } catch {
    throw Error(
      "SDK or acceptance workload changed since recorded tasks; review and rerun affected acceptance",
    );
  }
}

if (require.main === module) {
  try {
    console.log(check());
  } catch (e) {
    console.error("BLOCKED: " + e.message);
    process.exitCode = 2;
  }
}
module.exports = {
  validateHuman,
  validateAgent,
  validateProduct,
  check,
  requireUnchangedBehavior,
};
