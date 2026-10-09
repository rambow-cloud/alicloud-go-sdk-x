package codegen

import (
	"errors"
	"fmt"
	"go/token"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

type capabilityPolicy struct {
	SchemaVersion        int                        `json:"schemaVersion"`
	Product              string                     `json:"product"`
	Version              string                     `json:"version"`
	SourceManifestSHA256 string                     `json:"sourceManifestSHA256"`
	FieldNames           map[string]string          `json:"fieldNames,omitempty"`
	ModelNames           map[string]string          `json:"modelNames,omitempty"`
	Operations           map[string]operationPolicy `json:"operations"`
	Endpoints            []endpointPolicy           `json:"endpoints,omitempty"`
}
type operationPolicy struct {
	Idempotent      *bool             `json:"idempotent"`
	Evidence        []string          `json:"evidence"`
	Paginator       *nativePaginator  `json:"paginator,omitempty"`
	Waiter          *nativeWaiter     `json:"waiter,omitempty"`
	ClientToken     string            `json:"clientToken,omitempty"`
	Constraints     []inputConstraint `json:"constraints,omitempty"`
	SensitiveModels []string          `json:"sensitiveModels,omitempty"`
}
type nativePaginator struct {
	Mode        string `json:"mode"`
	Items       string `json:"items"`
	InputToken  string `json:"inputToken,omitempty"`
	OutputToken string `json:"outputToken,omitempty"`
	Limit       string `json:"limit,omitempty"`
	Page        string `json:"page,omitempty"`
	Size        string `json:"size,omitempty"`
	Total       string `json:"total,omitempty"`
	DefaultSize int    `json:"defaultSize"`
	MaximumSize int    `json:"maximumSize"`
}
type nativeWaiter struct {
	Name    string   `json:"name"`
	IDs     string   `json:"ids"`
	Items   string   `json:"items"`
	ID      string   `json:"id"`
	State   string   `json:"state"`
	Page    string   `json:"page"`
	Size    string   `json:"size"`
	MaxIDs  int      `json:"maxIDs"`
	Success string   `json:"success"`
	Retry   []string `json:"retry"`
}
type inputConstraint struct {
	Path          string `json:"path"`
	Minimum       *int64 `json:"minimum,omitempty"`
	Maximum       *int64 `json:"maximum,omitempty"`
	MinimumLength *int   `json:"minimumLength,omitempty"`
	MaximumLength *int   `json:"maximumLength,omitempty"`
	ASCII         bool   `json:"ascii,omitempty"`
	MaximumItems  int    `json:"maximumItems,omitempty"`
	NonemptyItems bool   `json:"nonemptyItems,omitempty"`
}
type capabilityCoverage struct {
	Status              string `json:"status"`
	Idempotent          bool   `json:"idempotent"`
	Paginator           string `json:"paginator,omitempty"`
	Waiter              string `json:"waiter,omitempty"`
	ClientToken         string `json:"clientToken,omitempty"`
	Validator           bool   `json:"validator"`
	SensitiveFormatting bool   `json:"sensitiveFormatting"`
	NamingException     bool   `json:"namingException"`
}

func (r *productRenderer) capabilityCoverage(op productOperation) *capabilityCoverage {
	result := &capabilityCoverage{Status: "not-reviewed"}
	cfg := r.opPolicy(op.Name)
	if cfg.Idempotent != nil {
		result.Status = "reviewed"
		result.Idempotent = *cfg.Idempotent
	}
	if cfg.Paginator != nil {
		result.Paginator = cfg.Paginator.Mode
	}
	if cfg.Waiter != nil {
		result.Waiter = cfg.Waiter.Name
	}
	result.ClientToken = cfg.ClientToken
	result.Validator = hasValidator(cfg)
	result.SensitiveFormatting = len(cfg.SensitiveModels) > 0
	if r.p.Policy != nil {
		for _, id := range op.ReachableModels {
			if r.p.Policy.ModelNames[id] != "" {
				result.NamingException = true
			}
			for key := range r.p.Policy.FieldNames {
				if strings.HasPrefix(key, id+"#") {
					result.NamingException = true
				}
			}
		}
	}
	return result
}

func loadCapabilityPolicies(root string, products []productIR) error {
	indexes := map[string]int{}
	for i, p := range products {
		indexes[p.Product] = i
	}
	if err := safeOutputPath(root, "policies/.scan"); err != nil {
		return err
	}
	entries, err := os.ReadDir(filepath.Join(root, "policies"))
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	for _, entry := range entries {
		file := "policies/" + entry.Name()
		if err := safeOutputPath(root, file); err != nil {
			return err
		}
		if entry.Name() == "README.md" || entry.Name() == "README.zh-CN.md" {
			continue
		}
		index, ok := indexes[strings.TrimSuffix(entry.Name(), ".json")]
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") || !ok {
			return fmt.Errorf("policy: unknown product/artifact %s", file)
		}
		var policy capabilityPolicy
		if err := readJSON(filepath.Join(root, filepath.FromSlash(file)), &policy, true); err != nil {
			return fmt.Errorf("policy %s: %w", file, err)
		}
		p := &products[index]
		if policy.SchemaVersion != 1 || policy.Product != p.Product || policy.Version != p.Version || policy.SourceManifestSHA256 != p.Provenance.SourceManifestSHA256 {
			return errors.New("policy: schema/product/version/source mismatch")
		}
		data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(file)))
		if err != nil {
			return err
		}
		p.Policy = &policy
		p.PolicySHA256 = digest(data)
	}
	return nil
}

