## Problem

- English and Chinese share files. Readers must scroll through both languages.
- Chinese guides use literal translations and dense technical shorthand.
- Issues and forms include duplicate Chinese text.
- Generated guides and CI still require mixed-language sections.

## Scope

- Keep English paths. Add linked name.zh-CN.md guides.
- Rewrite Chinese guidance and shorten English prose with bullet points.
- Make repository issues, issue drafts, forms and PR templates English only.
- Update AGENTS.md, contribution rules, generator templates and language checks.
- Preserve source files, license notices, runtime behavior and acceptance evidence.
- Do not publish a release or rerun cloud operations.

## Acceptance

- [ ] Every project guide has separate, linked English and Chinese files; each is usable on its own.
- [ ] Chinese wording is natural; English uses short sentences and bullets without losing constraints, commands, links or evidence.
- [ ] Existing GitHub issue bodies and repository issue files/forms use English only; states, labels, milestones and dependencies stay intact.
- [ ] Both generators emit separate guides and deterministic checks pass.
- [ ] CI rejects mixed guides, missing pairs, broken local links and Chinese issue text. Upstream documents remain unchanged.
- [ ] Frontend checks/tests, generator checks, doccheck, vet, tests, formatting and Linux race/Windows CI pass.

## Verification

- Follow docs/documentation-style.md.
- Review representative usage, development, acceptance and generated guides in both languages.
- Test language checks against missing pairs, fenced code, links and upstream exceptions.
- Check generated ownership and SDK/source changes before opening the PR.
