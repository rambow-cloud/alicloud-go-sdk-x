package ecs_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/rambow-cloud/alicloud-go-sdk-x/sdktest"
	"github.com/rambow-cloud/alicloud-go-sdk-x/service/ecs"
	"github.com/rambow-cloud/alicloud-go-sdk-x/waiter"
)

var timeOrigin = time.Date(2026, 10, 7, 0, 0, 0, 0, time.UTC)

type statusAPI func(context.Context, *ecs.DescribeInstanceStatusInput, ...func(*ecs.Options)) (*ecs.DescribeInstanceStatusOutput, error)

func (f statusAPI) DescribeInstanceStatus(ctx context.Context, in *ecs.DescribeInstanceStatusInput, opts ...func(*ecs.Options)) (*ecs.DescribeInstanceStatusOutput, error) {
	return f(ctx, in, opts...)
}
func observation(rows ...[2]string) *ecs.DescribeInstanceStatusOutput {
	out := &ecs.DescribeInstanceStatusOutput{InstanceStatuses: &ecs.DescribeInstanceStatusOutputInstanceStatuses{}}
	for _, row := range rows {
		out.InstanceStatuses.InstanceStatus = append(out.InstanceStatuses.InstanceStatus, ecs.DescribeInstanceStatusOutputInstanceStatusesInstanceStatus{InstanceID: pointer(row[0]), Status: pointer(row[1])})
	}
	return out
}

func TestWaiterAllIDsMissingTransitionsAndOwnedPolling(t *testing.T) {
	calls := 0
	clock := sdktest.NewClock(timeOrigin)
	in := &ecs.DescribeInstanceStatusInput{InstanceIDs: []string{"a", "b"}}
	api := statusAPI(func(ctx context.Context, owned *ecs.DescribeInstanceStatusInput, _ ...func(*ecs.Options)) (*ecs.DescribeInstanceStatusOutput, error) {
		calls++
		if owned.InstanceIDs[0] != "a" || *owned.PageNumber != 1 || *owned.PageSize != 50 {
			t.Error("poll snapshot/page changed")
		}
		owned.InstanceIDs[0] = "changed"
		switch calls {
		case 1:
			return observation([2]string{"a", "Running"}), nil
		case 2:
			return observation([2]string{"a", "Running"}, [2]string{"b", "Starting"}), nil
		default:
			return observation([2]string{"a", "Running"}, [2]string{"b", "Running"}), nil
		}
	})
	w, err := ecs.NewInstanceRunningWaiter(api, func(o *ecs.InstanceRunningWaiterOptions) { o.Now = clock.Now; o.Sleep = clock.Sleep })
	if err != nil {
		t.Fatal(err)
	}
	out, err := w.WaitForOutput(context.Background(), in, 10*time.Second)
	if err != nil || out == nil || calls != 3 || in.InstanceIDs[0] != "a" || in.PageNumber != nil {
		t.Fatal(out, err)
	}
}

