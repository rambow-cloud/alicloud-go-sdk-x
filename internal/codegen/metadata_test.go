package codegen

import (
	"context"
	"encoding/json/v2"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

type doer func(*http.Request) (*http.Response, error)

func (f doer) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

const synthetic = `{"methods":["post"],"schemes":["https"],"parameters":[{"name":"Name","in":"query","schema":{"type":"string","required":true,"description":"upstream prose"}}],"responses":{"200":{"schema":{"type":"object","properties":{"Value":{"type":"string","description":"upstream prose"},"example":{"type":"string"}}}}}}`

func TestMetadataFetchAndExtraction(t *testing.T) {
	closed := false
	client := &http.Client{Transport: doer(func(r *http.Request) (*http.Response, error) {
		if r.URL.Host != "api.aliyun.com" || r.Header.Get("Authorization") != "" {
			t.Fatal("invalid metadata request")
		}
		return &http.Response{StatusCode: 200, Body: &trackedBody{Reader: strings.NewReader(synthetic), closed: &closed}, Header: http.Header{}}, nil
	})}
	source, data, err := Fetch(context.Background(), client, "Demo", "2026-01-01", "Lookup", time.Unix(0, 0))
	if err != nil || !closed || source.RawSHA256 == source.SHA256 || strings.Contains(string(data), "upstream prose") || !strings.Contains(string(data), `"example":{"type":"string"}`) {
		t.Fatal(source, string(data), err)
	}
	if client.Timeout != 0 || client.CheckRedirect != nil {
		t.Fatal("mutated client")
	}
	for _, status := range []int{302, 404} {
		client.Transport = doer(func(*http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: status, Body: io.NopCloser(strings.NewReader("")), Header: http.Header{}}, nil
		})
		if _, _, err := Fetch(context.Background(), client, "Demo", "2026-01-01", "Lookup", time.Now()); err == nil {
			t.Fatal("accepted status", status)
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	client.Transport = doer(func(r *http.Request) (*http.Response, error) { return nil, r.Context().Err() })
	if _, _, err := Fetch(ctx, client, "Demo", "2026-01-01", "Lookup", time.Now()); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if _, err := Extract([]byte(`{"methods":[],"methods":[]}`)); err == nil {
		t.Fatal("duplicate JSON accepted")
	}
}

type trackedBody struct {
	io.Reader
	closed *bool
}

func (b *trackedBody) Close() error { *b.closed = true; return nil }

func copyFixture(t *testing.T, pkg string) string {
	t.Helper()
	dir := t.TempDir()
	source := filepath.Join("..", "..", "metadata", pkg)
	entries, err := os.ReadDir(source)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			data, err := os.ReadFile(filepath.Join(source, entry.Name()))
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(dir, entry.Name()), data, 0600); err != nil {
				t.Fatal(err)
			}
		}
	}
	return dir
}

func mutateSnapshot(t *testing.T, dir, operation string, change func(map[string]any)) {
	t.Helper()
	path := filepath.Join(dir, operation+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var document map[string]any
	if err := json.Unmarshal(data, &document); err != nil {
		t.Fatal(err)
	}
	change(document)
	data, err = json.Marshal(document, json.Deterministic(true))
	if err != nil {
		t.Fatal(err)
	}
	data = append(data, '\n')
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}
	var m Manifest
	if err := readJSON(filepath.Join(dir, "manifest.json"), &m, true); err != nil {
		t.Fatal(err)
	}
	for i := range m.Sources {
		if m.Sources[i].Operation == operation {
			m.Sources[i].SHA256 = digest(data)
		}
	}
	encoded, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "manifest.json"), encoded, 0600); err != nil {
		t.Fatal(err)
	}
}

