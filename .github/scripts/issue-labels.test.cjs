const { test } = require("node:test");
const assert = require("node:assert/strict");
const { classify, planLabels, run } = require("./issue-labels.cjs");

test("classifies structured areas and ignores arbitrary issue text", () => {
  assert.deepEqual(
    classify({
      title: "[Feature]: transport",
      body: "### Affected areas\n\ntransport, signing\n\n### Problem\n\ncredentials",
    }),
    { type: "enhancement", modules: ["module:transport", "module:signing"] },
  );
  assert.deepEqual(
    classify({ title: "Question about docs", body: "some mention of ecs" }),
    { type: null, modules: null },
  );
  assert.deepEqual(
    classify({
      title: "[Docs]: clearer bullets",
      body: "### Affected areas\n\n- tools\n- core\n\n### Scope\n\n- ecs",
    }),
    { type: "documentation", modules: ["module:tools", "module:core"] },
  );
});

test("preserves priority, active state and manual labels while updating areas", () => {
  const plan = planLabels(
    {
      title: "[Docs]: docs",
      body: "### Affected areas\n\ncore\n",
      state: "open",
      labels: [
        "enhancement",
        "module:ecs",
        "priority:p1",
        "status:in-progress",
        "help wanted",
      ],
    },
    "edited",
  );
  assert.deepEqual(plan.add, ["documentation", "module:core"]);
  assert.deepEqual(plan.remove, ["enhancement", "module:ecs"]);
});

test("closed issues become done and reopening requires triage", () => {
  const closed = planLabels(
    {
      title: "work",
      state: "closed",
      labels: ["priority:p1", "status:blocked"],
    },
    "closed",
  );
  assert.deepEqual(closed, {
    add: ["status:done"],
    remove: ["status:blocked"],
  });
  const reopened = planLabels(
    { title: "work", state: "open", labels: ["priority:p1", "status:done"] },
    "reopened",
  );
  assert.deepEqual(reopened, {
    add: ["status:triage"],
    remove: ["status:done"],
  });
});

test("new issues receive defaults; missing area fields preserve manual areas", () => {
  const plan = planLabels(
    {
      title: "[Bug]: failure",
      body: "",
      state: "open",
      labels: ["module:ecs"],
    },
    "opened",
  );
  assert.deepEqual(plan, {
    add: ["bug", "priority:p2", "status:triage"],
    remove: [],
  });
});

test("VPC structured areas replace stale modules and preserve active state", () => {
  const plan = planLabels(
    {
      title: "[Feature]: VPC paginator",
      body: "### Affected areas\n\nvpc, endpoints, tools\n\n### Scope\n\necs is only a reference",
      state: "open",
      labels: [
        "enhancement",
        "module:ecs",
        "priority:p2",
        "status:in-progress",
      ],
    },
    "edited",
  );
  assert.deepEqual(plan, {
    add: ["module:vpc", "module:endpoints", "module:tools"],
    remove: ["module:ecs"],
  });
});

test("non-English issues fail the workflow without rewriting text or active labels", async () => {
  const fs = require("node:fs");
  const catalog = JSON.parse(fs.readFileSync(".github/labels.json", "utf8"));
  const issue = {
    number: 70,
    title: "[Docs]: Writing",
    body: "### Problem\n\n中文说明",
    state: "open",
    labels: ["documentation", "priority:p1", "status:in-progress"],
  };
  const failures = [];
  await run({
    github: {
      paginate: async () => catalog,
      rest: {
        issues: {
          listLabelsForRepo() {},
          get: async () => ({ data: issue }),
          addLabels() {
            assert.fail("unexpected label mutation");
          },
          removeLabel() {
            assert.fail("unexpected label mutation");
          },
        },
      },
    },
    context: {
      repo: { owner: "owner", repo: "repo" },
      payload: { issue, action: "edited" },
    },
    core: {
      info() {},
      setFailed(message) {
        failures.push(message);
      },
    },
  });
  assert.equal(failures.length, 1);
  assert.match(failures[0], /Issue #70 must use English/);
  assert.equal(issue.body, "### Problem\n\n中文说明");
  assert.deepEqual(issue.labels, [
    "documentation",
    "priority:p1",
    "status:in-progress",
  ]);
});
