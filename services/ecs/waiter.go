package ecs

import (
	"context"
	"errors"
	"github.com/rambow-cloud/alicloud-go-sdk-x/waiter"
	"time"
)

// InstanceRunningWaiter waits until every explicitly requested instance is Running.
// Construct with NewInstanceRunningWaiter. One to fifty distinct IDs are supported.
type InstanceRunningWaiter struct {
	engine *waiter.Waiter[*DescribeInstanceStatusOutput]
}

// NewInstanceRunningWaiter copies input and options and uses one status page of
// size fifty. Empty/missing results and known transitional/stopped states retry;
// unknown states, duplicate response IDs and operation errors fail immediately.
func NewInstanceRunningWaiter(api DescribeInstanceStatusAPI, input *DescribeInstanceStatusInput, options waiter.Options, opts ...func(*Options)) (*InstanceRunningWaiter, error) {
	if api == nil || input == nil || len(input.InstanceIDs) == 0 || len(input.InstanceIDs) > 50 || input.PageNumber > 1 || input.PageNumber < 0 {
		return nil, errors.New("ecs: waiter requires one to fifty distinct IDs on the first page")
	}
	in := *input
	in.InstanceIDs = append([]string(nil), input.InstanceIDs...)
	in.PageNumber = 1
	in.PageSize = 50
	copiedOptions := append([]func(*Options){}, opts...)
	required := map[string]bool{}
	for _, id := range in.InstanceIDs {
		if id == "" || required[id] {
			return nil, errors.New("ecs: invalid waiter instance IDs")
		}
		required[id] = true
	}
	engine, err := waiter.New(func(ctx context.Context) (*DescribeInstanceStatusOutput, error) {
		request := in
		request.InstanceIDs = append([]string(nil), in.InstanceIDs...)
		return api.DescribeInstanceStatus(ctx, &request, copiedOptions...)
	}, func(out *DescribeInstanceStatusOutput, err error) waiter.Decision {
		if err != nil || out == nil {
			return waiter.Failure
		}
		states := map[string]string{}
		for _, entry := range out.InstanceStatuses {
			if !required[entry.InstanceID] {
				continue
			}
			if _, ok := states[entry.InstanceID]; ok {
				return waiter.Failure
			}
			states[entry.InstanceID] = entry.Status
		}
		running := 0
		for id := range required {
			state, present := states[id]
			if !present {
				continue
			}
			switch state {
			case "Running":
				running++
			case "Pending", "Starting", "Stopping", "Stopped":
			default:
				return waiter.Failure
			}
		}
		if running == len(required) {
			return waiter.Success
		}
		return waiter.Retry
	}, options)
	if err != nil {
		return nil, err
	}
	return &InstanceRunningWaiter{engine: engine}, nil
}

// Wait requires a positive total duration and returns the successful status page.
// Caller context bounds each underlying operation; waiter expiry is ErrTimeout.
func (w *InstanceRunningWaiter) Wait(ctx context.Context, maxWait time.Duration) (*DescribeInstanceStatusOutput, error) {
	return w.engine.Wait(ctx, maxWait)
}
