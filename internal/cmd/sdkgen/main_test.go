package main

import (
	"context"
	"errors"
	"testing"
)

func TestFullDSLCommandAliasesAndRemovedImport(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	for _, command := range []string{"generate", "check", "product-generate", "product-check"} {
		if err := run(ctx, []string{command, "-root", t.TempDir(), "-operations", "sts/AssumeRole"}); !errors.Is(err, context.Canceled) {
			t.Fatalf("%s did not reach the cancelable full-DSL backend: %v", command, err)
		}
	}
	if err := run(context.Background(), []string{"import"}); err == nil {
		t.Fatal("removed metadata import command accepted")
	}
}

func TestInvalidCommandArguments(t *testing.T) {
	for _, args := range [][]string{nil, {"unknown"}, {"import"}, {"check", "unexpected"}, {"generate", "unexpected"}, {"product-check", "unexpected"}, {"product-generate", "unexpected"}} {
		if err := run(context.Background(), args); err == nil {
			t.Fatal("accepted invalid arguments", args)
		}
	}
}
