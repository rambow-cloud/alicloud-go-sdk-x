package middleware_test

import (
	"context"
	"errors"
	"fmt"
	"github.com/rambow-cloud/alicloud-go-sdk-x/middleware"
	"reflect"
	"sync"
	"testing"
)

func TestOrderingCopyAndShortCircuit(t *testing.T) {
	var order []string
	makeHook := func(id string) middleware.Middleware {
		return middleware.Func(id, func(ctx context.Context, e *middleware.Exchange, next middleware.Handler) error {
			order = append(order, id+"+")
			err := next(ctx, e)
			order = append(order, id+"-")
			return err
		})
	}
	regs := []middleware.Registration{{Stage: middleware.Build, Middleware: makeHook("a")}, {Stage: middleware.Build, Middleware: makeHook("b")}}
	stack, err := middleware.NewStack(regs)
	if err != nil {
		t.Fatal(err)
	}
	regs[0].Middleware = makeHook("changed")
	sentinel := errors.New("stop")
	err = stack.Run(context.Background(), middleware.Build, &middleware.Exchange{}, func(context.Context, *middleware.Exchange) error { order = append(order, "final"); return sentinel })
	if !errors.Is(err, sentinel) || !reflect.DeepEqual(order, []string{"a+", "b+", "final", "b-", "a-"}) {
		t.Fatalf("order=%v error=%v", order, err)
	}
	if _, err = middleware.NewStack([]middleware.Registration{{Stage: middleware.Build, Middleware: makeHook("a")}, {Stage: middleware.Build, Middleware: makeHook("a")}}); err == nil {
		t.Fatal("duplicate accepted")
	}
	blocked, _ := middleware.NewStack([]middleware.Registration{{Stage: middleware.Finalize, Middleware: middleware.Func("block", func(context.Context, *middleware.Exchange, middleware.Handler) error { return sentinel })}})
	err = blocked.Run(context.Background(), middleware.Finalize, &middleware.Exchange{}, func(context.Context, *middleware.Exchange) error { t.Fatal("short circuit failed"); return nil })
	if !errors.Is(err, sentinel) {
		t.Fatal(err)
	}
}
func TestCancellationAndConcurrentUse(t *testing.T) {
	stack, _ := middleware.NewStack(nil)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := stack.Run(ctx, middleware.Build, &middleware.Exchange{}, func(context.Context, *middleware.Exchange) error { t.Fatal("called"); return nil }); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	var group sync.WaitGroup
	for range 20 {
		group.Go(func() {
			if err := stack.Run(context.Background(), middleware.Build, &middleware.Exchange{}, func(context.Context, *middleware.Exchange) error { return nil }); err != nil {
				t.Error(err)
			}
		})
	}
	group.Wait()
}
func Example() {
	stack, _ := middleware.NewStack([]middleware.Registration{{Stage: middleware.Initialize, Middleware: middleware.Func("greeting", func(ctx context.Context, e *middleware.Exchange, next middleware.Handler) error {
		fmt.Println(e.Operation)
		return next(ctx, e)
	})}})
	_ = stack.Run(context.Background(), middleware.Initialize, &middleware.Exchange{Operation: "DescribeRegions"}, func(context.Context, *middleware.Exchange) error { return nil })
	// Output: DescribeRegions
}
