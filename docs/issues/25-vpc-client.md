# [Feature]: Generate a real VPC DescribeVpcs client and page paginator

- GitHub issue: #25.

# [Feature]: Generate a real VPC DescribeVpcs client and page paginator

### Problem and evidence

- The first generator only covers selected ECS/STS shapes.
- VPC DescribeVpcs introduces meaningful absent/false boolean filters, repeatList tag objects, nested responses and page-only pagination.
- Source: https://help.aliyun.com/en/vpc/developer-reference/api-vpc-2016-04-28-describevpcs.
- Design and acceptance: docs/generator-expansion.md (written before code).

### Scope

- Pin Vpc 2016-04-28 DescribeVpcs protocol facts/provenance; add a reviewed overlay, generated selected models/client/mock API/Examples and page-only paginator.
- Add prose-only tag/paging validation and manually reviewed VPC endpoints for the existing five public regions.
- Maintain label catalog/forms/classification and bilingual support/usage documentation.

### Dependencies

- #24

### Acceptance criteria

- [ ] Signed offline fixtures assert bool omission/false/true, Tag.N.Key/Value including empty value, nested CIDR/tag/vSwitch response decoding and int64 owner IDs. Verify input ownership during middleware/retry, page traversal/defaults/errors/overflow/cancellation, APIError wrapping, endpoint rules and runnable Examples. Pass regeneration, public docs, language, vet, all tests, Linux race and Windows CI.
- [ ] Record exact commit and successful CI evidence before closure; no live account tests.

### Affected areas

- vpc, endpoints, tools
