package rpcmodel

import (
	"context"
	"testing"
)

func TestParameterLocationsRejectUnknownAndNestedLocations(t *testing.T) {
	for _, input := range []any{
		struct {
			Value string `json:"Value" rpcLocation:"header"`
		}{"x"},
		struct {
			Nested struct {
				Value string `json:"Value" rpcLocation:"form"`
			} `json:"Nested"`
		}{},
		struct {
			Value []int `json:"Value" rpc:"simple"`
		}{[]int{1}},
		struct {
			First  string `json:"Value" rpcLocation:"form"`
			Second string `json:"Value" rpcLocation:"form"`
		}{"x", "y"},
	} {
		if _, _, err := Parameters(context.Background(), input); err == nil {
			t.Fatal("invalid location/encoding/collision accepted")
		}
	}
}
