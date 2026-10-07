package ecs_test

import (
	"context"
	"errors"
	"github.com/rambow-cloud/alicloud-go-sdk-x/sdktest"
	"github.com/rambow-cloud/alicloud-go-sdk-x/services/ecs"
	"github.com/rambow-cloud/alicloud-go-sdk-x/waiter"
	"testing"
	"time"
)

type statusFunc func(context.Context, *ecs.DescribeInstanceStatusInput, ...func(*ecs.Options)) (*ecs.DescribeInstanceStatusOutput, error)

func (f statusFunc) DescribeInstanceStatus(ctx context.Context, in *ecs.DescribeInstanceStatusInput, opts ...func(*ecs.Options)) (*ecs.DescribeInstanceStatusOutput, error) {
	return f(ctx, in, opts...)
}
func TestRunningWaiterRequiresEveryInstanceAndCopiesInput(t *testing.T) {
	clock := sdktest.NewClock(time.Unix(0, 0))
	calls := 0
	api := statusFunc(func(ctx context.Context, in *ecs.DescribeInstanceStatusInput, opts ...func(*ecs.Options)) (*ecs.DescribeInstanceStatusOutput, error) {
		calls++
		if in.InstanceIDs[0] != "i-a" || in.PageSize != 50 {
			t.Error("waiter input")
		}
		in.InstanceIDs[0] = "mutated"
		out := &ecs.DescribeInstanceStatusOutput{}
		if calls > 1 {
			out.InstanceStatuses = []ecs.InstanceStatus{{InstanceID: "i-a", Status: "Running"}}
		}
		if calls > 2 {
			out.InstanceStatuses = append(out.InstanceStatuses, ecs.InstanceStatus{InstanceID: "i-b", Status: "Running"})
		}
		return out, nil
	})
	input := &ecs.DescribeInstanceStatusInput{InstanceIDs: []string{"i-a", "i-b"}}
	w, err := ecs.NewInstanceRunningWaiter(api, input, waiter.Options{Now: clock.Now, Sleep: clock.Sleep})
	if err != nil {
		t.Fatal(err)
	}
	input.InstanceIDs[0] = "caller-change"
	_, err = w.Wait(context.Background(), time.Minute)
	if err != nil || calls != 3 {
		t.Fatal(err, calls)
	}
}
func TestRunningWaiterTimeoutUnknownAndErrors(t *testing.T) {
	for _, state := range []string{"Starting", "Unknown"} {
		clock := sdktest.NewClock(time.Unix(0, 0))
		api := statusFunc(func(context.Context, *ecs.DescribeInstanceStatusInput, ...func(*ecs.Options)) (*ecs.DescribeInstanceStatusOutput, error) {
			return &ecs.DescribeInstanceStatusOutput{InstanceStatuses: []ecs.InstanceStatus{{InstanceID: "i", Status: state}}}, nil
		})
		w, _ := ecs.NewInstanceRunningWaiter(api, &ecs.DescribeInstanceStatusInput{InstanceIDs: []string{"i"}}, waiter.Options{Now: clock.Now, Sleep: clock.Sleep})
		_, err := w.Wait(context.Background(), 2*time.Second)
		expected := waiter.ErrTimeout
		if state == "Unknown" {
			expected = waiter.ErrFailure
		}
		if !errors.Is(err, expected) {
			t.Fatal(err)
		}
	}
	c := clientFor(sdktest.NewTransport(sdktest.Step{Body: `{"InstanceStatuses":{"InstanceStatus":[{"InstanceId":"i","Status":"Running"}]}}`}))
	w, _ := ecs.NewInstanceRunningWaiter(c, &ecs.DescribeInstanceStatusInput{InstanceIDs: []string{"i"}}, waiter.Options{})
	if _, err := w.Wait(context.Background(), time.Minute); err != nil {
		t.Fatal(err)
	}
	if _, err := ecs.NewInstanceRunningWaiter(c, &ecs.DescribeInstanceStatusInput{}, waiter.Options{}); err == nil {
		t.Fatal("empty IDs accepted")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := w.Wait(ctx, time.Minute)
	if !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
}
