package codegen

import (
	"bytes"
	"context"
	"encoding/json/v2"
	"os"
	"path/filepath"
	"testing"
)

func TestExplicitResponsePagePolicyFailsBeforeWrites(t *testing.T) {
	root := fullProductFixture(t)
	copyPolicyFixture(t, root)
	before := []byte(productGenerated + "package vpc\n// Existing output must survive invalid policies.\n")
	writeTestFile(t, root, "service/vpc/paginators.gen.go", before)
	for _, field := range []string{"UnknownPage", "TotalCount", "PageSize", "EcGrantRelations"} {
		t.Run(field, func(t *testing.T) {
			p := readPolicyProduct(t, "vpc")
			p.Policy.Operations["DescribeEcGrantRelation"].Paginator.OutputPage = field
			data, err := json.Marshal(p.Policy)
			if err != nil {
				t.Fatal(err)
			}
			writeTestFile(t, root, "policies/vpc.json", data)
			for _, check := range []bool{true, false} {
				if err := GenerateProducts(context.Background(), root, check, nil); err == nil {
					t.Fatal("invalid response page policy accepted")
				}
				got, err := os.ReadFile(filepath.Join(root, "service/vpc/paginators.gen.go"))
				if err != nil || !bytes.Equal(got, before) {
					t.Fatal("invalid policy changed owned output", err)
				}
			}
		})
	}
}

func TestPageNumericWidthsDoNotChangeOperationDiscovery(t *testing.T) {
	for _, product := range []string{"ecs", "vpc"} {
		p := readPolicyProduct(t, product)
		with, err := renderProduct(p)
		if err != nil {
			t.Fatal(err)
		}
		p.Policy.Operations = map[string]operationPolicy{}
		p.PolicySHA256 = ""
		without, err := renderProduct(p)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(with["service/"+product+"/types.gen.go"], without["service/"+product+"/types.gen.go"]) {
			t.Fatal("page policy changed model emission")
		}
		if len(p.Operations) == 0 {
			t.Fatal("operation fixture absent")
		}
	}
}
