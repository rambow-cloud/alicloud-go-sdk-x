# Official Darabonba frontend

- Tracks [#31](https://github.com/rambow-cloud/alicloud-go-sdk-x/issues/31), depending on accepted #29/#30.
- The development path was established in [darabonba-migration.md](../darabonba-migration.md) before implementation.

- Problem: metadata-only generation lacks official SDK request binding and model evidence.
- Reuse real official ECS/STS/VPC DSL, official parser/import tools and pinned modules, retaining our IR, Go backend and unified runtime.
- Scope is the existing five operations; unsupported selected behavior must fail before writes.

- Acceptance: full imported-module semantic checking; deterministic offline projection and regeneration; reviewed metadata/DSL decisions; checksum and unsupported-pattern rejection; unchanged Go APIs and runtime contracts; Go 1.27/JSON v2, pkg.go.dev comments, Examples and paired docs; frontend checks/tests, Go gates and CI on Linux/Windows.

- Implementation: official parser frontend and build-only tool lock, pinned source/module/ license evidence, protocol/binding/model projections, strict cross-source IR checks, machine-readable divergence approvals, and CI frontend regeneration.
- Source and projection corruption, changed types/bindings/protocol/requiredness, missing operations and unreviewed inputs are exercised before output writes.
- Prior runtime behavior is retained and checked against the metadata-only backend.

- Evidence and remaining browser verification are recorded in [decisions](../darabonba-decisions.md).
- Benchmark #20 and broader DSL profiles remain separate.
- Current state and verification results are maintained on the GitHub issue.
