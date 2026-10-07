const { test } = require('node:test');
const assert = require('node:assert/strict');
const { check, projectDocument } = require('./check-doc-language.cjs');
test('requires both ordered language sections with real content', () => {
  assert.equal(check('## English\nUsage.\n## 中文\n使用。'), true);
  assert.equal(check('## English\nUsage.'), false);
  assert.equal(check('## English\n\n## 中文\n使用。'), false);
  assert.equal(check('## English\nUsage.\n## 中文\n'), false);
});
test('excludes only preserved upstream READMEs, retaining our source guides', () => {
  assert.equal(projectDocument('sources/darabonba/modules/alibabacloud_OpenApi_0.3.23/README.md'), false);
  assert.equal(projectDocument('sources/darabonba/modules/darabonba_String_0.0.13/README-CN.md'), false);
  assert.equal(projectDocument('sources/darabonba/README.md'), true);
  assert.equal(projectDocument('docs/darabonba-migration.md'), true);
});
