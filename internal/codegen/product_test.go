package codegen

import (
	"bytes"
	"context"
	"encoding/json/v2"
	"errors"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"
)

func readProductIR(t *testing.T, pkg string) productIR {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "..", "models", pkg, "ir.json"))
	if err != nil {
		t.Fatal(err)
	}
	var p productIR
	if err := json.Unmarshal(data, &p); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestCompleteProductsDeterministicModelsMethodsAndExamples(t *testing.T) {
	for pkg, want := range map[string]int{"ecs": 283, "sts": 1, "vpc": 295} {
		t.Run(pkg, func(t *testing.T) {
			p := readProductIR(t, pkg)
			r, err := newProductRenderer(p)
			if err != nil {
				t.Fatal(err)
			}
			if len(r.operations) != want {
				t.Fatal("unexpected emission count", len(r.operations))
			}
			files, err := renderProduct(p)
			if err != nil {
				t.Fatal(err)
			}
			slices.Reverse(p.Operations)
			slices.Reverse(p.Models)
			again, err := renderProduct(p)
			if err != nil {
				t.Fatal(err)
			}
			for name, data := range files {
				if !bytes.Equal(data, again[name]) {
					t.Fatal("unstable output", name)
				}
			}
			methods := string(files["service/"+pkg+"/operations.gen.go"])
			examples := string(files["service/"+pkg+"/examples.gen_test.go"])
			types := string(files["service/"+pkg+"/types.gen.go"])
			if strings.Count(methods, "func (c *Client)") != want || strings.Count(examples, "func ExampleClient_") != want || strings.Count(types, "type ") != len(r.models) {
				t.Fatal("model/method/example omissions")
			}
			var report struct {
				Emitted    int                        `json:"emitted"`
				Operations []productCoverageOperation `json:"operations"`
			}
			if err := json.Unmarshal(files["docs/products/"+pkg+".coverage.json"], &report); err != nil {
				t.Fatal(err)
			}
			emitted := 0
			for _, op := range report.Operations {
				if op.Status == "emitted" {
					emitted++
				}
				if op.Source.File == "" || op.Source.Line < 1 {
					t.Fatal("missing source evidence")
				}
				if op.Status == "unsupported" && (len(op.Reasons) == 0 || op.Reasons[0].Source.Line < 1) {
					t.Fatal("missing unsupported evidence")
				}
			}
			if emitted != want || report.Emitted != want {
				t.Fatal("report conflates lowering/emission")
			}
			for _, op := range r.operations {
				if !strings.Contains(methods, "type "+op.Name+"API interface") || !strings.Contains(types, "type "+op.Name+"Input struct") || !strings.Contains(types, "type "+op.Name+"Output struct") {
					t.Fatal("missing typed operation", op.Name)
				}
			}
			if pkg == "ecs" {
				for _, fragment := range []string{"ImageOwnerID *int64", "PageNumber *int32", "DryRun *bool", "type DescribeImagesOutputImagesImageDiskDeviceMappingsDiskDeviceMapping struct"} {
					if !strings.Contains(types, fragment) {
						t.Fatal("full shape/width missing", fragment)
					}
				}
			}
		})
	}
}

func fullProductFixture(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	for _, base := range []string{"models", "sources/darabonba"} {
		source := filepath.Join("..", "..", filepath.FromSlash(base))
		if err := filepath.WalkDir(source, func(path string, entry fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if entry.IsDir() {
				return nil
			}
			relative, err := filepath.Rel(source, path)
			if err != nil {
				return err
			}
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			writeTestFile(t, root, filepath.Join(base, relative), data)
			return nil
		}); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

func TestProductSupportAndCorruptionFailBeforeWrites(t *testing.T) {
	root := fullProductFixture(t)
	for _, selected := range [][]string{{"ecs/Unknown"}, {"ecs/RunInstances"}, {"sts/GetCallerIdentity"}, {"DescribeImages"}, {"ecs/DescribeImages", "sts/AssumeRoleWithSAML"}} {
		if err := GenerateProducts(context.Background(), root, false, selected); err == nil {
			t.Fatal("invalid selection accepted", selected)
		}
		if _, err := os.Stat(filepath.Join(root, "service")); !errors.Is(err, os.ErrNotExist) {
			t.Fatal("wrote before validating selection")
		}
	}
	path := filepath.Join(root, "models", "ecs", "ir.json")
	original, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, append(original, ' '), 0600); err != nil {
		t.Fatal(err)
	}
	if err := GenerateProducts(context.Background(), root, false, nil); err == nil || !strings.Contains(err.Error(), "checksum") {
		t.Fatal("IR corruption accepted", err)
	}
	if _, err := os.Stat(filepath.Join(root, "service")); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("corruption wrote output")
	}
}