func TestWaiterFailuresTimeoutOverridesAndCancellation(t *testing.T) {
	for _, out := range []*ecs.DescribeInstanceStatusOutput{nil, observation([2]string{"a", "Unknown"}), observation([2]string{"a", "Running"}, [2]string{"a", "Running"}), {InstanceStatuses: &ecs.DescribeInstanceStatusOutputInstanceStatuses{InstanceStatus: []ecs.DescribeInstanceStatusOutputInstanceStatusesInstanceStatus{{InstanceID: pointer("a")}}}}} {
		w, err := ecs.NewInstanceRunningWaiter(statusAPI(func(context.Context, *ecs.DescribeInstanceStatusInput, ...func(*ecs.Options)) (*ecs.DescribeInstanceStatusOutput, error) {
			return out, nil
		}))
		if err != nil {
			t.Fatal(err)
		}
		if err := w.Wait(context.Background(), &ecs.DescribeInstanceStatusInput{InstanceIDs: []string{"a"}}, time.Second); !errors.Is(err, waiter.ErrFailure) {
			t.Fatal("terminal state not failed", err)
		}
	}
	clock := sdktest.NewClock(timeOrigin)
	w, err := ecs.NewInstanceRunningWaiter(statusAPI(func(context.Context, *ecs.DescribeInstanceStatusInput, ...func(*ecs.Options)) (*ecs.DescribeInstanceStatusOutput, error) {
		return observation(), nil
	}), func(o *ecs.InstanceRunningWaiterOptions) { o.Now = clock.Now; o.Sleep = clock.Sleep })
	if err != nil {
		t.Fatal(err)
	}
	if err := w.Wait(context.Background(), &ecs.DescribeInstanceStatusInput{InstanceIDs: []string{"a"}}, 2*time.Second); !errors.Is(err, waiter.ErrTimeout) || !errors.Is(err, context.DeadlineExceeded) {
		t.Fatal(err)
	}
	cause := errors.New("fetch failure")
	w, err = ecs.NewInstanceRunningWaiter(statusAPI(func(context.Context, *ecs.DescribeInstanceStatusInput, ...func(*ecs.Options)) (*ecs.DescribeInstanceStatusOutput, error) {
		return nil, cause
	}))
	if err != nil {
		t.Fatal(err)
	}
	if err := w.Wait(context.Background(), &ecs.DescribeInstanceStatusInput{InstanceIDs: []string{"a"}}, time.Second, func(o *ecs.InstanceRunningWaiterOptions) {
		o.Retryable = func(context.Context, *ecs.DescribeInstanceStatusInput, *ecs.DescribeInstanceStatusOutput, error) (bool, error) {
			return false, nil
		}
	}); !errors.Is(err, waiter.ErrFailure) || !errors.Is(err, cause) {
		t.Fatal("failed fetch became success", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := w.Wait(ctx, nil, time.Second); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if err := w.Wait(context.Background(), &ecs.DescribeInstanceStatusInput{InstanceIDs: []string{"a", "a"}}, time.Second); err == nil {
		t.Fatal("duplicate input IDs accepted")
	}
}

func TestReusableWaiterConcurrentOptionsAndAcceptorIsolation(t *testing.T) {
	api := statusAPI(func(_ context.Context, in *ecs.DescribeInstanceStatusInput, opts ...func(*ecs.Options)) (*ecs.DescribeInstanceStatusOutput, error) {
		o := ecs.Options{}
		for _, f := range opts {
			f(&o)
		}
		if o.Region != "local" || len(in.InstanceIDs) != 1 {
			t.Error("options/input shared")
		}
		return observation([2]string{in.InstanceIDs[0], "Running"}), nil
	})
	w, err := ecs.NewInstanceRunningWaiter(api)
	if err != nil {
		t.Fatal(err)
	}
	var group sync.WaitGroup
	for _, id := range []string{"a", "b", "c"} {
		group.Add(1)
		go func() {
			defer group.Done()
			in := &ecs.DescribeInstanceStatusInput{InstanceIDs: []string{id}}
			err := w.Wait(context.Background(), in, time.Second, func(o *ecs.InstanceRunningWaiterOptions) {
				o.ClientOptions = []func(*ecs.Options){func(o *ecs.Options) { o.Region = "local" }}
				o.Retryable = func(_ context.Context, owned *ecs.DescribeInstanceStatusInput, _ *ecs.DescribeInstanceStatusOutput, _ error) (bool, error) {
					owned.InstanceIDs[0] = "changed"
					return false, nil
				}
			})
			if err != nil || in.InstanceIDs[0] != id {
				t.Error("wait/acceptor leak", err)
			}
		}()
	}
	group.Wait()
	var typedNil *ecs.Client
	if _, err := ecs.NewInstanceRunningWaiter(typedNil); err == nil {
		t.Fatal("typed nil waiter API")
	}
}
