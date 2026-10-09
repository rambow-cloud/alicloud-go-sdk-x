# 原生页码字段类型

[English](native-page-widths.md)

- 对应 #88，扩展公共 Darabonba 策略生成器，不手写服务方法。
- 添加策略前审核下列固定 DSL。操作读取已有属性或授权关系，重试仍须调用方显式启用。
- 页码、页大小输入支持十进制字符串和 int64，响应保持原生类型；不引入 NextToken。
- 响应页码字段可以与输入不同，但必须显式审核，例如 `Page` 与 `PageNumber`。仍校验页大小及总数；可选页码值允许缺失，存在时须与请求对应。
- 数字先解析再转换，游标同时受平台 int 范围和原生输入类型限制，在溢出前停止。
- 输入和选项使用独占副本；失败、取消、无效响应均不能推进游标。

## 已审核操作

| 产品 | 操作 | 输入页码/页大小 | 响应页码/页大小/总数 | 列表路径 | 默认 / 最大 |
| --- | --- | --- | --- | --- | --- |
| ECS | DescribeInstanceAutoRenewAttribute | string | int32 | InstanceRenewAttributes.InstanceRenewAttribute | 10 / 100 |
| ECS | DescribeInstanceMaintenanceAttributes | int64 | int32 | MaintenanceAttributes.MaintenanceAttribute | 10 / 100 |
| VPC | DescribeEcGrantRelation | int64 | Page: int32 / int32 / int32 | EcGrantRelations | 10 / 50 |
| VPC | DescribeGrantRulesToEcr | int64 | int32 | EcrGrantRules | 10 / 50 |

- 固定来源提交为 `ec489e5c3deae95496daae2b41503ac58b221adb`：[ECS 自动续费属性](https://github.com/aliyun/alibabacloud-sdk/blob/ec489e5c3deae95496daae2b41503ac58b221adb/ecs-20140526/main.tea#L24076)、[ECS 运维属性](https://github.com/aliyun/alibabacloud-sdk/blob/ec489e5c3deae95496daae2b41503ac58b221adb/ecs-20140526/main.tea#L24493)、[VPC EC 授权](https://github.com/aliyun/alibabacloud-sdk/blob/ec489e5c3deae95496daae2b41503ac58b221adb/vpc-20160428/main.tea#L22219)、[VPC ECR 授权](https://github.com/aliyun/alibabacloud-sdk/blob/ec489e5c3deae95496daae2b41503ac58b221adb/vpc-20160428/main.tea#L24469)。

## 实现前确定的验收

- 四个操作各有生成的两页 Example，准确保留原生线字段值。
- 独立协议夹具验证默认值、指定起始页、字符串语法和范围拒绝、宽整数不截断、响应无页码、元数据不一致、终止页和空页、取消、失败游标不变、输入所有权。
- 无效策略类型或路径在写入前失败，操作和模型发现不变。
- 执行 Node 22 前端测试及检查、生成器 check/product-check、格式、doccheck、vet、Go 测试，最终提交通过 Linux race 与 Windows CI。
- 更新配对指南及确定性策略覆盖，其余候选仍未审核，不宣称已完成真实云验收。

## 实现证据

- 四个适配器及其外部两页 Example 均通过。ECS 现有 14 个分页器，审核了 380 个操作中的 16 个；VPC 有 15 个分页器，审核了 403 个操作中的 15 个。Waiter 数量不变。
- 公共数字解析保留 int64、字符串输入和 int32 响应，以及显式的 `Page` 别名。无效策略不会修改既有输出，分页策略不改变模型输出。
- Node 22.21.1 前端检查、87 项测试、两项生成器检查、16 个公共包的 doccheck、vet、格式及语言和链接检查通过。
- 所有 Go 包均通过。首轮发现两项过时的测试预期，修正后完整生成器测试包用时 84.398 秒并通过；运行时和服务测试未发生失败。
- 这四项操作尚未进行真实云验证。最终提交的 CI 证据在实现 PR 合并前记录；其余候选继续由 #88 跟踪。
