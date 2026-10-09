package nativewire_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json/v2"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/rambow-cloud/alicloud-go-sdk-x/internal/nativewire"
)

const registry = `package fixture
import r "reflect"
var catalog = make(map[string]r.Type)
func init() {
 catalog["ReadText"] = r.TypeOf(TextEnvelope{})
 catalog["ReadNested"] = r.TypeOf(NestedEnvelope{})
 catalog["Unbound"] = r.TypeOf(Missing{})
 catalog["Flat"] = r.TypeOf(FlatEnvelope{})
}
func lookup(key string) r.Type { return catalog[key] }
`

const models = "package fixture\n" +
	"type TextEnvelope struct { Value *string `json:\"value,omitempty\" xml:\"urn:fixture Location,omitempty\"` }\n" +
	"type NestedEnvelope struct { Result *Details `json:\"result\" xml:\"Result\"` }\n" +
	"type Details struct { Labels []*string `json:\"labels\" xml:\"Label\"`; Number int32 `json:\"number\" xml:\"Number\"` }\n" +
	"type FlatEnvelope struct { A string `json:\"a\" xml:\"A\"`; B string `json:\"b\" xml:\"B\"` }\n"

func source(name, value string) nativewire.Source {
	data := []byte(value)
	digest := sha256.Sum256(data)
	return nativewire.Source{Name: name, Bytes: data, SHA256: hex.EncodeToString(digest[:])}
}

func parse(ctx context.Context, r, m string) (nativewire.Inventory, error) {
	return nativewire.ParseXML(ctx, source("registry.go", r), source("models.go", m), "catalog")
}

func TestCompleteExplicitFacts(t *testing.T) {
	input := source("registry.go", registry)
	before := append([]byte(nil), input.Bytes...)
	result, err := nativewire.ParseXML(context.Background(), input, source("models.go", models), "catalog")
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(input.Bytes, before) || len(result.Models) != 4 || len(result.Roots) != 4 {
		t.Fatalf("incomplete or mutated facts: %+v", result)
	}
	if got := result.Roots[1]; got.Action != "ReadNested" || got.Kind != "structured" || got.Name != "Result" || got.NamespaceMatch != "local" || got.Field.Type.Name != "Details" || got.Source.Line != 6 {
		t.Fatalf("structured root: %+v", got)
	}
	if got := result.Roots[2]; got.Action != "ReadText" || got.Kind != "scalar" || got.Name != "Location" || got.Namespace != "urn:fixture" || got.NamespaceMatch != "exact" || got.Field.JSONName != "value" || !got.Field.Type.Optional {
		t.Fatalf("scalar root: %+v", got)
	}
	if result.Roots[0].Reason != "root requires one explicit wrapper/scalar field" || result.Roots[3].Reason != "registered model missing" {
		t.Fatal("unsupported reasons lost")
	}
	if got := result.Models[0].Fields[0]; got.Type.Kind != "array" || got.Type.Items.Kind != "scalar" || !got.Type.Items.Optional || got.Source.File != "models.go" {
		t.Fatalf("array/source facts: %+v", got)
	}
	again, err := parse(context.Background(), registry, models)
	if err != nil {
		t.Fatal(err)
	}
	a, _ := json.Marshal(result, json.Deterministic(true))
	b, _ := json.Marshal(again, json.Deterministic(true))
	if string(a) != string(b) {
		t.Fatal("inventory is nondeterministic")
	}
	result.Roots[1].Field.JSONName = "changed"
	if again.Roots[1].Field.JSONName != "result" {
		t.Fatal("calls share mutable facts")
	}
}

