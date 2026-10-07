package codegen

import (
	"errors"
	"fmt"
	"strings"
)

func constrainedField(spec FieldSpec, schema *Schema) (Field, error) {
	f := Field{FieldSpec: spec, Required: schema.Required, MaxItems: schema.MaxItems}
	var err error
	f.Minimum, err = number(schema.Minimum)
	if err != nil {
		return Field{}, err
	}
	f.Maximum, err = number(schema.Maximum)
	if err != nil {
		return Field{}, err
	}
	if f.Minimum != nil && f.Maximum != nil && *f.Minimum > *f.Maximum {
		return Field{}, errors.New("codegen: inconsistent numeric constraints")
	}
	typ := strings.TrimPrefix(spec.Type, "*")
	if (f.Minimum != nil || f.Maximum != nil) && typ != "int" && typ != "int64" {
		return Field{}, errors.New("codegen: numeric constraints on a non-integer input")
	}
	if f.MaxItems < 0 || (f.MaxItems != 0 && !strings.HasPrefix(spec.Type, "[]")) {
		return Field{}, errors.New("codegen: invalid array bound")
	}
	return f, nil
}

func inputModelFields(s Snapshot, root *Schema, specs []FieldSpec) ([]Field, error) {
	if err := fieldNames(specs); err != nil {
		return nil, err
	}
	selected := map[string]bool{}
	var fields []Field
	for _, spec := range specs {
		if spec.Encoding != "" || spec.Region || strings.Contains(spec.Wire, ".") || spec.Type == "bool" || strings.TrimPrefix(spec.Type, "*") == "time.Time" {
			return nil, errors.New("codegen: repeatList model only supports scalar presence fields")
		}
		schema, err := s.resolve(root.Properties[spec.Wire])
		if err != nil {
			return nil, err
		}
		if err := scalar(schema, spec.Type); err != nil {
			return nil, err
		}
		f, err := constrainedField(spec, schema)
		if err != nil {
			return nil, err
		}
		fields = append(fields, f)
		selected[spec.Wire] = true
	}
	for name, schema := range root.Properties {
		if schema != nil && schema.Required && !selected[name] {
			return nil, fmt.Errorf("codegen: newly required unselected model property %s", name)
		}
	}
	return fields, nil
}
