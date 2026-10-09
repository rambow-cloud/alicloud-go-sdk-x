package main

import (
	"github.com/alibabacloud-go/tea/dara"
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
	options := &dara.RuntimeOptions{Autoretry: dara.Bool(false)}
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		if err := identity(c, options); err != nil {
			b.Fatal(err)
		}
	}
}
