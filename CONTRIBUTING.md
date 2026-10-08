# Contributing

[中文](CONTRIBUTING.zh-CN.md)

## Workflow

- Read [AGENTS.md](AGENTS.md) and [development path](docs/development-path.md).

- Open an English issue before code.
- Include evidence, scope, dependencies, acceptance and docs.

- Use `issue/<number>-<slug>`.
- Keep one reviewable issue per branch.

- Use `Closes #N` for complete work and `Refs #N` for partial work.
- Write issues and PRs in English.

- Offline drafts go in docs/issues/.
- Sync them before opening a PR; draft IDs are not GitHub issue numbers.

## Documentation

- Follow [writing rules](docs/documentation-style.md).
- English: name.md; Chinese: name.zh-CN.md.
- Link the pair and update both.

- Use short English bullets and natural Chinese.
- Each guide must include its own commands, defaults, limits and evidence.

- Go docs use English.
- Add doc.go, comments for exported symbols/fields and offline external Examples with the implementation.

- Preserve upstream source/README/license bytes.

## Generated files

- Edit pinned metadata, reviewed policy, templates or handwritten validation; do not edit generated files.

- Regenerate with `go run ./internal/cmd/sdkgen generate` and, for full-DSL products, `go run ./internal/cmd/sdkgen product-generate`.
- Then run the checks.

- Importing sources needs explicit network access; generation/check run offline.
- See [generator workflow](docs/generator.md).

- Keep the accepted foundation/runtime contracts and review schema drift before public API checks.

## Verify

- Run relevant checks once after a meaningful change.
- Run Node 22 frontend checks before Go gates for generator changes.

```powershell
# Node 22; only needed for generator changes
npm --prefix tools/darabonba run check
npm --prefix tools/darabonba test

go run ./internal/cmd/sdkgen check
go run ./internal/cmd/sdkgen product-check
go run ./internal/cmd/doccheck
node .github/scripts/check-doc-language.cjs
node --test .github/scripts/*.test.cjs
go vet ./...
go test ./...
```

- Use gofmt once.
- Linux CI adds race detection; Windows CI checks portability.

- Unit tests must not use real credentials or call cloud APIs.
- Do not change global Go/Git configuration.

- Repeat only after a meaningful change or an understood failure.
- Checks verify structure; reviewers verify translation, behavior, compatibility and limits.
