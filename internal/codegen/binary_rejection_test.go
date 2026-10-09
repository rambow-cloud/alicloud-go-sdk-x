package codegen

import (
	"context"
	"encoding/json/v2"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestInvalidBinaryIRFailsBeforeWrites(t *testing.T) {
	for _, mode := range []string{"header-model", "header-location", "body-kind", "body-encoding", "stream-kind"} {
		t.Run(mode, func(t *testing.T) {
			root := fullProductFixture(t)
			product := readProductIR(t, "fc")
			var binary *productOperation
			for i := range product.Operations {
				if product.Operations[i].Name == "InvokeFunction" {
					binary = &product.Operations[i]
				}
			}
			if binary == nil {
				t.Fatal("missing fixture")
			}
			switch mode {
			case "header-model":
				binary.HeaderBindings.Model = "unknown"
			case "header-location":
				binary.HeaderBindings.Fields[0].Location = "inferred"
			case "body-encoding":
				for i := range binary.Bindings {
					if binary.Bindings[i].Location == "binary-body" {
						binary.Bindings[i].Encoding = "json"
					}
				}
			case "body-kind", "stream-kind":
				id := binary.Roots.Request.Ref
				if mode == "stream-kind" {
					id = binary.Roots.Body.Ref
				}
				for i := range product.Models {
					if product.Models[i].ID == id {
						for j := range product.Models[i].Fields {
							if product.Models[i].Fields[j].WireName == "body" {
								product.Models[i].Fields[j].Type = productType{Kind: "scalar", DSLType: "string", WireType: "string"}
							}
						}
					}
				}
			}
			encoded, err := json.Marshal(product)
			if err != nil {
				t.Fatal(err)
			}
			writeTestFile(t, root, "models/fc/ir.json", encoded)
			var pins productPins
			data, err := os.ReadFile(filepath.Join(root, "models/manifest.json"))
			if err != nil {
				t.Fatal(err)
			}
			if err := json.Unmarshal(data, &pins); err != nil {
				t.Fatal(err)
			}
			for i := range pins.Files {
				if pins.Files[i].File == "models/fc/ir.json" {
					pins.Files[i].SHA256 = digest(encoded)
				}
			}
			data, err = json.Marshal(pins)
			if err != nil {
				t.Fatal(err)
			}
			writeTestFile(t, root, "models/manifest.json", data)
			if err := GenerateProducts(context.Background(), root, false, []string{"fc/InvokeFunction"}); err == nil {
				t.Fatal("invalid binary IR emitted")
			}
			if _, err := os.Stat(filepath.Join(root, "service")); !errors.Is(err, os.ErrNotExist) {
				t.Fatal("invalid binary IR wrote outputs")
			}
		})
	}
}
