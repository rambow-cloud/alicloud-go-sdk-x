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

func mutateDSL(t *testing.T, root, pkg string, change func(*DSLProduct)) {
	t.Helper()
	dir := filepath.Join(root, "metadata", pkg)
	var projection DSLProduct
	if err := readJSON(filepath.Join(dir, "dsl.json"), &projection, true); err != nil {
		t.Fatal(err)
	}
	change(&projection)
	data, err := json.Marshal(projection)
	if err != nil {
		t.Fatal(err)
	}
	writeTestFile(t, root, "metadata/"+pkg+"/dsl.json", data)
	var manifest Manifest
	if err := readJSON(filepath.Join(dir, "manifest.json"), &manifest, true); err != nil {
		t.Fatal(err)
	}
	manifest.DSL.SHA256 = digest(data)
	data, err = json.Marshal(manifest)
	if err != nil {
		t.Fatal(err)
	}
	writeTestFile(t, root, "metadata/"+pkg+"/manifest.json", data)
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
			mutateDSL(t, root, "sts", change)
			if err := Generate(context.Background(), root, false); err == nil {
				t.Fatal("unreviewed change accepted")
			}
			if _, err := os.Stat(filepath.Join(root, "services")); !os.IsNotExist(err) {
				t.Fatal("wrote outputs before cross-check")
			}
		})
	}
}

func TestApprovedHiddenInputCannotBecomeRequired(t *testing.T) {
	root := fixtureRepository(t)
	mutateDSL(t, root, "ecs", func(d *DSLProduct) {
		for i := range d.Operations {
			if d.Operations[i].Name == "DescribeRegions" {
				for j := range d.Operations[i].Inputs {
					if d.Operations[i].Inputs[j].Wire == "OwnerId" {
						d.Operations[i].Inputs[j].Schema.Required = true
					}
				}
			}
		}
	})
	if err := Generate(context.Background(), root, false); err == nil || !strings.Contains(err.Error(), "unselected required DSL input") {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, "services")); !os.IsNotExist(err) {
		t.Fatal("wrote before validation")
	}
}

func TestApprovedRequirednessCannotReverseDirection(t *testing.T) {
	root := fixtureRepository(t)
	mutateDSL(t, root, "sts", func(d *DSLProduct) {
		for i := range d.Operations[0].Inputs {
			if d.Operations[0].Inputs[i].Wire == "RoleArn" {
				d.Operations[0].Inputs[i].Schema.Required = true
			}
		}
	})
	mutateSnapshot(t, filepath.Join(root, "metadata", "sts"), "AssumeRole", func(d map[string]any) {
		for _, raw := range d["parameters"].([]any) {
			p := raw.(map[string]any)
			if p["name"] == "RoleArn" {
				p["schema"].(map[string]any)["required"] = false
			}
		}
	})
	if err := Generate(context.Background(), root, false); err == nil || !strings.Contains(err.Error(), "unreviewed requiredness direction") {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, "services")); !os.IsNotExist(err) {
		t.Fatal("wrote before validation")
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

func TestIndexedBindingsRejectDriftBeforeOutputWrites(t *testing.T) {
	cases := map[string]func(*DSLInput){
		"missing alias": func(in *DSLInput) { in.MetadataBindings = in.MetadataBindings[:7] },
		"forged alias":  func(in *DSLInput) { in.MetadataBindings[0] = "Filter.0.Key" },
		"leaf type":     func(in *DSLInput) { in.Schema.Items.Properties["Key"].Type = "integer" },
		"leaf case": func(in *DSLInput) {
			in.Schema.Items.Properties["key"] = in.Schema.Items.Properties["Key"]
			delete(in.Schema.Items.Properties, "Key")
		},
		"leaf requiredness": func(in *DSLInput) { in.Schema.Items.Properties["Key"].Required = true },
		"root requiredness": func(in *DSLInput) { in.Schema.Required = true },
	}
	for name, change := range cases {
		t.Run(name, func(t *testing.T) {
			root := fixtureRepository(t)
			mutateDSL(t, root, "ecs", func(d *DSLProduct) {
				for i := range d.Operations {
					if d.Operations[i].Name != "DescribeInstances" {
						continue
					}
					for j := range d.Operations[i].Inputs {
						if d.Operations[i].Inputs[j].Wire == "Filter" {
							change(&d.Operations[i].Inputs[j])
						}
					}
				}
			})
			if err := Generate(context.Background(), root, false); err == nil {
				t.Fatal("indexed drift accepted")
			}
			if _, err := os.Stat(filepath.Join(root, "services")); !os.IsNotExist(err) {
				t.Fatal("wrote outputs before binding validation")
			}
		})
	}
}

func TestIndexedMetadataAndCaseExceptionsAreRevalidated(t *testing.T) {
	for _, name := range []string{"zero", "leading zero", "case", "type", "requiredness", "duplicate", "direct", "tag case", "tag type", "tag requiredness"} {
		t.Run(name, func(t *testing.T) {
			root := fixtureRepository(t)
			mutateSnapshot(t, filepath.Join(root, "metadata", "ecs"), "DescribeInstances", func(d map[string]any) {
				parameters := d["parameters"].([]any)
				for _, raw := range parameters {
					p := raw.(map[string]any)
					if p["name"] == "Tag" && name == "tag case" {
						properties := p["schema"].(map[string]any)["items"].(map[string]any)["properties"].(map[string]any)
						properties["KEY"] = properties["key"]
						delete(properties, "key")
					}
					if p["name"] == "Tag" && (name == "tag type" || name == "tag requiredness") {
						field := p["schema"].(map[string]any)["items"].(map[string]any)["properties"].(map[string]any)["key"].(map[string]any)
						if name == "tag type" {
							field["type"] = "integer"
						} else {
							field["required"] = true
						}
					}
					if p["name"] != "Filter.1.Key" {
						continue
					}
					switch name {
					case "zero":
						p["name"] = "Filter.0.Key"
					case "leading zero":
						p["name"] = "Filter.01.Key"
					case "case":
						p["name"] = "Filter.1.key"
					case "type":
						p["schema"].(map[string]any)["type"] = "integer"
					case "requiredness":
						p["schema"].(map[string]any)["required"] = true
					case "duplicate":
						d["parameters"] = append(parameters, p)
					case "direct":
						d["parameters"] = append(parameters, map[string]any{"name": "Filter", "in": "query", "schema": map[string]any{"type": "array", "items": map[string]any{"type": "object"}}})
					}
				}
			})
			if err := Generate(context.Background(), root, false); err == nil {
				t.Fatal("metadata drift accepted")
			}
			if _, err := os.Stat(filepath.Join(root, "services")); !os.IsNotExist(err) {
				t.Fatal("wrote before cross-check")
			}
		})
	}
}

func TestIndexedDSLLeafDoesNotInferArrayLimit(t *testing.T) {
	root := &DSLShape{Type: "array", Items: &DSLShape{Type: "object", Properties: map[string]*DSLShape{"Key": {Type: "string"}}}}
	leaf, err := indexedDSLLeaf(root, "100.Key")
	if err != nil || leaf.Type != "string" {
		t.Fatal(leaf, err)
	}
}
