'use strict';

const fs = require('node:fs');

const types = new Map([
  ['bug', 'bug'], ['feature', 'enhancement'], ['docs', 'documentation'],
  ['maintenance', 'maintenance'], ['ci', 'maintenance'], ['release', 'release'],
  ['question', 'question'],
]);
const areas = new Set(['core', 'credentials', 'transport', 'signing', 'ecs', 'tools']);

function classify(issue) {
  const match = issue.title.match(/^\[([^\]]+)\]:/);
  const type = match && types.get(match[1].toLowerCase());
  const areaSection = (issue.body || '').match(/(?:^|\n)### Affected areas[ \t]*\r?\n([\s\S]*?)(?=\r?\n#{1,3} |$)/);
  const modules = areaSection
    ? areaSection[1].trim().split(/[,\n]/).map(value => value.trim()).filter(value => areas.has(value)).map(value => `module:${value}`)
    : null;
  return { type, modules };
}

function planLabels(issue, eventAction) {
  const current = (issue.labels || []).map(label => typeof label === 'string' ? label : label.name);
  const desired = new Set(current);
  const { type, modules } = classify(issue);
  if (type) {
    for (const category of types.values()) desired.delete(category);
    desired.add(type);
  }
  // Only replace module labels when the structured area field is valid.
  if (modules && modules.length) {
    for (const label of current) if (label.startsWith('module:')) desired.delete(label);
    for (const label of modules) desired.add(label);
  }
  if (!current.some(label => label.startsWith('priority:'))) desired.add('priority:p2');
  if (issue.state === 'closed') {
    for (const label of current) if (label.startsWith('status:')) desired.delete(label);
    desired.add('status:done');
  } else if (eventAction === 'reopened' || desired.has('status:done')) {
    for (const label of current) if (label.startsWith('status:')) desired.delete(label);
    desired.add('status:triage');
  } else if (!current.some(label => label.startsWith('status:'))) {
    desired.add('status:triage');
  }
  return {
    add: [...desired].filter(label => !current.includes(label)),
    remove: current.filter(label => !desired.has(label)),
  };
}

async function run({ github, context, core }) {
  const { owner, repo } = context.repo;
  const catalog = JSON.parse(fs.readFileSync('.github/labels.json', 'utf8'));
  const existing = await github.paginate(github.rest.issues.listLabelsForRepo, { owner, repo, per_page: 100 });
  for (const label of catalog) {
    const old = existing.find(item => item.name === label.name);
    if (!old) await github.rest.issues.createLabel({ owner, repo, ...label });
    else if (old.color.toLowerCase() !== label.color.toLowerCase() || (old.description || '') !== label.description) {
      await github.rest.issues.updateLabel({ owner, repo, ...label });
    }
  }
  const issues = context.payload.issue
    ? [(await github.rest.issues.get({ owner, repo, issue_number: context.payload.issue.number })).data]
    : await github.paginate(github.rest.issues.listForRepo, { owner, repo, state: 'all', per_page: 100 });
  for (const issue of issues) {
    if (issue.pull_request) continue;
    const plan = planLabels(issue, context.payload.action);
    if (plan.add.length) await github.rest.issues.addLabels({ owner, repo, issue_number: issue.number, labels: plan.add });
    for (const name of plan.remove) await github.rest.issues.removeLabel({ owner, repo, issue_number: issue.number, name });
    core.info(`Issue #${issue.number}: added ${plan.add.length}, removed ${plan.remove.length} labels`);
  }
}

module.exports = { run, classify, planLabels };
