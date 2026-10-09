# Documentation gap completion

[中文](documentation-gap-completion.zh-CN.md)

- Issue: #93, staged delivery. Source inventory: STS 4, ECS 380, VPC 403 actions.
- Correct generated STS guidance for renewable OIDC/SAML providers, OIDC discovery and explicit long-lived opt-in.
- Fill all nine Chinese-only ECS operation summaries using [source-bound reviewed translations](../translations/README.md).
- Keep optional enrichment separate from automatic operation/model discovery and runtime policy. New services do not require translations to generate.
- Source-manifest/text hashes and exact parser coordinates reject drift before writes. Upstream source bytes remain unchanged.
- Generated comments identify English translations and preserve Apache attribution. They are project-reviewed translations, not official English text or independent human review.
- Every generated operation and source field has a Go contract comment. Fields without upstream English prose also link their exact DSL location.
- Reports distinguish original English prose, translated summaries and Go contract comments. Missing upstream field descriptions are not counted as completed business documentation.

## Remaining source prose

| Product | Original English operation prose | Reviewed English summaries | Original English field prose | Total source fields |
|---|---:|---:|---:|---:|
| STS | 4 / 4 | 0 | 55 | 78 |
| ECS | 371 / 380 | 9 | 4,234 | 9,626 |
| VPC | 403 / 403 | 0 | 5,036 | 9,017 |

- All generated action summaries now have original or reviewed English prose.
- 23 STS, 5,392 ECS and 3,981 VPC source fields still lack original English prose. Their exact wire names, absence/zero contracts and source coordinates are documented; no undocumented business defaults are invented.
- Broader field-description enrichment remains #93 work. It requires reviewed official text after source normalization; generic comments or source links do not replace missing service semantics.
- Real federation renewal remains #94 / NOT RUN. Release/indexing stays #61; neither is established by documentation checks.
