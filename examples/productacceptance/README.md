# Product consumer acceptance

[中文](README.zh-CN.md)

- Separate Go 1.27 module. Imports public SDK packages only; local replace selects the reviewed checkout.
- Uses handwritten synthetic responses and small mocks. No cloud account, credentials or network connection.
- The SDK operations, models, paginators and waiter are generated from pinned official Darabonba and reviewed policies.
- `imageIDs` and `instanceIDs` contain business projections and the standard HasMorePages/NextPage loop. They never advance cursors by hand.

```powershell
go -C examples/productacceptance test ./...
go -C examples/productacceptance run .
```

## ECS use

- Construct with config.LoadDefaultConfig and ecs.NewFromConfig. A native temporary Profile or cached STS provider can be shared by product clients.
- Use ecs.NewDescribeImagesPaginator for native page numbers.
- Use ecs.NewDescribeInstancesPaginator with MaxResults/NextToken for token mode, or PageSize/PageNumber for page mode. Do not mix modes.
- Each paginator snapshots input and accepts per-page operation options. It is single-consumer. A failed/canceled fetch leaves the cursor unchanged.
- Use ecs.NewInstanceRunningWaiter and Wait/WaitForOutput with all desired InstanceIDs and a positive duration. No application polling loop is needed.
- The waiter is reusable concurrently. Missing IDs retry; duplicate/unknown observations fail. Timeout supports errors.Is with waiter.ErrTimeout and context.DeadlineExceeded.
- Mock only ecs.DescribeImagesAPI, ecs.DescribeInstancesAPI or ecs.DescribeInstanceStatusAPI needed by the application.
- Inspect errors with errors.As for alicloud.APIError/OperationError; preserve errors.Is cancellation. Message can contain secrets and requires explicit handling.
- Retry is opt-in. Reviewed read policy permits transient retry; arbitrary writes stay conservative.
- Inject telemetry/otel middleware with an application-owned provider. Exported attributes omit request/response bodies, credentials and error messages.

## Scope and evidence

- [ECS acceptance plan](../../docs/ecs-product-acceptance.md) declares required cases before execution.
- Go test durations describe automation, not independent human UX or performance.
- Full generated inventory, selected offline contracts and historical selected-field live reads remain separate evidence.
- The local replace is an acceptance mechanism. A published consumer selects an immutable SDK version after #61.
- Runtime/test code uses the project's MIT license. Generated product packages retain Apache-2.0 provenance and their LICENSE/NOTICE.
