const { test } = require("node:test");
const assert = require("node:assert/strict");
const {
  check,
  projectDocument,
  englishText,
  prose,
} = require("./check-doc-language.cjs");
const files = new Set(["docs/a.md", "docs/a.zh-CN.md"]);
const valid = "# Usage\n\n[中文](a.zh-CN.md)\n\n- Use the client.\n";
test("separate files require real counterparts and correct language links", () => {
  assert.deepEqual(
    check("docs/a.md", valid, files, (p) => files.has(p)),
    [],
  );
  assert.deepEqual(
    check(
      "docs/a.zh-CN.md",
      "# 用法\n\n[English](a.md)\n\n- 使用客户端。",
      files,
      (p) => files.has(p),
    ),
    [],
  );
  assert.match(
    check("docs/a.md", valid, new Set(["docs/a.md"]), () => false).join(),
    /missing language counterpart/,
  );
  assert.match(
    check(
      "docs/a.zh-CN.md",
      "# 用法\n[English](a.zh-CN.md)",
      files,
      () => true,
    ).join(),
    /missing language link/,
  );
});
test("mixed prose, empty Chinese guides and broken local links fail", () => {
  for (const suffix of [
    "\n## English\nUsage\n## 中文\n使用",
    "\n中文内容",
    "\n[x](missing.md)",
  ])
    assert.ok(
      check("docs/a.md", valid + suffix, files, (p) => files.has(p)).length,
    );
  assert.ok(
    check(
      "docs/a.zh-CN.md",
      "# Guide\n[English](a.md)\nUsage",
      files,
      () => true,
    ).length,
  );
});
test("fenced examples and literal wire values are preserved without language false positives", () => {
  const text =
    valid +
    '\n````markdown\n## 中文\n[example](missing.md)\n```\n````\n\n~~~go\nvalue := "中文"\n~~~';
  assert.deepEqual(
    check("docs/a.md", text, files, () => true),
    [],
  );
  assert.equal(englishText("- Test `中文` as an input value."), true);
  assert.equal(prose("```go\n## 中文\n```\nEnglish"), "English");
});
test("issue drafts and PR templates are English only", () => {
  assert.deepEqual(
    check(
      "docs/issues/70-docs.md",
      "# Problem\n- Split guides.",
      files,
      () => true,
    ),
    [],
  );
  assert.ok(
    check("docs/issues/70-docs.md", "# 问题", files, () => true).length,
  );
  assert.ok(
    check("docs/issues/70-docs.zh-CN.md", "# 问题", files, () => true).length,
  );
  assert.ok(
    check(
      ".github/PULL_REQUEST_TEMPLATE.md",
      "# Problem\n中文",
      files,
      () => true,
    ).length,
  );
});
test("excludes only pinned upstream READMEs, not project source guides", () => {
  assert.equal(
    projectDocument(
      "sources/darabonba/modules/alibabacloud_OpenApi_0.3.23/README.md",
    ),
    false,
  );
  assert.equal(
    projectDocument(
      "sources/darabonba/modules/darabonba_String_0.0.13/README-CN.md",
    ),
    false,
  );
  assert.equal(projectDocument("sources/darabonba/README.md"), true);
  assert.equal(projectDocument("sources/darabonba/modules/x/extra.md"), true);
});
