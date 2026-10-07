'use strict';
const fs = require('node:fs');
const { execFileSync } = require('node:child_process');

function check(text) {
  const en = text.indexOf('## English\n');
  const zh = text.indexOf('## 中文\n');
  return en >= 0 && zh > en && /[A-Za-z]/.test(text.slice(en + 11, zh)) && /[\u3400-\u9fff]/.test(text.slice(zh + 6));
}

// Preserve byte-exact upstream documentation and license notices in the source lock.
function projectDocument(path) {
  return !/^sources\/darabonba\/modules\/[^/]+\/README(?:-CN)?\.md$/.test(path);
}

if (require.main === module) {
  const paths = execFileSync('git', ['ls-files', '--cached', '--others', '--exclude-standard', '--', '*.md'], { encoding: 'utf8' }).trim().split(/\r?\n/).filter(Boolean);
  const documents = paths.filter(projectDocument);
  const failed = documents.filter(path => !check(fs.readFileSync(path, 'utf8').replace(/\r\n/g, '\n')));
  if (failed.length) { console.error('Missing equivalent English/Chinese sections:', failed.join(', ')); process.exitCode = 1; }
  else console.log(`Bilingual structure passed for ${documents.length} project documents; ${paths.length-documents.length} byte-exact upstream documents retained. Semantic equivalence is reviewed.`);
}
module.exports = { check, projectDocument };
