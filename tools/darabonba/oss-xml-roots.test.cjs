"use strict";

const test = require("node:test"), assert = require("node:assert/strict");
const { compareBody, compareProduct } = require("./oss-xml-roots.cjs");
const parser = require("@darabonba/parser"), fs = require("node:fs"), path = require("node:path");
const field = name => ({ jsonName:name, xmlName:name, type:{kind:"scalar",name:"string"} });

test("renamed explicit roots normalize structured wrappers and preserve scalar roots", () => {
  const models = new Map([["DifferentNativeName", {fields:[field("Owner"),field("Grant"),field("Extra")]}]]);
  const root = { kind:"structured", field:{...field("RenamedEnvelope"),type:{kind:"model",name:"DifferentNativeName"}} };
  const body = { fields:[{wireName:"Owner"},{wireName:"Grant"}] };
  assert.deepEqual(compareBody(body,root,models), {status:"root-fields-match",normalization:"unwrap-structured-root",missingNativeFields:[],extraNativeFields:["Extra"]});
  assert.equal(compareBody({fields:[{wireName:"RenamedEnvelope"}]},root,models).normalization,"preserve-wrapper");
  const scalar = {kind:"scalar",field:field("Text")};
  assert.equal(compareBody({fields:[{wireName:"Text"}]},scalar,models).normalization,"scalar-field");
  assert.equal(compareBody({fields:[{wireName:"Guessed"}]},scalar,models).status,"divergent");
  assert.equal(compareBody(body,null,models).reason,"native registry action missing");
  assert.equal(compareBody(body,{kind:"unsupported",reason:"ambiguous root"},models).reason,"ambiguous root");
  const before=JSON.stringify(root);compareBody(body,root,models);assert.equal(JSON.stringify(root),before);
});

test("wire conflicts are reported without guessed names or silent correction", () => {
  const body={fields:[{wireName:"Absent"}]},root={kind:"structured",field:{...field("Root"),type:{kind:"model",name:"Nested"}}};
  assert.equal(compareBody(body,root,new Map()).status,"unsupported");
  assert.deepEqual(compareBody(body,root,new Map([["Nested",{fields:[field("Present")]}]])).missingNativeFields,["Absent"]);
  for(const fields of [[field("Same"),field("Same")],[{...field("A"),xmlName:"Other"}],[{...field("A"),xmlOptions:["attr"]}]]) {
    assert.equal(compareBody(body,root,new Map([["Nested",{fields}]])).status,"unsupported");
  }
});

test("official parser discovery uses explicit XML/action evidence across the complete DSL", () => {
  const file=path.resolve(__dirname,"../../sources/darabonba/products/sts/main.tea"),ast=parser.parse(fs.readFileSync(file,"utf8"),file);
  // All STS declarations are JSON: the root reader never invents XML behavior.
  const report=compareProduct(ast,"products/sts/main.tea",{schemaVersion:1,models:[],roots:[]});
  assert.equal(report.counts.discovered,4);assert.equal(report.counts.xmlResponses,0);
  assert.deepEqual(report.operations,[]);
  assert.throws(()=>compareProduct(ast,"fixture",{schemaVersion:1,models:[],roots:[{action:"A"},{action:"A"}]}),/Duplicate/);
});

test("complete parsed DSL XML declarations retain absent roots and facade boundaries", () => {
  const file=path.resolve(__dirname,"../../sources/darabonba/products/sts/main.tea");
  const source=fs.readFileSync(file,"utf8").replaceAll("bodyType = 'json'","bodyType = 'xml'");
  const ast=parser.parse(source,file);
  const inventory={schemaVersion:1,models:[],roots:[]};
  const report=compareProduct(ast,"fixture/main.tea",inventory);
  assert.equal(report.counts.xmlResponses,4);assert.equal(report.counts.unsupported,4);
  assert.ok(report.operations.every(o=>o.reason==="native registry action missing" && o.bodyModel && o.source.file==="fixture/main.tea"));
  // An explicitly declared scalar root can be compared without naming conventions.
  inventory.roots.push({action:"GetCallerIdentity",kind:"scalar",field:field("AccountId")});
  const next=compareProduct(ast,"fixture/main.tea",inventory);
  const identity=next.operations.find(o=>o.action==="GetCallerIdentity");
  assert.equal(identity.status,"divergent");assert.ok(identity.missingNativeFields.includes("RequestId"));
  assert.ok(identity.extraNativeFields.length===0);
});
