"use strict";
const {test}=require("node:test"),assert=require("node:assert/strict"),fs=require("node:fs"),path=require("node:path"),os=require("node:os"),crypto=require("node:crypto"),parser=require("@darabonba/parser");
const {buildProduct,run}=require("./discovery.cjs");
const {pinnedLibraries}=require("./import-product.cjs");
const root=path.resolve(__dirname,"../.."),file="products/fc/main.tea",main=path.join(root,"sources/darabonba",file),bytes=fs.readFileSync(main);
const decisions=JSON.parse(fs.readFileSync(path.join(root,"metadata/endpoint-source-decisions.json"))).decisions;
const sourceSHA256=crypto.createHash("sha256").update(bytes).digest("hex");
const ast=()=>parser.parse(bytes.toString("utf8"),main);
const build=(a,options={})=>buildProduct(a,{pkg:"fc",identifier:"fc-20230330",file,provenance:{sourceSHA256},endpointDecisions:decisions,...options});
const fn=(a,name)=>a.moduleBody.nodes.find(n=>n.functionName?.lexeme===name[0].toLowerCase()+name.slice(1)+"WithOptions");

test("complete FC discovery preserves native ROA and binary facades",()=>{
  const {ir,coverage}=build(ast());assert.equal(coverage.counts.discovered,73);assert.equal(coverage.counts.lowered,73);assert.equal(coverage.counts.unsupported,0);
  const binary=ir.operations.find(o=>o.name==="InvokeFunction");assert.equal(binary.status,"lowered");assert.equal(binary.protocol.bodyType,"binary");
  assert.equal(binary.bindings.find(b=>b.field==="body").encoding,"bytes");assert.equal(binary.headerBindings.model,"InvokeFunctionHeaders");
  assert.equal(ir.models.find(m=>m.id===binary.roots.request.ref).fields.find(f=>f.dslName==="body").type.kind,"bytes");
  assert.equal(ir.models.find(m=>m.id===binary.roots.body.ref).fields.find(f=>f.dslName==="body").type.kind,"stream");
  const create=ir.operations.find(o=>o.name==="CreateAlias");assert.equal(create.protocol.pathname,"/2023-03-30/functions/{functionName}/aliases");
  assert.deepEqual(create.bindings.map(b=>b.location).sort(),["body","headers","path"]);
  const input=ir.models.find(m=>m.id===create.roots.request.ref),output=ir.models.find(m=>m.id===create.roots.body.ref);
  assert.equal(input.origin,"operation-input-facade");assert.equal(output.origin,"operation-output-facade");assert.ok(input.fields.every(f=>f.source.line>0));
  assert.equal(input.fields.find(f=>f.dslName==="functionName").required,true);
  const list=ir.operations.find(o=>o.name==="ListInstances");assert.equal(list.bindings.find(b=>b.wire==="instanceIds").encoding,"json");
  const network=ir.models.find(m=>m.id==="CreateSessionNetworkConfig");assert.equal(network.fields.find(f=>f.dslName==="rules").type.values.kind,"array");
  assert.equal(ir.operations.filter(o=>o.status==="lowered"&&o.protocol.bodyType==="none").length,18);
});

test("renamed ROA product and operation use the same semantic lowering",()=>{
  const a=ast();const operation=fn(a,"GetAlias");operation.functionName.lexeme="inspectAliasWithOptions";
  const protocol=operation.functionBody.stmts.stmts.find(s=>s.id?.lexeme==="params");protocol.expr.object.fields.find(f=>f.fieldName.lexeme==="action").expr.value.string="InspectAlias";
  const {ir}=build(a,{pkg:"computecheck"});const lowered=ir.operations.find(o=>o.name==="InspectAlias");assert.equal(lowered.status,"lowered");assert.equal(lowered.protocol.pathname,"/2023-03-30/functions/{functionName}/aliases/{aliasName}");assert.equal(ir.product,"computecheck");
});

test("unknown ROA statements, transforms and guards stay explicitly unsupported",()=>{
  for(const [action,mutate]of [
    ["GetAlias",f=>f.functionBody.stmts.stmts.push({type:"return",expr:{type:"string",value:{string:"unexpected"}}})],
    ["ListInstances",f=>f.functionBody.stmts.stmts.find(s=>s.type==="if").stmts.stmts[0].expr.args[2].value.string="indexed"],
    ["DisableFunctionInvocation",f=>f.functionBody.stmts.stmts.find(s=>s.type==="if").condition.type="call"],
  ]){const a=ast();mutate(fn(a,action));const operation=build(a).ir.operations.find(o=>o.name===action);assert.equal(operation.status,"unsupported",action);assert.ok(operation.reasons.length>0);}
});

test("endpoint normalization is exact, source-bound and leaves official bytes untouched",()=>{
  const accepted=build(ast()).ir.endpoints.overrides.find(o=>o.region==="ap-southeast-5");assert.equal(accepted.url,"https://fcv3.ap-southeast-5.aliyuncs.com");assert.deepEqual(accepted.normalization,decisions[0]);
  for(const endpointDecisions of [[],[...decisions,...decisions],[{...decisions[0],line:22}],[{...decisions[0],sourceSHA256:"0".repeat(64)}],[{...decisions[0],after:"another.aliyuncs.com"}]])assert.throws(()=>build(ast(),{endpointDecisions}),/endpoint/);
  assert.deepEqual(fs.readFileSync(main),bytes);
});

