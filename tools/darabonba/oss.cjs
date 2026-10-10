"use strict";

const { lowerROA } = require("./roa.cjs");
const lex = value => value?.lexeme;
const variable = (value, name) => value?.type === "variable" && lex(value.id) === name;
const literal = value => value?.type === "string" ? value.value.string : null;
const virtual = (value, name) => value?.type === "virtualVariable" && lex(value.vid) === name;

function gatewayProduct(ast) {
  return ast.moduleBody.nodes.some(node => node.type === "init" && node.initBody.stmts.some(s => lex(s.left?.vid) === "@spi"));
}

// Accept the complete initializer, rather than finding a convenient assignment
// inside arbitrary control flow. The import lock supplies the Gateway identity.
function validateGateway(ast, requireProfile) {
  const check = (ok, reason) => requireProfile(ok, "OSS " + reason);
  const initializers = ast.moduleBody.nodes.filter(node => node.type === "init");
  check(initializers.length === 1, "Gateway initializer");
  const [superCall, client, spi, rule, ...rest] = initializers[0].initBody.stmts;
  check(superCall?.type === "super" && superCall.args?.length === 1 && variable(superCall.args[0], "config") &&
    client?.type === "assign" && virtual(client.left, "@client") && client.expr?.type === "construct" && lex(client.expr.aliasId) === "GatewayClient" && !client.expr.args.length &&
    spi?.type === "assign" && virtual(spi.left, "@spi") && virtual(spi.expr, "@client") &&
    rule?.type === "assign" && virtual(rule.left, "@endpointRule") && literal(rule.expr) === "" && !rest.length,
  "Gateway initializer");
}

function pathFacts(expression, check) {
  const text = expression?.type === "string" ? literal(expression) :
    expression?.type === "template_string" && expression.elements.every(e => e.type === "element") ? expression.elements.map(e => e.value.string).join("") : null;
  check(typeof text === "string" && text.startsWith("/") && !/[{}#\r\n\t]/.test(text), "static path");
  const index = text.indexOf("?"), pathname = index < 0 ? text : text.slice(0, index), query = [];
  if (index >= 0) {
    const seen = new Set();
    for (const item of text.slice(index + 1).split("&")) {
      // Dynamic, escaped or ambiguous subresource programs need a separate profile.
      const [name, value, ...extra] = item.split("=");
      check(/^[A-Za-z0-9-]+$/.test(name) && !seen.has(name) && !extra.length && (value === undefined || /^[A-Za-z0-9-]*$/.test(value)), "static subresource");
      seen.add(name); query.push({ name, value: value ?? "" });
    }
  }
  return { pathname, query };
}

function lowerOSS(ast, fn, operation, helpers, reviewXML) {
  const check = (ok, reason) => helpers.requireProfile(ok, "OSS " + reason);
  validateGateway(ast, helpers.requireProfile);
  check(typeof reviewXML === "function", "native XML facts missing");
  const cloned = structuredClone(fn), statements = cloned.functionBody.stmts.stmts;
  const protocol = statements.find(s => s.type === "declare" && lex(s.id) === "params");
  const constants = protocol?.expr?.object?.fields;
  check(Array.isArray(constants), "protocol constants");
  const fact = name => constants.filter(f => lex(f.fieldName) === name);
  check(["method", "reqBodyType", "bodyType", "pathname"].every(name => fact(name).length === 1), "protocol constants");
  check(literal(fact("method")[0].expr) === "GET" && literal(fact("reqBodyType")[0].expr) === "xml" && literal(fact("bodyType")[0].expr) === "xml", "bodyless XML read profile");
  const xmlResponse = reviewXML(operation, fn.returnType);
  const path = pathFacts(fact("pathname")[0].expr, check);
  const request = statements.find(s => s.type === "declare" && lex(s.id) === "req");
  check(request?.expr?.object?.fields && !request.expr.object.fields.some(f => ["body", "stream"].includes(lex(f.fieldName))), "request body is not supported");
  const inputParams = cloned.params.params.slice(0, -2), buckets = inputParams.filter(p => lex(p.paramName) === "bucket");
  check(buckets.length === 1 && lex(buckets[0].paramType) === "string" && !buckets[0].defaultValue && inputParams.every(p => ["bucket", "request"].includes(lex(p.paramName))), "bucket signature");
  const hostIndex = statements.findIndex(s => s.type === "declare" && lex(s.id) === "hostMap");
  const host = statements[hostIndex], assignment = statements[hostIndex + 1];
  check(hostIndex >= 0 && host.expr?.type === "object" && !host.expr.fields.length &&
    assignment?.type === "assign" && assignment.left?.type === "map_access" && lex(assignment.left.id) === "hostMap" && literal(assignment.left.accessKey) === "bucket" && variable(assignment.expr, "bucket"), "bucket host map");
  const hostFields = request.expr.object.fields.filter(f => lex(f.fieldName) === "hostMap");
  check(hostFields.length === 1 && variable(hostFields[0].expr, "hostMap"), "host map handoff");
  const handoff = statements.at(-1);
  check(handoff?.type === "return" && handoff.expr?.type === "call" && handoff.expr.left?.type === "method_call" && lex(handoff.expr.left.id) === "execute" && handoff.expr.args?.length === 3 && handoff.expr.args.every((arg, i) => variable(arg, ["params", "req", "runtime"][i])), "execute handoff");
  // A bounded normalization delegates the remaining query/model program to the
  // existing generic ROA checker. Nothing else is deleted or rewritten.
  statements.splice(hostIndex, 2);
  request.expr.object.fields = request.expr.object.fields.filter(f => f !== hostFields[0]);
  cloned.params.params = cloned.params.params.filter(p => p !== buckets[0]);
  handoff.expr.left.id.lexeme = "callApi";
  for (const name of ["reqBodyType", "bodyType"]) fact(name)[0].expr.value.string = "json";
  fact("pathname")[0].expr = { type: "string", value: { string: path.pathname } };
  const lowered = lowerROA(ast, cloned, operation, helpers);
  check(!lowered.inputs.some(input => path.query.some(q => q.name === input.wire && input.location === "query")), "query/subresource collision");
  lowered.protocol.reqBodyType = "xml";
  lowered.protocol.bodyType = "xml";
  lowered.inputs.push({ field: "bucket", wire: "bucket", location: "bucket", guard: "required", schema: { type: "string", required: true } });
  lowered.oss = { profile: "gateway-oss4-xml-read-v1", query: path.query, response: xmlResponse };
  return lowered;
}

module.exports = { gatewayProduct, validateGateway, lowerOSS, pathFacts };
