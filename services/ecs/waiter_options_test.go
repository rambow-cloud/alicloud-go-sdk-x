package ecs_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/rambow-cloud/alicloud-go-sdk-x/sdktest"
	"github.com/rambow-cloud/alicloud-go-sdk-x/services/ecs"
	"github.com/rambow-cloud/alicloud-go-sdk-x/waiter"
)

func TestWaiterConcurrentInputAndOptionIsolation(t *testing.T) {
	ready := make(chan struct{}, 2)
	release := make(chan struct{})
	results := make(chan error, 2)
	registrations := []func(*ecs.Options){func(o *ecs.Options) { o.Region = "base" }}
	api := statusFunc(func(_ context.Context, in *ecs.DescribeInstanceStatusInput, opts ...func(*ecs.Options)) (*ecs.DescribeInstanceStatusOutput, error) {
		var config ecs.Options
		for _, f := range opts {
			f(&config)
		}
		id := in.InstanceIDs[0]
		expected := "base"
		if id == "i-a" {
			expected = "override"
		}
		if config.Region != expected || in.PageNumber != 1 || in.PageSize != 50 {
			return nil, fmt.Errorf("unexpected options for %s", id)
		}
		in.InstanceIDs[0] = "mutated"
		opts[0] = func(o *ecs.Options) { o.Region = "leaked" }
		ready <- struct{}{}
		<-release
		return &ecs.DescribeInstanceStatusOutput{InstanceStatuses: []ecs.InstanceStatus{{InstanceID: id, Status: "Running"}}}, nil
	})
	w, err := ecs.NewInstanceRunningWaiter(api, func(o *ecs.InstanceRunningWaiterOptions) { o.ClientOptions = registrations })
	if err != nil {
		t.Fatal(err)
	}
	registrations[0] = func(o *ecs.Options) { o.Region = "caller mutation" }
	for _, id := range []string{"i-a", "i-b"} {
		go func(id string) {
			input := &ecs.DescribeInstanceStatusInput{InstanceIDs: []string{id}}
			var overrides []func(*ecs.InstanceRunningWaiterOptions)
			if id == "i-a" {
				overrides = append(overrides, func(o *ecs.InstanceRunningWaiterOptions) {
					o.ClientOptions[0] = func(c *ecs.Options) { c.Region = "override" }
				})
			}
			out, err := w.WaitForOutput(context.Background(), input, time.Minute, overrides...)
			if err == nil && (input.InstanceIDs[0] != id || out == nil || out.InstanceStatuses[0].InstanceID != id) {
				err = errors.New("wait input/output leaked")
			}
			results <- err
		}(id)
	}
	for range 2 {
		select {
		case <-ready:
		case err := <-results:
			t.Fatalf("wait failed before poll: %v", err)
		case <-time.After(5 * time.Second):
			t.Fatal("concurrent waits did not reach polling")
		}
	}
	close(release)
	for range 2 {
		if err := <-results; err != nil {
			t.Fatal(err)
		}
	}
}

func TestWaiterRetryableCustomizationAndErrorCauses(t *testing.T) {
	transient := errors.New("transient fetch")
	rejected := errors.New("acceptor rejection")
	clock := sdktest.NewClock(time.Unix(0, 0))
	calls := 0
	api := statusFunc(func(_ context.Context, in *ecs.DescribeInstanceStatusInput, opts ...func(*ecs.Options)) (*ecs.DescribeInstanceStatusOutput, error) {
		calls++
		if in.InstanceIDs[0] != "i" {
			t.Error("acceptor mutated poll input")
		}
		if calls == 1 {
			return nil, transient
		}
		return &ecs.DescribeInstanceStatusOutput{InstanceStatuses: []ecs.InstanceStatus{{InstanceID: "i", Status: "CustomReady"}}}, nil
	})
	w, err := ecs.NewInstanceRunningWaiter(api, func(o *ecs.InstanceRunningWaiterOptions) { o.Now = clock.Now; o.Sleep = clock.Sleep })
	if err != nil {
		t.Fatal(err)
	}
	input := &ecs.DescribeInstanceStatusInput{InstanceIDs: []string{"i"}}
	out, err := w.WaitForOutput(context.Background(), input, time.Minute, func(o *ecs.InstanceRunningWaiterOptions) {
		o.MinDelay = 2 * time.Second
		o.Retryable = func(ctx context.Context, in *ecs.DescribeInstanceStatusInput, out *ecs.DescribeInstanceStatusOutput, err error) (bool, error) {
			if _, ok := ctx.Deadline(); !ok {
				t.Error("unbounded acceptor context")
			}
			in.InstanceIDs[0] = "acceptor mutation"
			return errors.Is(err, transient), nil
		}
	})
	if err != nil || out == nil || calls != 2 || clock.Now().Sub(time.Unix(0, 0)) != 2*time.Second || input.InstanceIDs[0] != "i" {
		t.Fatal(out, err, calls)
	}
	// The invocation override must not persist: CustomReady fails the reviewed default.
	if err = w.Wait(context.Background(), input, time.Minute); !errors.Is(err, waiter.ErrFailure) {
		t.Fatal("override persisted", err)
	}
	out, err = w.WaitForOutput(context.Background(), input, time.Minute, func(o *ecs.InstanceRunningWaiterOptions) {
		o.Retryable = func(context.Context, *ecs.DescribeInstanceStatusInput, *ecs.DescribeInstanceStatusOutput, error) (bool, error) {
			return false, rejected
		}
	})
	if out != nil || !errors.Is(err, rejected) || !errors.Is(err, waiter.ErrFailure) {
		t.Fatal("lost acceptor error", out, err)
	}
	failure := statusFunc(func(context.Context, *ecs.DescribeInstanceStatusInput, ...func(*ecs.Options)) (*ecs.DescribeInstanceStatusOutput, error) {
		return nil, transient
	})
	w, err = ecs.NewInstanceRunningWaiter(failure, func(o *ecs.InstanceRunningWaiterOptions) { o.Now = clock.Now; o.Sleep = clock.Sleep })
	if err != nil {
		t.Fatal(err)
	}
	err = w.Wait(context.Background(), input, 2*time.Second, func(o *ecs.InstanceRunningWaiterOptions) {
		o.Retryable = func(context.Context, *ecs.DescribeInstanceStatusInput, *ecs.DescribeInstanceStatusOutput, error) (bool, error) {
			return true, nil
		}
	})
	if !errors.Is(err, waiter.ErrTimeout) || !errors.Is(err, transient) {
		t.Fatal("lost retryable fetch cause", err)
	}
	err = w.Wait(context.Background(), input, time.Minute, func(o *ecs.InstanceRunningWaiterOptions) {
		o.Retryable = func(context.Context, *ecs.DescribeInstanceStatusInput, *ecs.DescribeInstanceStatusOutput, error) (bool, error) {
			return false, nil
		}
	})
	if !errors.Is(err, waiter.ErrFailure) || !errors.Is(err, transient) {
		t.Fatal("failed fetch became success", err)
	}
}

