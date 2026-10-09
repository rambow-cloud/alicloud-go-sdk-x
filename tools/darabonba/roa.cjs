"use strict";

// Recognize reviewed ROA SDK programs through official semantic AST nodes.
const lex = t => t?.lexeme;
const variable = (e, name) => e?.type === "variable" && lex(e.id) === name;
const property = (e, owner = "request") => e?.type === "property_access" && lex(e.id) === owner && e.propertyPath.length === 1 ? lex(e.propertyPath[0]) : null;
const constant = e => e?.type === "string" ? e.value.string : null;
const call = (e, module, name, argument) => e?.type === "call" && e.left.type === "static_call" && lex(e.left.id) === module && e.left.propertyPath.length === 1 && lex(e.left.propertyPath[0]) === name && e.args.length === 1 && argument(e.args[0]);
const construct = (e, name) => e?.type === "construct_model" && lex(e.aliasId) === "OpenApi" && e.propertyPath.length === 1 && lex(e.propertyPath[0]) === name;
function roaStyle(fn) {
  return fn.functionBody.stmts.stmts.some(s => s.type === "declare" && lex(s.id) === "params" && construct(s.expr, "Params") && s.expr.object.fields.some(f => lex(f.fieldName) === "style" && constant(f.expr) === "ROA"));
}
function lowerROA(ast, fn, operation, {shape, reviewedAttributes, requireProfile}) {
  const check = (ok, reason) => requireProfile(ok, "ROA " + reason);
  const declaredBody = fn.functionBody.stmts.stmts.find(s => s.type === "declare" && lex(s.id) === "params")?.expr?.object?.fields?.find(f => lex(f.fieldName) === "bodyType");
  check(["json","none"].includes(constant(declaredBody?.expr)),"unsupported response body "+constant(declaredBody?.expr));
  const params = fn.params.params, headers = params.at(-2), runtime = params.at(-1);
  check(runtime && lex(runtime.paramName) === "runtime" && runtime.paramType.type === "moduleModel" && runtime.paramType.path.map(lex).join(".") === "Util.RuntimeOptions", "runtime signature");
  check(headers && lex(headers.paramName) === "headers" && headers.paramType.type === "map" && lex(headers.paramType.keyType) === "string" && lex(headers.paramType.valueType) === "string", "headers signature");
  const inputParams = params.slice(0,-2), requestParam = inputParams.find(p => ["request","tmpReq"].includes(lex(p.paramName)));
  const paths = inputParams.filter(p => p !== requestParam);
  check(paths.every(p => lex(p.paramType) === "string" && !p.defaultValue) && new Set(inputParams.map(p => lex(p.paramName))).size === inputParams.length, "path signature");
  const models = new Map(ast.moduleBody.nodes.filter(n => n.type === "model").map(n => [lex(n.modelName), n]));
  const request = requestParam ? models.get(lex(requestParam.paramType)) : null;
  check(!requestParam || request && !request.extendOn, "request model");
  const fields = new Map((request?.modelBody.nodes || []).map(f => [lex(f.fieldName), f]));
  const wire = f => f.attrs.find(a => lex(a.attrName) === "name")?.attrValue.string || lex(f.fieldName);
  for(const f of fields.values()) reviewedAttributes(f, "ROA unreviewed input attribute");
  const statements = fn.functionBody.stmts.stmts; let index = 0;
  const inputName=lex(requestParam?.paramName);
  if(request)check(call(statements[index++],"Util","validateModel",e => variable(e,inputName)),"model validation");
  let queryFields=fields;const transforms=new Map();
  if(inputName === "tmpReq") {
    const create=statements[index++];
    check(create?.type === "declare" && lex(create.id) === "request" && create.expr.type === "construct_model" && lex(create.expr.aliasId) === lex(requestParam.paramType).replace(/Request$/,"ShrinkRequest") && !create.expr.propertyPath.length && !create.expr.object.fields.length,"shrink construction");
    const shrink=models.get(lex(create.expr.aliasId));check(shrink && !shrink.extendOn,"shrink model");
    queryFields=new Map(shrink.modelBody.nodes.map(f=>[lex(f.fieldName),f]));
    for(const f of queryFields.values())reviewedAttributes(f,"ROA unreviewed shrink attribute");
    const convert=statements[index++];check(convert?.type === "call" && convert.left.type === "static_call" && lex(convert.left.id) === "OpenApiUtil" && convert.left.propertyPath.map(lex).join(".") === "convert" && convert.args.length === 2 && variable(convert.args[0],"tmpReq") && variable(convert.args[1],"request"),"shrink conversion");
    while(statements[index]?.type === "if") {
      const guard=statements[index++],original=property(guard.condition?.expr?.args?.[0],"tmpReq");
      check(guard.condition?.type === "not" && call(guard.condition.expr,"Util","isUnset",e=>!!property(e,"tmpReq")) && !guard.elseIfs.length && !guard.elseStmts && guard.stmts.stmts.length === 1,"shrink guard");
      const assignment=guard.stmts.stmts[0],target=assignment.left?.type === "property" ? property({...assignment.left,type:"property_access"}) : null,expr=assignment.expr;
      const originalField=fields.get(original),shrinkField=queryFields.get(target);
      check(assignment.type === "assign" && originalField && shrinkField && target === original+"Shrink" && shrinkField.fieldValue.fieldType === "string" && wire(shrinkField) === wire(originalField) && !transforms.has(target) && expr?.type === "call" && expr.left.type === "static_call" && lex(expr.left.id) === "OpenApiUtil" && expr.left.propertyPath.map(lex).join(".") === "arrayToStringWithSpecifiedStyle" && expr.args.length === 3 && property(expr.args[0],"tmpReq") === original && constant(expr.args[1]) === wire(originalField) && constant(expr.args[2]) === "json","shrink JSON transform");
      transforms.set(target,original);
    }
    check(queryFields.size === fields.size && [...queryFields].every(([name,f])=>{const original=fields.get(transforms.get(name)||name);return original && wire(original) === wire(f) && (transforms.has(name)||JSON.stringify(shape(original.fieldValue,models)) === JSON.stringify(shape(f.fieldValue,models)))}),"shrink correspondence");
  }
  const inputs = [], bound = new Set();
  function bind(name, nativeWire, location, encoding) {
    const original=transforms.get(name)||name,f = fields.get(original),encoded=queryFields.get(name);
    check(f && encoded && wire(encoded) === nativeWire && !bound.has(original), "binding identity"); bound.add(original);
    check(location !== "query" || transforms.has(name) || ["string","boolean","integer","number","long","int32","int64","float","double"].includes(shape(f.fieldValue,models).type), "query requires scalar or reviewed JSON transform");
    inputs.push({field:original,wire:nativeWire,location,guard:"isUnset",...(transforms.has(name)?{encoding:"json"}:encoding?{encoding}:{}),schema:{...shape(f.fieldValue,models),required:!!f.required}});
  }
  let hasQuery = false,hasBodyFields=false;
  while(statements[index]?.type === "declare" && ["query","body"].includes(lex(statements[index].id))) {
    const q=statements[index++],target=lex(q.id);check(q.expr.type === "object" && !q.expr.fields.length,"map initialization");
    check((target === "query" && !hasQuery)||(target === "body" && !hasBodyFields),"duplicate map");if(target === "query")hasQuery=true;else hasBodyFields=true;
    while(statements[index]?.type === "if") {
      const guard = statements[index++], name = property(guard.condition?.expr?.args?.[0]);
      check(guard.condition?.type === "not" && call(guard.condition.expr,"Util","isUnset",e => !!property(e)) && !guard.elseIfs.length && !guard.elseStmts && guard.stmts.stmts.length === 1,"query guard");
      const assignment = guard.stmts.stmts[0];
      check(assignment.type === "assign" && assignment.left.type === "map_access" && lex(assignment.left.id) === target && constant(assignment.left.accessKey) !== null && property(assignment.expr) === name,"map assignment");
      bind(name,constant(assignment.left.accessKey),target === "query" ? "query" : "body-member");
    }
  }
  const req = statements[index++]; check(req?.type === "declare" && lex(req.id) === "req" && construct(req.expr,"OpenApiRequest"),"request construction");
  const seen = new Set(); let queryUsed = false,bodyUsed=false;
  for(const f of req.expr.object.fields) {
    const name=lex(f.fieldName); check(f.type === "objectField" && !seen.has(name),"request member"); seen.add(name);
    if(name === "headers")check(variable(f.expr,"headers"),"headers handoff");
    else if(name === "query") {check(hasQuery && call(f.expr,"OpenApiUtil","query",e => variable(e,"query")),"query handoff");queryUsed=true;}
    else if(name === "body") {
      if(hasBodyFields) {check(call(f.expr,"OpenApiUtil","parseToMap",e=>variable(e,"body")),"JSON map body handoff");bodyUsed=true;continue;}
      const bodyName = property(f.expr?.args?.[0]);
      check(fields.has(bodyName) && call(f.expr,"OpenApiUtil","parseToMap",e => !!property(e)),"JSON body conversion");
      bind(bodyName,wire(fields.get(bodyName)),"body","json");
    } else check(false,"unsupported request member "+name);
  }
  check(seen.has("headers") && (!hasQuery || queryUsed) && (!hasBodyFields || bodyUsed) && bound.size === fields.size,"unbound input");
  inputs.push({field:"headers",wire:"headers",location:"headers",guard:"isUnset",schema:{type:"map",values:{type:"string"}}});
  const protocol=statements[index++]; check(protocol?.type === "declare" && lex(protocol.id) === "params" && construct(protocol.expr,"Params"),"protocol construction");
  const expected=["action","version","protocol","pathname","method","authType","style","reqBodyType","bodyType"], facts={}, usedPaths=new Set();
  for(const f of protocol.expr.object.fields) {
    const name=lex(f.fieldName);check(f.type === "objectField" && expected.includes(name) && !Object.hasOwn(facts,name),"protocol constants");
    if(name !== "pathname") {check(constant(f.expr) !== null,"protocol constants");facts[name]=constant(f.expr);continue;}
    check(f.expr.type === "template_string" || f.expr.type === "string","path template");
    let template="";
    for(const element of f.expr.type === "string" ? [{type:"element",value:{string:constant(f.expr)}}] : f.expr.elements) {
      if(element.type === "element") {check(typeof element.value.string === "string" && !/[{}?#\r\n\t]/.test(element.value.string),"path literal");template+=element.value.string;}
      else {
        const param=paths.find(p => variable(element.expr?.args?.[0],lex(p.paramName)));
        check(element.type === "expr" && param && call(element.expr,"OpenApiUtil","getEncodeParam",e => variable(e,lex(param.paramName))),"path interpolation");
        const name=lex(param.paramName);usedPaths.add(name);template+="{"+name+"}";
      }
    }
    check(template.startsWith("/"),"absolute path");facts.pathname=template;
  }
  check(Object.keys(facts).length === expected.length && facts.action === operation && facts.protocol === "HTTPS" && facts.authType === "AK" && facts.style === "ROA" && facts.reqBodyType === "json" && ["GET","POST","PUT","DELETE","PATCH"].includes(facts.method),"profile");
  check(["json","none"].includes(facts.bodyType),"unsupported response body "+facts.bodyType);
  check(usedPaths.size === paths.length,"unused path parameter");
  for(const p of paths)inputs.push({field:lex(p.paramName),wire:lex(p.paramName),location:"path",guard:"required",schema:{type:"string",required:true}});
  const result=statements[index++], handoff=result?.expr;
  check(result?.type === "return" && handoff?.type === "call" && handoff.left.type === "method_call" && lex(handoff.left.id) === "callApi" && handoff.args.length === 3 && handoff.args.every((e,i)=>variable(e,["params","req","runtime"][i])) && index === statements.length,"runtime handoff and trailing statements");
  const responseModel=models.get(lex(fn.returnType)); check(responseModel,"response model");
  const body=responseModel.modelBody.nodes.find(f => wire(f) === "body");
  check(facts.bodyType === "none" || body,"response body");
  return {name:operation,protocol:facts,inputs,facade:true,response:facts.bodyType === "none" ? {type:"object",properties:{}} : shape(body.fieldValue,models)};
}
module.exports={roaStyle,lowerROA};