type capabilityPath struct {
	Access   string
	Guards   []string
	Type     productType
	Optional bool
}

func (p capabilityPath) condition() string {
	if len(p.Guards) == 0 {
		return "true"
	}
	return strings.Join(p.Guards, "&&")
}
func (p capabilityPath) value() string {
	if p.Optional {
		return "*" + p.Access
	}
	return p.Access
}
func (r *productRenderer) fieldName(id string, f productField) string {
	if r.p.Policy != nil {
		if name := r.p.Policy.FieldNames[id+"#"+f.DSLName]; name != "" {
			return name
		}
	}
	return productName(f.DSLName)
}
func (r *productRenderer) resolve(root, wire, variable string) (capabilityPath, error) {
	result := capabilityPath{Access: variable}
	parts := strings.Split(wire, ".")
	if wire == "" {
		return result, errors.New("policy: empty wire path")
	}
	for i, part := range parts {
		model, ok := r.models[root]
		if !ok {
			return result, errors.New("policy: missing path model")
		}
		var field productField
		found := false
		for _, f := range model.Fields {
			if f.WireName == part {
				field = f
				found = true
				break
			}
		}
		if !found {
			return result, fmt.Errorf("policy: unknown exact wire path %s", wire)
		}
		result.Access += "." + r.fieldName(root, field)
		result.Type = field.Type
		result.Optional = !field.Required && (field.Type.Kind == "scalar" || field.Type.Kind == "model")
		if result.Optional {
			result.Guards = append(result.Guards, result.Access+"!=nil")
		}
		if i < len(parts)-1 {
			if field.Type.Kind != "model" {
				return result, errors.New("policy: cannot traverse non-model")
			}
			root = field.Type.Ref
		}
	}
	return result, nil
}
func (r *productRenderer) scalarPath(root, wire, variable, typ string) (capabilityPath, error) {
	p, err := r.resolve(root, wire, variable)
	if err != nil {
		return p, err
	}
	if p.Type.Kind != "scalar" || p.Type.DSLType != typ {
		return p, fmt.Errorf("policy: %s must be %s", wire, typ)
	}
	return p, nil
}

// Compare roles only within a model; request/response and separate adapters may
// intentionally use identical native paths. Empty paths belong to inactive modes.
func distinctCapabilityRoles(scope string, paths ...string) error {
	seen := map[string]bool{}
	for _, path := range paths {
		if path == "" {
			continue
		}
		if seen[path] {
			return fmt.Errorf("policy: %s roles must be distinct", scope)
		}
		seen[path] = true
	}
	return nil
}

