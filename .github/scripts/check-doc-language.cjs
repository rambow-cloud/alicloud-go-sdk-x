"use strict";
const fs = require("node:fs");
const path = require("node:path");
const { execFileSync } = require("node:child_process");
const cjk = /[\u3400-\u9fff]/;

// Source-lock READMEs are upstream artifacts, not project guides.
function projectDocument(filename) {
  return !/^sources\/darabonba\/modules\/[^/]+\/README(?:-CN)?\.md$/.test(
    filename,
  );
}
function englishOnly(filename) {
  return (
    filename === ".github/PULL_REQUEST_TEMPLATE.md" ||
    /^docs\/issues\/(?!README(?:\.zh-CN)?\.md$).+\.md$/.test(filename)
  );
}
function prose(text) {
  const out = [];
  let fence = null;
  for (const line of text.replace(/\r\n/g, "\n").split("\n")) {
    const marker = /^\s*(`{3,}|~{3,})/.exec(line);
    if (marker) {
      if (!fence) fence = marker[1];
      else if (marker[1][0] === fence[0] && marker[1].length >= fence.length)
        fence = null;
      continue;
    }
    if (!fence) out.push(line);
  }
  return out
    .join("\n")
    .replace(/<!--[^]*?-->/g, "")
    .replace(/`[^`\n]*`/g, "");
}
function englishText(text) {
  return !cjk.test(prose(text));
}
function counterpart(filename) {
  return filename.endsWith(".zh-CN.md")
    ? filename.replace(/\.zh-CN\.md$/, ".md")
    : filename.replace(/\.md$/, ".zh-CN.md");
}
function check(filename, text, documents, exists) {
  const errors = [],
    body = prose(text);
  if (/^#{1,6}\s+(?:English|中文)\s*$/m.test(body))
    errors.push("mixed-language section marker");
  if (englishOnly(filename)) {
    if (filename.endsWith(".zh-CN.md") || !englishText(text))
      errors.push("issue/template text must use English");
  } else {
    const peer = counterpart(filename),
      label = filename.endsWith(".zh-CN.md") ? "English" : "中文";
    const link = `[${label}](${path.posix.basename(peer)})`;
    if (!documents.has(peer))
      errors.push("missing language counterpart: " + peer);
    if (!body.includes(link)) errors.push("missing language link: " + link);
    if (filename.endsWith(".zh-CN.md")) {
      if (!cjk.test(body.replace(link, "")))
        errors.push("Chinese guide has no Chinese prose");
    } else if (!englishText(body.replace(link, "")))
      errors.push("English guide contains Chinese prose");
  }
  for (const match of body.matchAll(/!?\[[^\]\n]*\]\(([^\s)]+)\)/g)) {
    const target = match[1];
    if (/^(?:[a-z][a-z\d+.-]*:|#|\/)/i.test(target)) continue;
    const file = target.split(/[?#]/)[0];
    let decoded;
    try {
      decoded = decodeURIComponent(file);
    } catch {
      errors.push("invalid local link");
      continue;
    }
    const resolved = path.posix.normalize(
      path.posix.join(path.posix.dirname(filename), decoded),
    );
    if (!exists(resolved)) errors.push("broken local link: " + target);
  }
  return errors;
}
function run() {
  const paths = execFileSync(
    "git",
    ["ls-files", "--cached", "--others", "--exclude-standard", "--", "*.md"],
    { encoding: "utf8" },
  )
    .trim()
    .split(/\r?\n/)
    .filter(Boolean);
  const files = [...new Set(paths)].filter((filename) =>
    fs.existsSync(filename),
  );
  const documents = new Set(files.filter(projectDocument)),
    failed = [];
  for (const filename of documents) {
    const errors = check(
      filename,
      fs.readFileSync(filename, "utf8"),
      documents,
      (p) => fs.existsSync(p),
    );
    if (errors.length) failed.push(filename + ": " + errors.join("; "));
  }
  for (const filename of fs.readdirSync(".github/ISSUE_TEMPLATE")) {
    if (
      filename.endsWith(".yml") &&
      !englishText(
        fs.readFileSync(".github/ISSUE_TEMPLATE/" + filename, "utf8"),
      )
    )
      failed.push(
        ".github/ISSUE_TEMPLATE/" + filename + ": issue form must use English",
      );
  }
  if (failed.length) {
    console.error(failed.join("\n"));
    process.exitCode = 1;
  } else
    console.log(
      `Language files and local links passed: ${documents.size} project files; ${files.length - documents.size} upstream READMEs preserved. Translation accuracy requires review.`,
    );
}
if (require.main === module) run();
module.exports = {
  check,
  projectDocument,
  englishOnly,
  englishText,
  prose,
  counterpart,
};
