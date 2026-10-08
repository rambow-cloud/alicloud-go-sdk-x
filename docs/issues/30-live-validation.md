# [Maintenance]: Validate generated reads with local Profile and Explorer CLI references

- GitHub issue: #30.

### Problem and scope

- The current SDK commit 618303b has offline and Linux race/Windows acceptance, but its selected generated APIs have not been checked with the user's explicitly authorized local Aliyun CLI profile.
- Follow docs/live-validation.md, written before execution.
- Use CLI read-only references corresponding to OpenAPI Explorer examples, then inject the local temporary credential snapshot into the existing SDK.
- No general profile/OAuth provider and no resource mutations.

- Dependency: #29 (completed).

### Acceptance criteria

- [ ] Record sanitized real-account SDK/CLI comparisons for ECS DescribeRegions, DescribeInstances token/page modes, DescribeInstanceStatus and VPC DescribeVpcs; bound each traversal to two pages and distinguish permission/auth failures, empty accounts and live drift.
- [ ] Check Wait/WaitForOutput only with an existing Running identity, or document skip; STS AssumeRole requires an explicit role. Do not claim live validation of retries, tracing, renewal or Explorer browser UI.
- [ ] Read credentials only locally, check expiration, keep raw data/harness ignored, and publish no credentials or resource/account identifiers.
- [ ] Publish equivalent English/Chinese evidence and Explorer browser verification links; record source commit and CLI version.

### Affected areas

- core, ecs, vpc, tools
