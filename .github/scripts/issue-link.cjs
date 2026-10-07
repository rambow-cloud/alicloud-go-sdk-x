'use strict';

function references(body, owner, repo) {
  const text = (body || '').replace(/<!--[\s\S]*?-->/g, '').replace(/```[\s\S]*?```/g, '');
  const pattern = /\b(?:close[sd]?|fix(?:e[sd])?|resolve[sd]?|refs)\s+(?:#(\d+)|https:\/\/github\.com\/([^/]+)\/([^/]+)\/issues\/(\d+))\b/gi;
  const numbers = new Set();
  for (const match of text.matchAll(pattern)) {
    if (match[2] && (match[2].toLowerCase() !== owner.toLowerCase() || match[3].toLowerCase() !== repo.toLowerCase())) continue;
    const number = Number(match[1] || match[4]);
    if (Number.isSafeInteger(number) && number > 0) numbers.add(number);
  }
  return [...numbers];
}

async function run({ github, context, core }) {
  const { owner, repo } = context.repo;
  for (const issue_number of references(context.payload.pull_request.body, owner, repo)) {
    try {
      const { data } = await github.rest.issues.get({ owner, repo, issue_number });
      if (!data.pull_request) return;
    } catch (error) {
      if (error.status !== 404) throw error;
    }
  }
  core.setFailed('Reference an existing repository issue using Closes #<number>, or Refs #<number> for partial work.');
}

module.exports = { references, run };
