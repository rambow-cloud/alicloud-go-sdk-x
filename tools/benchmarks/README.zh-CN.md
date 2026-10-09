# SDK 对比工作负载

[English](README.md)

- [测量结果与适用范围](../../docs/runtime-comparison.zh-CN.md)。

- 对应 issue #20。独立 module 将官方依赖隔离在 SDK 运行时之外。
- 要求 Go 1.27+；测量脚本使用 Node 22+。
- 两个程序各执行一次签名 STS 身份查询，使用同一模拟响应和注入的 HTTP transport，不访问网络或账号。
- 密钥均为测试值，程序仅输出 `PASS identity`，不打印凭据。

## 执行

```powershell
go -C tools/benchmarks mod download
go -C tools/benchmarks test ./...
node tools/benchmarks/compare.cjs
```

- 官方版本固定为 STS v2.1.0、OpenAPI v2.1.13、Tea v1.3.13、credentials-go v1.4.5。
- 本 SDK 通过本地 replace 使用；报告记录 SDK 提交、Go、操作系统、架构、CPU 和构建设置。
- `go list -deps` 对比实际导入闭包，分别统计标准库、本项目和外部包、module。
- 冷构建使用相互独立的空缓存，包含标准库编译时间，不包含 module 下载。
- 热构建复用各自缓存，先删除输出二进制；构建参数一致，使用 `-trimpath`，不裁剪符号。
- 执行顺序固定为本 SDK 后官方 SDK。冷、热构建各测一次，耗时受磁盘、杀毒软件和主机负载影响。
- 运行时使用三个串行的一秒采样，报告 `benchmem`；客户端构造不计入，签名、编码、模拟传输和解码计入。
- 官方方法的 context 签名不同；本测试不评价取消、真实网络延迟或重试等价性。
- 结果只描述同一主机上的这个 STS 程序，不代表普遍性能结论或功能覆盖。
- SDK 文档、Example 和来源正确性验收与基准分别记录。
- 临时二进制和缓存位于 `.git/sdk-benchmark-*`；不删除全局缓存，也不持久修改环境配置。
