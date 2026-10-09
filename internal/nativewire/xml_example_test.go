package nativewire_test

import (
	"context"
	"fmt"

	"github.com/rambow-cloud/alicloud-go-sdk-x/internal/nativewire"
)

func ExampleParseXML() {
	// Source pins are supplied by a reviewed plan; no code is executed.
	inventory, err := nativewire.ParseXML(context.Background(), source("registry.go", registry), source("models.go", models), "catalog")
	if err != nil {
		panic(err)
	}
	for _, root := range inventory.Roots {
		if root.Name == "" {
			fmt.Println(root.Action, root.Kind)
		} else {
			fmt.Println(root.Action, root.Kind, root.Name)
		}
	}
	// Output:
	// Flat unsupported
	// ReadNested structured Result
	// ReadText scalar Location
	// Unbound unsupported
}
