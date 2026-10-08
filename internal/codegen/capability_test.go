package codegen

import (
	"bytes"
	"context"
	"encoding/json/v2"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func readPolicyProduct(t *testing.T, pkg string) productIR {
	t.Helper()
	// The loader validates the whole policy inventory, so supply every pinned product.
	products := []productIR{readProductIR(t, "ecs"), readProductIR(t, "sts"), readProductIR(t, "vpc")}
	if err := loadCapabilityPolicies(filepath.Join("..", ".."), products); err != nil {
		t.Fatal(err)
	}
	for _, p := range products {
		if p.Product == pkg {
			return p
		}
	}
	t.Fatal("missing product")
	return productIR{}
}

func TestPolicyRenderingCoverageAndExactWireNaming(t *testing.T) {
	for pkg, want := range map[string]int{"ecs": 5, "vpc": 1, "sts": 3} {
		p := readPolicyProduct(t, pkg)
		r, err := newProductRenderer(p)
		if err != nil {
			t.Fatal(err)
		}
		reviewed := 0
		for _, op := range r.operations {
			if r.capabilityCoverage(op).Status == "reviewed" {
				reviewed++
			}
		}
		if reviewed != want {
			t.Fatalf("%s reviewed=%d", pkg, reviewed)
		}
		first, err := renderProduct(p)
		if err != nil {
			t.Fatal(err)
		}
		again, err := renderProduct(p)
		if err != nil {
			t.Fatal(err)
		}
		for name, data := range first {
			if !bytes.Equal(data, again[name]) {
				t.Fatal("non-deterministic policy output", name)
			}
		}
		if p.PolicySHA256 == "" || !bytes.Contains(first["docs/products/"+pkg+".coverage.json"], []byte(p.PolicySHA256)) {
			t.Fatal("missing policy provenance")
		}
		if pkg == "ecs" {
			if !bytes.Contains(first["service/ecs/types.gen.go"], []byte("InstanceIDs []string `json:\"InstanceId,omitzero\"`")) {
				t.Fatal("naming exception changed wire binding")
			}
			if first["service/ecs/waiters.gen.go"] == nil {
				t.Fatal("missing waiter")
			}
		}
	}
}

func TestInvalidSparsePoliciesFailBeforeRendering(t *testing.T) {
	for name, mutate := range map[string]func(*capabilityPolicy){
		"unknown operation": func(p *capabilityPolicy) { p.Operations["RunInstances"] = p.Operations["DescribeImages"] },
		"unknown field":     func(p *capabilityPolicy) { p.FieldNames["DescribeImagesRequest#unknown"] = "Unknown" },
		"wrong wire case":   func(p *capabilityPolicy) { p.Operations["DescribeImages"].Paginator.Items = "images.Image" },
		"wrong field type":  func(p *capabilityPolicy) { p.Operations["DescribeImages"].Paginator.Total = "RequestId" },
		"invalid mode":      func(p *capabilityPolicy) { p.Operations["DescribeImages"].Paginator.Mode = "automatic" },
		"symbol collision":  func(p *capabilityPolicy) { p.Operations["DescribeInstanceStatus"].Waiter.Name = "Client" },
		"missing review": func(p *capabilityPolicy) {
			c := p.Operations["DescribeRegions"]
			c.Idempotent = nil
			p.Operations["DescribeRegions"] = c
		},
		"missing evidence": func(p *capabilityPolicy) {
			c := p.Operations["DescribeRegions"]
			c.Evidence = nil
			p.Operations["DescribeRegions"] = c
		},
		"unsafe token retries": func(p *capabilityPolicy) {
			c := p.Operations["AllocateDedicatedHosts"]
			c.Idempotent = pointerForTest(true)
			p.Operations["AllocateDedicatedHosts"] = c
		},
		"wide bounds": func(p *capabilityPolicy) {
			c := p.Operations["DescribeImages"]
			v := int64(2147483648)
			c.Constraints[0].Maximum = &v
			p.Operations["DescribeImages"] = c
		},
		"ambiguous states": func(p *capabilityPolicy) { p.Operations["DescribeInstanceStatus"].Waiter.Retry = []string{"Running"} },
		"unrelated sensitivity": func(p *capabilityPolicy) {
			c := p.Operations["DescribeImages"]
			c.SensitiveModels = []string{"AllocateDedicatedHostsRequest"}
			p.Operations["DescribeImages"] = c
		},
	} {
		t.Run(name, func(t *testing.T) {
			p := readPolicyProduct(t, "ecs")
			mutate(p.Policy)
			if _, err := renderProduct(p); err == nil {
				t.Fatal("invalid policy rendered")
			}
		})
	}
}
func pointerForTest[T any](v T) *T { return &v }

func TestConservativeReviewWithoutAdaptersProducesNoEmptyFiles(t *testing.T) {
	p := readPolicyProduct(t, "ecs")
	p.Policy.FieldNames = nil
	p.Policy.Operations = map[string]operationPolicy{"DescribeRegions": {Idempotent: pointerForTest(false), Evidence: []string{"docs/support.md"}}}
	files, err := renderProduct(p)
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range []string{"policy.gen.go", "paginators.gen.go", "waiters.gen.go", "capability_examples.gen_test.go"} {
		if _, ok := files["service/ecs/"+file]; ok {
			t.Fatal("empty capability artifact", file)
		}
	}
}

func copyPolicyFixture(t *testing.T, root string) {
	t.Helper()
	for _, pkg := range []string{"ecs", "sts", "vpc"} {
		data, err := os.ReadFile(filepath.Join("..", "..", "policies", pkg+".json"))
		if err != nil {
			t.Fatal(err)
		}
		writeTestFile(t, root, "policies/"+pkg+".json", data)
	}
}

func TestPolicyGuidePairsDoNotPermitUnknownArtifacts(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, root, "policies/README.md", []byte("English guide"))
	writeTestFile(t, root, "policies/README.zh-CN.md", []byte("Chinese guide"))
	if err := loadCapabilityPolicies(root, nil); err != nil {
		t.Fatal("paired guides rejected", err)
	}
	writeTestFile(t, root, "policies/unreviewed.md", []byte("unexpected artifact"))
	if err := loadCapabilityPolicies(root, nil); err == nil {
		t.Fatal("unknown artifact accepted")
	}
}

