// Command xmltraits reads verified native helper source and prints XML root facts.
// It does not fetch or execute source or write SDK files.
package main

import (
	"context"
	"encoding/json/v2"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"

	"github.com/rambow-cloud/alicloud-go-sdk-x/internal/nativewire"
)

type sourcePin struct {
	Name   string `json:"name"`
	SHA256 string `json:"sha256"`
}
type pins struct {
	SchemaVersion  int       `json:"schemaVersion"`
	RegistrySymbol string    `json:"registrySymbol"`
	Registry       sourcePin `json:"registry"`
	Models         sourcePin `json:"models"`
}

func read(path string) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, 8<<20+1))
	if err != nil {
		return nil, err
	}
	if len(data) > 8<<20 {
		return nil, nativewire.ErrInvalid
	}
	return data, nil
}

func run(ctx context.Context, pinFile, registryFile, modelsFile string, output io.Writer) error {
	data, err := read(pinFile)
	if err != nil {
		return err
	}
	var p pins
	if err = json.Unmarshal(data, &p); err != nil || p.SchemaVersion != 1 {
		return nativewire.ErrInvalid
	}
	registry, err := read(registryFile)
	if err != nil {
		return err
	}
	models, err := read(modelsFile)
	if err != nil {
		return err
	}
	inventory, err := nativewire.ParseXML(ctx, nativewire.Source{Name: p.Registry.Name, SHA256: p.Registry.SHA256, Bytes: registry}, nativewire.Source{Name: p.Models.Name, SHA256: p.Models.SHA256, Bytes: models}, p.RegistrySymbol)
	if err != nil {
		return err
	}
	encoded, err := json.Marshal(inventory, json.Deterministic(true))
	if err != nil {
		return err
	}
	encoded = append(encoded, '\n')
	if err := ctx.Err(); err != nil {
		return err
	}
	n, err := output.Write(encoded)
	if err == nil && n != len(encoded) {
		return io.ErrShortWrite
	}
	return err
}

func main() {
	pinFile := flag.String("pins", "", "reviewed native source pins")
	registry := flag.String("registry", "", "complete native registry source file")
	models := flag.String("models", "", "complete native model source file")
	flag.Parse()
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	if err := run(ctx, *pinFile, *registry, *models, os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "xmltraits: extraction failed (%T)\n", err)
		os.Exit(1)
	}
}
