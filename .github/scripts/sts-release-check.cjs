"use strict";

// Read-only release gate. It never creates tags, publishes, or infers human UX.
const fs = require("node:fs");
const path = require("node:path");
const { execFileSync } = require("node:child_process");
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

function check(repository = root) {
  function readEvidence(relative, label) {
    try {
      return JSON.parse(fs.readFileSync(path.join(repository, relative)));
    } catch {
      throw Error(label + " evidence cannot be read or decoded");
    }
  }
  const human = readEvidence(
    "docs/acceptance/sts-independent-result.json",
    "Independent developer",
  );
  const problem = validateHuman(human);
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
  requireUnchangedBehavior(repository, human.sdkCommit);
  return "PASS local acceptance guard; final-main CI, closed #60, remote-tag absence, publication and browser indexing must still be verified separately. No mutation performed.";
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
        "docs/products",
        "docs/credentials.md",
        "docs/default-configuration.md",
        "docs/sts-credentials.md",
        "docs/sts-anonymous-rpc.md",
        "docs/sts-consumer-acceptance.md",
        ".github/workflows/ci.yml",
      ],
      { cwd: repository, stdio: "pipe" },
    );
  } catch {
    throw Error(
      "SDK or acceptance workload changed since independent tasks; review and rerun affected acceptance",
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
module.exports = { validateHuman, check, requireUnchangedBehavior };
