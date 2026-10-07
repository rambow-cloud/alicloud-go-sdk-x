GitHub issue: #30.

## English

### Problem and scope

The current SDK commit 618303b has offline and Linux race/Windows acceptance, but its selected generated APIs have not been checked with the user's explicitly authorized local Aliyun CLI profile. Follow docs/live-validation.md, written before execution. Use CLI read-only references corresponding to OpenAPI Explorer examples, then inject the local temporary credential snapshot into the existing SDK. No general profile/OAuth provider and no resource mutations.

Dependency: #29 (completed).

### Acceptance criteria

- [ ] Record sanitized real-account SDK/CLI comparisons for ECS DescribeRegions, DescribeInstances token/page modes, DescribeInstanceStatus and VPC DescribeVpcs; bound each traversal to two pages and distinguish permission/auth failures, empty accounts and live drift.
- [ ] Check Wait/WaitForOutput only with an existing Running identity, or document skip; STS AssumeRole requires an explicit role. Do not claim live validation of retries, tracing, renewal or Explorer browser UI.
- [ ] Read credentials only locally, check expiration, keep raw data/harness ignored, and publish no credentials or resource/account identifiers.
- [ ] Publish equivalent English/Chinese evidence and Explorer browser verification links; record source commit and CLI version.

### Affected areas

core, ecs, vpc, tools

## 中文

基于已通过离线及 Linux race/Windows 验收的 618303b，在用户授权的本地 Profile 下对照 Explorer CLI 示例执行只读验证；路径先写在 docs/live-validation.md，依赖已完成 #29。

验收 ECS 地域/实例 token 与页码/状态及 VPC 读取、每种最多两页、脱敏结果；有 Running 实例才验收 Wait/WaitForOutput，否则跳过；AssumeRole 缺少目标 role 不执行，不宣称重试/tracing/刷新及浏览器 UI 的真实验收。凭据仅本地注入并检查期限，原始数据/脚本忽略，公开证据不含凭据及资源/账号标识。记录提交、CLI 版本及双语证据和网页核验链接，不修改资源、不增加通用 Profile/OAuth provider。