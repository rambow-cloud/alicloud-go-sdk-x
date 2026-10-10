"use strict";

const test = require("node:test"), assert = require("node:assert/strict");
const fs = require("node:fs"), os = require("node:os"), path = require("node:path"), crypto = require("node:crypto"), tar = require("tar");
const { prepareModules, writeModules, archiveFiles } = require("./import-modules.cjs");
const { verifySources } = require("./frontend.cjs");
const sha = (b, algorithm = "sha256") => crypto.createHash(algorithm).update(b).digest("hex");

async function fixture(t, imports = { OpenApi: "alibabacloud:OpenApi:*" }) {
  const temporary = fs.mkdtempSync(path.join(os.tmpdir(), "alicloud-module-import-"));
  t.after(() => {
    // Resolve and verify the task directory before recursive cleanup on Windows.
    assert.equal(path.dirname(path.resolve(temporary)), path.resolve(os.tmpdir()));
    assert.ok(path.basename(temporary).startsWith("alicloud-module-import-"));
    fs.rmSync(temporary, { recursive: true, force: true });
  });
  const root = path.join(temporary,"source"), input = path.join(temporary,"archive");
  fs.cpSync(path.resolve(__dirname,"../../sources/darabonba"),root,{recursive:true});
  fs.mkdirSync(input);
  fs.writeFileSync(path.join(input,"Teafile"),JSON.stringify({ scope:"alibabacloud", name:"Fixture", version:"1.0.0", main:"./main.tea", libraries:imports }));
  fs.writeFileSync(path.join(input,"main.tea"),"static function fixture(): string;\n");
  fs.writeFileSync(path.join(input,"README.md"),"Original fixture license declaration: Apache-2.0\n");
  const chunks=[];
  for await (const chunk of tar.c({cwd:input,gzip:true},["Teafile","main.tea","README.md"]))chunks.push(chunk);
  const bytes=Buffer.concat(chunks), directory="modules/alibabacloud_Fixture_1.0.0";
  const pin={ spec:"alibabacloud:Fixture:*", directory,scope:"alibabacloud",name:"Fixture",version:"1.0.0",url:"https://darabonba-module-prod.oss-cn-zhangjiakou.aliyuncs.com/alibabacloud/Fixture-1.0.0.tar.gz",archiveSHA1:sha(bytes,"sha1"),archiveSHA256:sha(bytes),license:{spdx:"Apache-2.0",file:directory+"/README.md",evidence:"archive-declaration"} };
  const before=verifySources(root);
  return {root,bytes,plan:{schemaVersion:1,sourceManifestSHA256:before.hash,modules:[pin]},before};
}

test("explicit extension preserves existing pins and every transitive product map", async t => {
  const {root,bytes,plan,before}=await fixture(t);
  const prepared=await prepareModules(root,plan,async pin => {pin.directory="mutated-loader-input";return bytes});
  assert.equal(verifySources(root).hash,before.hash,"preflight must not write");
  assert.deepEqual(prepared.manifest.modules.filter(m=>m.name!=="Fixture"),before.manifest.modules,"existing versions/provenance are unchanged");
  writeModules(root,prepared);
  const after=verifySources(root);
  assert.equal(after.manifest.modules.length,before.manifest.modules.length+1);
  for(const product of Object.keys({...after.manifest.products,...after.manifest.stagedProducts})){
    const libraries=JSON.parse(fs.readFileSync(path.join(root,"products",product,".libraries.json")));
    assert.equal(Object.keys(libraries).length,after.manifest.modules.length);
    assert.equal(libraries[plan.modules[0].spec],"../../"+plan.modules[0].directory);
  }
  assert.equal(fs.readFileSync(path.join(root,plan.modules[0].directory,"README.md"),"utf8"),"Original fixture license declaration: Apache-2.0\n");
});

test("invalid plans and archive checksums leave the corpus unchanged", async t => {
  const {root,bytes,plan,before}=await fixture(t);
  for(const change of [
    p=>{p.sourceManifestSHA256="0".repeat(64)},
    p=>{p.modules[0].archiveSHA256="0".repeat(64)},
    p=>{p.modules[0].url="https://example.invalid/module.tgz"},
    p=>{p.modules[0].license.spdx="unknown"},
    p=>{p.modules[0].license.file="../README.md"},
    p=>{p.modules[0].version="2.0.0";p.modules[0].directory="modules/alibabacloud_Fixture_2.0.0";p.modules[0].url=p.modules[0].url.replace("1.0.0","2.0.0")},
    p=>{p.modules.push(structuredClone(p.modules[0]))},
    p=>{p.modules[0]=structuredClone(before.manifest.modules[0])},
    p=>{p.modules[0].license={spdx:"Apache-2.0",file:"licenses/unknown.NOTICE",evidence:"module-source-repository"}},
  ]) {
    const candidate=structuredClone(plan);change(candidate);
    await assert.rejects(prepareModules(root,candidate,async()=>bytes));
    assert.equal(verifySources(root).hash,before.hash);
    assert.ok(!fs.existsSync(path.join(root,plan.modules[0].directory)));
  }
});

test("unknown transitive import fails before source writes", async t => {
  const {root,bytes,plan,before}=await fixture(t,{ Unknown:"alibabacloud:Unknown:*" });
  await assert.rejects(prepareModules(root,plan,async()=>bytes),/unresolved module import/);
  assert.equal(verifySources(root).hash,before.hash);
});

test("corpus drift after asynchronous preflight prevents writes", async t => {
  const {root,bytes,plan}=await fixture(t);
  const prepared=await prepareModules(root,plan,async()=>bytes);
  const file=path.join(root,"manifest.json");
  fs.appendFileSync(file,"\n");
  const changed=verifySources(root).hash;
  assert.throws(()=>writeModules(root,prepared),/source corpus changed after preflight/);
  assert.equal(verifySources(root).hash,changed);
  assert.ok(!fs.existsSync(path.join(root,plan.modules[0].directory)));
});

test("archive paths, duplicate names and links are rejected", async () => {
  // Build tar headers directly so the writer cannot normalize unsafe paths away.
  const archive=(entries)=>Buffer.concat(entries.flatMap(({name,type="File",body=""})=>{
    const bytes=Buffer.from(body),header=new tar.Header({path:name,type,linkpath:type==="SymbolicLink"?"target":undefined,mode:0o644,size:bytes.length,mtime:new Date(0)});
    header.encode();return [header.block,bytes,Buffer.alloc((512-bytes.length%512)%512)];
  }).concat([Buffer.alloc(1024)]));
  for(const entries of [
    [{name:"../escape"}], [{name:"C:/escape"}], [{name:"safe\\escape"}], [{name:"con.txt"}], [{name:"trailing."}],
    [{name:"README.md"},{name:"readme.md"}], [{name:"link",type:"SymbolicLink"}],
    [{name:"folder"},{name:"folder/file"}], [{name:"folder/file"},{name:"folder"}],
  ])await assert.rejects(archiveFiles(archive(entries)),/unsafe archive path|duplicate archive path|unsupported archive entry|archive file\/directory conflict/,JSON.stringify(entries));
  assert.equal((await archiveFiles(archive([{name:"./",type:"Directory"},{name:"main.tea",body:"original"}]))).get("main.tea").toString(),"original");
  await assert.rejects(archiveFiles(Buffer.alloc((8<<20)+1)),/archive exceeds limit/);
});
