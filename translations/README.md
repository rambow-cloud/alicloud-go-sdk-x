# Reviewed documentation translations

[中文](README.zh-CN.md)

- Issue: #93. Optional prose enrichment; it never selects operations, fields or runtime capabilities.
- Official Darabonba discovery remains automatic. Products need no translation file to generate.
- Each `<product>.json` binds to the exact source manifest and contains reviewed English translations of existing Chinese operation summaries.
- Entries record the attribute, original source coordinates, SHA256 of the original parser text, English text and original Chinese text.
- Source drift, unknown operations, duplicate/unsupported attributes, invalid language or attempts to replace existing English prose fail before generation writes.
- Generated comments identify translations explicitly, preserve Apache attribution and link the original DSL.
- Coverage reports count upstream English, reviewed translations and Go contract comments separately. Missing upstream field prose remains visible.
- Translations add no validation, pagination, waiter or retry policy. Never translate account/resource examples into runnable tests.
