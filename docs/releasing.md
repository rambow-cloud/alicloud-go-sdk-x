# 发布与 pkg.go.dev

公开 module：`github.com/rambow-cloud/alicloud-go-sdk-x`。使用标准 MIT LICENSE。
本次初始化不创建版本标签或 GitHub Release；首次发布仍需检查实际功能范围。

## 版本发布

1. 在 issue 中确定发布范围；检查 CI、LICENSE、module 路径、导入和文档。
2. 更新 README 与版本说明，列出已支持操作、未支持内容和破坏性变更。
3. 首个初始化预览可用 `v0.1.0`；仅在维护者明确批准发布后创建并 push tag。
   标签索引后不可复用或改写；错误版本使用新版本及必要的 retract 修正。
4. v0 阶段 API 可以变动但必须记录；v1 承诺兼容；v2 及以后遵循 Go module
   主版本路径规则。不要在未支持产品调用时宣称可替代官方 SDK。

## 索引与浏览器验收

pkg.go.dev 从 Go Module Mirror 与 Index 获取源码，注释并不需要单独上传。
公开仓库和可访问的 module 是前提；有版本标签能提高可预测性。

打开 <https://pkg.go.dev/github.com/rambow-cloud/alicloud-go-sdk-x>。
若尚未收录，点击 Request，等待索引后重新打开。
随后打开 <https://pkg.go.dev/github.com/rambow-cloud/alicloud-go-sdk-x/credentials>。

确认：module 路径、版本、MIT 许可证、Overview、APIError 字段说明、credentials
包说明、Provider 的取消/并发契约、静态凭据 Example 与可展开的输出。
README badge 指向文档入口；其存在不代表已经被索引。初始化阶段页面可能尚未出现。

无需 curl 验证页面。若索引失败，按网页提示检查公开可见性、module 路径、支持的 Go
版本和许可证，修复根因后再请求；不要重复点击代替诊断。

参考：[索引与文档说明](https://pkg.go.dev/about#adding-a-package)、
[LICENSE 检测规则](https://pkg.go.dev/license-policy)、
[原生 Go 文档语法](https://go.dev/doc/comment)。
