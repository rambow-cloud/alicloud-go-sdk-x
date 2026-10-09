"use strict";
const fs=require("node:fs"),path=require("node:path"),crypto=require("node:crypto"),parser=require("@darabonba/parser");
const {buildProduct}=require("./discovery.cjs");

// Parse renamed source in memory. Original pinned bytes remain unchanged.
function binaryReuseFixture(root) {
  const file=path.join(root,"sources/darabonba/products/fc/main.tea");
  const source=fs.readFileSync(file,"utf8").replaceAll("InvokeFunction","ExecutePayload").replaceAll("invokeFunction","executePayload")
    .replace("body?: readable(name='body', description=","payload?: readable(name='body', description=")
    .replace(/body = request\.body,(\r?\n\s*)stream = request\.body,/,"body = request.payload,$1stream = request.payload,").replace("body?: readable(name='body'),","result?: readable(name='body'),")
    .replace("res.body = respBody;","res.result = respBody;").replaceAll("commonHeaders","sharedHeaders");
  const sourceSHA256=crypto.createHash("sha256").update(source).digest("hex");
  const decisions=JSON.parse(fs.readFileSync(path.join(root,"metadata/endpoint-source-decisions.json"))).decisions.map(d=>({...d,file:"products/binaryfixture/main.tea",sourceSHA256}));
  const result=buildProduct(parser.parse(source,file),{pkg:"binaryfixture",identifier:"binaryfixture-20230330",file:"products/binaryfixture/main.tea",endpointDecisions:decisions,provenance:{repository:"synthetic-fixture",revision:"in-memory-test",license:"Apache-2.0",parserVersion:"2.2.1",sourceManifestSHA256:"synthetic-not-for-publication",sourceSHA256}});
  result.ir.operations=result.ir.operations.filter(op=>op.name==="ExecutePayload");
  if(result.ir.operations.length!==1||result.ir.operations[0].status!=="lowered")throw new Error("renamed binary fixture did not lower");
  return result;
}
module.exports={binaryReuseFixture};
if(require.main===module)process.stdout.write(JSON.stringify(binaryReuseFixture(process.argv[2]).ir));
