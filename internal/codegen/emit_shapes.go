package codegen

import (
	"bytes"
	"fmt"
	"strings"
)

func emitPresenceDoc(b *bytes.Buffer, f Field) {
	if strings.HasPrefix(f.Type, "*") {
		b.WriteString("// Nil omits the field; a non-nil pointer preserves an explicit zero value.\n")
	}
}
func hasNested(fields []Field) bool {
	for _, f := range fields {
		if strings.Contains(f.Wire, ".") {
			return true
		}
	}
	return false
}

func emitModelJSON(b *bytes.Buffer, m Model) {
	root := wireTree(m.Fields)
	fmt.Fprintf(b, "// UnmarshalJSON decodes selected wire fields with JSON v2; errors leave the receiver unchanged.\nfunc(out *%s)UnmarshalJSON(data []byte)error{var wire ", m.Name)
	emitWire(b, root)
	b.WriteString("\nif err:=json.Unmarshal(data,&wire);err!=nil{return err}\n")
	fmt.Fprintf(b, "*out=%s{", m.Name)
	for _, f := range m.Fields {
		fmt.Fprintf(b, "%s:%s,", f.Name, wireAccess(root, f.Wire))
	}
	b.WriteString("};return nil}\n")
	fmt.Fprintf(b, "// MarshalJSON restores reviewed wire containers with JSON v2.\nfunc(in %s)MarshalJSON()([]byte,error){var wire ", m.Name)
	emitWire(b, root)
	b.WriteString("\n")
	for _, f := range m.Fields {
		fmt.Fprintf(b, "%s=in.%s\n", wireAccess(root, f.Wire), f.Name)
	}
	b.WriteString("return json.Marshal(wire)}\n")
}

func emitQueryField(b *bytes.Buffer, f Field, value, key string) {
	typ := strings.TrimPrefix(f.Type, "*")
	pointer := strings.HasPrefix(f.Type, "*")
	if f.Required {
		condition := value + "==0"
		if pointer {
			condition = value + "==nil"
		} else if typ == "string" {
			condition = value + `==""`
		}
		fmt.Fprintf(b, "if %s{return fail(errors.New(%q))}\n", condition, "required parameter "+f.Wire)
	}
	guard := value + "!=0"
	expression := value
	if pointer {
		guard = value + "!=nil"
		expression = "*" + value
	}
	if f.Minimum != nil {
		fmt.Fprintf(b, "if %s&&%s<%d{return fail(errors.New(%q))}\n", guard, expression, *f.Minimum, "parameter below minimum: "+f.Wire)
	}
	if f.Maximum != nil {
		fmt.Fprintf(b, "if %s&&%s>%d{return fail(errors.New(%q))}\n", guard, expression, *f.Maximum, "parameter above maximum: "+f.Wire)
	}
	if typ == "string" && !pointer {
		guard = value + `!=""`
	}
	encoded := expression
	switch typ {
	case "int":
		encoded = "strconv.Itoa(" + expression + ")"
	case "int64":
		encoded = "strconv.FormatInt(" + expression + ",10)"
	case "bool":
		encoded = "strconv.FormatBool(" + expression + ")"
	}
	fmt.Fprintf(b, "if %s{q.Set(%s,%s)}\n", guard, key, encoded)
}
