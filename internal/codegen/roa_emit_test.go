package codegen

import (
	"bytes"
	"context"
	"encoding/json/v2"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestROAEmitterReusesFacadesForAnotherProduct(t *testing.T) {
	p := readProductIR(t, "fc")
	data, err := json.Marshal(p)
	if err != nil {
		t.Fatal(err)
	}
	data = bytes.ReplaceAll(data, []byte("products/fc/main.tea"), []byte("products/computecheck/main.tea"))
	if err := json.Unmarshal(data, &p); err != nil {
		t.Fatal(err)
	}
	p.Product = "computecheck"
	files, err := renderProduct(p)
	if err != nil {
		t.Fatal(err)
	}
	methods := files["service/computecheck/operations.gen.go"]
	types := files["service/computecheck/types.gen.go"]
	for _, comment := range []string{"// Body is the native JSON request body, encoded without an extra wrapper.", "// Headers supplies custom service HTTP headers; nil supplies none.", "// FunctionName is the required nonempty path parameter functionName."} {
		if !bytes.Contains(types, []byte(comment)) {
			t.Fatal("ROA role contract lost", comment)
		}
	}
	if !bytes.Contains(methods, []byte("invokeROA[")) || !bytes.Contains(methods, []byte("ResponseBody: alicloud.ResponseBodyNone")) {
		t.Fatal("ROA runtime composition lost")
	}
	for _, language := range []string{".md", ".zh-CN.md"} {
		guide := files["docs/products/computecheck"+language]
		if bytes.Contains(guide, []byte("fc-roa-product")) {
			t.Fatal("FC-only route leaked into another product")
		}
	}
}

func TestROAUnsupportedSelectionFailsBeforeWrites(t *testing.T) {
	root := fullProductFixture(t)
	if err := GenerateProducts(context.Background(), root, false, []string{"fc/UnknownBinary"}); err == nil {
		t.Fatal("unknown operation emitted")
	}
	if _, err := os.Stat(filepath.Join(root, "service")); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("unsupported selection wrote files")
	}
}

func TestROAEmitterRejectsUnboundAndUnknownBehavior(t *testing.T) {
	for _, name := range []string{"path", "stray-brace", "guard", "location", "encoding", "response"} {
		t.Run(name, func(t *testing.T) {
			p := readProductIR(t, "fc")
			for i := range p.Operations {
				op := &p.Operations[i]
				if op.Name != "CreateAlias" {
					continue
				}
				switch name {
				case "path":
					op.Protocol.Path = "/unbound/{unknown}"
				case "stray-brace":
					op.Protocol.Path = "/}/{functionName}"
				case "guard":
					op.Bindings[0].Guard = "inferred"
				case "location":
					op.Bindings[0].Location = "unknown"
				case "encoding":
					op.Bindings[0].Encoding = "binary"
				case "response":
					op.Protocol.BodyType = "xml"
				}
			}
			if _, err := renderProduct(p); err == nil {
				t.Fatal("unknown ROA behavior emitted")
			}
		})
	}
}
