"use strict";

// Extend a pinned source corpus without resolving wildcard imports again.
const fs=require("node:fs"),path=require("node:path"),crypto=require("node:crypto");
const {execFileSync}=require("node:child_process");
const {verifySources}=require("./frontend.cjs");
const root=path.resolve(__dirname,"../../sources/darabonba");
const sha=b=>crypto.createHash("sha256").update(b).digest("hex");
function pinnedLibraries(manifest,meta){
  for(const spec of Object.values(meta.libraries||{}))if(!manifest.modules.some(m=>m.spec===spec))throw Error("import-product: import is not already pinned: "+spec);
  // The semantic parser resolves transitive imports through the product map too.
  return Buffer.from(JSON.stringify(Object.fromEntries(manifest.modules.map(m=>[m.spec,"../../"+m.directory]).sort(([a],[b])=>a.localeCompare(b))),null,2)+"\n");
}
function importProduct(product,identifier){
  if(!/^[a-z][a-z0-9]*$/.test(product)||!/^[a-z][a-z0-9]*-\d{8}$/.test(identifier))throw Error("import-product: unsafe product identifier");
  const verified=verifySources(root),manifest=structuredClone(verified.manifest);
  if(manifest.products[product]||fs.existsSync(path.join(root,"products",product)))throw Error("import-product: product already exists");
  const files=new Map();
  for(const name of ["main.tea","Teafile","api-info.json"]){
    const remote=`${identifier}/${name}`;
    const result=JSON.parse(execFileSync("gh",["api",`repos/aliyun/alibabacloud-sdk/contents/${remote}?ref=${manifest.revision}`],{encoding:"utf8",windowsHide:true,timeout:40000,maxBuffer:8<<20}));
    if(result.type!=="file"||result.encoding!=="base64"||result.path!==remote)throw Error("import-product: unexpected source response");
    const bytes=Buffer.from(result.content.replace(/\s/g,""),"base64");
    const blob=crypto.createHash("sha1").update(Buffer.from(`blob ${bytes.length}\0`)).update(bytes).digest("hex");
    if(blob!==result.sha)throw Error("import-product: git blob checksum differs");
    const file=`products/${product}/${name}`;files.set(file,bytes);
    manifest.files.push({file,url:`https://raw.githubusercontent.com/aliyun/alibabacloud-sdk/${manifest.revision}/${remote}`,sha256:sha(bytes)});
  }
  const meta=JSON.parse(files.get(`products/${product}/Teafile`));
  const libraries=`products/${product}/.libraries.json`,bytes=pinnedLibraries(manifest,meta);
  files.set(libraries,bytes);manifest.files.push({file:libraries,url:"local:pinned-import-projection",sha256:sha(bytes)});
  manifest.products[product]=identifier;manifest.products=Object.fromEntries(Object.entries(manifest.products).sort(([a],[b])=>a.localeCompare(b)));manifest.files.sort((a,b)=>a.file.localeCompare(b.file));
  // Network, source and import preflight precede every write.
  for(const [file,bytes]of files){fs.mkdirSync(path.dirname(path.join(root,file)),{recursive:true});fs.writeFileSync(path.join(root,file),bytes);}
  fs.writeFileSync(path.join(root,"manifest.json"),JSON.stringify(manifest,null,2)+"\n");
  console.log(`Imported ${product} at ${manifest.revision}; existing imports remain pinned. Regenerate/review IR and capability source bindings before emission.`);
}
if(require.main===module){try{if(process.argv.length!==4)throw Error("import-product: use <product> <official-directory>");importProduct(process.argv[2],process.argv[3]);}catch(e){console.error(e.message);process.exitCode=1;}}
module.exports={importProduct,pinnedLibraries};
