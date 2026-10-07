package codegen

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// exampleLiteral validates the supported declarative values and emits Go 1.27
// expressions. Presence pointers use the builtin new(expr), without SDK helpers.
func exampleLiteral(typ string, value any, pkg string, models map[string]ModelSpec) (string, error) {
	if strings.HasPrefix(typ, "*") {
		if value == nil {
			return "nil", nil
		}
		base := strings.TrimPrefix(typ, "*")
		literal, err := exampleLiteral(base, value, pkg, models)
		if err != nil {
			return "", err
		}
		if base == "int64" {
			literal = "int64(" + literal + ")"
		}
		return "new(" + literal + ")", nil
	}
	switch typ {
	case "string":
		v, ok := value.(string)
		if !ok {
			return "", errors.New("codegen: example string required")
		}
		return quote(v), nil
	case "bool":
		v, ok := value.(bool)
		if !ok {
			return "", errors.New("codegen: example bool required")
		}
		return strconv.FormatBool(v), nil
	case "int", "int64":
		v, ok := value.(float64)
		if !ok || v != float64(int64(v)) {
			return "", errors.New("codegen: example integer required")
		}
		return strconv.FormatInt(int64(v), 10), nil
	}
	if strings.HasPrefix(typ, "[]") {
		entries, ok := value.([]any)
		if !ok {
			return "", errors.New("codegen: example array required")
		}
		itemType := strings.TrimPrefix(typ, "[]")
		goType := itemType
		if _, ok := models[itemType]; ok {
			goType = pkg + "." + goType
		}
		var b strings.Builder
		b.WriteString("[]" + goType + "{")
		for _, entry := range entries {
			literal, err := exampleLiteral(itemType, entry, pkg, models)
			if err != nil {
				return "", err
			}
			b.WriteString(literal + ",")
		}
		b.WriteString("}")
		return b.String(), nil
	}
	model, ok := models[typ]
	if !ok || model.Location != "input" {
		return "", errors.New("codegen: unsupported example input type")
	}
	object, ok := value.(map[string]any)
	if !ok {
		return "", errors.New("codegen: example model object required")
	}
	fields := map[string]FieldSpec{}
	for _, f := range model.Fields {
		fields[f.Name] = f
	}
	var b strings.Builder
	fmt.Fprintf(&b, "%s.%s{", pkg, typ)
	for _, name := range sortedKeys(object) {
		field, ok := fields[name]
		if !ok {
			return "", errors.New("codegen: unknown example model field")
		}
		literal, err := exampleLiteral(field.Type, object[name], pkg, models)
		if err != nil {
			return "", err
		}
		fmt.Fprintf(&b, "%s:%s,", name, literal)
	}
	b.WriteString("}")
	return b.String(), nil
}
