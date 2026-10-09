package codegen

import (
	"context"
	"encoding/json/v2"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEndpointGenerationRejectsInvalidProjectionAndPolicyBeforeWrites(t *testing.T) {
	for name, mutate := range map[string]func(*productIR){
		"changed profile":     func(p *productIR) { p.Endpoints.Profile = "unreviewed" },
		"missing coordinates": func(p *productIR) { p.Endpoints.Overrides[0].Source.Line = 0 },
		"invalid origin":      func(p *productIR) { p.Endpoints.Overrides[0].URL = "https://user:secret@example.invalid" },
		"duplicate mapping":   func(p *productIR) { p.Endpoints.Overrides = append(p.Endpoints.Overrides, p.Endpoints.Overrides[0]) },
		"unreviewed network": func(p *productIR) {
			p.Policy = &capabilityPolicy{Endpoints: []endpointPolicy{{Network: "inner", Region: "cn-hangzhou", URL: "https://ecs-inner.cn-hangzhou.aliyuncs.com", Evidence: []string{"https://www.alibabacloud.com/help/en/sdk"}}}}
		},
		"missing evidence": func(p *productIR) {
			p.Policy = &capabilityPolicy{Endpoints: []endpointPolicy{{Network: "vpc", Region: "cn-hangzhou", URL: "https://ecs-vpc.cn-hangzhou.aliyuncs.com"}}}
		},
	} {
		t.Run(name, func(t *testing.T) {
			p := readProductIR(t, "ecs")
			mutate(&p)
			if _, err := renderEndpoints([]productIR{p}); err == nil {
				t.Fatal("invalid endpoint input accepted")
			}
		})
	}
	root := fullProductFixture(t)
	p := readProductIR(t, "ecs")
	p.Endpoints.Profile = "unreviewed"
	data, err := json.Marshal(p)
	if err != nil {
		t.Fatal(err)
	}
	writeTestFile(t, root, "models/ecs/ir.json", data)
	lockPath := filepath.Join(root, "models/manifest.json")
	lockData, err := os.ReadFile(lockPath)
	if err != nil {
		t.Fatal(err)
	}
	var pins productPins
	if err = json.Unmarshal(lockData, &pins); err != nil {
		t.Fatal(err)
	}
	for i := range pins.Files {
		if pins.Files[i].File == "models/ecs/ir.json" {
			pins.Files[i].SHA256 = digest(data)
		}
	}
	lockData, err = json.Marshal(pins)
	if err != nil {
		t.Fatal(err)
	}
	writeTestFile(t, root, "models/manifest.json", lockData)
	if err = GenerateProducts(context.Background(), root, false, nil); err == nil || !strings.Contains(err.Error(), "endpoint") {
		t.Fatal("invalid endpoint projection did not fail generation", err)
	}
	for _, dir := range []string{"service", "endpoint"} {
		if _, err = os.Stat(filepath.Join(root, dir)); !errors.Is(err, os.ErrNotExist) {
			t.Fatal("endpoint failure wrote partial files")
		}
	}
}
