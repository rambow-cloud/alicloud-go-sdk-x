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

func aliasedRolePolicies() []struct {
	name   string
	mutate func(*capabilityPolicy)
} {
	return []struct {
		name   string
		mutate func(*capabilityPolicy)
	}{
		{"page and size", func(p *capabilityPolicy) {
			p.Operations["DescribeImages"].Paginator.Page = "PageSize"
		}},
		{"response total and page", func(p *capabilityPolicy) {
			p.Operations["DescribeImages"].Paginator.Total = "PageNumber"
		}},
		{"response total and size", func(p *capabilityPolicy) {
			p.Operations["DescribeImages"].Paginator.Total = "PageSize"
		}},
		{"dual page and size", func(p *capabilityPolicy) {
			p.Operations["DescribeInstances"].Paginator.Page = "PageSize"
		}},
		{"dual limit and page", func(p *capabilityPolicy) {
			p.Operations["DescribeInstances"].Paginator.Limit = "PageNumber"
		}},
		{"dual limit and size", func(p *capabilityPolicy) {
			p.Operations["DescribeInstances"].Paginator.Limit = "PageSize"
		}},
		{"waiter page and size", func(p *capabilityPolicy) {
			p.Operations["DescribeInstanceStatus"].Waiter.Page = "PageSize"
		}},
		{"waiter member ID and state", func(p *capabilityPolicy) {
			p.Operations["DescribeInstanceStatus"].Waiter.State = "InstanceId"
		}},
	}
}

func TestAliasedCapabilityRolesRejectedBeforeRendering(t *testing.T) {
	for _, test := range aliasedRolePolicies() {
		t.Run(test.name, func(t *testing.T) {
			p := readPolicyProduct(t, "ecs")
			test.mutate(p.Policy)
			if _, err := newProductRenderer(p); err == nil || !strings.Contains(err.Error(), "roles must be distinct") {
				t.Fatalf("expected policy role rejection, got %v", err)
			}
		})
	}
}

func TestCapabilityRolesMaySharePathsAcrossModelsAndAdapters(t *testing.T) {
	p := readPolicyProduct(t, "ecs")
	status := p.Policy.Operations["DescribeInstanceStatus"]
	if status.Paginator.Page != status.Waiter.Page || status.Paginator.Size != status.Waiter.Size {
		t.Fatal("fixture must share fields across separate adapters")
	}
	paginator := p.Policy.Operations["DescribeInstances"].Paginator
	if paginator.InputToken != paginator.OutputToken {
		t.Fatal("fixture must share a token path across request and response models")
	}
	if _, err := renderProduct(p); err != nil {
		t.Fatal("valid cross-model or cross-adapter paths rejected", err)
	}
	// Token-only policy has no page roles and still permits matching native token names.
	paginator.Mode = "tokens"
	paginator.Page, paginator.Size, paginator.Total = "", "", ""
	if _, err := renderProduct(p); err != nil {
		t.Fatal("valid token-only policy rejected", err)
	}
}

func TestAliasedCapabilityRolesLeaveExistingAndStaleOutputsUntouched(t *testing.T) {
	root := fullProductFixture(t)
	copyPolicyFixture(t, root)
	before := map[string][]byte{
		"service/ecs/paginators.gen.go":   []byte(productGenerated + "package ecs\n// Existing owned output.\n"),
		"service/ecs/stale.gen.go":        []byte(productGenerated + "package ecs\n// Stale owned output.\n"),
		"service/ecs/manual.go":           []byte("package ecs\n// Manual output.\n"),
		"docs/products/ecs.coverage.json": []byte(productJSON + "\n  \"existing\": true\n}\n"),
	}
	for path, data := range before {
		writeTestFile(t, root, path, data)
	}
	for _, test := range aliasedRolePolicies() {
		t.Run(test.name, func(t *testing.T) {
			p := readPolicyProduct(t, "ecs")
			test.mutate(p.Policy)
			data, err := json.Marshal(p.Policy)
			if err != nil {
				t.Fatal(err)
			}
			writeTestFile(t, root, "policies/ecs.json", data)
			for _, check := range []bool{true, false} {
				if err := GenerateProducts(context.Background(), root, check, nil); err == nil || !strings.Contains(err.Error(), "roles must be distinct") {
					t.Fatalf("check=%t: expected policy role rejection, got %v", check, err)
				}
				for path, expected := range before {
					actual, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(path)))
					if err != nil || !bytes.Equal(expected, actual) {
						t.Fatalf("check=%t modified/deleted %s: %v", check, path, err)
					}
				}
				if _, err := os.Stat(filepath.Join(root, "service", "vpc")); !os.IsNotExist(err) {
					t.Fatal("invalid policy wrote another product's outputs", err)
				}
			}
		})
	}
}
