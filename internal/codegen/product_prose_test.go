package codegen

import (
	"bytes"
	"context"
	"encoding/json/v2"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func proseProducts(t *testing.T) []productIR {
	t.Helper()
	products := []productIR{readPolicyProduct(t, "sts"), readPolicyProduct(t, "ecs"), readPolicyProduct(t, "vpc")}
	if err := loadProductProse(filepath.Join("..", ".."), products); err != nil {
		t.Fatal(err)
	}
	return products
}
func TestOptionalProseChangesOnlyCommentsDocsAndNotices(t *testing.T) {
	for _, p := range proseProducts(t) {
		with, err := renderProduct(p)
		if err != nil {
			t.Fatal(err)
		}
		count := len(p.MetadataProse)
		p.MetadataProse = nil
		p.MetadataProseRevision = ""
		without, err := renderProduct(p)
		if err != nil {
			t.Fatal(err)
		}
		for file, data := range without {
			if strings.HasSuffix(file, "/types.gen.go") {
				fingerprint := func(source []byte) []byte {
					f, err := parser.ParseFile(token.NewFileSet(), "types.go", source, parser.SkipObjectResolution)
					if err != nil {
						t.Fatal(err)
					}
					var b bytes.Buffer
					err = ast.Fprint(&b, nil, f, func(name string, value reflect.Value) bool {
						return value.Type() != reflect.TypeOf(token.Pos(0)) && name != "Doc" && name != "Comment" && name != "Comments"
					})
					if err != nil {
						t.Fatal(err)
					}
					return b.Bytes()
				}
				if !bytes.Equal(fingerprint(data), fingerprint(with[file])) {
					t.Fatal("prose changed model AST", file)
				}
				continue
			}
			if strings.HasSuffix(file, "/NOTICE") || file == "docs/products/"+p.Product+".md" || file == "docs/products/"+p.Product+".zh-CN.md" || strings.HasSuffix(file, ".documentation.json") {
				continue
			}
			if !bytes.Equal(data, with[file]) {
				t.Fatal("prose changed non-documentation output", file)
			}
		}
		var coverage productDocCoverage
		if json.Unmarshal(with["docs/products/"+p.Product+".documentation.json"], &coverage) != nil || coverage.MetadataEnglishFields != count || len(coverage.MetadataSources) != count {
			t.Fatal("metadata coverage lost")
		}
		if !bytes.Contains(with["service/"+p.Product+"/NOTICE"], []byte("aliyun-openapi-meta")) {
			t.Fatal("missing metadata attribution")
		}
	}
}

func proseFixture(t *testing.T) string {
	t.Helper()
	root := fullProductFixture(t)
	production := filepath.Join("..", "..")
	var source proseSourceManifest
	if err := readJSON(filepath.Join(production, "sources", "openapi-meta", "prose", "manifest.json"), &source, true); err != nil {
		t.Fatal(err)
	}
	var projection productProseFile
	if err := readJSON(filepath.Join(production, "metadata", "prose", "sts.json"), &projection, true); err != nil {
		t.Fatal(err)
	}
	var key string
	for _, k := range sortedKeys(projection.Fields) {
		if projection.Fields[k].Operation == "AssumeRole" {
			key = k
			break
		}
	}
	if key == "" {
		t.Fatal("no AssumeRole prose")
	}
	entry := projection.Fields[key]
	pins := source.Files[:0]
	for _, pin := range source.Files {
		if pin.File == "LICENSE" || pin.File == entry.File {
			pins = append(pins, pin)
			b, err := os.ReadFile(filepath.Join(production, "sources", "openapi-meta", "prose", filepath.FromSlash(pin.File)))
			if err != nil {
				t.Fatal(err)
			}
			writeTestFile(t, root, "sources/openapi-meta/prose/"+pin.File, b)
		}
	}
	source.Files = pins
	manifestBytes, err := json.Marshal(source)
	if err != nil {
		t.Fatal(err)
	}
	writeTestFile(t, root, "sources/openapi-meta/prose/manifest.json", manifestBytes)
	for _, product := range []string{"sts", "ecs", "vpc"} {
		ir, err := os.ReadFile(filepath.Join(root, "models", product, "ir.json"))
		if err != nil {
			t.Fatal(err)
		}
		p := readProductIR(t, product)
		projection.Product = product
		projection.Version = p.Version
		projection.IRSHA256 = digest(ir)
		projection.CorpusManifestSHA256 = digest(manifestBytes)
		projection.Fields = map[string]metadataProse{}
		projection.Reasons = nil
		if product == "sts" {
			projection.Fields[key] = entry
		}
		b, err := json.Marshal(projection)
		if err != nil {
			t.Fatal(err)
		}
		writeTestFile(t, root, "metadata/prose/"+product+".json", b)
	}
	return root
}
func TestProseDriftAndInvalidBindingsLeaveOwnedOutputsUntouched(t *testing.T) {
	for _, test := range []string{"source", "unknown-field", "source-text", "existing-English", "unlisted-source"} {
		t.Run(test, func(t *testing.T) {
			root := proseFixture(t)
			path := filepath.Join(root, "metadata", "prose", "sts.json")
			var p productProseFile
			if err := readJSON(path, &p, true); err != nil {
				t.Fatal(err)
			}
			var key string
			for k := range p.Fields {
				key = k
			}
			entry := p.Fields[key]
			switch test {
			case "source":
				writeTestFile(t, root, "sources/openapi-meta/prose/"+entry.File, []byte("modified"))
			case "unlisted-source":
				writeTestFile(t, root, "sources/openapi-meta/prose/canonical/extra.json", []byte("{}"))
			case "unknown-field":
				delete(p.Fields, key)
				p.Fields["Unknown#field"] = entry
			case "source-text":
				entry.English = "Different prose."
				entry.TextSHA256 = digest([]byte(entry.English))
				p.Fields[key] = entry
			case "existing-English":
				entry.FieldSource = productSource{}
				p.Fields["AssumeRoleRequest#roleArn"] = entry
			}
			b, err := json.Marshal(p)
			if err != nil {
				t.Fatal(err)
			}
			writeTestFile(t, root, "metadata/prose/sts.json", b)
			before := []byte(productGenerated + "package sts\n// Existing owned output.\n")
			writeTestFile(t, root, "service/sts/types.gen.go", before)
			if err := GenerateProducts(context.Background(), root, false, nil); err == nil {
				t.Fatal("invalid prose accepted")
			}
			got, err := os.ReadFile(filepath.Join(root, "service", "sts", "types.gen.go"))
			if err != nil || !bytes.Equal(got, before) {
				t.Fatal("invalid prose changed output", err)
			}
		})
	}
}
