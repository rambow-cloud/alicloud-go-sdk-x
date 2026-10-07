const { test } = require('node:test');
const assert = require('node:assert/strict');
const { classify, planLabels } = require('./issue-labels.cjs');

test('classifies structured areas and ignores arbitrary issue text', () => {
  assert.deepEqual(classify({ title: '[Feature]: transport', body: '### Affected areas\n\ntransport, signing\n\n### Problem\n\ncredentials' }),
    { type: 'enhancement', modules: ['module:transport', 'module:signing'] });
  assert.deepEqual(classify({ title: 'Question about docs', body: 'some mention of ecs' }), { type: null, modules: null });
});

test('preserves priority, active state and manual labels while updating areas', () => {
  const plan = planLabels({ title: '[Docs]: docs', body: '### Affected areas\n\ncore\n', state: 'open',
    labels: ['enhancement', 'module:ecs', 'priority:p1', 'status:in-progress', 'help wanted'] }, 'edited');
  assert.deepEqual(plan.add, ['documentation', 'module:core']);
  assert.deepEqual(plan.remove, ['enhancement', 'module:ecs']);
});

test('closed issues become done and reopening requires triage', () => {
  const closed = planLabels({ title: 'work', state: 'closed', labels: ['priority:p1', 'status:blocked'] }, 'closed');
  assert.deepEqual(closed, { add: ['status:done'], remove: ['status:blocked'] });
  const reopened = planLabels({ title: 'work', state: 'open', labels: ['priority:p1', 'status:done'] }, 'reopened');
  assert.deepEqual(reopened, { add: ['status:triage'], remove: ['status:done'] });
});

test('new issues receive defaults; missing area fields preserve manual areas', () => {
  const plan = planLabels({ title: '[Bug]: failure', body: '', state: 'open', labels: ['module:ecs'] }, 'opened');
  assert.deepEqual(plan, { add: ['bug', 'priority:p2', 'status:triage'], remove: [] });
});
