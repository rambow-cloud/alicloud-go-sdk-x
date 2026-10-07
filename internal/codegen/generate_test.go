package codegen

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func fixtureRepository(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	for _, pkg := range []string{"ecs", "sts"} {
		dir := copyFixture(t, pkg)
		entries, err := os.ReadDir(dir)
		if err != nil {
			t.Fatal(err)
		}
		for _, entry := range entries {
			data, err := os.ReadFile(filepath.Join(dir, entry.Name()))
			if err != nil {
				t.Fatal(err)
			}
			writeTestFile(t, root, filepath.Join("metadata", pkg, entry.Name()), data)
		}
	}
	return root
}

func TestCheckNeverWritesAndWriteRepairsOwnedDrift(t *testing.T) {
	root := fixtureRepository(t)
	if err := Generate(context.Background(), root, true); err == nil {
		t.Fatal("missing outputs accepted")
	}
	if _, err := os.Stat(filepath.Join(root, "services")); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("check created output")
	}
	if err := Generate(context.Background(), root, false); err != nil {
		t.Fatal(err)
	}
	if err := Generate(context.Background(), root, true); err != nil {
		t.Fatal(err)
	}
	changed := "services/ecs/sdk.gen.go"
	original, err := os.ReadFile(filepath.Join(root, changed))
	if err != nil {
		t.Fatal(err)
	}
	writeTestFile(t, root, changed, append(append([]byte(nil), original...), []byte("// manually edited\n")...))
	missing := "services/sts/examples.gen_test.go"
	if err := os.Remove(filepath.Join(root, missing)); err != nil {
		t.Fatal(err)
	}
	stale := "services/ecs/obsolete.gen.go"
	writeTestFile(t, root, stale, []byte(generated+"package ecs\n"))
	manual := "services/ecs/extension.go"
	writeTestFile(t, root, manual, []byte("package ecs\n"))
	var drift *DriftError
	if err := Generate(context.Background(), root, true); !errors.As(err, &drift) || len(drift.Files) != 3 {
		t.Fatal(drift, err)
	}
	if data, _ := os.ReadFile(filepath.Join(root, changed)); bytes.Equal(data, original) {
		t.Fatal("check repaired file")
	}
	if _, err := os.Stat(filepath.Join(root, stale)); err != nil {
		t.Fatal("check removed stale file")
	}
	if err := Generate(context.Background(), root, false); err != nil {
		t.Fatal(err)
	}
	if data, _ := os.ReadFile(filepath.Join(root, changed)); !bytes.Equal(data, original) {
		t.Fatal("write did not repair")
	}
	if _, err := os.Stat(filepath.Join(root, stale)); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("stale file retained")
	}
	if _, err := os.Stat(filepath.Join(root, manual)); err != nil {
		t.Fatal("manual extension lost")
	}
}

func TestPreflightFailuresDoNotPartiallyWrite(t *testing.T) {
	root := fixtureRepository(t)
	protected := "services/sts/sdk.gen.go"
	writeTestFile(t, root, protected, []byte("package sts\n"))
	if err := Generate(context.Background(), root, false); err == nil {
		t.Fatal("overwrote manual file")
	}
	if _, err := os.Stat(filepath.Join(root, "docs", "generated")); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("wrote before ownership preflight completed")
	}
	if data, _ := os.ReadFile(filepath.Join(root, protected)); string(data) != "package sts\n" {
		t.Fatal("modified protected file")
	}
	root = fixtureRepository(t)
	if err := Generate(context.Background(), root, false); err != nil {
		t.Fatal(err)
	}
	before, err := os.ReadFile(filepath.Join(root, "services/ecs/sdk.gen.go"))
	if err != nil {
		t.Fatal(err)
	}
	writeTestFile(t, root, "metadata/sts/overlay.json", []byte("{}"))
	if err := Generate(context.Background(), root, false); err == nil {
		t.Fatal("invalid last product accepted")
	}
	after, _ := os.ReadFile(filepath.Join(root, "services/ecs/sdk.gen.go"))
	if !bytes.Equal(before, after) {
		t.Fatal("earlier product updated before full validation")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := Generate(ctx, root, false); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
}

func TestOutputConfinement(t *testing.T) {
	root := t.TempDir()
	for _, name := range []string{"../outside.go", "services\\ecs\\file.go"} {
		if err := safeOutputPath(root, name); err == nil {
			t.Fatal("unsafe output accepted", name)
		}
	}
	if err := os.MkdirAll(filepath.Join(root, "services"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(t.TempDir(), filepath.Join(root, "services", "ecs")); err != nil {
		t.Skip("symlink unavailable")
	}
	if err := safeOutputPath(root, "services/ecs/sdk.gen.go"); err == nil {
		t.Fatal("symlink escape accepted")
	}
}
