# Releasing and pkg.go.dev

[中文](releasing.zh-CN.md)

## Current delivery route (2026-10-09)

- Follow [STS/ECS/VPC delivery](sts-ecs-vpc-path.md): finish #60, then ECS #74 and VPC #75, then #61 release/indexing.
- #60 now requires truthful implementation-agent consumer acceptance. Independent human usability moves to optional follow-up #76; its record remains NOT RUN and does not block this release.
- Existing STS-only first-release scheduling and required-human #60 gates below are historical and superseded by this decision.
- Preserve actual technical/live/source evidence. Agent test duration is not human task time.
- Publication remains pending until both product acceptance issues pass. No tag is created by this change.

- The first release target is [v0.1.0 STS](sts-v0.1.0.md), parent #57 / release #61, milestone v0.1.0 / Project 3.
- Apply its explicitly scoped experimental gate before publication; the older multi-product candidate Beta schedule is not a prerequisite for this version.
- Its broader acceptance remains open.
- Include all four pinned STS actions' actual coverage and OIDC/SAML successful live federation NOT RUN in paired release notes.
- Other packages stay available with their separate acceptance limits.
- This planning update does not publish a tag, release or indexing claim.

- Canonical public module: github.com/rambow-cloud/alicloud-go-sdk-x, Go 1.27.
- Original runtime/tooling use MIT; generated product definitions/prose retain Apache-2.0, with package LICENSE/NOTICE and pinned-source references.
- Inspect both before release.
- This work does not create a release tag.
- Release only the documented operation coverage after CI and foundation acceptance; update equivalent English/Chinese release notes.
- Use v0 for experimental APIs.
- Never overwrite indexed tags; use a new version and retract when necessary. v1 promises compatibility; later major modules use Go's versioned paths.

- Open https://pkg.go.dev/github.com/rambow-cloud/alicloud-go-sdk-x in a browser.
- If absent, click Request.
- Inspect version, license, overview, exported field docs and external Examples, then inspect credentials and the other public packages.
- The badge is an indexing entry point, not evidence that indexing occurred.
- Do not use local curl/DNS to verify the page.
- If indexing fails, diagnose public access, module path, Go support and license before retrying.
- Sources: https://pkg.go.dev/about#adding-a-package and https://pkg.go.dev/license-policy.

- Use the prepared [paired release notes](releases/v0.1.0.md) and [exact release/indexing checklist](sts-v010-release-checklist.md).
- Required independent acceptance remains NOT RUN; no tag or indexing is claimed.
- The read-only release guard and same-version browser URLs are recorded there.
- Existing authorization covers completion after gates pass; missing human evidence is not an invitation to reconfirm publication permission.
