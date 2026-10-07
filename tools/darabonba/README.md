# Darabonba build tools / Darabonba 构建工具

## English

The [product roadmap](../../docs/product-generator-roadmap.md) overrides conflicting
older per-operation prerequisites. #31 remains a five-operation bridge; #34 adds
[source normalization](../../docs/source-normalization.md), and #35 discovers
complete products without per-operation snapshots/overlays. Its offline command and
coverage/IR contract are in [product discovery](../../docs/product-discovery.md).
The #36 [batch Go backend](../../docs/batch-go-emission.md) consumes this complete IR
without legacy overlays. Run `go run ./internal/cmd/sdkgen product-generate` and
read-only `go run ./internal/cmd/sdkgen product-check` from the root. Outputs use
`service/`; the legacy bridge still uses `services/`. Product capabilities are #37.

Issue [#31](https://github.com/rambow-cloud/alicloud-go-sdk-x/issues/31) replaces the
metadata-only production frontend with real official product DSL. Requires Node 22
and Go 1.27. The official parser 2.2.1 performs syntax and imported-module semantic
analysis. Official repo-client 1.0.3 resolves explicit network imports; tar 7.5.22
extracts checked archives. Exact packages and transitive integrity hashes are locked
in package-lock.json. Node/Tea dependencies serve build tooling only, outside go.mod.

From the repository root:

```sh
cd tools/darabonba
npm ci --ignore-scripts --no-audit --no-fund
cd ../..
node tools/darabonba/frontend.cjs generate
go run ./internal/cmd/sdkgen generate
node tools/darabonba/frontend.cjs check
node tools/darabonba/discovery.cjs generate
node tools/darabonba/discovery.cjs report ecs
node tools/darabonba/discovery.cjs check
node --test tools/darabonba/frontend.test.cjs tools/darabonba/normalization.test.cjs tools/darabonba/discovery.test.cjs
go run ./internal/cmd/sdkgen check
```

Installing tools requires the package registry; after installation, projection,
generation, checks and tests use local files. Normal Go consumers need no Node.
Go generation consumes checksum-pinned projections, verifies the source lock and
cross-checks metadata/policy; CI independently re-parses the DSL on Linux and Windows.
Both frontends preflight every product before output writes. OS write failures can
leave a partial update; resolve the failure and regenerate.

Only explicit import accesses upstream services:

```sh
node tools/darabonba/import.cjs
node tools/darabonba/import-canonical.cjs
```

The product commit is a constant in import.cjs. Imported Teafile wildcard specs are
preserved, while manifest.json and .libraries.json lock actual versions and local
paths. Import may resolve newer registry modules; review source/module/license
changes and update metadata/darabonba-decisions.json under an issue before projection.
The importer writes source artifacts as it fetches them; network failure may require
repairing the source lock. Generation never fetches or refreshes dependencies.

Recognized functions are async operationWithOptions(request, runtime), guarded
direct query bindings, OpenApiRequest query encoding, constant OpenApi.Params and
callApi handoff. They lower to our runtime instead of emitting the imported Tea
program. Selected models support bounded named/anonymous objects, arrays and scalars.
Unknown behavior, language overrides, new model attributes, incomplete bindings,
dynamic parameters and unsupported selected shapes fail. Native pagination, waiters,
Go names, presence, redaction and idempotency remain reviewed overlay/runtime policy.
This is a bounded frontend for five operations, not a general Darabonba compiler.

Read [migration](../../docs/darabonba-migration.md),
[decisions](../../docs/darabonba-decisions.md) and
[sources/licenses](../../sources/darabonba/README.md) before adding operations.

## 中文

[产品路线](../../docs/product-generator-roadmap.md) 优先于冲突旧逐操作前置要求。
#31 仍为五操作兼容桥；#34 增加[来源规范化](../../docs/source-normalization.md)，
#35 已提供无需逐操作快照/overlay 的完整产品发现；离线命令与覆盖/IR 契约见
[产品发现](../../docs/product-discovery.md)。
#36 [批量 Go 后端](../../docs/batch-go-emission.md) 无旧 overlay 地消费完整 IR。
在根目录执行 `go run ./internal/cmd/sdkgen product-generate` 和只读
`go run ./internal/cmd/sdkgen product-check`，输出到 `service/`；旧桥仍使用 `services/`。
产品能力策略属于 #37。

Issue [#31](https://github.com/rambow-cloud/alicloud-go-sdk-x/issues/31) 将生产前端
从纯元数据改为真实官方产品 DSL。要求 Node 22、Go 1.27；官方 parser 2.2.1 执行语法
及导入模块语义检查，官方 repo-client 1.0.3 负责显式联网解析，tar 7.5.22 解包经校验
的归档。package-lock.json 固定直接版本及传递依赖完整性；Node/Tea 依赖仅服务于
构建工具，不进入 go.mod。

仓库根目录按英文章节命令安装工具、导出投影、生成 Go、检查前端、运行前端测试、
检查 Go 再生成。安装需要包仓库；安装后投影、生成、检查、测试使用本地文件，普通
Go 使用者无需 Node。Go 生成消费哈希固定的投影、验证来源锁并对照元数据/策略，
CI 在 Linux/Windows 独立重解析 DSL。两层均先完成所有产品检查再写输出；系统写入
失败可能留下部分更新，解决原因后再生成。

只有英文章节两条显式 import 命令访问上游；import-canonical 固定 CLI 样本和许可证。
产品 commit 固定在 import.cjs，
Teafile 保留原始通配符，manifest.json 与 .libraries.json 固定实际模块版本及本地路径。
重新 import 可能解析到新版模块，必须在 issue 下审核源码、模块、许可变化，更新
metadata/darabonba-decisions.json 后再投影。导入随获取写源码，网络失败可能需要修复
来源锁；生成绝不联网刷新依赖。

支持模式为 async operationWithOptions(request, runtime)、判空后的直接 query 绑定、
OpenApiRequest query 编码、常量 OpenApi.Params 和 callApi 交接；降低到本项目运行时，
不输出导入的 Tea 程序。选定模型支持有界具名/匿名对象、数组及标量。不支持行为、
语言覆盖、新模型属性、漏绑、动态参数及不支持的选定结构明确失败。原生分页、waiter、
Go 命名、缺失、脱敏及幂等性仍由审核的 overlay/runtime 管理。这是五个操作的有界前端，
不是通用 Darabonba 编译器。

新增操作前阅读[迁移](../../docs/darabonba-migration.md)、
[决策](../../docs/darabonba-decisions.md)及[来源/许可](../../sources/darabonba/README.md)。
