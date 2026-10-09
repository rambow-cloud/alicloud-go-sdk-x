package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json/v2"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/rambow-cloud/alicloud-go-sdk-x/internal/nativewire"
)

func TestRunValidatesBeforeOutput(t *testing.T) {
	dir := t.TempDir()
	r := []byte("package fixture\nimport \"reflect\"\nvar roots = make(map[string]reflect.Type)\nfunc init() { roots[\"Read\"] = reflect.TypeOf(Envelope{}) }\n")
	m := []byte("package fixture\ntype Envelope struct { Text string `json:\"text\" xml:\"Text\"` }\n")
	hash := func(b []byte) string { sum := sha256.Sum256(b); return hex.EncodeToString(sum[:]) }
	p := pins{SchemaVersion: 1, RegistrySymbol: "roots", Registry: sourcePin{"registry.go", hash(r)}, Models: sourcePin{"models.go", hash(m)}}
	paths := []string{filepath.Join(dir, "pins.json"), filepath.Join(dir, "registry.go"), filepath.Join(dir, "models.go")}
	data, err := json.Marshal(p)
	if err != nil {
		t.Fatal(err)
	}
	for i, b := range [][]byte{data, r, m} {
		if err := os.WriteFile(paths[i], b, 0600); err != nil {
			t.Fatal(err)
		}
	}
	var output bytes.Buffer
	if err := run(context.Background(), paths[0], paths[1], paths[2], &output); err != nil {
		t.Fatal(err)
	}
	var facts nativewire.Inventory
	if err := json.Unmarshal(output.Bytes(), &facts); err != nil || len(facts.Roots) != 1 || facts.Roots[0].Name != "Text" {
		t.Fatalf("invalid facts: %+v %v", facts, err)
	}
	if err := run(context.Background(), paths[0], paths[1], paths[2], shortWriter{}); !errors.Is(err, io.ErrShortWrite) {
		t.Fatal("short output ignored", err)
	}
	if err := os.WriteFile(paths[2], append(m, ' '), 0600); err != nil {
		t.Fatal(err)
	}
	output.Reset()
	if err := run(context.Background(), paths[0], paths[1], paths[2], &output); !errors.Is(err, nativewire.ErrInvalid) || output.Len() != 0 {
		t.Fatal("drift published partial output", err)
	}
}

type shortWriter struct{}

func (shortWriter) Write(b []byte) (int, error) { return len(b) - 1, nil }
