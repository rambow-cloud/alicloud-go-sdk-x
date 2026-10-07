"use strict";

// Explicit network import; normalization and generation use only pinned local files.
const fs = require("node:fs");
const path = require("node:path");
const crypto = require("node:crypto");
const revision = "51286a65c79d008436eb314e636f9c9ad4b1ca08";
const root = path.resolve(__dirname, "../../sources/openapi-meta");
const files = [
  "LICENSE",
  "canonical/ecs/2014-05-26/version.json",
  ...["DescribeImages", "DescribeInstances", "DescribeRegions"].map(
    (name) => `canonical/ecs/2014-05-26/${name}.json`,
  ),
].sort();
async function main() {
  const fetched = [];
  for (const file of files) {
    const url = `https://raw.githubusercontent.com/aliyun/aliyun-openapi-meta/${revision}/${file}`;
    const response = await fetch(url, { signal: AbortSignal.timeout(40000) });
    if (!response.ok)
      throw new Error(`canonical: HTTP ${response.status}: ${file}`);
    const data = Buffer.from(await response.arrayBuffer());
    fetched.push({ file, url, data });
  }
  // Fetch the complete fixture set before replacing its lock or source bytes.
  for (const { file, data } of fetched) {
    const target = path.join(root, file);
    fs.mkdirSync(path.dirname(target), { recursive: true });
    fs.writeFileSync(target, data);
  }
  const manifest = {
    schemaVersion: 1,
    repository: "https://github.com/aliyun/aliyun-openapi-meta",
    revision,
    license: "Apache-2.0",
    adapterVersion: 1,
    files: fetched.map(({ file, url, data }) => ({
      file,
      url,
      sha256: crypto.createHash("sha256").update(data).digest("hex"),
    })),
  };
  fs.writeFileSync(
    path.join(root, "manifest.json"),
    JSON.stringify(manifest, null, 2) + "\n",
  );
  console.log(`Pinned ${fetched.length} canonical source artifacts.`);
}
if (require.main === module)
  main().catch((error) => {
    console.error(error.message);
    process.exitCode = 1;
  });
