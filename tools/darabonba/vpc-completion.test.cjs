"use strict";
const { test } = require("node:test");
const assert = require("node:assert/strict");
const fs = require("node:fs");
const path = require("node:path");
const parser = require("@darabonba/parser");
const { lowerOperation } = require("./frontend.cjs");
const { buildProduct } = require("./discovery.cjs");
const main = path.resolve(__dirname, "../../sources/darabonba/products/vpc/main.tea");
const ast = parser.parse(fs.readFileSync(main, "utf8"), main);
const fn = (a, name) => a.moduleBody.nodes.find(n => n.functionName?.lexeme === name[0].toLowerCase()+name.slice(1)+"WithOptions");
const statements = (a, name) => fn(a, name).functionBody.stmts.stmts;
const formActions = ["CreateVpnAttachment", "CreateVpnConnection", "ModifyVpnAttachmentAttribute", "ModifyVpnConnectionAttribute"];

test("complete pinned VPC lowers all 403 actions with native locations and methods", () => {
  const { ir, coverage } = buildProduct(ast, { pkg:"vpc", identifier:"vpc-20160428", file:"products/vpc/main.tea", info:JSON.parse(fs.readFileSync(path.join(path.dirname(main), "api-info.json"))), provenance:{fixture:true} });
  assert.equal(ir.schemaVersion, 6);
  assert.equal(coverage.counts.discovered, 403);
  assert.equal(coverage.counts.lowered, 403);
  for (const name of formActions) {
    const op=ir.operations.find(o=>o.name===name);
    assert.equal(op.bindings.find(b=>b.wire==="TunnelOptionsSpecification").location,"form");
    assert.equal(op.bindings.find(b=>b.wire==="RegionId").location,"query");
  }
  assert.equal(ir.operations.find(o=>o.name==="DescribeVpnGatewayAvailableZones").protocol.method,"GET");
  for(const name of ["GrantInstanceToVbr","RevokeInstanceFromVbr"]) {
    const b=ir.operations.find(o=>o.name===name).bindings.find(b=>b.wire==="VbrInstanceIds");
    assert.equal(b.encoding,"simple");
    assert.equal(b.field,"vbrInstanceIds");
  }
});

for(const [label, mutate] of [
  ["body initialization", ss=>ss.find(s=>s.id?.lexeme==="body").expr.fields.push({type:"objectField"})],
  ["body source", ss=>ss.find(s=>s.type==="if"&&s.stmts.stmts[0].left?.id?.lexeme==="bodyFlat").stmts.stmts[0].expr.propertyPath[0].lexeme="regionId"],
  ["body wire alias", ss=>ss.find(s=>s.type==="if"&&s.stmts.stmts[0].left?.id?.lexeme==="bodyFlat").stmts.stmts[0].left.accessKey.value.string="Wrong"],
  ["body merge", ss=>ss.find(s=>s.type==="assign").expr.fields[1].expr.args[0].id.lexeme="query"],
  ["body parse helper", ss=>ss.find(s=>s.id?.lexeme==="req").expr.object.fields[1].expr.left.propertyPath[0].lexeme="query"],
  ["body on GET", ss=>ss.find(s=>s.id?.lexeme==="params").expr.object.fields.find(f=>f.fieldName.lexeme==="method").expr.value.string="GET"],
  ["duplicate location", ss=>{const guard=structuredClone(ss.find(s=>s.type==="if"));guard.stmts.stmts[0].left.id.lexeme="bodyFlat";const i=ss.findIndex(s=>s.type==="assign");ss.splice(i,0,guard);}],
]) test("form program rejects " + label, () => {
  const a=structuredClone(ast);mutate(statements(a,"CreateVpnAttachment"));
  assert.throws(()=>lowerOperation(a,"CreateVpnAttachment"),/unsupported SDK pattern/);
});

test("whole-model query conversion rejects altered source/helper and unsupported method",()=>{
  for(const change of [ss=>ss[1].expr.args[0].left.propertyPath[0].lexeme="stringify",ss=>ss[1].expr.args[0].args[0].id.lexeme="runtime",ss=>ss.find(s=>s.id?.lexeme==="params").expr.object.fields.find(f=>f.fieldName.lexeme==="method").expr.value.string="DELETE"]){
    const a=structuredClone(ast);change(statements(a,"DescribeVpnGatewayAvailableZones"));
    assert.throws(()=>lowerOperation(a,"DescribeVpnGatewayAvailableZones"),/unsupported SDK pattern/);
  }
});

test("simple shrink rejects unknown style and a non-string element",()=>{
  const a=structuredClone(ast);statements(a,"GrantInstanceToVbr")[3].stmts.stmts[0].expr.args[2].value.string="pipeDelimited";
  assert.throws(()=>lowerOperation(a,"GrantInstanceToVbr"),/unsupported SDK pattern/);
  const b=structuredClone(ast);b.moduleBody.nodes.find(n=>n.modelName?.lexeme==="GrantInstanceToVbrRequest").modelBody.nodes.find(f=>f.fieldName.lexeme==="vbrInstanceIds").fieldValue.fieldItemType="int32";
  assert.throws(()=>lowerOperation(b,"GrantInstanceToVbr"),/unsupported SDK pattern/);
});

test("renamed form, GET and simple actions reuse shared lowering",()=>{
  for(const [name, renamed] of [["CreateVpnAttachment","CreateTunnel"],["DescribeVpnGatewayAvailableZones","InspectZones"],["GrantInstanceToVbr","AuthorizeRouters"]]){
    const a=structuredClone(ast);const f=fn(a,name);f.functionName.lexeme=renamed[0].toLowerCase()+renamed.slice(1)+"WithOptions";
    f.functionBody.stmts.stmts.find(s=>s.id?.lexeme==="params").expr.object.fields.find(v=>v.fieldName.lexeme==="action").expr.value.string=renamed;
    assert.equal(lowerOperation(a,renamed).protocol.action,renamed);
  }
});
