# STS runtime comparison

[中文](runtime-comparison.zh-CN.md)

- Issue: #20. [Run the isolated workloads](../tools/benchmarks/README.md).
- [Raw results](../tools/benchmarks/results/windows-go1.27.1.json), recorded 2026-10-09.
- SDK revision: `863001133b5d7135c3ef48a38decc11505ed26bf`. Later integration commits were not measured.
- Host: Windows/amd64, Intel Core i5-12500H, Go 1.27.1, CGO disabled, GOAMD64 v1.
- Official versions: STS v2.1.0, OpenAPI v2.1.13, Tea v1.3.13, credentials-go v1.4.5.

| Metric | This SDK | Official SDK |
|---|---:|---:|
| Imported packages, including standard library | 208 | 236 |
| External packages / modules | 0 / 0 | 25 / 14 |
| Cold build, seconds | 10.524 | 13.544 |
| Warm build, seconds | 1.770 | 2.206 |
| Binary bytes | 10,176,512 | 12,431,872 |
| Median nanoseconds per identity call | 14,512 | 62,135 |
| Median bytes allocated per call | 15,344 | 28,211 |
| Allocations per call, all three samples | 145 | 387 |

## What was measured

- One generated STS `GetCallerIdentity` call with synthetic temporary credentials and the same injected response transport.
- Signing, encoding, transport dispatch and response decoding are included. Client construction is outside the runtime loop.
- Three serial one-second runtime samples. Raw sample values and complete import closures are in the report.
- Cold and warm builds each have one observation. Separate empty caches include standard-library compilation and exclude module downloads.
- Both binaries use `go build -trimpath`, without linker stripping. Timing order is this SDK, then official SDK.
- No network, account, credentials, resource data or global-cache deletion is involved.

## Limits

- These results describe one STS workload on one Windows host. They do not predict other services, live latency or production throughput.
- Build timing depends on machine load, storage and antivirus. No statistical significance or universal speed claim is made.
- The official method has a different context signature. Cancellation, retries, service coverage and credential renewal are separate acceptance tests.
- CI checks workload behavior and vet on Linux and Windows. It does not enforce host-specific timing or size thresholds.
- Official dependencies stay in a separate Go module. The SDK runtime keeps its standard-library-only core.
