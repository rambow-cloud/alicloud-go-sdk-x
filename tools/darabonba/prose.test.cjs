"use strict";
const { test } = require("node:test"), assert = require("node:assert/strict"), fs = require("node:fs"), path = require("node:path"), os = require("node:os");
const { annotate, nativeShape, matchProse, project, run } = require("./prose.cjs");
const { requestShape, responseShape, indexedLeaf } = require("./normalization.cjs");
const root=path.resolve(__dirname,"../..");
function fixture() {
 const field=(dslName,wireName,type,documentation=[])=>({dslName,wireName,type,documentation,source:{file:"main.tea",line:1,column:1,endLine:1,endColumn:2}});
 const scalar={kind:"scalar",dslType:"string",wireType:"string"};
 return nativeShape({kind:"model",ref:"Root"},new Map([
  ["Root",{id:"Root",fields:[field("items","Items",{kind:"model",ref:"Wrapper"}),field("known","Known",scalar,[{attribute:"description",text:"Existing English"}])]}],
  ["Wrapper",{id:"Wrapper",fields:[field("item","Item",{kind:"array",items:{kind:"model",ref:"Item"}})]}],
  ["Item",{id:"Item",fields:[field("id","Id",scalar)]}]
 ]));
}
test("itemName and indexed input prose bind exact native fields and source pointers",()=>{
 const raw={type:"object",properties:{Items:{type:"array",itemName:"Item",items:{type:"object",properties:{Id:{type:"string",description_en:"The item identifier."}}}}}};
 annotate(raw,"/responses/200/schema");const out=new Map(),reasons=[];
 matchProse(responseShape(raw),fixture(),{file:"fixture.json"},out,reasons);
 assert.equal(out.get("Item#id").pointer,"/responses/200/schema/properties/Items/items/properties/Id/description_en");assert.equal(reasons.length,0);
 const parameter={type:"string",raw_name:"Items.Item.1.Id",help_en:"The requested identifier."};annotate(parameter,"/parameters/0");
 const indexed=new Map();matchProse(requestShape(parameter),indexedLeaf(fixture(),parameter.raw_name),{file:"fixture.json"},indexed,[]);
 assert.equal(indexed.get("Item#id").pointer,"/parameters/0/help_en");
});
test("existing prose, wrong case/type, missing English and conflicting mappings do not invent descriptions",()=>{
 const raw={type:"object",properties:{Known:{type:"string",description_en:"Replacement."},items:{type:"string",description_en:"Wrong case."},Items:{type:"integer",description_en:"Wrong type."}}};annotate(raw);const out=new Map(),reasons=[];
 matchProse(responseShape(raw),fixture(),{file:"a.json"},out,reasons);assert.equal(out.size,0);assert.equal(reasons.length,2);
 const native=indexedLeaf(fixture(),"Items.Item.1.Id");
 for(const text of ["First meaning.","Different meaning."]){const node={type:"string",help_en:text};annotate(node,"/parameters/0");matchProse(requestShape(node),native,{file:"a.json"},out,reasons);}
 assert.equal(out.get("Item#id").conflict,true);
 const missing=new Map();matchProse(requestShape({type:"string",help_zh:"只有中文"}),native,{file:"a.json"},missing,[]);assert.equal(missing.size,0);
});
test("real corpus projects deterministically and keeps missing mappings explicit",()=>{
 const first=project(root),second=project(root);assert.deepEqual(first,second);
 const counts={sts:11,ecs:253,vpc:11};for(const p of Object.keys(counts)){const d=JSON.parse(first["metadata/prose/"+p+".json"]);assert.equal(Object.keys(d.fields).length,counts[p]);assert.ok(d.matchedOperations>0);assert.match(d.irSHA256,/^[a-f0-9]{64}$/);}
 const ecs=JSON.parse(first["metadata/prose/ecs.json"]);assert.equal(ecs.reasons.filter(r=>r.reason==="metadata-absent-at-pinned-revision").length,9);
});
test("corpus absence does not require operation or field configuration",()=>{
 const temp=fs.mkdtempSync(path.join(os.tmpdir(),"sdk-prose-absence-"));assert.deepEqual(project(temp),{});assert.deepEqual(run(temp,true),{});
});
test("source drift fails before replacing existing projections",()=>{
 const temp=fs.mkdtempSync(path.join(os.tmpdir(),"sdk-prose-drift-")),corpus=path.join(temp,"sources/openapi-meta/prose");
 const manifest=JSON.parse(fs.readFileSync(path.join(root,"sources/openapi-meta/prose/manifest.json")));manifest.files=manifest.files.filter(f=>f.file==="LICENSE"||f.file==="canonical/sts/2015-04-01/AssumeRole.json");
 for(const pin of manifest.files){const dest=path.join(corpus,pin.file);fs.mkdirSync(path.dirname(dest),{recursive:true});fs.copyFileSync(path.join(root,"sources/openapi-meta/prose",pin.file),dest);}
 fs.writeFileSync(path.join(corpus,"manifest.json"),JSON.stringify(manifest));fs.mkdirSync(path.join(temp,"models/sts"),{recursive:true});fs.copyFileSync(path.join(root,"models/sts/ir.json"),path.join(temp,"models/sts/ir.json"));
 run(temp,false);const target=path.join(temp,"metadata/prose/sts.json"),before=fs.readFileSync(target);
 fs.writeFileSync(path.join(corpus,"canonical/sts/2015-04-01/AssumeRole.json"),"modified");assert.throws(()=>run(temp,false),/checksum mismatch/);assert.ok(fs.readFileSync(target).equals(before));
});
