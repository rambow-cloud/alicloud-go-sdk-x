package codegen

import (
	"bytes"
	"context"
	"encoding/json/v2"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestOfficialDSLJoinsAllProductsWithoutChangingGoContracts(t *testing.T) {
	for _, pkg := range []string{"ecs", "sts", "vpc"} {
		t.Run(pkg, func(t *testing.T) {
			dir := filepath.Join("..", "..", "metadata", pkg)
			original, err := Load(dir)
			if err != nil {
				t.Fatal(err)
			}
			product, err := LoadProduct(dir)
			if err != nil {
				t.Fatal(err)
			}
			if product.DSL == nil {
				t.Fatal("official frontend bypassed")
			}
			before, err := Render(original)
			if err != nil {
				t.Fatal(err)
			}
			after, err := Render(product)
			if err != nil {
				t.Fatal(err)
			}
			for name, content := range before {
				if strings.HasSuffix(name, ".go") && !bytes.Equal(content, after[name]) {
					t.Fatalf("public contract changed: %s", name)
				}
			}
			for _, op := range product.Operations {
				if op.Protocol.Action != op.Name || op.Protocol.Version != product.Manifest.Version || op.Protocol.Method != "POST" {
					t.Fatal("protocol not lowered", op.Name)
				}
			}
		})
	}
}

func mutateDSL(t *testing.T, root string, change func(*DSLProduct)) {
	t.Helper()
	dir := filepath.Join(root, "metadata", "sts")
	var projection DSLProduct
	if err := readJSON(filepath.Join(dir, "dsl.json"), &projection, true); err != nil {
		t.Fatal(err)
	}
	change(&projection)
	data, err := json.Marshal(projection)
	if err != nil {
		t.Fatal(err)
	}
	writeTestFile(t, root, "metadata/sts/dsl.json", data)
	var manifest Manifest
	if err := readJSON(filepath.Join(dir, "manifest.json"), &manifest, true); err != nil {
		t.Fatal(err)
	}
	manifest.DSL.SHA256 = digest(data)
	data, err = json.Marshal(manifest)
	if err != nil {
		t.Fatal(err)
	}
	writeTestFile(t, root, "metadata/sts/manifest.json", data)
}

func TestDSLProjectionRejectsUnreviewedContracts(t *testing.T) {
	cases := map[string]func(*DSLProduct){
		"wire type": func(d *DSLProduct) {
			d.Operations[0].Response.Properties["Credentials"].Properties["Expiration"].Type = "integer"
		},
		"binding location":  func(d *DSLProduct) { d.Operations[0].Inputs[0].Location = "body" },
		"protocol":          func(d *DSLProduct) { d.Operations[0].Protocol.Method = "GET" },
		"revision":          func(d *DSLProduct) { d.Revision = strings.Repeat("0", 40) },
		"missing operation": func(d *DSLProduct) { d.Operations = nil },
		"requiredness":      func(d *DSLProduct) { d.Operations[0].Inputs[0].Schema.Required = true },
		"hidden input": func(d *DSLProduct) {
			d.Operations[0].Inputs = append(d.Operations[0].Inputs, DSLInput{Wire: "Hidden", Location: "query", Guard: "isUnset", Schema: &DSLShape{Type: "string"}})
		},
	}
	for name, change := range cases {
		t.Run(name, func(t *testing.T) {
			root := fixtureRepository(t)
			mutateDSL(t, root, change)
			if err := Generate(context.Background(), root, false); err == nil {
				t.Fatal("unreviewed change accepted")
			}
			if _, err := os.Stat(filepath.Join(root, "services")); !os.IsNotExist(err) {
				t.Fatal("wrote outputs before cross-check")
			}
		})
	}
}

func TestDSLTamperingFailsBeforeAnyOutputWrites(t *testing.T) {
	for _, name := range []string{"sources/darabonba/products/sts/main.tea", "metadata/sts/dsl.json", "metadata/darabonba-decisions.json", "sources/darabonba/modules/shadow.tea"} {
		t.Run(name, func(t *testing.T) {
			root := fixtureRepository(t)
			writeTestFile(t, root, name, []byte("tampered"))
			if err := Generate(context.Background(), root, false); err == nil {
				t.Fatal("tampering accepted")
			}
			if _, err := os.Stat(filepath.Join(root, "services")); !os.IsNotExist(err) {
				t.Fatal("wrote outputs before source validation")
			}
		})
	}
}
