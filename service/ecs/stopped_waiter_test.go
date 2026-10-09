package ecs_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/rambow-cloud/alicloud-go-sdk-x/sdktest"
	"github.com/rambow-cloud/alicloud-go-sdk-x/service/ecs"
	"github.com/rambow-cloud/alicloud-go-sdk-x/waiter"
)

func TestStoppedWaiterRequiresEveryID(t *testing.T) {
	clock := sdktest.NewClock(timeOrigin)
	calls := 0
	w, err := ecs.NewInstanceStoppedWaiter(statusAPI(func(context.Context, *ecs.DescribeInstanceStatusInput, ...func(*ecs.Options)) (*ecs.DescribeInstanceStatusOutput, error) {
		calls++
		if calls == 1 {
			return observation([2]string{"a", "Stopped"}, [2]string{"b", "Stopping"}), nil
		}
		return observation([2]string{"a", "Stopped"}, [2]string{"b", "Stopped"}), nil
	}), func(o *ecs.InstanceStoppedWaiterOptions) { o.Now = clock.Now; o.Sleep = clock.Sleep })
	if err != nil {
		t.Fatal(err)
	}
	in := &ecs.DescribeInstanceStatusInput{InstanceIDs: []string{"a", "b"}}
	if err = w.Wait(context.Background(), in, time.Minute); err != nil || calls != 2 || in.PageNumber != nil {
		t.Fatal("all-ID stopped transition failed", err)
	}
	unknown, err := ecs.NewInstanceStoppedWaiter(statusAPI(func(context.Context, *ecs.DescribeInstanceStatusInput, ...func(*ecs.Options)) (*ecs.DescribeInstanceStatusOutput, error) {
		return observation([2]string{"a", "Deleted"}), nil
	}))
	if err != nil {
		t.Fatal(err)
	}
	if err = unknown.Wait(context.Background(), &ecs.DescribeInstanceStatusInput{InstanceIDs: []string{"a"}}, time.Second); !errors.Is(err, waiter.ErrFailure) {
		t.Fatal("unknown state accepted", err)
	}
}

func TestTokenOnlyPaginatorRejectsDeprecatedPageInputs(t *testing.T) {
	client := new(ecs.Client)
	if _, err := ecs.NewDescribeDisksPaginator(client, &ecs.DescribeDisksInput{PageNumber: pointer(int32(1))}); err == nil {
		t.Fatal("legacy page inputs accepted by token-only adapter")
	}
	if err := ecs.ValidateDescribeDisksInput(&ecs.DescribeDisksInput{PageNumber: pointer(int32(1))}); err != nil {
		t.Fatal("raw legacy request was rejected", err)
	}
}
