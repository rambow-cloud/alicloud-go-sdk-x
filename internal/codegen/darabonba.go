package codegen

import (
	"encoding/json/v2"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

// DSLSource pins an official parser projection and its complete source manifest.
type DSLSource struct {
	File                 string `json:"file"`
	SHA256               string `json:"sha256"`
	SourceManifestSHA256 string `json:"sourceManifestSHA256"`
	DecisionsSHA256      string `json:"decisionsSHA256"`
}

// WireProtocol contains reviewed constants lowered from OpenApi.Params.
type WireProtocol struct {
	Action          string `json:"action"`
	Version         string `json:"version"`
	Protocol        string `json:"protocol"`
	Path            string `json:"pathname"`
	Method          string `json:"method"`
	AuthType        string `json:"authType"`
	Style           string `json:"style"`
	RequestBodyType string `json:"reqBodyType"`
	BodyType        string `json:"bodyType"`
}

// DSLShape is a wire model; Required records DSL optionality, not API requiredness.
type DSLShape struct {
	Type       string               `json:"type"`
	Required   bool                 `json:"required,omitempty"`
	Properties map[string]*DSLShape `json:"properties,omitempty"`
	Items      *DSLShape            `json:"items,omitempty"`
}

// DSLInput records a recognized guarded query binding.
type DSLInput struct {
	Wire     string    `json:"wire"`
	Location string    `json:"location"`
	Guard    string    `json:"guard"`
	Schema   *DSLShape `json:"schema"`
}

// DSLOperation records a selected operation's lowered protocol and wire shapes.
type DSLOperation struct {
	Name     string       `json:"name"`
	Protocol WireProtocol `json:"protocol"`
	Inputs   []DSLInput   `json:"inputs"`
	Response *DSLShape    `json:"response"`
}

// DSLProduct is the deterministic build-time official parser projection.
type DSLProduct struct {
	SchemaVersion        int            `json:"schemaVersion"`
	ParserVersion        string         `json:"parserVersion"`
	SourceManifestSHA256 string         `json:"sourceManifestSHA256"`
	Revision             string         `json:"revision"`
	Product              string         `json:"product"`
	Version              string         `json:"version"`
	Operations           []DSLOperation `json:"operations"`
}

type dslManifest struct {
	SchemaVersion int    `json:"schemaVersion"`
	Repository    string `json:"repository"`
	Revision      string `json:"revision"`
	License       string `json:"license"`
	ParserVersion string `json:"parserVersion"`
	Files         []struct {
		File   string `json:"file"`
		SHA256 string `json:"sha256"`
	} `json:"files"`
}

func verifyDSLSource(root string, pins map[string]bool) error {
	if len(pins) == 0 {
		return nil
	}
	sourceRoot := filepath.Join(root, "sources", "darabonba")
	if err := safeOutputPath(sourceRoot, "manifest.json"); err != nil {
		return err
	}
	data, err := os.ReadFile(filepath.Join(sourceRoot, "manifest.json"))
	if err != nil {
		return fmt.Errorf("darabonba: source manifest: %w", err)
	}
	if len(pins) != 1 || !pins[digest(data)] {
		return errors.New("darabonba: source manifest checksum mismatch")
	}
	var manifest dslManifest
	if err = json.Unmarshal(data, &manifest); err != nil {
		return err
	}
	if manifest.SchemaVersion != 1 || manifest.Repository != "https://github.com/aliyun/alibabacloud-sdk" || manifest.License != "Apache-2.0" || manifest.ParserVersion != "2.2.1" || len(manifest.Files) == 0 {
		return errors.New("darabonba: unsupported source manifest")
	}
	seen := map[string]bool{}
	for _, file := range manifest.Files {
		if seen[file.File] || file.File == "manifest.json" {
			return errors.New("darabonba: duplicate source file")
		}
		seen[file.File] = true
		if err = safeOutputPath(sourceRoot, file.File); err != nil {
			return err
		}
		content, err := os.ReadFile(filepath.Join(sourceRoot, filepath.FromSlash(file.File)))
		if err != nil {
			return err
		}
		if digest(content) != file.SHA256 {
			return fmt.Errorf("darabonba: source checksum mismatch: %s", file.File)
		}
	}
	for _, directory := range []string{"products", "modules", "licenses"} {
		if err := filepath.WalkDir(filepath.Join(sourceRoot, directory), func(path string, entry fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if entry.Type()&fs.ModeSymlink != 0 {
				return errors.New("darabonba: symlink source")
			}
			if entry.IsDir() {
				return nil
			}
			relative, err := filepath.Rel(sourceRoot, path)
			if err != nil {
				return err
			}
			if !seen[filepath.ToSlash(relative)] {
				return errors.New("darabonba: unlisted source file")
			}
			return nil
		}); err != nil {
			return err
		}
	}
	return nil
}

type dslDecisions struct {
	SchemaVersion int    `json:"schemaVersion"`
	Revision      string `json:"revision"`
	Operations    map[string]struct {
		DSLOnlyInputs  []string `json:"dslOnlyInputs"`
		RequiredInputs []string `json:"requiredInputs"`
	} `json:"operations"`
}

// LoadProduct joins official DSL facts, public metadata and reviewed SDK policy.
// Legacy metadata-only synthetic products remain supported for backend tests.
func LoadProduct(dir string) (Product, error) {
	p, err := Load(dir)
	if err != nil || p.Manifest.DSL == nil {
		return p, err
	}
	source := p.Manifest.DSL
	if source.File != "dsl.json" {
		return Product{}, errors.New("darabonba: invalid projection path")
	}
	data, err := os.ReadFile(filepath.Join(dir, source.File))
	if err != nil {
		return Product{}, err
	}
	if digest(data) != source.SHA256 {
		return Product{}, errors.New("darabonba: projection checksum mismatch")
	}
	var d DSLProduct
	if err = readJSON(filepath.Join(dir, source.File), &d, true); err != nil {
		return Product{}, err
	}
	if d.SchemaVersion != 1 || d.ParserVersion != "2.2.1" || d.SourceManifestSHA256 != source.SourceManifestSHA256 || d.Product != p.Manifest.Product || d.Version != p.Manifest.Version || len(d.Revision) != 40 || len(d.Operations) != len(p.Operations) {
		return Product{}, errors.New("darabonba: projection provenance or coverage mismatch")
	}
	repository := filepath.Dir(filepath.Dir(dir))
	manifestData, err := os.ReadFile(filepath.Join(repository, "sources", "darabonba", "manifest.json"))
	if err != nil {
		return Product{}, err
	}
	var manifest dslManifest
	if err := json.Unmarshal(manifestData, &manifest); err != nil {
		return Product{}, err
	}
	if digest(manifestData) != d.SourceManifestSHA256 || manifest.Revision != d.Revision {
		return Product{}, errors.New("darabonba: source revision mismatch")
	}
	decisionPath := filepath.Join(repository, "metadata", "darabonba-decisions.json")
	decisionData, err := os.ReadFile(decisionPath)
	if err != nil {
		return Product{}, err
	}
	if digest(decisionData) != source.DecisionsSHA256 {
		return Product{}, errors.New("darabonba: decision checksum mismatch")
	}
	var decisions dslDecisions
	if err := readJSON(decisionPath, &decisions, true); err != nil {
		return Product{}, err
	}
	if decisions.SchemaVersion != 1 || decisions.Revision != d.Revision {
		return Product{}, errors.New("darabonba: decision revision mismatch")
	}
	operations := map[string]DSLOperation{}
	for _, op := range d.Operations {
		w := op.Protocol
		if _, exists := operations[op.Name]; exists || w.Action != op.Name || w.Version != d.Version || w.Protocol != "HTTPS" || w.Path != "/" || w.Method != "POST" || w.AuthType != "AK" || w.Style != "RPC" || w.RequestBodyType != "formData" || w.BodyType != "json" {
			return Product{}, errors.New("darabonba: unsupported operation protocol")
		}
		operations[op.Name] = op
	}
	for i := range p.Operations {
		op := &p.Operations[i]
		dsl, ok := operations[op.Name]
		if !ok {
			return Product{}, fmt.Errorf("darabonba: missing operation %s", op.Name)
		}
		var snapshot Snapshot
		if err = readJSON(filepath.Join(dir, op.Name+".json"), &snapshot, false); err != nil {
			return Product{}, err
		}
		inputs := map[string]*DSLShape{}
		for _, in := range dsl.Inputs {
			if inputs[in.Wire] != nil || in.Location != "query" || in.Guard != "isUnset" || in.Schema == nil {
				return Product{}, errors.New("darabonba: unsupported/duplicate input binding")
			}
			inputs[in.Wire] = in.Schema
		}
		var extra, required []string
		parameters := map[string]*Schema{}
		for _, param := range snapshot.Parameters {
			schema, err := snapshot.resolve(param.Schema)
			if err != nil {
				return Product{}, err
			}
			parameters[param.Name] = schema
		}
		for name, input := range inputs {
			if schema := parameters[name]; schema == nil {
				extra = append(extra, name)
			} else if schema.Required != input.Required {
				required = append(required, name)
			}
		}
		slices.Sort(extra)
		slices.Sort(required)
		decision, reviewed := decisions.Operations[p.Manifest.Package+"/"+op.Name]
		if !reviewed || !slices.Equal(extra, decision.DSLOnlyInputs) || !slices.Equal(required, decision.RequiredInputs) {
			return Product{}, fmt.Errorf("darabonba: unreviewed metadata/DSL difference: %s", op.Name)
		}
		for _, f := range op.Inputs {
			var schema *Schema
			for _, param := range snapshot.Parameters {
				if param.Name == f.Wire {
					schema, err = snapshot.resolve(param.Schema)
					break
				}
			}
			if err != nil {
				return Product{}, err
			}
			if err = compatibleDSL(schema, inputs[f.Wire]); err != nil {
				return Product{}, fmt.Errorf("darabonba: %s input %s: %w", op.Name, f.Wire, err)
			}
		}
		response, err := snapshot.responseRoot()
		if err != nil {
			return Product{}, err
		}
		for _, f := range op.Outputs {
			if err = checkDSLField(snapshot, response, dsl.Response, f.Wire); err != nil {
				return Product{}, fmt.Errorf("darabonba: %s output %s: %w", op.Name, f.Wire, err)
			}
		}
		for _, m := range p.Models {
			if m.Operation != op.Name {
				continue
			}
			wireRoot := dsl.Response
			if m.Location == "input" {
				wireRoot = &DSLShape{Type: "object", Properties: inputs}
			}
			wireRoot, err = dslPath(wireRoot, m.Path)
			if err != nil {
				return Product{}, err
			}
			metadataRoot, err := snapshot.modelRoot(m.ModelSpec)
			if err != nil {
				return Product{}, err
			}
			for _, f := range m.Fields {
				if err = checkDSLField(snapshot, metadataRoot, wireRoot, f.Wire); err != nil {
					return Product{}, fmt.Errorf("darabonba: model %s.%s: %w", m.Name, f.Name, err)
				}
			}
		}
		op.Protocol = dsl.Protocol
	}
	p.DSL = &d
	return p, nil
}

func compatibleDSL(metadata *Schema, wire *DSLShape) error {
	if metadata == nil || wire == nil {
		return errors.New("selected field is absent from one source")
	}
	if metadata.Type != wire.Type {
		return fmt.Errorf("wire type differs: metadata=%s DSL=%s", metadata.Type, wire.Type)
	}
	if wire.Type == "array" {
		return compatibleDSL(metadata.Items, wire.Items)
	}
	if wire.Type == "unsupported" {
		return errors.New("unsupported selected DSL shape")
	}
	return nil
}

func dslPath(root *DSLShape, location string) (*DSLShape, error) {
	if location == "" {
		return root, nil
	}
	for _, part := range strings.Split(location, ".") {
		array := strings.HasSuffix(part, "[]")
		part = strings.TrimSuffix(part, "[]")
		if root == nil || root.Type != "object" || root.Properties[part] == nil {
			return nil, fmt.Errorf("darabonba: response/model path absent: %s", location)
		}
		root = root.Properties[part]
		if array {
			if root.Type != "array" || root.Items == nil {
				return nil, errors.New("darabonba: expected array model")
			}
			root = root.Items
		}
	}
	return root, nil
}

func checkDSLField(snapshot Snapshot, metadata *Schema, wire *DSLShape, location string) error {
	s, err := schemaAt(metadata, location, snapshot.resolve)
	if err != nil {
		return err
	}
	d, err := dslPath(wire, location)
	if err != nil {
		return err
	}
	return compatibleDSL(s, d)
}