func TestAmbiguousRegistryFailsWithoutPartialFacts(t *testing.T) {
	for name, edit := range map[string]func(string) string{
		"dynamic key":       func(s string) string { return strings.Replace(s, `catalog["ReadText"]`, `catalog[key]`, 1) },
		"short declaration": func(s string) string { return strings.Replace(s, `catalog["ReadText"] =`, `catalog["ReadText"] :=`, 1) },
		"variadic call": func(s string) string {
			return strings.Replace(s, "r.TypeOf(TextEnvelope{})", "r.TypeOf(TextEnvelope{}...)", 1)
		},
		"duplicate": func(s string) string { return strings.Replace(s, `catalog["ReadNested"]`, `catalog["ReadText"]`, 1) },
		"conditional": func(s string) string {
			return strings.Replace(s, `catalog["ReadText"] = r.TypeOf(TextEnvelope{})`, `if true { catalog["ReadText"] = r.TypeOf(TextEnvelope{}) }`, 1)
		},
		"outside init":   func(s string) string { return strings.Replace(s, "func init()", "func modify()", 1) },
		"alias":          func(s string) string { return s + "func alias() { x := (catalog); _ = x }\n" },
		"global alias":   func(s string) string { return s + "var alias = catalog\n" },
		"whole map":      func(s string) string { return s + "func replace() { catalog = make(map[string]r.Type) }\n" },
		"delete":         func(s string) string { return s + "func remove() { delete(catalog, \"ReadText\") }\n" },
		"clear":          func(s string) string { return s + "func remove() { clear((catalog)) }\n" },
		"wrong reflect":  func(s string) string { return strings.Replace(s, `"reflect"`, `"example.invalid/reflect"`, 1) },
		"shadow reflect": func(s string) string { return strings.Replace(s, "func init() {", "func init() { r := fake;", 1) },
		"shadow catalog": func(s string) string {
			return strings.Replace(s, "func init() {", "func init() { catalog := make(map[string]r.Type);", 1)
		},
		"shadow model": func(s string) string {
			return strings.Replace(s, "func init() {", "func init() { type TextEnvelope struct{};", 1)
		},
		"populated literal":   func(s string) string { return strings.Replace(s, "TextEnvelope{}", `TextEnvelope{Value: nil}`, 1) },
		"unknown initializer": func(s string) string { return strings.Replace(s, "make(map[string]r.Type)", "createCatalog()", 1) },
		"wrong map type":      func(s string) string { return strings.Replace(s, "map[string]r.Type", "map[int]r.Type", 1) },
		"missing registry":    func(s string) string { return strings.ReplaceAll(s, "catalog", "other") },
		"syntax":              func(s string) string { return s + "source-secret-parse-error" },
	} {
		t.Run(name, func(t *testing.T) {
			got, err := parse(context.Background(), edit(registry), models)
			if !errors.Is(err, nativewire.ErrInvalid) || !reflect.DeepEqual(got, nativewire.Inventory{}) || strings.Contains(err.Error(), "source-secret") {
				t.Fatalf("partial facts or unsafe error: %+v %v", got, err)
			}
		})
	}
}

func TestSourceBindingAndCancellation(t *testing.T) {
	r, m := source("registry.go", registry), source("models.go", models)
	r.Bytes = append(r.Bytes, ' ')
	if got, err := nativewire.ParseXML(context.Background(), r, m, "catalog"); !errors.Is(err, nativewire.ErrInvalid) || len(got.Roots) != 0 {
		t.Fatal("drift accepted")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if got, err := parse(ctx, registry, models); !errors.Is(err, context.Canceled) || len(got.Roots) != 0 {
		t.Fatal("cancellation lost")
	}
	mid := &cancelDuringRead{Context: context.Background()}
	if got, err := parse(mid, registry, models); !errors.Is(err, context.Canceled) || !reflect.DeepEqual(got, nativewire.Inventory{}) {
		t.Fatal("mid-discovery cancellation published partial facts")
	}
	if _, err := parse(context.Background(), registry, strings.Replace(models, "package fixture", "package another", 1)); !errors.Is(err, nativewire.ErrInvalid) {
		t.Fatal("package mismatch accepted")
	}
	if _, err := parse(context.Background(), registry, models+"var alias = catalog\n"); !errors.Is(err, nativewire.ErrInvalid) {
		t.Fatal("model file registry alias accepted")
	}
}

type cancelDuringRead struct {
	context.Context
	checks int
}

func (c *cancelDuringRead) Err() error {
	c.checks++
	if c.checks >= 3 {
		return context.Canceled
	}
	return nil
}

func TestUnsupportedRootTagsAreNotGuessed(t *testing.T) {
	for _, tag := range []string{"", "-", "A>B", "Root,attr", "relative Root"} {
		m := "package fixture\ntype TextEnvelope struct { Value string `json:\"value\" xml:\"" + tag + "\"` }\n"
		r := strings.Replace(registry, " catalog[\"ReadNested\"] = r.TypeOf(NestedEnvelope{})\n catalog[\"Unbound\"] = r.TypeOf(Missing{})\n catalog[\"Flat\"] = r.TypeOf(FlatEnvelope{})\n", "", 1)
		got, err := parse(context.Background(), r, m)
		if err != nil {
			t.Fatal(err)
		}
		if got.Roots[0].Kind != "unsupported" {
			t.Fatalf("unsupported tag %q guessed: %+v", tag, got.Roots[0])
		}
	}
}
