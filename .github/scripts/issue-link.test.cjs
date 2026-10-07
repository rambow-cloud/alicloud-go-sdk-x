const { test } = require('node:test');
const assert = require('node:assert/strict');
const { references, run } = require('./issue-link.cjs');

test('accepts closing and partial references for this repository only', () => {
  const body = 'Closes #1\nRefs #2\nFixes https://github.com/RAMBOW-CLOUD/alicloud-go-sdk-x/issues/3\nCloses https://github.com/other/repo/issues/4\nCloses #1';
  assert.deepEqual(references(body, 'rambow-cloud', 'alicloud-go-sdk-x'), [1, 2, 3]);
});

test('ignores templates, fenced examples, invalid numbers and bare mentions', () => {
  const body = '<!-- Closes #12 -->\n```text\nCloses #13\n```\nRelated #14\nCloses #0\nCloses #9007199254740992';
  assert.deepEqual(references(body, 'rambow-cloud', 'alicloud-go-sdk-x'), []);
});

test('PR numbers and missing issues do not satisfy the gate', async () => {
  let failure;
  const github = { rest: { issues: { get: async ({ issue_number }) => {
    if (issue_number === 1) return { data: { pull_request: {} } };
    throw Object.assign(new Error('missing'), { status: 404 });
  } } } };
  await run({ github, context: { repo: { owner: 'rambow-cloud', repo: 'alicloud-go-sdk-x' }, payload: { pull_request: { body: 'Closes #1\nRefs #2' } } },
    core: { setFailed: message => { failure = message; } } });
  assert.match(failure, /existing repository issue/);
});

test('existing issue satisfies the gate without requiring premature closure', async () => {
  await run({ github: { rest: { issues: { get: async () => ({ data: {} }) } } },
    context: { repo: { owner: 'rambow-cloud', repo: 'alicloud-go-sdk-x' }, payload: { pull_request: { body: 'Refs #3' } } },
    core: { setFailed: () => { assert.fail('valid issue rejected'); } } });
});
