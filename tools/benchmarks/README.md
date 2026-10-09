# SDK comparison workload

[中文](README.zh-CN.md)

- Issue: #20. This isolated module keeps official dependencies outside the SDK runtime.
- [Measured results and limits](../../docs/runtime-comparison.md).
- Go 1.27+; Node 22+ for measurement orchestration.
- Both programs make one signed STS identity call using the same synthetic response and injected transport. No network or account is used.
- Keys are fixtures. Program output is `PASS identity`; no credentials are printed.

## Run

```powershell
go -C tools/benchmarks mod download
go -C tools/benchmarks test ./...
node tools/benchmarks/compare.cjs
```

- Official pins: STS v2.1.0, OpenAPI v2.1.13, Tea v1.3.13, credentials-go v1.4.5.
- Our module uses a local replace; reports record the SDK commit, Go/OS/architecture, CPU and build settings.
- `go list -deps` compares actual import closures, including standard, project and external packages/modules.
- Cold builds use separate empty caches. Standard-library compile cost is included; module downloads are excluded.
- Warm builds reuse each cache and remove the output binary first. Build arguments match: `-trimpath`, no linker stripping.
- The order is fixed: ours, then official. Cold/warm build times are single observations and depend on disk, antivirus and host load.
- Runtime measurements use three serial one-second samples with `benchmem`. Client construction is outside the timed loop; transport, signing, encoding and response decode are included.
- The official method lacks the same context signature. This workload does not test cancellation, live latency or retry equivalence.
- Results describe this STS program on one host; they do not establish a universal performance or feature-coverage claim.
- Existing SDK docs/Examples and source correctness remain separate from benchmarks.
- Temporary binaries/caches remain under `.git/sdk-benchmark-*`; no global cache is removed or environment configuration persisted.
