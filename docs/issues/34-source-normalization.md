# Issue 34: source normalization

- Actual issue: [#34](https://github.com/rambow-cloud/alicloud-go-sdk-x/issues/34), stage of parent [#33](https://github.com/rambow-cloud/alicloud-go-sdk-x/issues/33).
- Depends on #31 and its unmerged PR #32; branch issue/34-source-normalization is stacked on issue/31-official-darabonba-frontend.
- Scope, commands and evidence limits are in [source-normalization.md](../source-normalization.md).
- The authoritative route was committed before implementation at a206ea0, with actual issue links at 92095fe.

- Delivery includes pinned Apache-2.0 canonical fixtures, the versioned Node adapter, exact CLI/wire case separation, recursive itemName wrappers, verified Filter aliases in Node/Go and an explicit optional-string exception for metadata-only Tag[].key/value.
- New discrepancies and source tampering fail before writes; no generated Go API or runtime changes.
- Regression tests include signed-wire/public contract tests already in the repository and cross-backend generated Go byte comparison.

- Acceptance uses frontend offline check/tests, sdkgen check, doccheck, vet, Go tests, formatting and paired-document checks; record actual results in the linked PR.
- Linux race/Windows CI remain required; local Windows checks do not replace them.
- Browser/live behavior is unverified for this stage. #35 owns full operation/model discovery; #36 owns batch emission.
- Parent #33 stays open for remaining stages.
