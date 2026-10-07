package codegen

import (
	"bytes"
	_ "embed"
	"fmt"
	"go/format"
	"text/template"
)

//go:embed adapters.go.tmpl
var adapterTemplate string

func emitAdapters(p Product) ([]byte, error) {
	imports := map[string]string{"context": "", "errors": ""}
	if len(p.Overlay.paginatorSpecs()) > 0 {
		imports[module+"/pagination"] = ""
	}
	if len(p.Overlay.waiterSpecs()) > 0 {
		imports[module+"/waiter"] = ""
		imports["time"] = ""
	}
	var b bytes.Buffer
	preamble(&b, p.Manifest.Package, imports)
	functions := template.FuncMap{
		"bounds": func(operation, name, expression string) string {
			var code bytes.Buffer
			for _, op := range p.Operations {
				if op.Name != operation {
					continue
				}
				for _, f := range op.Inputs {
					if f.Name != name {
						continue
					}
					if f.Minimum != nil {
						fmt.Fprintf(&code, "if %s<%d{return nil,errors.New(\"paginator parameter below minimum\")};", expression, *f.Minimum)
					}
					if f.Maximum != nil {
						fmt.Fprintf(&code, "if %s>%d{return nil,errors.New(\"paginator parameter above maximum\")};", expression, *f.Maximum)
					}
				}
			}
			return code.String()
		},
		"q": quote,
		"copies": func(operation, dst, src string) string {
			var code bytes.Buffer
			for _, op := range p.Operations {
				if op.Name == operation {
					copySlices(&code, op.Inputs, dst, src, p.Models)
				}
			}
			return code.String()
		},
		"validate": func(operation string) string {
			for _, op := range p.Operations {
				if op.Name == operation && op.Validator != "" {
					return fmt.Sprintf("if err:=%s(in);err!=nil{return nil,err}", op.Validator)
				}
			}
			return ""
		},
	}
	t, err := template.New("adapters").Funcs(functions).Parse(adapterTemplate)
	if err != nil {
		return nil, err
	}
	policies := struct {
		Paginators []PaginatorSpec
		Waiters    []WaiterSpec
	}{p.Overlay.paginatorSpecs(), p.Overlay.waiterSpecs()}
	if err := t.Execute(&b, policies); err != nil {
		return nil, err
	}
	code, err := format.Source(b.Bytes())
	if err != nil {
		return nil, fmt.Errorf("format adapters: %w", err)
	}
	return code, nil
}
