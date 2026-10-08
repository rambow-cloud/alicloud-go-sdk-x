"use strict";
const { test } = require("node:test");
const assert = require("node:assert/strict");
const fs = require("node:fs");
const path = require("node:path");
const { validateHuman } = require("./sts-release-check.cjs");

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
