package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDocumentationPolicyRejectsMissingContract(t *testing.T) {
	dir := t.TempDir()
	fixtures := map[string]string{
		"doc.go":    "package sample\n",
		"sample.go": "package sample\nimport _ \"encoding/json\"\ntype Input struct { Name string }\ntype Provider interface { Retrieve() error }\n",
	}
	for name, source := range fixtures {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(source), 0600); err != nil {
			t.Fatal(err)
		}
	}
	failures, err := checkPackage(packageInfo{Dir: dir, ImportPath: "example.com/sample", Name: "sample", GoFiles: []string{"doc.go", "sample.go"}})
	if err != nil {
		t.Fatal(err)
	}
	got := strings.Join(failures, "\n")
	for _, want := range []string{"encoding/json/v2", "Input needs", "Name field", "Provider needs", "Retrieve field", "doc.go package overview", "external Example"} {
		if !strings.Contains(got, want) {
			t.Errorf("policy failed to detect %q", want)
		}
	}
}

func TestDocumentationPolicyAcceptsExternalExample(t *testing.T) {
	dir := t.TempDir()
	fixtures := map[string]string{
		"doc.go":          "// Package sample demonstrates a fixture.\npackage sample\n",
		"sample.go":       "package sample\n// Input holds data.\ntype Input struct {\n// Name identifies data.\nName string\n}\n// Provider retrieves data.\ntype Provider interface {\n// Retrieve returns data.\nRetrieve() error\n}\n",
		"example_test.go": "package sample_test\nimport \"fmt\"\nfunc Example() { fmt.Println(\"sample\")\n// Output: sample\n}\n",
	}
	for name, source := range fixtures {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(source), 0600); err != nil {
			t.Fatal(err)
		}
	}
	failures, err := checkPackage(packageInfo{Dir: dir, ImportPath: "example.com/sample", Name: "sample", GoFiles: []string{"doc.go", "sample.go"}, XTestGoFiles: []string{"example_test.go"}})
	if err != nil {
		t.Fatal(err)
	}
	if len(failures) != 0 {
		t.Fatalf("valid documented package rejected: %v", failures)
	}
}
