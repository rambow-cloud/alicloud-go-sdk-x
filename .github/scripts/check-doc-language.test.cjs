const { test } = require('node:test');
const assert = require('node:assert/strict');
const { check } = require('./check-doc-language.cjs');
test('requires both ordered language sections with real content', () => {
  assert.equal(check('## English\nUsage.\n## 中文\n使用。'), true);
  assert.equal(check('## English\nUsage.'), false);
  assert.equal(check('## English\n\n## 中文\n使用。'), false);
  assert.equal(check('## English\nUsage.\n## 中文\n'), false);
});