func TestPolicyLoaderStrictSourceBindingAndAtomicWrites(t *testing.T) {
	root := fullProductFixture(t)
	copyPolicyFixture(t, root)
	path := filepath.Join(root, "policies", "ecs.json")
	original, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	for _, data := range [][]byte{bytes.Replace(original, []byte(`"schemaVersion": 1`), []byte(`"schemaVersion": 2`), 1), bytes.Replace(original, []byte(`"schemaVersion": 1`), []byte(`"unknown": true, "schemaVersion": 1`), 1), bytes.Replace(original, []byte(`7119b2b63b79f769ae06ee8c544230cc7f60e603e70358aa7d960599770ab1dd`), []byte(strings.Repeat("0", 64)), 1)} {
		if err := os.WriteFile(path, data, 0600); err != nil {
			t.Fatal(err)
		}
		if err := GenerateProducts(context.Background(), root, false, nil); err == nil {
			t.Fatal("invalid policy accepted")
		}
		if _, err := os.Stat(filepath.Join(root, "service")); !errors.Is(err, os.ErrNotExist) {
			t.Fatal("invalid policy wrote outputs")
		}
	}
}

func TestRemovedPolicyReconcilesOnlyOwnedAdapters(t *testing.T) {
	root := fullProductFixture(t)
	copyPolicyFixture(t, root)
	ctx := context.Background()
	if err := GenerateProducts(ctx, root, false, nil); err != nil {
		t.Fatal(err)
	}
	writeTestFile(t, root, "service/ecs/manual.go", []byte("package ecs\n"))
	path := filepath.Join(root, "policies", "ecs.json")
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	var drift *DriftError
	if err := GenerateProducts(ctx, root, true, nil); !errors.As(err, &drift) {
		t.Fatal("policy removal drift undetected", err)
	}
	if _, err := os.Stat(filepath.Join(root, "service", "ecs", "paginators.gen.go")); err != nil {
		t.Fatal("check deleted adapters")
	}
	if err := GenerateProducts(ctx, root, false, nil); err != nil {
		t.Fatal(err)
	}
	for _, file := range []string{"paginators.gen.go", "waiters.gen.go", "policy.gen.go", "capability_examples.gen_test.go"} {
		if _, err := os.Stat(filepath.Join(root, "service", "ecs", file)); !errors.Is(err, os.ErrNotExist) {
			t.Fatal("stale adapter retained", file)
		}
	}
	if _, err := os.Stat(filepath.Join(root, "service", "ecs", "manual.go")); err != nil {
		t.Fatal("manual file lost")
	}
}

func TestPolicyUnknownJSONMembersRejected(t *testing.T) {
	var p capabilityPolicy
	if err := json.Unmarshal([]byte(`{"unknown":1}`), &p, json.RejectUnknownMembers(true)); err == nil {
		t.Fatal("unknown policy member accepted")
	}
}