func TestWaiterValidationCancellationAndDuplicateStates(t *testing.T) {
	calls := 0
	api := statusFunc(func(context.Context, *ecs.DescribeInstanceStatusInput, ...func(*ecs.Options)) (*ecs.DescribeInstanceStatusOutput, error) {
		calls++
		return &ecs.DescribeInstanceStatusOutput{InstanceStatuses: []ecs.InstanceStatus{{InstanceID: "i", Status: "Running"}, {InstanceID: "i", Status: "Running"}}}, nil
	})
	w, err := ecs.NewInstanceRunningWaiter(api)
	if err != nil {
		t.Fatal(err)
	}
	for _, input := range []*ecs.DescribeInstanceStatusInput{nil, {}, {InstanceIDs: []string{"i", "i"}}, {InstanceIDs: []string{""}}, {InstanceIDs: []string{"i"}, PageNumber: 2}} {
		if err = w.Wait(context.Background(), input, time.Minute); err == nil || calls != 0 {
			t.Fatal("invalid input polled", err)
		}
	}
	input := &ecs.DescribeInstanceStatusInput{InstanceIDs: []string{"i"}}
	if err = w.Wait(context.Background(), input, time.Minute, nil); err == nil || calls != 0 {
		t.Fatal("nil override accepted")
	}
	if err = w.Wait(context.Background(), input, time.Minute, func(o *ecs.InstanceRunningWaiterOptions) { o.MinDelay = -1 }); err == nil || calls != 0 {
		t.Fatal("negative delay polled")
	}
	if err = w.Wait(context.Background(), input, time.Minute); !errors.Is(err, waiter.ErrFailure) {
		t.Fatal("duplicate state accepted", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	err = w.Wait(ctx, input, time.Minute, func(o *ecs.InstanceRunningWaiterOptions) {
		o.Retryable = func(context.Context, *ecs.DescribeInstanceStatusInput, *ecs.DescribeInstanceStatusOutput, error) (bool, error) {
			cancel()
			return false, nil
		}
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatal("acceptor swallowed cancellation", err)
	}
	for _, f := range []func(*ecs.InstanceRunningWaiterOptions){nil, func(o *ecs.InstanceRunningWaiterOptions) { o.MinDelay = 6 * time.Second }, func(o *ecs.InstanceRunningWaiterOptions) { o.ClientOptions = []func(*ecs.Options){nil} }} {
		if _, err = ecs.NewInstanceRunningWaiter(api, f); err == nil {
			t.Fatal("invalid constructor option accepted")
		}
	}
}

func ExampleInstanceRunningWaiter() {
	api := statusFunc(func(_ context.Context, in *ecs.DescribeInstanceStatusInput, _ ...func(*ecs.Options)) (*ecs.DescribeInstanceStatusOutput, error) {
		return &ecs.DescribeInstanceStatusOutput{InstanceStatuses: []ecs.InstanceStatus{{InstanceID: in.InstanceIDs[0], Status: "Running"}}}, nil
	})
	running, err := ecs.NewInstanceRunningWaiter(api)
	if err != nil {
		panic(err)
	}
	if err = running.Wait(context.Background(), &ecs.DescribeInstanceStatusInput{InstanceIDs: []string{"i-a"}}, time.Minute); err != nil {
		panic(err)
	}
	out, err := running.WaitForOutput(context.Background(), &ecs.DescribeInstanceStatusInput{InstanceIDs: []string{"i-b"}}, time.Minute)
	if err != nil {
		panic(err)
	}
	fmt.Println(out.InstanceStatuses[0].InstanceID)
	// Output: i-b
}
