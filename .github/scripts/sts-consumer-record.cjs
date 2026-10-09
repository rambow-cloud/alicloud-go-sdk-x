"use strict";

const requiredTests = {
  identity: [
    "TestExternalDefaultProfileWorkload",
    "TestExternalOAuthRefreshPersistenceAndReuse",
    "TestExternalLongLivedProfileRequiresOptIn",
    "TestConsumerReviewIdentityOptionsAndInflightCancellation",
  ],
  "provider-cache": [
    "TestExternalGeneratedWorkload",
    "TestConsumerReviewNativeRoleCacheAndCanceledWaiter",
  ],
  "mock-errors": [
    "TestExternalSmallMockStructuredErrorsAndCancellation",
    "TestConsumerReviewSmallMockAndErrorIdentity",
    "TestConsumerReviewIssuanceDoesNotRetryAndErrorRedaction",
  ],
  anonymous: [
    "TestConsumerReviewAnonymousTokenPreservationAndProviderIsolation",
  ],
  comparison: ["TestPinnedOfficialWorkload"],
};
const consumerPackage =
  "github.com/rambow-cloud/alicloud-go-sdk-x/examples/stsacceptance";

function collect(log, metadata) {
  let events;
  try {
    events = log
      .trim()
      .split(/\r?\n/)
      .map((line) => JSON.parse(line));
  } catch {
    throw Error("consumer test events cannot be decoded");
  }
  if (events.some((e) => !e || typeof e !== "object" || Array.isArray(e)))
    throw Error("consumer test event shape is invalid");
  if (
    events.some((e) => e.Action === "fail") ||
    !events.some(
      (e) => e.Package === consumerPackage && !e.Test && e.Action === "pass",
    )
  )
    throw Error("consumer package did not pass");
  const tasks = Object.entries(requiredTests).map(([id, names]) => {
    const tests = names.map((name) => {
      const passed = events.filter(
        (e) =>
          e.Package === consumerPackage &&
          e.Test === name &&
          e.Action === "pass",
      );
      if (
        passed.length !== 1 ||
        typeof passed[0].Elapsed !== "number" ||
        !Number.isFinite(passed[0].Elapsed) ||
        passed[0].Elapsed < 0
      )
        throw Error(
          "required consumer test is missing or has invalid timing: " + name,
        );
      return { name, status: "PASS", elapsedSeconds: passed[0].Elapsed };
    });
    return {
      id,
      status: "PASS",
      elapsedSeconds: tests.reduce((total, t) => total + t.elapsedSeconds, 0),
      tests,
    };
  });
  return {
    ...metadata,
    schemaVersion: 1,
    status: "PASS",
    reviewKind: "implementation-agent",
    independentHuman: false,
    officialSTSVersion: "v2.1.0",
    evidence: "docs/sts-agent-closeout.md",
    timingKind: "automated-test-execution",
    tasks,
  };
}

module.exports = { requiredTests, collect };
