# Project working agreements

## Product goal

Build an independent Alibaba Cloud SDK for Go with idiomatic APIs, small dependencies,
predictable request behavior, and documentation available on pkg.go.dev. This project
is not an official Alibaba Cloud SDK. Read `docs/research.md` and `docs/design.md`
before changing the public API.

## Issue-driven development

- Every code or behavior change starts with an issue containing the problem, evidence,
  scope, acceptance criteria, and documentation requirements. Use GitHub issues in
  `rambow-cloud/alicloud-go-sdk-x`; when offline, create a draft under `docs/issues/`
  and synchronize it before opening a PR.
- Keep one independently reviewable issue per branch: `issue/<number>-<slug>`.
- Before implementing, establish how the acceptance criteria will be verified.
- PRs link the issue using `Closes #<number>` for complete work, or `Refs #<number>`
  for partial work, and include behavior, validation, and docs.
- A change is complete only when code, relevant checks, examples, and package docs
  satisfy the issue. Do not close roadmap issues just because an interface exists.
- Never invent issue numbers, completed checks, compatibility, or service coverage.

## Go API rules

- Every blocking operation accepts `context.Context` as its first argument and
  preserves cancellation/deadline errors for `errors.Is`.
- Use standard `net/http`, injectable HTTP clients, and `time.Duration` for timeouts.
- Require Go 1.27 or newer. Import `encoding/json/v2` directly for JSON; do not add
  legacy `encoding/json`, third-party JSON libraries, or an experiment flag.
- Keep client configuration private after construction; do not mutate shared state
  or caller requests. Document concurrency contracts of extension interfaces.
- Use concrete request/response types and `errors.As` for service errors. Pointers
  represent meaningful absence; do not require helpers for every ordinary value.
- Retry only when the operation's idempotency and error policy allow it. Bound total
  attempts and elapsed time, respect context during backoff, and never retry arbitrary
  writes by default.
- Add dependencies only with issue justification. Keep the runtime core standard
  library only unless a concrete requirement prevents it.
- Keep signing, encoding, and retry machinery internal until real use requires an
  extension contract. Do not implement cloud operations without protocol evidence.
- Never log credentials, authorization headers, or raw request/response bodies by
  default. Do not commit real credentials or make live cloud calls in unit tests.

## pkg.go.dev definition of done

- Every public package has a `doc.go` overview explaining usage and limitations.
- Every exported declaration, field, and interface method has a Go doc comment
  beginning with its name. Explain zero values, defaults, errors, cancellation,
  ownership, and concurrency when relevant. Use native Go comments, not `@param`.
- Every public package has at least one runnable external `Example` with deterministic
  output and no account/network requirement. Add operation examples as operations land.
- Keep `LICENSE`, canonical module imports, and runnable README snippets current.
- Run `go run ./internal/cmd/doccheck`, `go vet ./...`, and `go test ./...` once for
  a meaningful change. CI runs the race detector on Linux. Check formatting once.
- `docs/releasing.md` describes indexing; local docs passing does not mean the package
  has already been published or indexed.

## Validation and communication

- Avoid redundant validation and repeated retries. Repeat only after a meaningful
  state change or a failure whose reason is newly understood.
- If a result can be verified by opening it in a browser, tell the user the exact URL
  and what to inspect instead of checking it using curl or similar local requests.
- For DNS, server-side CloudFormation status and outputs are the source of truth.
  Do not validate DNS locally; provide user verification steps only if needed.
- Do not spawn sub-agents unless the user explicitly requests delegation.
