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
	if p.Overlay.Paginator != nil {
		imports[module+"/pagination"] = ""
	}
	if p.Overlay.Waiter != nil {
		imports[module+"/waiter"] = ""
		imports["time"] = ""
	}
	var b bytes.Buffer
	preamble(&b, p.Manifest.Package, imports)
	functions := template.FuncMap{
		"q": quote,
		"copies": func(operation, dst, src string) string {
			var code bytes.Buffer
			for _, op := range p.Operations {
				if op.Name == operation {
					copySlices(&code, op.Inputs, dst, src)
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
	if err := t.Execute(&b, p.Overlay); err != nil {
		return nil, err
	}
	code, err := format.Source(b.Bytes())
	if err != nil {
		return nil, fmt.Errorf("format adapters: %w", err)
	}
	return code, nil
}
