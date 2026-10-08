# Documentation style

[中文](documentation-style.zh-CN.md)

- Write English in `name.md` and Chinese in `name.zh-CN.md`.
- Link each guide to its language counterpart.
- Update both in the same change.
- Use short sentences and bullet points.
- Keep code samples and comparison tables.
- Write natural Chinese.
- Explain technical terms on first use; keep API names unchanged.
- Include usage, defaults, limits and evidence in each version.
- Do not make readers switch languages to find a command or source.
- Write GitHub issues, issue drafts, forms and PR templates in English only.
- Keep Go comments, diagnostics and identifiers in English.
- Update generator templates before regenerating guides.
- Do not edit generated files.
- Preserve upstream source documents and license notices byte for byte.
- CI checks file pairs, language links, local links and English-only issue files.
- People review translation accuracy.

This policy follows the user's 2026-10-09 instruction and replaces earlier mixed-language Markdown rules. It does not change SDK behavior or release acceptance.

## Migration scope

- 69 project guides have separate English and Chinese files.
- 49 historical GitHub issues were edited to use English only; their states, labels and milestones were preserved.
- Both generators emit separate language guides.
- Runtime code, pinned DSL/metadata, source licenses and seven upstream READMEs are unchanged.
- Native Profile/OAuth and STS acceptance evidence stays recorded. Independent developer acceptance and release/indexing are still pending.
