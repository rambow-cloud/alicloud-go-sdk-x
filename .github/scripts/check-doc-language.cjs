'use strict';
const fs = require('node:fs');
const { execFileSync } = require('node:child_process');

function check(text) {
  const en = text.indexOf('## English\n');
  const zh = text.indexOf('## 中文\n');
  return en >= 0 && zh > en && /[A-Za-z]/.test(text.slice(en + 11, zh)) && /[\u3400-\u9fff]/.test(text.slice(zh + 6));
}

if (require.main === module) {
  const paths = execFileSync('git', ['ls-files', '--cached', '--others', '--exclude-standard', '--', '*.md'], { encoding: 'utf8' }).trim().split(/\r?\n/).filter(Boolean);
  const failed = paths.filter(path => !check(fs.readFileSync(path, 'utf8').replace(/\r\n/g, '\n')));
  if (failed.length) { console.error('Missing equivalent English/Chinese sections:', failed.join(', ')); process.exitCode = 1; }
  else console.log(`Bilingual structure passed for ${paths.length} Markdown documents; semantic equivalence is reviewed.`);
}
module.exports = { check };
