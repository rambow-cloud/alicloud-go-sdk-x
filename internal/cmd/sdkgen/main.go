// Command sdkgen imports explicit metadata snapshots and generates repository RPC clients.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"github.com/rambow-cloud/alicloud-go-sdk-x/internal/codegen"
	"os"
	"os/signal"
	"strings"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	if err := run(ctx, os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return errors.New("usage: sdkgen import|generate|check [flags]")
	}
	switch args[0] {
	case "generate", "check":
		flags := flag.NewFlagSet(args[0], flag.ContinueOnError)
		root := flags.String("root", ".", "Repository root containing metadata and generated outputs")
		if err := flags.Parse(args[1:]); err != nil {
			return err
		}
		if flags.NArg() != 0 {
			return errors.New("codegen: no positional arguments expected")
		}
		if err := codegen.Generate(ctx, *root, args[0] == "check"); err != nil {
			return err
		}
		fmt.Println("Generation " + args[0] + " passed (offline).")
		return nil
	case "import":
		flags := flag.NewFlagSet("import", flag.ContinueOnError)
		product := flags.String("product", "", "Official product code")
		version := flags.String("version", "", "Official API version")
		pkg := flags.String("package", "", "Go service package")
		operations := flags.String("operations", "", "Comma-separated reviewed operations")
		raw := flags.String("raw-dir", "", "Optional already downloaded operation documents")
		out := flags.String("out", "", "Metadata output directory (required)")
		if err := flags.Parse(args[1:]); err != nil {
			return err
		}
		if flags.NArg() != 0 || *out == "" {
			return errors.New("codegen: import requires -out and no positional arguments")
		}
		return codegen.Import(ctx, nil, codegen.Manifest{SchemaVersion: 1, Product: *product, Version: *version, Style: "RPC", Package: *pkg, Service: *pkg}, strings.Split(*operations, ","), *raw, *out)
	default:
		return errors.New("codegen: unknown command")
	}
}
