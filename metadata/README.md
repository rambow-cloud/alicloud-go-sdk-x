# Pinned protocol metadata

[中文](README.zh-CN.md)

- Production manifests also pin the official parser's dsl.json projection, source lock and reviewed divergence decisions.
- Refreshing metadata alone requires re-projecting DSL and reviewing differences before Go generation; see [Darabonba tools](../tools/darabonba/README.md).

- Each product directory contains manifest.json, a reviewed overlay.json and protocol-only operation snapshots.
- Official source URLs, retrieval times, raw-source SHA-256 and snapshot SHA-256 are in the manifest.
- Generation verifies snapshot hashes offline; the raw hash is provenance, not a promise that today's remote response is identical.

- Snapshots derive structural protocol facts from Alibaba Cloud's public metadata API.
- Upstream descriptions/examples are excluded and no upstream documentation license is asserted.
- Original overlays and generated documentation follow the repository MIT LICENSE.
- See [the design and supported profile](../docs/generator.md).

- To refresh ECS explicitly (network access, replaces pinned source files):

```sh
go run ./internal/cmd/sdkgen import -product Ecs -version 2014-05-26 -package ecs -operations DescribeRegions,DescribeInstances,DescribeInstanceStatus -out metadata/ecs
```

- For STS use product Sts, version 2015-04-01, package sts and operation AssumeRole.
- For VPC use product Vpc, version 2016-04-28, package vpc and operation DescribeVpcs.
- The VPC overlay selects pointer boolean filters, Tag repeatList objects, nested response models and a page-only paginator; see [the VPC guide](../docs/vpc.md).
- Nonempty components.schemas definitions are retained for reviewed local references, without prose. `-raw-dir PATH` extracts already downloaded operation-name.json files offline and records their modification time as retrieval time.
- Review the metadata and overlay diff under an issue before regenerating.
- Product/style selection is reviewed: the operation API does not itself carry product Info/style; import supports RPC only.
- Do not import credentials, account responses, upstream prose or guessed endpoint rules.
