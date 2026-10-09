"use strict";
const {test}=require("node:test"),assert=require("node:assert/strict"),fs=require("node:fs"),path=require("node:path"),parser=require("@darabonba/parser");
const {buildProduct}=require("./discovery.cjs");
const main=path.resolve(__dirname,"../../sources/darabonba/products/sts/main.tea");
const ast=()=>parser.parse(fs.readFileSync(main,"utf8"),main);
const build=(a)=>buildProduct(a,{pkg:"sts",identifier:"sts-20150401",file:"products/sts/main.tea",provenance:{fixture:true}}).ir;
test("endpoint projection preserves source coordinates and exact map precedence",()=>{
  const ep=build(ast()).endpoints;assert.equal(ep.profile,"map-before-regional-v1");assert.equal(ep.rule,"regional");assert.equal(ep.productCode,"sts");assert.equal(ep.overrides.length,32);assert.ok(ep.overrides.every(r=>r.source.line>0&&r.url==="https://sts.aliyuncs.com"));assert.equal(ep.source.line,14);assert.equal(ep.resolverSource.line,54);
});
test("changed initialization or endpoint precedence fails projection before writes",()=>{
  for(const mutate of [
    a=>a.moduleBody.nodes.find(n=>n.type==="init").initBody.stmts.push({type:"assign"}),
    a=>a.moduleBody.nodes.find(n=>n.type==="init").initBody.stmts[1].expr.value.string="dynamic",
    a=>a.moduleBody.nodes.find(n=>n.type==="init").initBody.stmts[2].expr.fields[0].expr.type="call",
    a=>a.moduleBody.nodes.find(n=>n.type==="init").initBody.stmts[4].expr.args[3].vid.lexeme="@suffix",
    a=>a.moduleBody.nodes.find(n=>n.functionName?.lexeme==="getEndpoint").functionBody.stmts.stmts.reverse(),
    a=>a.moduleBody.nodes.find(n=>n.functionName?.lexeme==="getEndpoint").functionBody.stmts.stmts[0].condition.type="call",
  ]){const a=ast();mutate(a);assert.throws(()=>build(a),/discovery: endpoint/)}
});
