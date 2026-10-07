GitHub issue: #25.

# [Feature]: Generate a real VPC DescribeVpcs client and page paginator

## English

### Problem and evidence

The first generator only covers selected ECS/STS shapes. VPC DescribeVpcs introduces meaningful absent/false boolean filters, repeatList tag objects, nested responses and page-only pagination. Source: https://help.aliyun.com/en/vpc/developer-reference/api-vpc-2016-04-28-describevpcs. Design and acceptance: docs/generator-expansion.md (written before code).

### Scope

Pin Vpc 2016-04-28 DescribeVpcs protocol facts/provenance; add a reviewed overlay, generated selected models/client/mock API/Examples and page-only paginator. Add prose-only tag/paging validation and manually reviewed VPC endpoints for the existing five public regions. Maintain label catalog/forms/classification and bilingual support/usage documentation.

### Dependencies

#24

### Acceptance criteria

- [ ] Signed offline fixtures assert bool omission/false/true, Tag.N.Key/Value including empty value, nested CIDR/tag/vSwitch response decoding and int64 owner IDs. Verify input ownership during middleware/retry, page traversal/defaults/errors/overflow/cancellation, APIError wrapping, endpoint rules and runnable Examples. Pass regeneration, public docs, language, vet, all tests, Linux race and Windows CI.
- [ ] Record exact commit and successful CI evidence before closure; no live account tests.

### Affected areas

vpc, endpoints, tools

## 中文

首版生成器仅覆盖 ECS/STS 子集，VPC 带来布尔缺失/false、tag 对象数组、嵌套响应和纯页码需求；设计先于代码，见 docs/generator-expansion.md。依赖 #24。

固定 Vpc 2016-04-28 DescribeVpcs 元数据与来源，通过 overlay 生成模型/客户端/mock API/Example/纯页码 paginator；添加说明中的 tag/分页校验及五个核实公网端点。维护 labels、表单、分类和双语文档。离线验收签名线编码、布尔存在语义、tag 空值、嵌套响应、int64、输入所有权、重试、分页边界/取消/错误和 CI。

关闭前记录真实提交与成功 CI，不使用真实账号测试。ROA/body、基准 #20、发布为后续阶段。
