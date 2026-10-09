# Unified waiters

[中文](waiters.zh-CN.md)

- #37 emits this contract for full-DSL `service/ecs` using reviewed sparse policies; the earlier services/ecs reference is removed in #81.
- Import the chosen package explicitly.
- See [capability policy](capability-policy.md) for coverage and source binding.

- Construct `ecs.NewInstanceRunningWaiter(api, optFns...)` once, then call `Wait(ctx, input, maxWait, optFns...) error` or `WaitForOutput(ctx, input, maxWait, optFns...) (*ecs.DescribeInstanceStatusOutput, error)`.
- Concurrent waits have private inputs, polling state and option snapshots.
- Inputs and option registrations are copied per invocation and poll; callers must not mutate their input concurrently.
- APIs, transports, clocks and callbacks remain shared and must be concurrency safe.
- Constructor validation returns an error, an intentional early-v0 difference from AWS alongside the similar calling convention.

- `InstanceRunningWaiterOptions` supplies MinDelay/MaxDelay (zero selects 1s/5s), ClientOptions applied to every poll, deterministic Now/Sleep seams and Retryable.
- Invocation overrides do not persist.
- Nil callbacks and invalid delays fail before polling.
- Retryable receives the bounded poll context, a fresh input copy, output and fetch error: true retries, false succeeds, and an error fails with its inspectable cause.
- Nil retains the reviewed default.
- A failed fetch cannot become success; cancellation/expiry are checked before and after acceptance.
- Callbacks must not block.
- Changing acceptance is an explicit caller policy, not a new service guarantee.

- Each invocation requires 1..50 nonempty distinct IDs and a first-page request.
- It polls page 1 at size 50.
- The default requires every requested ID present and Running.
- Missing IDs, empty results and Pending/Starting/Stopping/Stopped retry.
- Unknown states, duplicate response IDs and API errors fail.
- Absence never implies success or deletion.
- Larger sets require separate waits.
- A custom acceptor retains these input/page bounds.

```go
running, err := ecs.NewInstanceRunningWaiter(client)
if err != nil { return err }
input := &ecs.DescribeInstanceStatusInput{InstanceIDs: []string{"i-example"}}
if err := running.Wait(ctx, input, time.Minute); err != nil { return err }
out, err := running.WaitForOutput(ctx, input, time.Minute)
```

- The shared `waiter.New[T]` engine combines a context-aware fetcher with Retry/Success/ Failure acceptance.
- Its Wait continues returning `(T, error)`.
- A positive maximum duration includes fetches, operation retries and capped exponential sleeps.
- Caller cancellation/deadlines pass through; own expiry supports ErrTimeout and context.DeadlineExceeded, preserving the last retryable fetch error.
- FailureError supports ErrFailure and a fetch/custom-acceptor cause.
- Failed generated waits return nil output.
- Offline examples and regression fixtures cover behavior; no live cloud validation is claimed.

- Migration: the old generated constructor bound input and waiter.Options, and Wait returned output.
- Move input/maxWait to each call, configure dedicated options through functional callbacks, and use WaitForOutput when the response is needed.
- The generic engine API is preserved.
- See [the remediation path](aws-style-remediation.md).
