# Issue management

[中文](issue-management.zh-CN.md)

- Follow the type/module classification of rambow-cloud/powertools-lambda-go, with explicit priority and status labels. .github/labels.json is the managed catalog; automation updates catalog labels without deleting outside manual labels. module labels denote responsibilities, not separate go.mod files.
- Keep new areas synchronized in catalog, forms and classifier.

- Types: bug, enhancement, documentation, maintenance, question, release.
- Supplement with cicd, help wanted or good first issue only when justified.
- Keep exactly one priority: p1 foundation/current milestone, p2 planned follow-up, p3 uncommitted future work.
- Keep one status: triage, ready, in-progress, blocked, done.
- Record blockers in issue bodies.

- English-only issue forms include Affected areas.
- Automation classifies title prefixes and this structured field, preserves manual priority/active status, sets done on closure, and triage on reopening.
- Done means delivered acceptance, not release/indexing.
- The workflow handles issue events and manual catalog backfill without posting comments.

- Translate legacy issues without losing historical completion evidence.
- Complete work uses Closes #N, partial work Refs #N; the PR gate checks actual repository issues, excluding comment templates, fenced examples and PR numbers.
- Follow development-path.md and the issue index.
- Keep generator blocked until the eleven-capability foundation passes.

- The issue workflow rejects Chinese prose in issue titles or bodies.
- Code samples may preserve literal API values.
- Language-only edits keep issue states, labels, milestones, links and acceptance evidence.
