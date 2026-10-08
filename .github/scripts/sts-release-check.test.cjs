"use strict";
const { test } = require("node:test");
const assert = require("node:assert/strict");
const fs = require("node:fs");
const path = require("node:path");
const os = require("node:os");
const { execFileSync } = require("node:child_process");
const {
  validateHuman,
  check,
  requireUnchangedBehavior,
} = require("./sts-release-check.cjs");

test("behavior guard detects deleted root Go files after the independent revision", (t) => {
  const dir = fs.mkdtempSync(path.join(os.tmpdir(), "release-tree-"));
  t.after(() => {
    assert.equal(path.dirname(path.resolve(dir)), path.resolve(os.tmpdir()));
    fs.rmSync(dir, { recursive: true, force: true });
  });
  const git = (args) =>
    execFileSync(
      "git",
      [
        "-c",
        "user.name=Release unit fixture",
        "-c",
        "user.email=fixture@example.invalid",
        "-c",
        "core.hooksPath=" + path.join(dir, "disabled-hooks"),
        ...args,
      ],
      { cwd: dir, encoding: "utf8", stdio: ["ignore", "pipe", "pipe"] },
    );
  git(["init", "-q"]);
  const file = path.join(dir, "runtime_fixture.go");
  fs.writeFileSync(file, "package fixture\n");
  git(["add", "runtime_fixture.go"]);
  git(["commit", "--no-gpg-sign", "-qm", "synthetic baseline"]);
  const revision = git(["rev-parse", "HEAD"]).trim();
  requireUnchangedBehavior(dir, revision);
  fs.unlinkSync(file);
  git(["add", "-u"]);
  git(["commit", "--no-gpg-sign", "-qm", "synthetic removal"]);
  assert.throws(
    () => requireUnchangedBehavior(dir, revision),
    /SDK or acceptance workload changed/,
  );
});

test("malformed acceptance reports never expose raw evidence in diagnostics", (t) => {
  const dir = fs.mkdtempSync(path.join(os.tmpdir(), "release-evidence-"));
  t.after(() => {
    assert.equal(path.dirname(path.resolve(dir)), path.resolve(os.tmpdir()));
    fs.rmSync(dir, { recursive: true, force: true });
  });
  const target = path.join(dir, "docs/acceptance");
  fs.mkdirSync(target, { recursive: true });
  fs.writeFileSync(
    path.join(target, "sts-independent-result.json"),
    "{secret-unit-fixture",
  );
  assert.throws(
    () => check(dir),
    (e) =>
      e.message === "Independent developer evidence cannot be read or decoded",
  );
});

// Synthetic unit fixture only; never written to project acceptance evidence.
function passed() {
  return {
    schemaVersion: 1,
    status: "PASS",
    developer: { alias: "test-only-developer", independent: true },
    sdkCommit: "a".repeat(40),
    evidence: "synthetic unit fixture",
    goVersion: "go1.27.1",
    os: "fixture",
    officialSTSVersion: "v2.1.0",
    tasks: [
      "identity",
      "provider-cache",
      "mock-errors",
      "anonymous",
      "comparison",
    ].map((id) => ({ id, status: "PASS", elapsedMinutes: 1 })),
  };
}
test("record shape accepts complete synthetic evidence without claiming actual human success", () =>
  assert.equal(validateHuman(passed()), null));
test("missing independence, version, tasks, times and pinned source block publication guard", () => {
  for (const mutate of [
    (h) => (h.status = "NOT RUN"),
    (h) => (h.developer.independent = false),
    (h) => (h.sdkCommit = ""),
    (h) => (h.evidence = ""),
    (h) => (h.goVersion = "go1.26.9"),
    (h) => h.tasks.pop(),
    (h) => (h.tasks[1].id = "identity"),
    (h) => (h.tasks[0].status = "FAIL"),
    (h) => (h.tasks[0].elapsedMinutes = null),
    (h) => (h.tasks[0].elapsedMinutes = -1),
  ]) {
    const h = passed();
    mutate(h);
    assert.ok(validateHuman(h));
  }
});
test("working acceptance template remains a truthful pending independent gate", () => {
  const h = JSON.parse(
    fs.readFileSync(
      path.resolve(
        __dirname,
        "../../docs/acceptance/sts-independent-result.json",
      ),
    ),
  );
  if (h.status !== "PASS") assert.match(validateHuman(h), /not PASS/);
  else assert.equal(validateHuman(h), null);
});

test("native Profile/OAuth evidence cannot be bypassed by completed synthetic human evidence", (t) => {
  const dir = fs.mkdtempSync(
    path.join(os.tmpdir(), "release-profile-evidence-"),
  );
  t.after(() => {
    assert.equal(path.dirname(path.resolve(dir)), path.resolve(os.tmpdir()));
    fs.rmSync(dir, { recursive: true, force: true });
  });
  const target = path.join(dir, "docs/acceptance"),
    products = path.join(dir, "docs/products");
  fs.mkdirSync(target, { recursive: true });
  fs.mkdirSync(products, { recursive: true });
  fs.writeFileSync(
    path.join(target, "sts-independent-result.json"),
    JSON.stringify(passed()),
  );
  fs.copyFileSync(
    path.join(__dirname, "../../docs/products/sts.coverage.json"),
    path.join(products, "sts.coverage.json"),
  );
  fs.writeFileSync(
    path.join(target, "sts-identity-live.json"),
    JSON.stringify({ Status: "PASS", IdentityFieldsCompared: 6 }),
  );
  fs.writeFileSync(
    path.join(target, "sts-source-rehearsal.json"),
    JSON.stringify({
      verification: { candidateCompilationAndIndependentContracts: "PASS" },
    }),
  );
  fs.writeFileSync(
    path.join(target, "profile-oauth-live.json"),
    JSON.stringify({
      Status: "BLOCKED",
      NativeCLIConfig: true,
      CLISubprocess: false,
      SuccessfulNativeOAuthExchange: "NOT RUN",
    }),
  );
  assert.throws(
    () => check(dir),
    /required native Profile\/OAuth acceptance is incomplete/,
  );
});