func TestProductReconciliationAndIndependentOwnership(t *testing.T) {
	root := fullProductFixture(t)
	ctx := context.Background()
	if err := GenerateProducts(ctx, root, true, nil); err == nil {
		t.Fatal("missing output accepted")
	}
	if _, err := os.Stat(filepath.Join(root, "service")); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("check wrote")
	}
	writeTestFile(t, root, "services/ecs/sdk.gen.go", []byte(generated+"package ecs\n"))
	if err := GenerateProducts(ctx, root, false, []string{"ecs/DescribeImages", "sts/AssumeRole"}); err != nil {
		t.Fatal(err)
	}
	legacy, _ := os.ReadFile(filepath.Join(root, "services", "ecs", "sdk.gen.go"))
	if string(legacy) != generated+"package ecs\n" {
		t.Fatal("legacy output changed")
	}
	if _, err := os.Stat(filepath.Join(root, "service", "vpc", "operations.gen.go")); err != nil {
		t.Fatal("selection silently narrowed products")
	}
	writeTestFile(t, root, "service/ecs/obsolete.gen.go", []byte(productGenerated+"package ecs\n"))
	writeTestFile(t, root, "service/ecs/manual.go", []byte("package ecs\n"))
	changed := filepath.Join(root, "service", "sts", "types.gen.go")
	original, err := os.ReadFile(changed)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(changed, append(slices.Clone(original), []byte("// drift\n")...), 0600); err != nil {
		t.Fatal(err)
	}
	var drift *DriftError
	if err := GenerateProducts(ctx, root, true, nil); !errors.As(err, &drift) || len(drift.Files) != 2 {
		t.Fatal("drift not recorded", err)
	}
	if data, _ := os.ReadFile(changed); bytes.Equal(data, original) {
		t.Fatal("check repaired drift")
	}
	if err := GenerateProducts(ctx, root, false, nil); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, "service", "ecs", "obsolete.gen.go")); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("stale output retained")
	}
	if _, err := os.Stat(filepath.Join(root, "service", "ecs", "manual.go")); err != nil {
		t.Fatal("handwritten file lost")
	}
	if data, _ := os.ReadFile(changed); !bytes.Equal(data, original) {
		t.Fatal("write did not repair")
	}
}

func TestProductPreflightRejectsProtectedFilesAndSymlinks(t *testing.T) {
	root := fullProductFixture(t)
	writeTestFile(t, root, "service/sts/types.gen.go", []byte("package sts\n"))
	if err := GenerateProducts(context.Background(), root, false, nil); err == nil {
		t.Fatal("manual file overwritten")
	}
	if _, err := os.Stat(filepath.Join(root, "docs", "products")); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("partial output before preflight")
	}
	if err := os.Remove(filepath.Join(root, "service", "sts", "types.gen.go")); err != nil {
		t.Fatal(err)
	}
	out := t.TempDir()
	if err := os.Symlink(out, filepath.Join(root, "service", "ecs")); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	if err := GenerateProducts(context.Background(), root, false, nil); err == nil {
		t.Fatal("symlink accepted")
	}
	entries, err := os.ReadDir(out)
	if err != nil || len(entries) != 0 {
		t.Fatal("symlink target changed", err)
	}
}

func TestProductRejectsInvalidBindingProtocolShapeAndCollision(t *testing.T) {
	for _, mutate := range []func(*productIR){
		func(p *productIR) { p.Operations[0].Protocol.Method = "GET" },
		func(p *productIR) { p.Operations[0].Bindings[0].Wire = "wrong" },
		func(p *productIR) { p.Operations[0].ReachableModels = nil },
		func(p *productIR) { p.Models[0].Fields[0].Type.Kind = "stream" },
		func(p *productIR) { p.Models[0].Fields = append(p.Models[0].Fields, p.Models[0].Fields[0]) },
		func(p *productIR) {
			p.Models[0].Fields[0].DSLName = "metadata"
			p.Models[0].ID = "AssumeRoleResponseBody"
		},
	} {
		p := readProductIR(t, "sts")
		mutate(&p)
		if _, err := renderProduct(p); err == nil {
			t.Fatal("invalid IR rendered")
		}
	}
}

func TestFullProductEmissionCompilesInIsolatedModule(t *testing.T) {
	if testing.Short() {
		t.Skip("compiler integration")
	}
	root := t.TempDir()
	repository, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	for _, relative := range []string{".", "credentials", "endpoint", "middleware", "retry", "pagination", "waiter", "sdktest", "internal/signing", "internal/rpcmodel"} {
		entries, err := os.ReadDir(filepath.Join(repository, relative))
		if err != nil {
			t.Fatal(err)
		}
		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
				continue
			}
			data, err := os.ReadFile(filepath.Join(repository, relative, entry.Name()))
			if err != nil {
				t.Fatal(err)
			}
			writeTestFile(t, root, filepath.Join(relative, entry.Name()), data)
		}
	}
	writeTestFile(t, root, "go.mod", []byte("module "+module+"\n\ngo 1.27.0\n"))
	for _, pkg := range []string{"ecs", "sts", "vpc"} {
		files, err := renderProduct(readPolicyProduct(t, pkg))
		if err != nil {
			t.Fatal(err)
		}
		for name, data := range files {
			writeTestFile(t, root, name, data)
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 180*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "go", "test", "-p", "1", "./service/...")
	cmd.Dir = root
	cmd.Env = append(os.Environ(), "GOPROXY=off", "GOSUMDB=off", "GOWORK=off")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("isolated products do not compile/run: %v\n%s", err, output)
	}
}
