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

func readReuseFixture(t *testing.T) productIR {
	t.Helper()
	repository, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "node", filepath.Join(repository, "tools", "darabonba", "reuse-fixture.cjs"), repository)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("official parser reuse fixture failed: %v\n%s", err, output)
	}
	var p productIR
	if err := json.Unmarshal(output, &p); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestRenamedProductUsesSharedEmitterAndGuide(t *testing.T) {
	p := readReuseFixture(t)
	files, err := renderProduct(p)
	if err != nil {
		t.Fatal(err)
	}
	methods := string(files["service/authfixture/operations.gen.go"])
	if strings.Count(methods, "Authentication: alicloud.AuthenticationAnonymousRPC") != 2 {
		t.Fatal("anonymous protocol not reused")
	}
	if !strings.Contains(string(files["service/authfixture/types.gen.go"]), "type InspectIdentityInput struct{}") {
		t.Fatal("requestless convention not reused")
	}
	for _, language := range []string{".md", ".zh-CN.md"} {
		guide := string(files["docs/products/authfixture"+language])
		for _, name := range []string{"ExchangeIdentity", "ExchangeAssertion"} {
			if !strings.Contains(guide, "`"+name+"`") {
				t.Fatal("guide omits native anonymous action", name)
			}
		}
		if strings.Contains(guide, "sts-anonymous-rpc") || strings.Contains(guide, "v0.1.0") || strings.Contains(guide, "AssumeRoleWithOIDC/SAML") {
			t.Fatal("STS-only acceptance leaked into another product")
		}
	}
}