func (r *productRenderer) validateCapabilityPolicy() error {
	policy := r.p.Policy
	if policy == nil {
		return nil
	}
	for id, name := range policy.ModelNames {
		if _, ok := r.models[id]; !ok || name == "" {
			return errors.New("policy: unknown model naming exception")
		}
		for _, op := range r.operations {
			if id == op.Roots.Request.Ref || id == op.Roots.Body.Ref || id == op.Roots.Response.Ref {
				return errors.New("policy: operation root names are fixed")
			}
		}
	}
	for key, name := range policy.FieldNames {
		parts := strings.Split(key, "#")
		if len(parts) != 2 || !exportedName.MatchString(name) {
			return errors.New("policy: invalid field naming exception")
		}
		model, ok := r.models[parts[0]]
		found := false
		for _, f := range model.Fields {
			found = found || f.DSLName == parts[1]
		}
		if !ok || !found {
			return errors.New("policy: unknown field naming exception")
		}
	}
	ops := map[string]productOperation{}
	names := map[string]bool{"Client": true, "Options": true, "New": true, "NewFromConfig": true}
	for _, name := range r.names {
		names[name] = true
	}
	for _, op := range r.operations {
		ops[op.Name] = op
		names[op.Name+"API"] = true
	}
	for _, name := range sortedKeys(policy.Operations) {
		cfg := policy.Operations[name]
		op, ok := ops[name]
		if !ok {
			return fmt.Errorf("policy: unknown/unsupported action %s", name)
		}
		if cfg.Idempotent == nil || len(cfg.Evidence) == 0 {
			return errors.New("policy: explicit idempotency and evidence required")
		}
		for _, evidence := range cfg.Evidence {
			if strings.HasPrefix(evidence, "https://") {
				u, err := url.Parse(evidence)
				if err != nil || u.Host == "" {
					return errors.New("policy: invalid evidence URL")
				}
			} else if !strings.HasPrefix(evidence, "docs/") && !strings.HasPrefix(evidence, "sources/darabonba/") {
				return errors.New("policy: invalid evidence reference")
			}
		}
		reserve := func(symbol string) error {
			if !exportedName.MatchString(symbol) || token.Lookup(symbol).IsKeyword() || names[symbol] {
				return fmt.Errorf("policy: symbol collision %s", symbol)
			}
			names[symbol] = true
			return nil
		}
		request, body := op.Roots.Request.Ref, op.Roots.Body.Ref
		if cfg.Paginator != nil {
			p := cfg.Paginator
			if p.Mode != "tokens" && p.Mode != "pages" && p.Mode != "dual" {
				return errors.New("policy: invalid paginator mode")
			}
			if p.DefaultSize < 1 || p.MaximumSize < p.DefaultSize || p.MaximumSize > 2147483647 {
				return errors.New("policy: invalid paginator size")
			}
			items, err := r.resolve(body, p.Items, "out")
			if err != nil {
				return err
			}
			if items.Type.Kind != "array" {
				return errors.New("policy: paginator collection is not an array")
			}
			if p.Mode != "pages" {
				if strings.Contains(p.InputToken, ".") || strings.Contains(p.Limit, ".") || strings.Contains(p.OutputToken, ".") {
					return errors.New("policy: token cursors must be root fields")
				}
				for _, s := range []struct{ root, path, typ string }{{request, p.InputToken, "string"}, {body, p.OutputToken, "string"}, {request, p.Limit, "int32"}} {
					field, err := r.scalarPath(s.root, s.path, "in", s.typ)
					if err != nil {
						return err
					}
					if p.Mode == "dual" && s.root == request && !field.Optional {
						return errors.New("policy: dual-mode cursors require optional fields")
					}
				}
			} else if p.InputToken != "" || p.OutputToken != "" || p.Limit != "" {
				return errors.New("policy: page mode cannot declare token fields")
			}
			if p.Mode != "tokens" {
				for _, s := range []struct{ root, path string }{{request, p.Page}, {request, p.Size}, {body, p.Page}, {body, p.Size}, {body, p.Total}} {
					if strings.Contains(s.path, ".") {
						return errors.New("policy: cursor must be a root field")
					}
					field, err := r.scalarPath(s.root, s.path, "in", "int32")
					if err != nil {
						return err
					}
					if p.Mode == "dual" && s.root == request && !field.Optional {
						return errors.New("policy: dual-mode cursors require optional fields")
					}
				}
			} else if p.Page != "" || p.Size != "" || p.Total != "" {
				return errors.New("policy: token mode cannot declare page fields")
			}
			if err := distinctCapabilityRoles("paginator request", p.Page, p.Size, p.Limit); err != nil {
				return err
			}
			if err := distinctCapabilityRoles("paginator response", p.Page, p.Size, p.Total); err != nil {
				return err
			}
			for _, symbol := range []string{name + "Paginator", name + "PaginatorOptions", "New" + name + "Paginator"} {
				if err := reserve(symbol); err != nil {
					return err
				}
			}
		}
		if cfg.Waiter != nil {
			w := cfg.Waiter
			if w.MaxIDs < 1 || w.MaxIDs > 2147483647 || w.Success == "" || len(w.Retry) == 0 {
				return errors.New("policy: invalid waiter bounds/states")
			}
			ids, err := r.resolve(request, w.IDs, "in")
			if err != nil {
				return err
			}
			if strings.Contains(w.IDs, ".") || ids.Type.Kind != "array" || ids.Type.Items == nil || ids.Type.Items.Kind != "scalar" || ids.Type.Items.DSLType != "string" {
				return errors.New("policy: waiter IDs must be a root string array")
			}
			items, err := r.resolve(body, w.Items, "out")
			if err != nil {
				return err
			}
			if items.Type.Kind != "array" || items.Type.Items == nil || items.Type.Items.Kind != "model" {
				return errors.New("policy: waiter needs model collection")
			}
			for _, path := range []string{w.ID, w.State} {
				if _, err := r.scalarPath(items.Type.Items.Ref, path, "item", "string"); err != nil {
					return err
				}
			}
			for _, path := range []string{w.Page, w.Size} {
				if strings.Contains(path, ".") {
					return errors.New("policy: waiter page fields must be root fields")
				}
				if _, err := r.scalarPath(request, path, "in", "int32"); err != nil {
					return err
				}
			}
			if err := distinctCapabilityRoles("waiter request", w.Page, w.Size); err != nil {
				return err
			}
			if err := distinctCapabilityRoles("waiter member", w.ID, w.State); err != nil {
				return err
			}
			states := map[string]bool{w.Success: true}
			for _, state := range w.Retry {
				if state == "" || states[state] {
					return errors.New("policy: ambiguous waiter states")
				}
				states[state] = true
			}
			for _, symbol := range []string{w.Name, w.Name + "Options", "New" + w.Name} {
				if err := reserve(symbol); err != nil {
					return err
				}
			}
		}
		if cfg.ClientToken != "" {
			if *cfg.Idempotent {
				return errors.New("policy: token-bearing writes remain conservative")
			}
			if strings.Contains(cfg.ClientToken, ".") {
				return errors.New("policy: client token must be root field")
			}
			p, err := r.scalarPath(request, cfg.ClientToken, "in", "string")
			if err != nil {
				return err
			}
			if !p.Optional {
				return errors.New("policy: client token requires optional pointer")
			}
		}
		seen := map[string]bool{}
		for _, c := range cfg.Constraints {
			if w := cfg.Waiter; w != nil {
				fixed := int64(0)
				if c.Path == w.Page {
					fixed = 1
				}
				if c.Path == w.Size {
					fixed = int64(w.MaxIDs)
				}
				if fixed > 0 && ((c.Minimum != nil && fixed < *c.Minimum) || (c.Maximum != nil && fixed > *c.Maximum)) {
					return errors.New("policy: waiter forced fields violate constraints")
				}
				if c.Path == w.IDs && c.MaximumItems > 0 && w.MaxIDs > c.MaximumItems {
					return errors.New("policy: waiter IDs violate array bound")
				}
			}
			p, err := r.resolve(request, c.Path, "in")
			if err != nil {
				return err
			}
			if seen[c.Path] {
				return errors.New("policy: duplicate constraint")
			}
			seen[c.Path] = true
			if c.Minimum == nil && c.Maximum == nil && c.MinimumLength == nil && c.MaximumLength == nil && !c.ASCII && c.MaximumItems == 0 && !c.NonemptyItems {
				return errors.New("policy: empty constraint")
			}
			if c.Minimum != nil || c.Maximum != nil {
				if p.Type.Kind != "scalar" || p.Type.WireType != "integer" || c.MinimumLength != nil || c.MaximumLength != nil || c.ASCII || c.MaximumItems != 0 || c.NonemptyItems {
					return errors.New("policy: numeric constraint type mismatch")
				}
				if c.Minimum != nil && c.Maximum != nil && *c.Minimum > *c.Maximum {
					return errors.New("policy: inverted bounds")
				}
				low, high := int64(-9223372036854775808), int64(9223372036854775807)
				switch p.Type.DSLType {
				case "int8":
					low, high = -128, 127
				case "int16":
					low, high = -32768, 32767
				case "int32", "integer", "number":
					low, high = -2147483648, 2147483647
				case "int64", "long":
				case "uint8":
					low, high = 0, 255
				case "uint16":
					low, high = 0, 65535
				case "uint32":
					low, high = 0, 4294967295
				case "uint64":
					low = 0
				default:
					return errors.New("policy: unsupported integer constraint width")
				}
				for _, bound := range []*int64{c.Minimum, c.Maximum} {
					if bound != nil && (*bound < low || *bound > high) {
						return errors.New("policy: bound outside native integer width")
					}
				}
			}
			if c.MinimumLength != nil || c.MaximumLength != nil || c.ASCII {
				if p.Type.Kind != "scalar" || p.Type.DSLType != "string" || c.MaximumItems != 0 || c.NonemptyItems {
					return errors.New("policy: string constraint type mismatch")
				}
				if (c.MinimumLength != nil && *c.MinimumLength < 0) || (c.MaximumLength != nil && *c.MaximumLength < 0) || (c.MinimumLength != nil && c.MaximumLength != nil && *c.MinimumLength > *c.MaximumLength) {
					return errors.New("policy: invalid string bounds")
				}
			}
			if c.MaximumItems != 0 || c.NonemptyItems {
				if p.Type.Kind != "array" || c.MaximumItems < 0 {
					return errors.New("policy: array constraint type mismatch")
				}
				if c.NonemptyItems && (p.Type.Items == nil || p.Type.Items.Kind != "scalar" || p.Type.Items.DSLType != "string") {
					return errors.New("policy: nonempty items require strings")
				}
			}
		}
		if cfg.ClientToken != "" {
			found := false
			for _, c := range cfg.Constraints {
				if c.Path == cfg.ClientToken && c.ASCII && c.MinimumLength != nil && *c.MinimumLength == 1 && c.MaximumLength != nil && *c.MaximumLength == 64 {
					found = true
				}
			}
			if !found {
				return errors.New("policy: client token requires explicit ASCII/1..64 constraint")
			}
		}
		if len(cfg.Constraints) > 0 || (cfg.Paginator != nil && cfg.Paginator.Mode == "dual") {
			if err := reserve("Validate" + name + "Input"); err != nil {
				return err
			}
		}
		for _, id := range cfg.SensitiveModels {
			if _, ok := r.models[id]; !ok {
				return errors.New("policy: unknown sensitive model")
			}
			found := false
			for _, reachable := range op.ReachableModels {
				found = found || reachable == id
			}
			if !found {
				return errors.New("policy: sensitivity must belong to operation")
			}
			for _, field := range r.models[id].Fields {
				if name := r.fieldName(id, field); name == "String" || name == "GoString" {
					return errors.New("policy: sensitive formatting method conflicts with field")
				}
			}
		}
	}
	return nil
}
