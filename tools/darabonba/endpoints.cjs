"use strict";

// Recognize the pinned endpoint initialization through the official semantic AST.
const lex=(t)=>t?.lexeme;
const {gatewayProduct,validateGateway}=require("./oss.cjs");
const variable=(name)=>({type:"variable",id:{lexeme:name,type:"variable"}});
const call=(module,name,args)=>({type:"call",left:{type:"static_call",id:{lexeme:module,type:"module"},propertyPath:[{lexeme:name}]},args});
const not=(expr)=>({type:"not",expr});
const returned=(expr)=>({type:"stmts",stmts:[{type:"return",expr}]});
const mapAccess={type:"map_access",id:{lexeme:"endpointMap",type:"variable"},accessKey:variable("regionId")};
const expected={type:"stmts",stmts:[
  {type:"if",condition:not(call("Util","empty",[variable("endpoint")])),stmts:returned(variable("endpoint")),elseIfs:[]},
  {type:"if",condition:{type:"and",left:not(call("Util","isUnset",[variable("endpointMap")])),right:not(call("Util","empty",[mapAccess]))},stmts:returned(mapAccess),elseIfs:[]},
  {type:"return",expr:call("EndpointUtil","getEndpointRules",["productId","regionId","endpointRule","network","suffix"].map(variable))},
]};
const semanticKeys=new Set(["type","left","right","expr","condition","stmts","args","id","vid","lexeme","string","value","propertyPath","accessKey","op","elseStmts","elseIfs"]);
function syntax(value){
  if(Array.isArray(value))return value.map(syntax);
  if(value&&typeof value==="object")return Object.fromEntries(Object.entries(value).filter(([k])=>semanticKeys.has(k)).sort(([a],[b])=>a.localeCompare(b)).map(([k,v])=>[k,syntax(v)]));
  return value;
}
function projectEndpoints(ast,file,source,decisions=[],sourceSHA256=""){
  const fail=(ok,message)=>{if(!ok)throw Error("discovery: endpoint "+message)};
  if(gatewayProduct(ast)) {
    validateGateway(ast,(ok,reason)=>fail(ok,reason));
    fail(!decisions.length,"Gateway endpoint decisions are not supported");
    const initializer=ast.moduleBody.nodes.find(n=>n.type==="init");
    return {profile:"gateway-explicit-origin-v1",source:source(file,initializer.initBody.stmts[3].left.vid),overrides:[]};
  }
  const initializers=ast.moduleBody.nodes.filter(n=>n.type==="init");fail(initializers.length===1,"initializer must be unique");
  // Signing declarations are validated separately by operation lowering. A
  // constant initializer does not alter endpoint flow, including historical v2.
  const statements=initializers[0].initBody.stmts.filter(s=>!(s.type==="assign"&&lex(s.left?.vid)==="@signatureAlgorithm"&&s.expr?.type==="string"));
  fail(statements.length===5&&statements[0].type==="super"&&statements[3].type==="call"&&lex(statements[3].left?.id)==="checkConfig","unsupported initializer control flow");
  const assignment=(name)=>{const found=statements.filter(s=>s.type==="assign"&&lex(s.left?.vid)===name);fail(found.length===1,"assignment must be unique: "+name);return found[0]};
  const rule=assignment("@endpointRule"),map=assignment("@endpointMap"),endpoint=assignment("@endpoint");
  fail(rule.expr.type==="string"&&["regional","global"].includes(rule.expr.value.string),"unsupported deployment rule");
  fail(map.expr.type==="object","map must be constant");
  const seen=new Set(),overrides=map.expr.fields.map(field=>{
    const region=field.fieldName?.string;fail(typeof region==="string"&&!seen.has(region)&&field.expr?.type==="string","nonconstant or duplicate map entry");seen.add(region);
    let host=field.expr.value.string;
    const location=source(file,field.fieldName),decision=decisions.find(d=>d.region===region);
    if(decision){
      fail(decision.sourceSHA256===sourceSHA256&&decision.file===file&&decision.line===location.line&&decision.before===host&&decision.after===host.trimEnd()&&decision.before!==decision.after&&decision.reason==="trim-reviewed-trailing-whitespace","source-bound normalization differs");
      host=decision.after;
    }
    fail(host.split('.').every(label=>/^[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?$/.test(label))&&host.endsWith(".aliyuncs.com"),"invalid mapped hostname");
    return {region,url:"https://"+host,source:location,...(decision?{normalization:decision}:{})};
  }).sort((a,b)=>a.region.localeCompare(b.region,"en"));
  fail(decisions.every(d=>d.file===file&&overrides.some(o=>o.region===d.region&&o.normalization===d))&&new Set(decisions.map(d=>d.region)).size===decisions.length,"unused or duplicate normalization");
  const expression=endpoint.expr;
  fail(expression.type==="call"&&lex(expression.left?.id)==="getEndpoint"&&expression.args?.length===7&&expression.args[0].type==="string","unsupported endpoint handoff");
  fail(expression.args.slice(1).every((arg,i)=>arg.type==="virtualVariable"&&lex(arg.vid)===["@regionId","@endpointRule","@network","@suffix","@endpointMap","@endpoint"][i]),"endpoint handoff arguments differ");
  const getter=ast.moduleBody.nodes.filter(n=>lex(n.functionName)==="getEndpoint");
  fail(getter.length===1&&getter[0].params?.params?.length===7&&getter[0].params.params.every((param,i)=>lex(param.paramName)===["productId","regionId","endpointRule","network","suffix","endpointMap","endpoint"][i]),"resolver parameters differ");
  fail(getter.length===1&&JSON.stringify(syntax(getter[0].functionBody?.stmts))===JSON.stringify(syntax(expected)),"resolver precedence is unsupported");
  return {profile:"map-before-regional-v1",productCode:expression.args[0].value.string,rule:rule.expr.value.string,source:source(file,rule.left.vid),resolverSource:source(file,getter[0].functionName),overrides};
}
module.exports={projectEndpoints};
