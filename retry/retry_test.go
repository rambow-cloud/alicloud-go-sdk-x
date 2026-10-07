package retry_test

import (
	"context"
	"errors"
	"fmt"
	"github.com/rambow-cloud/alicloud-go-sdk-x/retry"
	"net/http"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestClassificationBoundsAndBudget(t *testing.T) {
	p, _ := retry.NewStandard(retry.Options{Budget: 2, Jitter: func(d time.Duration) time.Duration { return d }})
	a := retry.Attempt{Number: 1, StatusCode: 503, Err: errors.New("server"), Idempotent: true, Replayable: true}
	for _, bad := range []retry.Attempt{{Number: 1, StatusCode: 503, Err: a.Err}, {Number: 1, StatusCode: 403, Err: a.Err, Idempotent: true, Replayable: true}, {Number: 1, StatusCode: 503, Err: context.Canceled, Idempotent: true, Replayable: true}, {Number: 3, StatusCode: 503, Err: a.Err, Idempotent: true, Replayable: true}} {
		if p.ShouldRetry(bad) {
			t.Fatal("unsafe retry")
		}
	}
	if !p.ShouldRetry(a) || !p.ShouldRetry(a) || p.ShouldRetry(a) {
		t.Fatal("budget not bounded")
	}
	p.RecordSuccess()
	if !p.ShouldRetry(a) {
		t.Fatal("success did not refund")
	}
	if p.Delay(a) != 200*time.Millisecond {
		t.Fatal("base delay")
	}
	a.Number = 100
	a.RetryAfter = time.Hour
	if p.Delay(a) != 20*time.Second {
		t.Fatal("uncapped delay")
	}
}
func TestConcurrentBudgetAndCancellation(t *testing.T) {
	p, _ := retry.NewStandard(retry.Options{Budget: 7})
	var count atomic.Int32
	var wg sync.WaitGroup
	for range 40 {
		wg.Go(func() {
			if p.ShouldRetry(retry.Attempt{Number: 1, StatusCode: 429, Err: errors.New("throttling"), Idempotent: true, Replayable: true}) {
				count.Add(1)
			}
		})
	}
	wg.Wait()
	if count.Load() != 7 {
		t.Fatal(count.Load())
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := retry.Wait(ctx, time.Hour); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	if retry.ParseRetryAfter("2", now) != 2*time.Second || retry.ParseRetryAfter(now.Add(time.Minute).Format(http.TimeFormat), now) != time.Minute || retry.ParseRetryAfter("bad", now) != 0 {
		t.Fatal("retry-after parsing")
	}
}
func ExampleNewStandard() {
	p, _ := retry.NewStandard(retry.Options{})
	fmt.Println(p.MaxAttempts())
	// Output: 3
}
