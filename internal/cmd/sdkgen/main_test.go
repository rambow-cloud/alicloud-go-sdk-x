package main

import (
	"context"
	"testing"
)

func TestInvalidCommandArguments(t *testing.T) {
	for _, args := range [][]string{nil, {"unknown"}, {"import"}, {"check", "unexpected"}, {"generate", "unexpected"}, {"product-check", "unexpected"}, {"product-generate", "unexpected"}, {"import", "-out", t.TempDir(), "unexpected"}} {
		if err := run(context.Background(), args); err == nil {
			t.Fatal("accepted invalid arguments", args)
		}
	}
}
