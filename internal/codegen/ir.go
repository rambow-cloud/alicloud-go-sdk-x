package codegen

import (
	"encoding/json/v2"
	"errors"
	"fmt"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// Doc is original, equivalent English/Chinese guidance supplied by reviewers.
type Doc struct {
	English string `json:"english"`
	Chinese string `json:"chinese"`
}

// FieldSpec selects and names one schema field. Wire is a dotted response path
// or an exact parameter/property name; Type may only use the supported profile.
type FieldSpec struct {
	Name     string `json:"name"`
	Wire     string `json:"wire"`
	Type     string `json:"type"`
	Encoding string `json:"encoding,omitempty"`
	Region   bool   `json:"region,omitempty"`
	Doc      Doc    `json:"doc"`
}

// ModelSpec projects one metadata object into a named public Go model.
type ModelSpec struct {
	Name      string      `json:"name"`
	Operation string      `json:"operation"`
	Path      string      `json:"path"`
	Fields    []FieldSpec `json:"fields"`
	Redact    bool        `json:"redact,omitempty"`
	Doc       Doc         `json:"doc"`
}

// ExampleSpec supplies a deterministic offline operation example.
type ExampleSpec struct {
	Input    map[string]any `json:"input"`
	Response string         `json:"response"`
	Print    string         `json:"print"`
	Output   string         `json:"output"`
}

// OperationSpec contains reviewed semantics absent from wire metadata.
type OperationSpec struct {
	Name          string      `json:"name"`
	Idempotent    *bool       `json:"idempotent"`
	InputRequired bool        `json:"inputRequired,omitempty"`
	Validator     string      `json:"validator,omitempty"`
	Inputs        []FieldSpec `json:"inputs"`
	Outputs       []FieldSpec `json:"outputs"`
	Doc           Doc         `json:"doc"`
	Example       ExampleSpec `json:"example"`
}

// PaginatorSpec selects the reviewed dual token/page-number profile.
type PaginatorSpec struct {
	Operation   string `json:"operation"`
	Items       string `json:"items"`
	Token       string `json:"token"`
	Limit       string `json:"limit"`
	Page        string `json:"page"`
	Size        string `json:"size"`
	Total       string `json:"total"`
	DefaultSize int    `json:"defaultSize"`
}

// WaiterSpec defines an all-requested-identities-on-one-page acceptor.
type WaiterSpec struct {
	Name      string   `json:"name"`
	Operation string   `json:"operation"`
	IDs       string   `json:"ids"`
	Items     string   `json:"items"`
	ID        string   `json:"id"`
	State     string   `json:"state"`
	Page      string   `json:"page"`
	Size      string   `json:"size"`
	MaxIDs    int      `json:"maxIDs"`
	Success   string   `json:"success"`
	Retry     []string `json:"retry"`
}

// Overlay is a strict, versioned declaration; it contains no executable code.
type Overlay struct {
	SchemaVersion int             `json:"schemaVersion"`
	Doc           Doc             `json:"doc"`
	Models        []ModelSpec     `json:"models"`
	Operations    []OperationSpec `json:"operations"`
	Paginator     *PaginatorSpec  `json:"paginator,omitempty"`
	Waiter        *WaiterSpec     `json:"waiter,omitempty"`
}

// Field is a validated field with its source constraints retained.
type Field struct {
	FieldSpec
	Required         bool
	Minimum, Maximum *int64
	MaxItems         int
}

// Model is a validated public model projection.
type Model struct {
	ModelSpec
	Fields []Field
}

// Operation is a validated RPC operation with an immutable source projection.
type Operation struct {
	OperationSpec
	Inputs, Outputs []Field
}

// Product is the normalized input to emitters. It contains no network clients.
type Product struct {
	Manifest   Manifest
	Overlay    Overlay
	Models     []Model
	Operations []Operation
}

var exportedName = regexp.MustCompile(`^[A-Z][A-Za-z0-9]*$`)
var packageName = regexp.MustCompile(`^[a-z][a-z0-9]*$`)

// Load verifies pinned sources and overlays before returning sorted IR.
func Load(dir string) (Product, error) {
	var m Manifest
	var o Overlay
	if err := readJSON(filepath.Join(dir, "manifest.json"), &m, true); err != nil {
		return Product{}, err
	}
	if err := readJSON(filepath.Join(dir, "overlay.json"), &o, true); err != nil {
		return Product{}, err
	}
	if m.SchemaVersion != 1 || o.SchemaVersion != 1 || m.Style != "RPC" || !packageName.MatchString(m.Package) || token.Lookup(m.Package).IsKeyword() || m.Service != m.Package {
		return Product{}, errors.New("codegen: unsupported manifest/overlay version, style, package or service")
	}
	if err := checkDoc(o.Doc, "package"); err != nil {
		return Product{}, err
	}
	snapshots := map[string]Snapshot{}
	for _, source := range m.Sources {
		if !exportedName.MatchString(source.Operation) || source.File != source.Operation+".json" {
			return Product{}, errors.New("codegen: invalid source file or operation")
		}
		if _, exists := snapshots[source.Operation]; exists {
			return Product{}, errors.New("codegen: duplicate source operation")
		}
		data, err := os.ReadFile(filepath.Join(dir, source.File))
		if err != nil {
			return Product{}, err
		}
		if len(data) > maxMetadataBytes {
			return Product{}, errors.New("codegen: snapshot too large")
		}
		if err := verifySource(m, source, data); err != nil {
			return Product{}, fmt.Errorf("%s: %w", source.Operation, err)
		}
		var s Snapshot
		if err := json.Unmarshal(data, &s); err != nil {
			return Product{}, err
		}
		if err := checkProtocol(s); err != nil {
			return Product{}, fmt.Errorf("%s: %w", source.Operation, err)
		}
		snapshots[source.Operation] = s
	}
	p := Product{Manifest: m, Overlay: o}
	names := map[string]bool{"Client": true, "Options": true, "New": true}
	models := map[string]ModelSpec{}
	claim := func(name string) error {
		if !exportedName.MatchString(name) || names[name] {
			return fmt.Errorf("codegen: invalid or duplicate Go declaration %q", name)
		}
		names[name] = true
		return nil
	}
	for _, model := range o.Models {
		if err := claim(model.Name); err != nil {
			return Product{}, err
		}
		if err := checkDoc(model.Doc, model.Name); err != nil {
			return Product{}, err
		}
		models[model.Name] = model
	}
	for _, model := range o.Models {
		s, ok := snapshots[model.Operation]
		if !ok {
			return Product{}, fmt.Errorf("codegen: missing model operation %s", model.Operation)
		}
		schema, err := schemaAt(s.Responses["200"].Schema, model.Path)
		if err != nil {
			return Product{}, err
		}
		if err := object(schema); err != nil {
			return Product{}, fmt.Errorf("model %s: %w", model.Name, err)
		}
		fields, err := responseFields(schema, model.Fields, models, model.Operation, model.Path)
		if err != nil {
			return Product{}, err
		}
		for _, field := range fields {
			if strings.Contains(field.Wire, ".") {
				return Product{}, errors.New("codegen: nested model projections require separate models")
			}
		}
		p.Models = append(p.Models, Model{ModelSpec: model, Fields: fields})
	}
	opNames := map[string]bool{}
	for _, spec := range o.Operations {
		if opNames[spec.Name] || !exportedName.MatchString(spec.Name) || spec.Idempotent == nil {
			return Product{}, errors.New("codegen: duplicate/invalid operation or missing explicit idempotency")
		}
		opNames[spec.Name] = true
		for _, suffix := range []string{"Input", "Output", "API"} {
			if err := claim(spec.Name + suffix); err != nil {
				return Product{}, err
			}
		}
		if err := checkDoc(spec.Doc, spec.Name); err != nil {
			return Product{}, err
		}
		if spec.Validator != "" && (!token.IsIdentifier(spec.Validator) || token.Lookup(spec.Validator).IsKeyword()) {
			return Product{}, errors.New("codegen: validator must name a local function")
		}
		s, ok := snapshots[spec.Name]
		if !ok {
			return Product{}, fmt.Errorf("codegen: missing source for %s", spec.Name)
		}
		inputs, err := inputFields(s, spec.Inputs)
		if err != nil {
			return Product{}, fmt.Errorf("%s input: %w", spec.Name, err)
		}
		outputs, err := responseFields(s.Responses["200"].Schema, spec.Outputs, models, spec.Name, "")
		if err != nil {
			return Product{}, fmt.Errorf("%s output: %w", spec.Name, err)
		}
		for _, f := range outputs {
			if f.Name == "Metadata" {
				return Product{}, errors.New("codegen: Metadata is reserved")
			}
		}
		if err := checkExample(spec, inputs, outputs, models); err != nil {
			return Product{}, err
		}
		p.Operations = append(p.Operations, Operation{OperationSpec: spec, Inputs: inputs, Outputs: outputs})
	}
	if len(opNames) != len(snapshots) || len(opNames) == 0 {
		return Product{}, errors.New("codegen: every pinned operation needs one overlay")
	}
	if err := checkPolicies(p, claim); err != nil {
		return Product{}, err
	}
	sort.Slice(p.Models, func(i, j int) bool { return p.Models[i].Name < p.Models[j].Name })
	sort.Slice(p.Operations, func(i, j int) bool { return p.Operations[i].Name < p.Operations[j].Name })
	return p, nil
}

func checkDoc(doc Doc, name string) error {
	if strings.TrimSpace(doc.English) == "" || strings.TrimSpace(doc.Chinese) == "" || strings.ContainsAny(doc.English, "\r\n") || strings.ContainsAny(doc.Chinese, "\r\n") {
		return fmt.Errorf("codegen: %s requires single-line bilingual guidance", name)
	}
	return nil
}

func object(s *Schema) error {
	if err := shape(s); err != nil {
		return err
	}
	if s.Type != "object" || len(s.Properties) == 0 {
		return errors.New("codegen: object properties required")
	}
	return nil
}
func shape(s *Schema) error {
	if s == nil || s.Ref != "" || len(s.OneOf)+len(s.AllOf)+len(s.AnyOf) > 0 || len(s.AdditionalProperties) > 0 {
		return errors.New("codegen: missing or unsupported schema reference/composition/map")
	}
	return nil
}

func schemaAt(s *Schema, path string) (*Schema, error) {
	if path == "" {
		return s, nil
	}
	for _, part := range strings.Split(path, ".") {
		if err := object(s); err != nil {
			return nil, err
		}
		array := strings.HasSuffix(part, "[]")
		key := strings.TrimSuffix(part, "[]")
		if key == "" {
			return nil, errors.New("codegen: empty response path")
		}
		s = s.Properties[key]
		if s == nil {
			return nil, fmt.Errorf("codegen: removed response path %q", path)
		}
		if array {
			if err := shape(s); err != nil {
				return nil, err
			}
			if s.Type != "array" {
				return nil, errors.New("codegen: response path array changed")
			}
			s = s.Items
		}
	}
	if err := shape(s); err != nil {
		return nil, err
	}
	return s, nil
}

func fieldNames(fields []FieldSpec) error {
	names := map[string]bool{}
	wires := map[string]bool{}
	for _, f := range fields {
		if !exportedName.MatchString(f.Name) || names[f.Name] || f.Wire == "" || wires[f.Wire] {
			return errors.New("codegen: invalid/duplicate field name or wire path")
		}
		names[f.Name] = true
		wires[f.Wire] = true
		if err := checkDoc(f.Doc, f.Name); err != nil {
			return err
		}
	}
	return nil
}

func scalar(s *Schema, typ string) error {
	if err := shape(s); err != nil {
		return err
	}
	ok := typ == "string" && s.Type == "string" || typ == "time.Time" && s.Type == "string" || typ == "int" && s.Type == "integer" && (s.Format == "" || s.Format == "int32") || typ == "int64" && s.Type == "integer" && (s.Format == "int64" || s.Format == "")
	if !ok {
		return fmt.Errorf("codegen: wire type %s/%s cannot become %s", s.Type, s.Format, typ)
	}
	return nil
}

func inputFields(s Snapshot, specs []FieldSpec) ([]Field, error) {
	if err := fieldNames(specs); err != nil {
		return nil, err
	}
	parameters := map[string]Parameter{}
	for _, param := range s.Parameters {
		if _, exists := parameters[param.Name]; exists {
			return nil, errors.New("codegen: duplicate parameter")
		}
		parameters[param.Name] = param
	}
	selected := map[string]bool{}
	regionCount := 0
	var result []Field
	for _, spec := range specs {
		param, ok := parameters[spec.Wire]
		if !ok {
			return nil, fmt.Errorf("codegen: removed parameter %s", spec.Wire)
		}
		if param.In != "query" {
			return nil, errors.New("codegen: only query parameters supported")
		}
		if spec.Type == "[]string" {
			if spec.Encoding == "jsonArray" {
				if param.Style != "" || param.Schema == nil || param.Schema.Type != "string" {
					return nil, errors.New("codegen: JSON string array wire changed")
				}
				if err := shape(param.Schema); err != nil {
					return nil, err
				}
			} else if spec.Encoding == "repeatList" {
				if param.Style != "repeatList" || param.Schema == nil || param.Schema.Type != "array" {
					return nil, errors.New("codegen: repeatList wire changed")
				}
				if err := shape(param.Schema); err != nil {
					return nil, err
				}
				if err := scalar(param.Schema.Items, "string"); err != nil {
					return nil, err
				}
			} else {
				return nil, errors.New("codegen: array needs explicit supported encoding")
			}
		} else {
			if spec.Encoding != "" || param.Style != "" || spec.Type == "time.Time" {
				return nil, errors.New("codegen: unsupported scalar input encoding")
			}
			if err := scalar(param.Schema, spec.Type); err != nil {
				return nil, err
			}
		}
		if spec.Region {
			regionCount++
			if spec.Type != "string" || spec.Wire != "RegionId" {
				return nil, errors.New("codegen: region binding must be the RegionId string")
			}
		}
		field := Field{FieldSpec: spec, Required: param.Schema.Required, MaxItems: param.Schema.MaxItems}
		var err error
		field.Minimum, err = number(param.Schema.Minimum)
		if err != nil {
			return nil, err
		}
		field.Maximum, err = number(param.Schema.Maximum)
		if err != nil {
			return nil, err
		}
		if field.Minimum != nil && field.Maximum != nil && *field.Minimum > *field.Maximum {
			return nil, errors.New("codegen: inconsistent numeric constraints")
		}
		selected[spec.Wire] = true
		result = append(result, field)
	}
	for _, param := range s.Parameters {
		if param.Schema == nil {
			return nil, errors.New("codegen: missing parameter schema")
		}
		if param.Schema.Required && !selected[param.Name] {
			return nil, fmt.Errorf("codegen: newly required unselected parameter %s", param.Name)
		}
	}
	if regionCount > 1 {
		return nil, errors.New("codegen: multiple region bindings")
	}
	return result, nil
}

func number(raw []byte) (*int64, error) {
	if len(raw) == 0 {
		return nil, nil
	}
	text := strings.Trim(string(raw), `"`)
	value, err := strconv.ParseInt(text, 10, 64)
	if err != nil {
		return nil, errors.New("codegen: non-integer numeric constraint")
	}
	return &value, nil
}

func responseFields(root *Schema, specs []FieldSpec, models map[string]ModelSpec, operation, prefix string) ([]Field, error) {
	if err := object(root); err != nil {
		return nil, err
	}
	if err := fieldNames(specs); err != nil {
		return nil, err
	}
	var result []Field
	for _, spec := range specs {
		if spec.Encoding != "" || spec.Region {
			return nil, errors.New("codegen: input options on response field")
		}
		s, err := schemaAt(root, spec.Wire)
		if err != nil {
			return nil, err
		}
		typ := spec.Type
		path := spec.Wire
		if prefix != "" {
			path = prefix + "." + path
		}
		if strings.HasPrefix(typ, "[]") {
			if s.Type != "array" {
				return nil, errors.New("codegen: expected response array")
			}
			s = s.Items
			typ = strings.TrimPrefix(typ, "[]")
			path += "[]"
		}
		if model, ok := models[typ]; ok {
			if model.Operation != operation || model.Path != path {
				return nil, fmt.Errorf("codegen: model %s bound to incompatible schema path", typ)
			}
			if err := object(s); err != nil {
				return nil, err
			}
		} else {
			if err := scalar(s, typ); err != nil {
				return nil, err
			}
		}
		result = append(result, Field{FieldSpec: spec})
	}
	// Prefix collisions cannot be represented by a single wire struct.
	for i, a := range result {
		for j, b := range result {
			if i != j && strings.HasPrefix(b.Wire, a.Wire+".") {
				return nil, errors.New("codegen: overlapping response projections")
			}
		}
	}
	return result, nil
}

func checkExample(spec OperationSpec, inputs, outputs []Field, models map[string]ModelSpec) error {
	var body map[string]any
	if err := json.Unmarshal([]byte(spec.Example.Response), &body); err != nil || body == nil {
		return errors.New("codegen: example requires JSON object response")
	}
	if spec.Example.Output == "" || strings.ContainsAny(spec.Example.Output, "\r\n") {
		return errors.New("codegen: single-line example output required")
	}
	fields := map[string]Field{}
	for _, f := range inputs {
		fields[f.Name] = f
	}
	for name, value := range spec.Example.Input {
		f, ok := fields[name]
		if !ok {
			return errors.New("codegen: example input field missing")
		}
		switch f.Type {
		case "string":
			if _, ok := value.(string); !ok {
				return errors.New("codegen: example string required")
			}
		case "int", "int64":
			n, ok := value.(float64)
			if !ok || n != float64(int64(n)) {
				return errors.New("codegen: example integer required")
			}
		case "[]string":
			v, ok := value.([]any)
			if !ok {
				return errors.New("codegen: example string array required")
			}
			for _, entry := range v {
				if _, ok := entry.(string); !ok {
					return errors.New("codegen: example array entry required")
				}
			}
		default:
			return errors.New("codegen: unsupported example input")
		}
	}
	for _, f := range inputs {
		if f.Required && !f.Region {
			if _, ok := spec.Example.Input[f.Name]; !ok {
				return errors.New("codegen: example missing required input")
			}
		}
	}
	for i, part := range strings.Split(spec.Example.Print, ".") {
		if i == 0 {
			fields = map[string]Field{}
			for _, f := range outputs {
				fields[f.Name] = f
			}
		}
		array := strings.HasSuffix(part, "[0]")
		name := strings.TrimSuffix(part, "[0]")
		f, ok := fields[name]
		if !ok {
			return errors.New("codegen: invalid example output path")
		}
		typ := f.Type
		if array {
			if !strings.HasPrefix(typ, "[]") {
				return errors.New("codegen: example index requires array")
			}
			typ = strings.TrimPrefix(typ, "[]")
		}
		fields = map[string]Field{}
		if model, ok := models[typ]; ok {
			for _, child := range model.Fields {
				fields[child.Name] = Field{FieldSpec: child}
			}
		} else if i != len(strings.Split(spec.Example.Print, "."))-1 {
			return errors.New("codegen: example path traverses scalar")
		}
	}
	return nil
}

func checkPolicies(p Product, claim func(string) error) error {
	find := func(name string) (Operation, error) {
		for _, op := range p.Operations {
			if op.Name == name {
				return op, nil
			}
		}
		return Operation{}, errors.New("codegen: policy operation missing")
	}
	field := func(fields []Field, name, typ string) error {
		for _, f := range fields {
			if f.Name == name && f.Type == typ {
				return nil
			}
		}
		return fmt.Errorf("codegen: policy field %s must be %s", name, typ)
	}
	if spec := p.Overlay.Paginator; spec != nil {
		op, err := find(spec.Operation)
		if err != nil {
			return err
		}
		if spec.DefaultSize <= 0 {
			return errors.New("codegen: positive paginator size required")
		}
		if err := claim(spec.Operation + "Paginator"); err != nil {
			return err
		}
		if err := claim("New" + spec.Operation + "Paginator"); err != nil {
			return err
		}
		for _, name := range []string{spec.Limit, spec.Page, spec.Size} {
			if err := field(op.Inputs, name, "int"); err != nil {
				return err
			}
		}
		if err := field(op.Inputs, spec.Token, "string"); err != nil {
			return err
		}
		for _, name := range []string{spec.Page, spec.Size, spec.Total} {
			if err := field(op.Outputs, name, "int"); err != nil {
				return err
			}
		}
		if err := field(op.Outputs, spec.Token, "string"); err != nil {
			return err
		}
		found := false
		for _, f := range op.Outputs {
			if f.Name == spec.Items && strings.HasPrefix(f.Type, "[]") {
				found = true
			}
		}
		if !found {
			return errors.New("codegen: paginator items missing")
		}
	}
	if spec := p.Overlay.Waiter; spec != nil {
		op, err := find(spec.Operation)
		if err != nil {
			return err
		}
		if err := claim(spec.Name); err != nil {
			return err
		}
		if err := claim("New" + spec.Name); err != nil {
			return err
		}
		if spec.MaxIDs <= 0 || spec.Success == "" || len(spec.Retry) == 0 {
			return errors.New("codegen: explicit waiter bounds and states required")
		}
		states := map[string]bool{spec.Success: true}
		for _, state := range spec.Retry {
			if state == "" || states[state] {
				return errors.New("codegen: duplicate/empty waiter states")
			}
			states[state] = true
		}
		if err := field(op.Inputs, spec.IDs, "[]string"); err != nil {
			return err
		}
		for _, name := range []string{spec.Page, spec.Size} {
			if err := field(op.Inputs, name, "int"); err != nil {
				return err
			}
		}
		modelName := ""
		for _, f := range op.Outputs {
			if f.Name == spec.Items && strings.HasPrefix(f.Type, "[]") {
				modelName = strings.TrimPrefix(f.Type, "[]")
			}
		}
		found := false
		for _, m := range p.Models {
			if m.Name == modelName {
				found = true
				if err := field(m.Fields, spec.ID, "string"); err != nil {
					return err
				}
				if err := field(m.Fields, spec.State, "string"); err != nil {
					return err
				}
			}
		}
		if !found {
			return errors.New("codegen: waiter item model missing")
		}
	}
	return nil
}
