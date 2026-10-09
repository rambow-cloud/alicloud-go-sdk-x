package main

import (
	"context"
	"testing"
)

func Example() {
	if err := run(); err != nil {
		panic(err)
	}
	// Output: PASS identity
}
func BenchmarkIdentity(b *testing.B) {
	c, _, err := client()
	if err != nil {
		b.Fatal(err)
	}
	ctx := context.Background()
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		if err := identity(ctx, c); err != nil {
			b.Fatal(err)
		}
	}
}
