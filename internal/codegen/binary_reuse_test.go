package codegen

import (
	"context"
	"encoding/json/v2"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

func readBinaryReuseFixture(t *testing.T) productIR {
	t.Helper()
	repository, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "node", filepath.Join(repository, "tools/darabonba/binary-reuse-fixture.cjs"), repository)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("binary parser reuse fixture failed: %v\n%s", err, output)
	}
	var product productIR
	if err := json.Unmarshal(output, &product); err != nil {
		t.Fatal(err)
	}
	return product
}

const binaryReuseRuntimeTest = `package binaryfixture_test
import("context";"crypto/sha256";"encoding/hex";"errors";"io";"net/http";"testing";alicloud "github.com/rambow-cloud/alicloud-go-sdk-x";"github.com/rambow-cloud/alicloud-go-sdk-x/credentials";"github.com/rambow-cloud/alicloud-go-sdk-x/sdktest";"github.com/rambow-cloud/alicloud-go-sdk-x/service/binaryfixture")
func TestRenamedBinaryProgramUsesSharedRuntime(t *testing.T){
 provider,_:=credentials.NewStaticProvider(credentials.Credentials{AccessKeyID:"fixture",AccessKeySecret:"fixture"})
 tail:="Tail";transport:=sdktest.NewTransport(sdktest.Step{Body:"raw-response",Check:func(r *http.Request)error{
 data,err:=io.ReadAll(r.Body);if err!=nil{return err};hash:=sha256.Sum256(data)
 if string(data)!="raw-request"||r.Header.Get("X-Acs-Action")!="ExecutePayload"||r.Header.Get("Content-Type")!="application/octet-stream"||r.Header.Get("X-Fc-Log-Type")!="Tail"||r.Header.Get("X-Custom")!="copied"||r.Header.Get("X-Acs-Content-Sha256")!=hex.EncodeToString(hash[:]){return errors.New("renamed binary wire mismatch")};return nil}})
 client,err:=binaryfixture.NewFromConfig(alicloud.Config{CredentialsProvider:provider,BaseEndpoint:"https://example.invalid",HTTPClient:&http.Client{Transport:transport}});if err!=nil{t.Fatal(err)}
 var api binaryfixture.ExecutePayloadAPI=client
 out,err:=api.ExecutePayload(context.Background(),&binaryfixture.ExecutePayloadInput{FunctionName:"example",Payload:[]byte("raw-request"),Headers:&binaryfixture.ExecutePayloadHeaders{SharedHeaders:map[string]string{"X-Custom":"copied"},XFcLogType:&tail}});if err!=nil{t.Fatal(err)};defer out.Result.Close()
 data,err:=io.ReadAll(out.Result);if err!=nil||string(data)!="raw-response"||out.Metadata.HTTPStatusCode!=200{t.Fatal(data,err)}
}
`