func TestPinnedProductsAndSchemaDrift(t *testing.T) {
	for _, pkg := range []string{"ecs", "sts"} {
		p, err := Load(filepath.Join("..", "..", "metadata", pkg))
		if err != nil {
			t.Fatal(pkg, err)
		}
		if len(p.Operations) == 0 || len(p.Models) == 0 {
			t.Fatal("empty IR")
		}
	}
	cases := map[string]func(map[string]any){
		"removed selected": func(d map[string]any) { d["parameters"] = []any{} },
		"changed type": func(d map[string]any) {
			for _, raw := range d["parameters"].([]any) {
				p := raw.(map[string]any)
				if p["name"] == "AcceptLanguage" {
					p["schema"].(map[string]any)["type"] = "boolean"
				}
			}
		},
		"new required": func(d map[string]any) {
			d["parameters"] = append(d["parameters"].([]any), map[string]any{"name": "NewRequired", "in": "query", "schema": map[string]any{"type": "string", "required": true}})
		},
		"reference": func(d map[string]any) {
			for _, raw := range d["parameters"].([]any) {
				p := raw.(map[string]any)
				if p["name"] == "AcceptLanguage" {
					p["schema"].(map[string]any)["$ref"] = "#/components/schemas/Foo"
				}
			}
		},
	}
	for name, change := range cases {
		t.Run(name, func(t *testing.T) {
			dir := copyFixture(t, "ecs")
			mutateSnapshot(t, dir, "DescribeRegions", change)
			if _, err := Load(dir); err == nil {
				t.Fatal("accepted incompatible metadata")
			}
		})
	}
	dir := copyFixture(t, "ecs")
	mutateSnapshot(t, dir, "DescribeRegions", func(d map[string]any) {
		d["parameters"] = append(d["parameters"].([]any), map[string]any{"name": "NewOptional", "in": "body", "schema": map[string]any{"type": "object", "required": false}})
	})
	if _, err := Load(dir); err != nil {
		t.Fatal("unselected optional addition rejected", err)
	}
	dir = copyFixture(t, "ecs")
	path := filepath.Join(dir, "DescribeRegions.json")
	data, _ := os.ReadFile(path)
	os.WriteFile(path, append(data, ' '), 0600)
	if _, err := Load(dir); err == nil {
		t.Fatal("tampering accepted")
	}
	dir = copyFixture(t, "ecs")
	path = filepath.Join(dir, "overlay.json")
	data, _ = os.ReadFile(path)
	data = append([]byte(`{"typo":true,`), data[1:]...)
	os.WriteFile(path, data, 0600)
	if _, err := Load(dir); err == nil {
		t.Fatal("unknown overlay member accepted")
	}
}

func TestSyntheticServiceIR(t *testing.T) {
	dir := t.TempDir()
	rawDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(rawDir, "Lookup.json"), []byte(synthetic), 0600); err != nil {
		t.Fatal(err)
	}
	m := Manifest{SchemaVersion: 1, Product: "Demo", Version: "2026-01-01", Style: "RPC", Package: "demo", Service: "demo"}
	if err := Import(context.Background(), nil, m, []string{"Lookup"}, rawDir, dir); err != nil {
		t.Fatal(err)
	}
	truth := true
	doc := Doc{"Looks up a named value.", "查询具名值。"}
	o := Overlay{SchemaVersion: 1, Doc: doc, Operations: []OperationSpec{{Name: "Lookup", Idempotent: &truth, Inputs: []FieldSpec{{Name: "Name", Wire: "Name", Type: "string", Doc: doc}}, Outputs: []FieldSpec{{Name: "Value", Wire: "Value", Type: "string", Doc: doc}}, Doc: doc, Example: ExampleSpec{Input: map[string]any{"Name": "example"}, Response: `{"Value":"result"}`, Print: "Value", Output: "result"}}}}
	data, _ := json.Marshal(o)
	os.WriteFile(filepath.Join(dir, "overlay.json"), data, 0600)
	p, err := Load(dir)
	if err != nil || p.Manifest.Service != "demo" || !p.Operations[0].Inputs[0].Required {
		t.Fatal(p, err)
	}
}