func TestCompleteProductsDeterministicModelsMethodsAndExamples(t *testing.T) {
	for pkg, want := range map[string]int{"ecs": 380, "sts": 4, "vpc": 403, "fc": 72} {
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
			// Count declaration lines rather than words inside licensed prose.
			emptyInputs := 0
			for _, op := range r.operations {
				if op.Roots.Request.Kind == "empty" {
					emptyInputs++
				}
			}
			if strings.Count(methods, "\nfunc (c *Client)") != want || strings.Count(examples, "\nfunc ExampleClient_") != want || strings.Count(types, "\ntype ") != len(r.models)+emptyInputs {
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

func TestShrinkBindingsRejectUnknownEncodingAndDynamicQueryValues(t *testing.T) {
	for _, encoding := range []string{"base64", ""} {
		t.Run(encoding, func(t *testing.T) {
			p := readProductIR(t, "ecs")
			for i := range p.Operations {
				if p.Operations[i].Name != "InvokeCommand" {
					continue
				}
				for j := range p.Operations[i].Bindings {
					if p.Operations[i].Bindings[j].Wire == "Parameters" {
						p.Operations[i].Bindings[j].Encoding = encoding
					}
				}
			}
			if _, err := renderProduct(p); err == nil {
				t.Fatal("unsafe encoding reached emission")
			}
		})
	}
	p := readProductIR(t, "ecs")
	files, err := renderProduct(p)
	if err != nil {
		t.Fatal(err)
	}
	types := string(files["service/ecs/types.gen.go"])
	if !strings.Contains(types, "rpc:\"json\"") || !strings.Contains(types, "// Deprecated: The upstream DSL") {
		t.Fatal("encoding or deprecation contract missing")
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
	for _, selected := range [][]string{{"ecs/Unknown"}, {"vpc/DoesNotExist"}, {"DescribeImages"}, {"ecs/DescribeImages", "vpc/DoesNotExist"}} {
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

func TestProductSchemaVersionsFailBeforeWrites(t *testing.T) {
	root := fullProductFixture(t)
	path := filepath.Join(root, "models", "manifest.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var pins productPins
	if err := json.Unmarshal(data, &pins); err != nil {
		t.Fatal(err)
	}
	writePins := func() {
		t.Helper()
		encoded, err := json.Marshal(pins)
		if err != nil {
			t.Fatal(err)
		}
		writeTestFile(t, root, "models/manifest.json", encoded)
	}
	checkRejected := func(want string) {
		t.Helper()
		if err := GenerateProducts(context.Background(), root, false, nil); err == nil || !strings.Contains(err.Error(), want) {
			t.Fatal("incompatible IR accepted", err)
		}
		if _, err := os.Stat(filepath.Join(root, "service")); !errors.Is(err, os.ErrNotExist) {
			t.Fatal("version rejection wrote output")
		}
	}
	pins.SchemaVersion = 1
	writePins()
	checkRejected("unsupported IR manifest")
	pins.SchemaVersion = 2
	writePins()
	checkRejected("unsupported IR manifest")
	pins.SchemaVersion = 3
	writePins()
	checkRejected("unsupported IR manifest")
	pins.SchemaVersion = 4
	writePins()
	checkRejected("unsupported IR manifest")
	pins.SchemaVersion = 5
	p := readProductIR(t, "ecs")
	p.SchemaVersion = 2
	encoded, err := json.Marshal(p)
	if err != nil {
		t.Fatal(err)
	}
	writeTestFile(t, root, "models/ecs/ir.json", encoded)
	for i := range pins.Files {
		if pins.Files[i].File == "models/ecs/ir.json" {
			pins.Files[i].SHA256 = digest(encoded)
		}
	}
	writePins()
	checkRejected("IR provenance mismatch")
}

func TestProductSourceIntegrityFailsBeforeWrites(t *testing.T) {
	for _, file := range []string{"products/sts/main.tea", "modules/shadow.tea", "manifest.json"} {
		t.Run(file, func(t *testing.T) {
			root := fullProductFixture(t)
			writeTestFile(t, root, "sources/darabonba/"+file, []byte("tampered"))
			if err := GenerateProducts(context.Background(), root, false, nil); err == nil {
				t.Fatal("invalid source lock accepted")
			}
			if _, err := os.Stat(filepath.Join(root, "service")); !errors.Is(err, os.ErrNotExist) {
				t.Fatal("source failure wrote output")
			}
		})
	}
}

func TestProductReconciliationAndManualOwnership(t *testing.T) {
	root := fullProductFixture(t)
	ctx := context.Background()
	if err := GenerateProducts(ctx, root, true, nil); err == nil {
		t.Fatal("missing output accepted")
	}
	if _, err := os.Stat(filepath.Join(root, "service")); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("check wrote")
	}
	if err := GenerateProducts(ctx, root, false, []string{"ecs/DescribeImages", "sts/AssumeRole"}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, "services")); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("generation recreated removed bridge")
	}
	if _, err := os.Stat(filepath.Join(root, "service", "vpc", "operations.gen.go")); err != nil {
		t.Fatal("selection silently narrowed products")
	}
	writeTestFile(t, root, "service/ecs/obsolete.gen.go", []byte(productGenerated+"package ecs\n"))
	writeTestFile(t, root, "service/obsolete/LICENSE", []byte(productGenerated+"\nold terms\n"))
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
	if err := GenerateProducts(ctx, root, true, nil); !errors.As(err, &drift) || len(drift.Files) != 3 {
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
	if _, err := os.Stat(filepath.Join(root, "service", "obsolete", "LICENSE")); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("stale license retained")
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
	for _, file := range []string{"types.gen.go", "LICENSE", "NOTICE"} {
		writeTestFile(t, root, "service/sts/"+file, []byte("unmarked manual content\n"))
		if err := GenerateProducts(context.Background(), root, false, nil); err == nil {
			t.Fatal("manual file overwritten", file)
		}
		if _, err := os.Stat(filepath.Join(root, "docs", "products")); !errors.Is(err, os.ErrNotExist) {
			t.Fatal("partial output before preflight")
		}
		if err := os.Remove(filepath.Join(root, "service", "sts", file)); err != nil {
			t.Fatal(err)
		}
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
		func(p *productIR) { p.Operations[0].Protocol.Method = "DELETE" },
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
	for _, relative := range []string{".", "credentials", "endpoint", "middleware", "retry", "pagination", "waiter", "sdktest", "internal/signing", "internal/rpcmodel", "internal/roamodel"} {
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
	for _, pkg := range []string{"ecs", "sts", "vpc", "fc"} {
		files, err := renderProduct(readPolicyProduct(t, pkg))
		if err != nil {
			t.Fatal(err)
		}
		for name, data := range files {
			writeTestFile(t, root, name, data)
		}
	}
	fixtureFiles, err := renderProduct(readReuseFixture(t))
	if err != nil {
		t.Fatal(err)
	}
	for name, data := range fixtureFiles {
		writeTestFile(t, root, name, data)
	}
	writeTestFile(t, root, "service/authfixture/reuse_test.go", []byte(reuseRuntimeTest))
	ctx, cancel := context.WithTimeout(context.Background(), 180*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "go", "test", "-p", "1", "./service/...")
	cmd.Dir = root
	cmd.Env = append(os.Environ(), "GOPROXY=off", "GOSUMDB=off", "GOWORK=off")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("isolated products do not compile/run: %v\n%s", err, output)
	}
}

const reuseRuntimeTest = `package authfixture_test

import (
 "context"
 "encoding/json/v2"
 "fmt"
 "net/http"
 "sync/atomic"
 "testing"
 alicloud "github.com/rambow-cloud/alicloud-go-sdk-x"
 "github.com/rambow-cloud/alicloud-go-sdk-x/credentials"
 "github.com/rambow-cloud/alicloud-go-sdk-x/sdktest"
 "github.com/rambow-cloud/alicloud-go-sdk-x/service/authfixture"
)

func TestRenamedActionsKeepWireAndAuthentication(t *testing.T) {
 var reads atomic.Int32
 source := credentials.ProviderFunc(func(context.Context)(credentials.Credentials,error){
  reads.Add(1)
  return credentials.Credentials{AccessKeyID:"synthetic-key",AccessKeySecret:"synthetic-secret"},nil
 })
 step := func(action, wire, value string, signed bool) sdktest.Step {
  return sdktest.Step{Body:"{}",Check:func(r *http.Request)error{
   if r.Header.Get("x-acs-action")!=action{return fmt.Errorf("wrong native action")}
   if (r.Header.Get("Authorization")!="")!=signed{return fmt.Errorf("wrong authentication")}
   if wire!="" && r.URL.Query().Get(wire)!=value{return fmt.Errorf("wire value lost")}
   if action=="InspectIdentity" && len(r.URL.Query())!=0{return fmt.Errorf("requestless query changed")}
   return nil
  }}
 }
 transport:=sdktest.NewTransport(
  step("ObtainRole","RoleArn","acs:ram::123:role/example",true),
  step("InspectIdentity","","",true),
  step("ExchangeIdentity","OIDCToken","synthetic-oidc",false),
  step("ExchangeAssertion","SAMLAssertion","synthetic-saml",false),
 )
 client,err:=authfixture.NewFromConfig(alicloud.Config{Region:"cn-hangzhou",BaseEndpoint:"https://example.invalid",CredentialsProvider:source,HTTPClient:&http.Client{Transport:transport}})
 if err!=nil{t.Fatal(err)}
 role:=new(authfixture.ObtainRoleInput)
 if err:=json.Unmarshal([]byte("{\"RoleArn\":\"acs:ram::123:role/example\",\"RoleSessionName\":\"example\"}"),role);err!=nil{t.Fatal(err)}
 if _,err:=client.ObtainRole(context.Background(),role);err!=nil{t.Fatal(err)}
 if _,err:=client.InspectIdentity(context.Background(),nil);err!=nil{t.Fatal(err)}
 oidc:=new(authfixture.ExchangeIdentityInput)
 if err:=json.Unmarshal([]byte("{\"OIDCToken\":\"synthetic-oidc\"}"),oidc);err!=nil{t.Fatal(err)}
 if _,err:=client.ExchangeIdentity(context.Background(),oidc);err!=nil{t.Fatal(err)}
 saml:=new(authfixture.ExchangeAssertionInput)
 if err:=json.Unmarshal([]byte("{\"SAMLAssertion\":\"synthetic-saml\"}"),saml);err!=nil{t.Fatal(err)}
 if _,err:=client.ExchangeAssertion(context.Background(),saml);err!=nil{t.Fatal(err)}
 if reads.Load()!=2 || transport.Calls()!=4{t.Fatal("anonymous calls retrieved signing credentials or calls were lost")}
}
`

func TestAnonymousProductRejectsAuthHandoffDrift(t *testing.T) {
	for _, mutate := range []func(*productOperation){func(o *productOperation) { o.Handoff = "callApi" }, func(o *productOperation) { o.Protocol.AuthType = "AK" }, func(o *productOperation) { o.Protocol.AuthType = "FutureAuth" }} {
		p := readProductIR(t, "sts")
		for i := range p.Operations {
			if p.Operations[i].Name == "AssumeRoleWithOIDC" {
				mutate(&p.Operations[i])
			}
		}
		if _, err := renderProduct(p); err == nil {
			t.Fatal("unreviewed authentication rendered")
		}
	}
}
