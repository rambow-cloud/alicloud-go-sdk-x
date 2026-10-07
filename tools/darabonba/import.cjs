"use strict";

// Explicit network import. Generation never resolves wildcard dependencies online.
const fs = require("node:fs");
const path = require("node:path");
const crypto = require("node:crypto");
const tar = require("tar");
const {
  Config,
  DownloadModuleObject,
  default: RepoClient,
} = require("@darabonba/repo-client");
const repository = path.resolve(__dirname, "../..");
const root = path.join(repository, "sources/darabonba");
const revision = "ec489e5c3deae95496daae2b41503ac58b221adb";
const products = {
  ecs: "ecs-20140526",
  sts: "sts-20150401",
  vpc: "vpc-20160428",
};
const hash = (data, algorithm = "sha256") =>
  crypto.createHash(algorithm).update(data).digest("hex");
const readMeta = (dir) =>
  JSON.parse(
    fs.readFileSync(
      path.join(
        dir,
        fs.existsSync(path.join(dir, "Darafile")) ? "Darafile" : "Teafile",
      ),
      "utf8",
    ),
  );
const licenseRepos = {
  Credential: "credentials-nodejs",
  EndpointUtil: "endpoint-util",
  GatewayPOP: "alibabacloud-gateway",
  GatewaySPI: "alibabacloud-gateway",
  OpenApi: "darabonba-openapi",
  OpenApiUtil: "darabonba-openapi-util",
  Paginator: "alibabacloud-openapi-paginator",
  Array: "darabonba-array",
  EncodeUtil: "darabonba-crypto-util",
  Map: "darabonba-map",
  SignatureUtil: "darabonba-crypto-util",
  String: "darabonba-string",
  Util: "tea-util",
  XML: "tea-xml",
};
async function pinLicenses(manifest) {
  const cache = new Map();
  for (const module of manifest.modules) {
    const meta = readMeta(path.join(root, module.directory));
    module.scope = meta.scope;
    module.name = meta.name;
    module.version = meta.version;
    const repo = "aliyun/" + licenseRepos[meta.name];
    if (!cache.has(repo)) {
      const response = await fetch(
        "https://api.github.com/repos/" + repo + "/license",
        { signal: AbortSignal.timeout(40000) },
      );
      let evidence;
      if (response.status === 404) {
        evidence = JSON.parse(
          await download(
            "https://api.github.com/repos/" + repo + "/contents/README.md",
          ),
        );
        if (
          !Buffer.from(evidence.content || "", "base64")
            .toString()
            .includes("Apache-2.0")
        )
          throw new Error("Missing source license declaration: " + repo);
        evidence.license = { spdx_id: "Apache-2.0" };
      } else {
        if (!response.ok)
          throw new Error("License HTTP " + response.status + ": " + repo);
        evidence = await response.json();
      }
      if (
        !["Apache-2.0", "MIT"].includes(evidence.license?.spdx_id) ||
        evidence.encoding !== "base64"
      )
        throw new Error("Unreviewed module license: " + repo);
      const data = Buffer.from(evidence.content, "base64"),
        file = "licenses/" + repo.slice(7) + ".NOTICE";
      write(file, data);
      const record = {
        file,
        sha256: hash(data),
        url:
          "https://api.github.com/repos/" + repo + "/git/blobs/" + evidence.sha,
      };
      manifest.files.push(record);
      cache.set(repo, {
        repository: "https://github.com/" + repo,
        spdx: evidence.license.spdx_id,
        file,
        gitBlobSHA: evidence.sha,
      });
    }
    module.license = {
      ...cache.get(repo),
      evidence:
        meta.name === "Credential"
          ? "related-runtime-repository"
          : "module-source-repository",
    };
    if (meta.name === "Credential") module.license.sourceSPDX = "NOASSERTION";
  }
  manifest.files.sort((a, b) => a.file.localeCompare(b.file));
}
async function download(url) {
  const response = await fetch(url, { signal: AbortSignal.timeout(40000) });
  if (!response.ok) throw new Error("Import HTTP " + response.status);
  return Buffer.from(await response.arrayBuffer());
}
function write(relative, data) {
  const file = path.join(root, relative);
  fs.mkdirSync(path.dirname(file), { recursive: true });
  fs.writeFileSync(file, data);
}
async function main() {
  const files = [];
  for (const [pkg, upstream] of Object.entries(products))
    for (const name of ["main.tea", "Teafile", "api-info.json"]) {
      const relative = "products/" + pkg + "/" + name,
        url = `https://raw.githubusercontent.com/aliyun/alibabacloud-sdk/${revision}/${upstream}/${name}`;
      const data = await download(url);
      write(relative, data);
      files.push({ file: relative, sha256: hash(data), url });
    }
  const licenseURL = `https://raw.githubusercontent.com/aliyun/alibabacloud-sdk/${revision}/LICENSE`;
  const license = await download(licenseURL);
  write("LICENSE.upstream", license);
  files.push({
    file: "LICENSE.upstream",
    sha256: hash(license),
    url: licenseURL,
  });
  const client = new RepoClient(
    new Config({
      endpoint: "darabonba.api.aliyun.com",
      protocol: "https",
      auth: "",
    }),
  );
  const resolutions = {},
    modules = [];
  const pending = new Set();
  for (const pkg of Object.keys(products))
    for (const spec of Object.values(
      readMeta(path.join(root, "products", pkg)).libraries || {},
    ))
      pending.add(spec);
  for (let depth = 0; pending.size; depth++) {
    if (depth > 24) throw new Error("Module dependency depth exceeded");
    const requested = [...pending].filter((s) => !resolutions[s]).sort();
    pending.clear();
    if (!requested.length) break;
    const response = await client.downloadModule(
      new DownloadModuleObject({ specs: requested.join(",") }),
    );
    if (!response.ok)
      throw new Error(
        "Official module registry rejected dependency resolution",
      );
    for (const info of Object.values(response.download_list || {})) {
      if (!info?.dist_tarball) continue;
      const relative = "modules/" + info.dist_dir.replaceAll(":", "_");
      if (
        relative.includes("..") ||
        path.isAbsolute(info.dist_dir) ||
        relative.includes("\\")
      )
        throw new Error("Unsafe module directory");
      resolutions[info.version] = relative;
      if (!modules.some((m) => m.spec === info.version)) {
        const data = await download(info.dist_tarball);
        if (hash(data, "sha1") !== info.dist_shasum)
          throw new Error("Official module archive checksum mismatch");
        const archive = path.join(repository, ".git", "darabonba-module.tgz");
        fs.writeFileSync(archive, data);
        const target = path.join(root, relative);
        fs.mkdirSync(target, { recursive: true });
        await tar.x({
          file: archive,
          cwd: target,
          strict: true,
          filter: (name, entry) => {
            if (
              path.isAbsolute(name) ||
              name.split(/[\\/]/).includes("..") ||
              !["File", "Directory", "OldFile"].includes(entry.type)
            )
              throw new Error("Unsupported archive entry");
            return true;
          },
        });
        modules.push({
          spec: info.version,
          directory: relative,
          archiveSHA256: hash(data),
          archiveSHA1: info.dist_shasum,
          url: info.dist_tarball,
        });
        for (const spec of Object.values(readMeta(target).libraries || {}))
          if (!spec.startsWith(".") && !resolutions[spec]) pending.add(spec);
      }
    }
    for (const spec of requested)
      if (!resolutions[spec])
        throw new Error("No pinned resolution returned for " + spec);
    console.log("Resolved " + modules.length + " official modules.");
  }
  for (const pkg of Object.keys(products))
    write(
      "products/" + pkg + "/.libraries.json",
      JSON.stringify(
        Object.fromEntries(
          Object.entries(resolutions)
            .sort()
            .map(([spec, dir]) => [spec, "../../" + dir]),
        ),
        null,
        2,
      ) + "\n",
    );
  function walk(dir) {
    for (const entry of fs.readdirSync(path.join(root, dir), {
      withFileTypes: true,
    })) {
      const relative = dir + "/" + entry.name;
      if (entry.isDirectory()) walk(relative);
      else
        files.push({
          file: relative,
          sha256: hash(fs.readFileSync(path.join(root, relative))),
        });
    }
  }
  walk("modules");
  for (const pkg of Object.keys(products)) {
    const file = "products/" + pkg + "/.libraries.json";
    files.push({ file, sha256: hash(fs.readFileSync(path.join(root, file))) });
  }
  const manifest = {
    schemaVersion: 1,
    repository: "https://github.com/aliyun/alibabacloud-sdk",
    revision,
    license: "Apache-2.0",
    parserVersion: require("@darabonba/parser/package.json").version,
    products,
    modules: modules.sort((a, b) => a.spec.localeCompare(b.spec)),
    files: files.sort((a, b) => a.file.localeCompare(b.file)),
  };
  await pinLicenses(manifest);
  write("manifest.json", JSON.stringify(manifest, null, 2) + "\n");
  console.log(
    "Pinned " + files.length + " files from " + modules.length + " modules.",
  );
}
if (require.main === module)
  main().catch((error) => {
    console.error(error.message);
    process.exitCode = 1;
  });
module.exports = { pinLicenses };
