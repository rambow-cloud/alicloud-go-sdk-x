package codegen

import (
	"errors"
	"fmt"
	"strings"
)

// resolve follows only local reference chains; it never mutates source schemas.
// Selected nested model dependencies are checked separately, without expanding
// unrelated recursive components. Required is an allowed metadata annotation.
func (s Snapshot) resolve(schema *Schema) (*Schema, error) {
	seen := map[string]bool{}
	required := false
	for depth := 0; depth < 32; depth++ {
		if schema == nil {
			return nil, errors.New("codegen: missing schema")
		}
		required = required || schema.Required
		if schema.Ref == "" {
			copy := *schema
			copy.Required = required
			return &copy, nil
		}
		if schema.Type != "" || schema.Format != "" || schema.Properties != nil || schema.Items != nil || len(schema.OneOf)+len(schema.AllOf)+len(schema.AnyOf) > 0 || len(schema.AdditionalProperties)+len(schema.Minimum)+len(schema.Maximum) > 0 || schema.MaxItems != 0 || len(schema.Extra) > 0 {
			return nil, errors.New("codegen: structural siblings of $ref are unsupported")
		}
		const prefix = "#/components/schemas/"
		if !strings.HasPrefix(schema.Ref, prefix) {
			return nil, errors.New("codegen: only offline local component references supported")
		}
		if seen[schema.Ref] {
			return nil, errors.New("codegen: cyclic schema reference")
		}
		seen[schema.Ref] = true
		key := strings.TrimPrefix(schema.Ref, prefix)
		if key == "" || strings.Contains(key, "/") {
			return nil, errors.New("codegen: invalid component reference path")
		}
		var decoded strings.Builder
		for i := 0; i < len(key); i++ {
			if key[i] != '~' {
				decoded.WriteByte(key[i])
				continue
			}
			if i+1 >= len(key) || key[i+1] != '0' && key[i+1] != '1' {
				return nil, errors.New("codegen: invalid JSON pointer escape")
			}
			i++
			if key[i] == '0' {
				decoded.WriteByte('~')
			} else {
				decoded.WriteByte('/')
			}
		}
		schema = s.Components.Schemas[decoded.String()]
		if schema == nil {
			return nil, errors.New("codegen: dangling local schema reference")
		}
	}
	return nil, errors.New("codegen: schema reference chain exceeds 32 nodes")
}

func (s Snapshot) responseRoot() (*Schema, error) { return s.resolve(s.Responses["200"].Schema) }

func (s Snapshot) modelRoot(model ModelSpec) (*Schema, error) {
	if model.Location == "" || model.Location == "output" {
		root, err := s.responseRoot()
		if err != nil {
			return nil, err
		}
		return schemaAt(root, model.Path, s.resolve)
	}
	if model.Location != "input" || !strings.HasSuffix(model.Path, "[]") {
		return nil, errors.New("codegen: input model must bind repeatList items")
	}
	name := strings.TrimSuffix(model.Path, "[]")
	for _, param := range s.Parameters {
		if param.Name == name {
			if param.In != "query" || param.Style != "repeatList" {
				return nil, errors.New("codegen: input model must be query repeatList")
			}
			array, err := s.resolve(param.Schema)
			if err != nil {
				return nil, err
			}
			if array.Type != "array" {
				return nil, errors.New("codegen: input model array changed")
			}
			return s.resolve(array.Items)
		}
	}
	return nil, fmt.Errorf("codegen: removed input model parameter %s", name)
}

func modelDependencies(models []Model) error {
	byName := map[string]Model{}
	for _, m := range models {
		byName[m.Name] = m
	}
	visiting := map[string]bool{}
	done := map[string]bool{}
	var visit func(string) error
	visit = func(name string) error {
		if done[name] {
			return nil
		}
		if visiting[name] {
			return errors.New("codegen: cyclic named model dependency")
		}
		visiting[name] = true
		for _, f := range byName[name].Fields {
			dependency := strings.TrimPrefix(strings.TrimPrefix(f.Type, "[]"), "*")
			if _, ok := byName[dependency]; ok {
				if err := visit(dependency); err != nil {
					return err
				}
			}
		}
		visiting[name] = false
		done[name] = true
		return nil
	}
	for _, m := range models {
		if err := visit(m.Name); err != nil {
			return err
		}
	}
	return nil
}
