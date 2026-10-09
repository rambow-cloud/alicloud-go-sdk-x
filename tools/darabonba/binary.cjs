"use strict";

// Reviewed header-model and binary result programs, independent of action names.
function lowerHeaderProgram(statements, index, model, helpers) {
  const {check,lex,variable,property,constant,call,wire,reviewedAttributes}=helpers;
  check(model && !model.extendOn,"header model");
  const fields=new Map(model.modelBody.nodes.map(f=>[lex(f.fieldName),f]));
  for(const f of fields.values()){reviewedAttributes(f,"ROA unreviewed header attribute");check(!f.required,"required header model member");}
  const init=statements[index++];
  check(init?.type==="declare"&&lex(init.id)==="realHeaders"&&init.expr.type==="object"&&!init.expr.fields.length,"header map initialization");
  const bindings=[];let common=false;
  while(statements[index]?.type==="if") {
    const guard=statements[index++],name=property(guard.condition?.expr?.args?.[0],"headers"),field=fields.get(name);
    check(field&&guard.condition.type==="not"&&call(guard.condition.expr,"Util","isUnset",e=>property(e,"headers")===name)&&!guard.elseIfs.length&&!guard.elseStmts&&guard.stmts.stmts.length===1&&!bindings.some(b=>b.field===name),"header guard");
    const assignment=guard.stmts.stmts[0];check(assignment.type==="assign","header assignment");
    if(assignment.left.type==="variable") {
      const type=field.fieldValue;
      check(!common&&!bindings.length&&variable(assignment.left,"realHeaders")&&property(assignment.expr,"headers")===name&&type.fieldType==="map"&&lex(type.keyType)==="string"&&lex(type.valueType)==="string","common headers");
      common=true;bindings.push({field:name,wire:wire(field),location:"headers"});
    } else {
      check(common&&field.fieldValue.fieldType==="string"&&assignment.left.type==="map_access"&&lex(assignment.left.id)==="realHeaders"&&constant(assignment.left.accessKey)===wire(field)&&call(assignment.expr,"Util","toJSONString",e=>property(e,"headers")===name),"string header conversion");
      bindings.push({field:name,wire:wire(field),location:"header"});
    }
  }
  check(common&&bindings.length===fields.size&&new Set(bindings.filter(b=>b.location==="header").map(b=>b.wire.toLowerCase())).size===bindings.length-1,"unbound or duplicate header model");
  return {index,bindings};
}

function lowerBinaryResult(statements,index,response,helpers) {
  const {check,lex,variable,property,call,wire,reviewedAttributes}=helpers;
  const fields=new Map(response.modelBody.nodes.map(f=>[wire(f),f]));
  check(!response.extendOn&&fields.size===3&&fields.get("body")?.fieldValue.fieldType==="readable"&&fields.get("statusCode")?.fieldValue.fieldType==="int32","binary response shape");
  const header=fields.get("headers")?.fieldValue;
  check(header?.fieldType==="map"&&lex(header.keyType)==="string"&&lex(header.valueType)==="string","binary response headers");
  for(const f of fields.values()){reviewedAttributes(f,"ROA unreviewed response attribute");check(!f.required,"required binary response member");}
  const result=statements[index++];
  check(result?.type==="declare"&&lex(result.id)==="res"&&result.expr.type==="construct_model"&&lex(result.expr.aliasId)===lex(response.modelName)&&!result.expr.propertyPath.length&&!result.expr.object.fields.length,"binary response construction");
  const tmp=statements[index++];
  check(tmp?.type==="declare"&&lex(tmp.id)==="tmp"&&call(tmp.expr,"Util","assertAsMap",e=>e?.type==="call"&&e.left.type==="method_call"&&lex(e.left.id)==="callApi"&&e.args.length===3&&e.args.every((arg,i)=>variable(arg,["params","req","runtime"][i]))),"binary runtime handoff");
  const seen=new Set();
  while(statements[index]?.type==="if") {
    const guard=statements[index++],name=property(guard.condition?.expr?.args?.[0],"tmp");
    check(fields.has(name)&&!seen.has(name)&&guard.condition.type==="not"&&call(guard.condition.expr,"Util","isUnset",e=>property(e,"tmp")===name)&&!guard.elseIfs.length&&!guard.elseStmts&&guard.stmts.stmts.length===2,"binary response guard");seen.add(name);
    const [conversion,assignment]=guard.stmts.stmts,helper={body:"assertAsReadable",headers:"assertAsMap",statusCode:"assertAsInteger"}[name];
    check(conversion.type==="declare"&&call(conversion.expr,"Util",helper,e=>property(e,"tmp")===name)&&assignment.type==="assign"&&assignment.left.type==="property"&&property({...assignment.left,type:"property_access"},"res")===lex(fields.get(name).fieldName),"binary response assignment");
    const value=lex(conversion.id);
    check(name==="headers"?call(assignment.expr,"Util","stringifyMapValue",e=>variable(e,value)):variable(assignment.expr,value),"binary response conversion");
  }
  const end=statements[index++];check(seen.size===3&&end?.type==="return"&&variable(end.expr,"res")&&index===statements.length,"binary trailing statements");
}

module.exports={lowerHeaderProgram,lowerBinaryResult};
