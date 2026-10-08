# SDK research

[中文](research.zh-CN.md)

- Research date: 2026-10-07.
- Reports concern specific upstream versions and are not local reproductions or claims that every latest Alibaba SDK has the same issue.
- V1 ended support on 2025-03-01 and was archived on 2026-02-11: https://github.com/aliyun/alibaba-cloud-sdk-go.
- V2 products are under alibabacloud-go.

| Evidence                                                                                              | Design response                   |
| ----------------------------------------------------------------------------------------------------- | --------------------------------- |
| [V2 #72](https://github.com/aliyun/alibabacloud-go-sdk/issues/72): Config.HttpClient report           | injectable net/http               |
| [OpenAPI #177](https://github.com/aliyun/darabonba-openapi/issues/177): tea dependency compile report | standard-library core             |
| [V2 #68](https://github.com/aliyun/alibabacloud-go-sdk/issues/68): header state mutation report       | private immutable configuration   |
| [V1 #612](https://github.com/aliyun/alibaba-cloud-sdk-go/issues/612): endpoint resolver request       | public resolver contract          |
| [V1 #669](https://github.com/aliyun/alibaba-cloud-sdk-go/issues/669): NextToken coverage report       | unified engine, product policies  |
| [OpenAPI #221](https://github.com/aliyun/darabonba-openapi/issues/221): swag comment parser report    | native Go comments and doc checks |
| [OpenAPI #225](https://github.com/aliyun/darabonba-openapi/issues/225): dependency license report     | license/provenance review         |

- ECS generated code illustrates pointer-heavy configuration and methods without explicit ctx in the inspected examples: https://github.com/alibabacloud-go/ecs-20140526/blob/master/client/client.go.
- This does not prove all V2 lacks context.
- OSS V2 provides good ctx-first patterns: https://github.com/aliyun/alibabacloud-oss-go-sdk-v2/blob/master/DEVGUIDE-CN.md.
- Do not generalize old migration docs to every current product or infer maturity from handwritten line count.

- Use AWS v2's middleware, small testing interfaces and provider/cache contracts as design references, preserving Alibaba-specific protocols.
- Complete the shared runtime before generation.
- Do not claim performance/size improvements without pinned, reproducible benchmarks.
