package pagination_test

import (
	"context"
	"errors"
	"fmt"
	"github.com/rambow-cloud/alicloud-go-sdk-x/pagination"
	"testing"
)

func TestErrorDoesNotAdvanceAndEmptyPageCanContinue(t *testing.T) {
	sentinel := errors.New("temporary")
	calls := 0
	p, _ := pagination.New(pagination.Cursor{}, func(ctx context.Context, c pagination.Cursor) (pagination.Page[[]int], error) {
		calls++
		if calls == 1 {
			return pagination.Page[[]int]{}, sentinel
		}
		if c.Token == "" {
			return pagination.Page[[]int]{Next: pagination.Cursor{Token: "next"}, HasMore: true}, nil
		}
		return pagination.Page[[]int]{Value: []int{42}}, nil
	})
	_, err := p.NextPage(context.Background())
	if !errors.Is(err, sentinel) || !p.HasMorePages() {
		t.Fatal(err)
	}
	first, err := p.NextPage(context.Background())
	if err != nil || len(first) != 0 || !p.HasMorePages() {
		t.Fatal(first, err)
	}
	last, err := p.NextPage(context.Background())
	if err != nil || last[0] != 42 || p.HasMorePages() {
		t.Fatal(last, err)
	}
	_, err = p.NextPage(context.Background())
	if !errors.Is(err, pagination.ErrNoMorePages) || calls != 3 {
		t.Fatal("exhaustion fetched again")
	}
}
func TestCyclesAndCancellation(t *testing.T) {
	calls := 0
	p, _ := pagination.New(pagination.Cursor{PageNumber: 1}, func(ctx context.Context, c pagination.Cursor) (pagination.Page[int], error) {
		calls++
		next := 2
		if c.PageNumber == 2 {
			next = 1
		}
		return pagination.Page[int]{Value: c.PageNumber, Next: pagination.Cursor{PageNumber: next}, HasMore: true}, nil
	})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := p.NextPage(ctx)
	if !errors.Is(err, context.Canceled) || calls != 0 {
		t.Fatal(err)
	}
	_, _ = p.NextPage(context.Background())
	value, err := p.NextPage(context.Background())
	if err != nil || value != 2 || p.HasMorePages() {
		t.Fatal("cycle not stopped")
	}
}

func TestDuplicateProtectionOptOut(t *testing.T) {
	p, _ := pagination.New(pagination.Cursor{Token: "same"}, func(context.Context, pagination.Cursor) (pagination.Page[int], error) {
		return pagination.Page[int]{Value: 42, HasMore: true, Next: pagination.Cursor{Token: "same"}}, nil
	}, func(o *pagination.Options) { o.StopOnDuplicateCursor = false })
	for range 2 {
		value, err := p.NextPage(context.Background())
		if err != nil || value != 42 || !p.HasMorePages() {
			t.Fatal(value, err)
		}
	}
	if _, err := pagination.New(pagination.Cursor{}, func(context.Context, pagination.Cursor) (pagination.Page[int], error) {
		return pagination.Page[int]{}, nil
	}, nil); err == nil {
		t.Fatal("nil option accepted")
	}
}
func ExampleNew() {
	p, _ := pagination.New(pagination.Cursor{PageNumber: 1}, func(ctx context.Context, c pagination.Cursor) (pagination.Page[int], error) {
		return pagination.Page[int]{Value: c.PageNumber, Next: pagination.Cursor{PageNumber: c.PageNumber + 1}, HasMore: c.PageNumber < 2}, nil
	})
	for p.HasMorePages() {
		page, err := p.NextPage(context.Background())
		if err != nil {
			panic(err)
		}
		fmt.Println(page)
	}
	// Output:
	// 1
	// 2
}