test("unknown selections and invalid endpoint approvals fail before IR writes",()=>{
  const temp=fs.mkdtempSync(path.join(os.tmpdir(),"sdk-roa-"));try{
    fs.cpSync(path.join(root,"sources/darabonba"),path.join(temp,"sources/darabonba"),{recursive:true});fs.mkdirSync(path.join(temp,"metadata"));
    const approval=path.join(temp,"metadata/endpoint-source-decisions.json");fs.copyFileSync(path.join(root,"metadata/endpoint-source-decisions.json"),approval);
    assert.throws(()=>run("generate",{root:temp,selected:["fc/UnknownBinary"]}),/unknown|selected/);assert.equal(fs.existsSync(path.join(temp,"models")),false);
    fs.writeFileSync(approval,JSON.stringify({schemaVersion:1,decisions:[{...decisions[0],line:22}]}));
    assert.throws(()=>run("generate",{root:temp}),/endpoint/);assert.equal(fs.existsSync(path.join(temp,"models")),false);
  }finally{fs.rmSync(temp,{recursive:true,force:true});}
});

test("binary ROA programs reuse lowering after product/action renames",()=>{
  const a=ast(),operation=fn(a,"InvokeFunction");operation.functionName.lexeme="executePayloadWithOptions";
  operation.functionBody.stmts.stmts.find(s=>s.id?.lexeme==="params").expr.object.fields.find(f=>f.fieldName.lexeme==="action").expr.value.string="ExecutePayload";
  const renamed=build(a,{pkg:"computecheck"}).ir.operations.find(o=>o.name==="ExecutePayload");assert.equal(renamed.status,"lowered");assert.equal(renamed.protocol.bodyType,"binary");assert.equal(renamed.headerBindings.fields.length,4);
});

test("selected changed binary source fails before IR writes",()=>{
  const temp=fs.mkdtempSync(path.join(os.tmpdir(),"sdk-binary-drift-"));try {
    fs.cpSync(path.join(root,"sources/darabonba"),path.join(temp,"sources/darabonba"),{recursive:true});fs.mkdirSync(path.join(temp,"metadata"));
    const sourceFile=path.join(temp,"sources/darabonba",file),changed=fs.readFileSync(sourceFile,"utf8").replace("stream = request.body,","");
    const changedHash=crypto.createHash("sha256").update(changed).digest("hex");fs.writeFileSync(sourceFile,changed);
    const manifestFile=path.join(temp,"sources/darabonba/manifest.json"),manifest=JSON.parse(fs.readFileSync(manifestFile));manifest.files.find(f=>f.file===file).sha256=changedHash;fs.writeFileSync(manifestFile,JSON.stringify(manifest));
    fs.writeFileSync(path.join(temp,"metadata/endpoint-source-decisions.json"),JSON.stringify({schemaVersion:1,decisions:decisions.map(d=>({...d,sourceSHA256:changedHash}))}));
    assert.throws(()=>run("generate",{root:temp,selected:["fc/InvokeFunction"]}),/unsupported/);assert.equal(fs.existsSync(path.join(temp,"models")),false);
  }finally{fs.rmSync(temp,{recursive:true,force:true});}
});

test("binary programs reject changed body, headers, conversions and signature algorithms",()=>{
  for(const mutate of [
    f=>f.functionBody.stmts.stmts.find(s=>s.id?.lexeme==="req").expr.object.fields.find(f=>f.fieldName.lexeme==="stream").expr.propertyPath[0].lexeme="qualifier",
    f=>f.functionBody.stmts.stmts.find(s=>s.type==="if"&&s.stmts.stmts[0]?.expr?.left?.propertyPath?.[0]?.lexeme==="toJSONString").stmts.stmts[0].expr.left.propertyPath[0].lexeme="unknown",
    f=>f.functionBody.stmts.stmts.find(s=>s.type==="if"&&s.stmts.stmts[0]?.expr?.left?.propertyPath?.[0]?.lexeme==="assertAsReadable").stmts.stmts[0].expr.left.propertyPath[0].lexeme="assertAsString",
    f=>f.functionBody.stmts.stmts.push({type:"return",expr:{type:"variable",id:{lexeme:"res"}}}),
  ]){const a=ast();mutate(fn(a,"InvokeFunction"));assert.equal(build(a).ir.operations.find(o=>o.name==="InvokeFunction").status,"unsupported");}
  const a=ast(),init=a.moduleBody.nodes.find(n=>n.type==="init");init.initBody.stmts.push({type:"assign",left:{type:"virtualVariable",vid:{lexeme:"@signatureAlgorithm"}},expr:{type:"string",value:{string:"v2"}}});
  assert.equal(build(a).ir.operations.find(o=>o.name==="InvokeFunction").status,"unsupported");
});

test("product import retains all pinned transitive modules without wildcard resolution",()=>{
  const manifest=JSON.parse(fs.readFileSync(path.join(root,"sources/darabonba/manifest.json"))),meta=JSON.parse(fs.readFileSync(path.join(root,"sources/darabonba/products/fc/Teafile")));
  const libraries=JSON.parse(pinnedLibraries(manifest,meta));assert.equal(Object.keys(libraries).length,manifest.modules.length);assert.ok(Object.keys(libraries).some(k=>k.startsWith("alibabacloud:Credential:")));
  assert.throws(()=>pinnedLibraries(manifest,{libraries:{Unknown:"unknown:Missing:*"}}),/not already pinned/);
});
