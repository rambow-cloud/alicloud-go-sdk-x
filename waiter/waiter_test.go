package waiter_test

import (
	"context"
	"errors"
	"fmt"
	"github.com/rambow-cloud/alicloud-go-sdk-x/sdktest"
	"github.com/rambow-cloud/alicloud-go-sdk-x/waiter"
	"testing"
	"time"
)

func TestSuccessFailureAndTimeout(t *testing.T) {
	clock := sdktest.NewClock(time.Unix(0, 0))
	calls := 0
	w, _ := waiter.New(func(context.Context) (int, error) { calls++; return calls, nil }, func(n int, err error) waiter.Decision {
		if n == 3 {
			return waiter.Success
		}
		return waiter.Retry
	}, waiter.Options{Now: clock.Now, Sleep: clock.Sleep})
	value, err := w.Wait(context.Background(), time.Minute)
	if err != nil || value != 3 || clock.Now().Sub(time.Unix(0, 0)) != 3*time.Second {
		t.Fatal(value, err)
	}
	sentinel := errors.New("fetch failure")
	w, _ = waiter.New(func(context.Context) (int, error) { return 0, sentinel }, func(int, error) waiter.Decision { return waiter.Failure }, waiter.Options{})
	_, err = w.Wait(context.Background(), time.Minute)
	if !errors.Is(err, waiter.ErrFailure) || !errors.Is(err, sentinel) {
		t.Fatal(err)
	}
	w, _ = waiter.New(func(context.Context) (int, error) { return 0, sentinel }, func(int, error) waiter.Decision { return waiter.Retry }, waiter.Options{Now: clock.Now, Sleep: clock.Sleep})
	_, err = w.Wait(context.Background(), 2*time.Second)
	if !errors.Is(err, waiter.ErrTimeout) || !errors.Is(err, context.DeadlineExceeded) || !errors.Is(err, sentinel) {
		t.Fatal(err)
	}
}
func TestCancellationAndFetchDeadline(t *testing.T) {
	w, _ := waiter.New(func(ctx context.Context) (int, error) { <-ctx.Done(); return 0, ctx.Err() }, func(int, error) waiter.Decision { return waiter.Retry }, waiter.Options{})
	_, err := w.Wait(context.Background(), time.Millisecond)
	if !errors.Is(err, waiter.ErrTimeout) {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = w.Wait(ctx, time.Hour)
	if !errors.Is(err, context.Canceled) || errors.Is(err, waiter.ErrTimeout) {
		t.Fatal(err)
	}
	parent, stop := context.WithDeadline(context.Background(), time.Unix(0, 0))
	defer stop()
	_, err = w.Wait(parent, time.Hour)
	if !errors.Is(err, context.DeadlineExceeded) || errors.Is(err, waiter.ErrTimeout) {
		t.Fatal(err)
	}
}
func ExampleNew() {
	clock := sdktest.NewClock(time.Unix(0, 0))
	n := 0
	w, _ := waiter.New(func(context.Context) (int, error) { n++; return n, nil }, func(value int, err error) waiter.Decision {
		if value == 2 {
			return waiter.Success
		}
		return waiter.Retry
	}, waiter.Options{Now: clock.Now, Sleep: clock.Sleep})
	value, err := w.Wait(context.Background(), time.Minute)
	if err != nil {
		panic(err)
	}
	fmt.Println(value)
	// Output: 2
}
