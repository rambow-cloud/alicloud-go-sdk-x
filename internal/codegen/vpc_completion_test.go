package codegen

import (
	"strings"
	"testing"
)

func TestVPCCompletionRejectsUnsafeIRAndReusesRenamedActions(t *testing.T) {
	for _, mutation := range []func(*productIR){
		func(p *productIR) {
			for i := range p.Operations {
				if p.Operations[i].Name == "CreateVpnAttachment" {
					p.Operations[i].Protocol.Method = "GET"
				}
			}
		},
		func(p *productIR) {
			for i := range p.Operations {
				for j := range p.Operations[i].Bindings {
					if p.Operations[i].Bindings[j].Location == "form" {
						p.Operations[i].Bindings[j].Location = "header"
					}
				}
			}
		},
		func(p *productIR) {
			for i := range p.Operations {
				if p.Operations[i].Name == "DescribeVpnGatewayAvailableZones" {
					p.Operations[i].Protocol.Method = "DELETE"
				}
			}
		},
		func(p *productIR) {
			for i := range p.Models {
				if p.Models[i].ID == "GrantInstanceToVbrRequest" {
					for j := range p.Models[i].Fields {
						if p.Models[i].Fields[j].DSLName == "vbrInstanceIds" {
							p.Models[i].Fields[j].Type.Items.DSLType = "long"
						}
					}
				}
			}
		},
	} {
		p := readProductIR(t, "vpc")
		mutation(&p)
		if _, err := renderProduct(p); err == nil {
			t.Fatal("invalid wire contract reached emission")
		}
	}
	p := readProductIR(t, "vpc")
	for i := range p.Operations {
		if p.Operations[i].Name == "DescribeVpnGatewayAvailableZones" {
			p.Operations[i].Name = "InspectZones"
			p.Operations[i].Protocol.Action = "InspectZones"
		}
	}
	files, err := renderProduct(p)
	if err != nil {
		t.Fatal(err)
	}
	methods := string(files["service/vpc/operations.gen.go"])
	if !strings.Contains(methods, "func (c *Client) InspectZones(") || !strings.Contains(methods, `"GET",`) {
		t.Fatal("renamed GET did not use shared backend")
	}
	types := string(files["service/vpc/types.gen.go"])
	if !strings.Contains(types, `rpcLocation:"form"`) || !strings.Contains(types, `rpc:"simple"`) {
		t.Fatal("location or encoding not emitted")
	}
}
