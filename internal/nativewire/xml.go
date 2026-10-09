// Package nativewire reads source-bound native serialization declarations as data.
// It never executes native code or supplies operation/schema/runtime policy.
package nativewire

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"net/url"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

// ErrInvalid indicates a source hash, syntax or static registry ambiguity.
// Error text never contains source contents.
var ErrInvalid = errors.New("nativewire: invalid pinned source or registry")

// Source is a complete original source file bound to a reviewed SHA256.
type Source struct {
	// Name is the source-relative filename used in output coordinates.
	Name string
	// Bytes contains the complete file; it is not mutated or retained.
	Bytes []byte
	// SHA256 is the exact lowercase hexadecimal file digest.
	SHA256 string
}

// Coordinate identifies an original source declaration.
type Coordinate struct {
	// File is the source-relative file.
	File string `json:"file"`
	// Line is the one-based source line.
	Line int `json:"line"`
	// Column is the one-based source column.
	Column int `json:"column"`
}

// Type records native field shape without choosing a DSL/Go output type.
type Type struct {
	// Kind is scalar, model, array or unsupported.
	Kind string `json:"kind"`
	// Name is the exact primitive or referenced native model name.
	Name string `json:"name,omitempty"`
	// Optional records pointer presence in the native declaration.
	Optional bool `json:"optional,omitempty"`
	// Items is the repeated-element type, when Kind is array.
	Items *Type `json:"items,omitempty"`
}

// Field records explicit JSON/XML wire names and source/type facts.
type Field struct {
	// GoName is the native field identifier, never used to infer a root.
	GoName string `json:"goName"`
	// JSONName is the explicit JSON tag name; absent tags remain empty.
	JSONName string `json:"jsonName,omitempty"`
	// XMLName is the explicit local XML tag name/path; absent tags remain empty.
	XMLName string `json:"xmlName,omitempty"`
	// XMLOptions preserves explicit native tag options for compatibility review.
	XMLOptions []string `json:"xmlOptions,omitempty"`
	// Namespace is the explicit XML tag namespace URI, or empty when unspecified.
	Namespace string `json:"namespace,omitempty"`
	// Type records the native type facts.
	Type Type `json:"type"`
	// Source identifies the field declaration.
	Source Coordinate `json:"source"`
}

// Model records a complete native struct's field declarations.
type Model struct {
	// Name is the native model identifier.
	Name string `json:"name"`
	// Fields preserves declaration order, including unsupported fields.
	Fields []Field `json:"fields"`
}

// Root records a static registry action's explicit XML root candidate.
type Root struct {
	// Action is the exact static registry key.
	Action string `json:"action"`
	// Model is the exact registered native struct.
	Model string `json:"model"`
	// Kind is structured, scalar or unsupported; it does not prove nested compatibility.
	Kind string `json:"kind"`
	// Name is the explicit local XML root tag.
	Name string `json:"name,omitempty"`
	// NamespaceMatch is local when a native tag omits its namespace, or exact.
	NamespaceMatch string `json:"namespaceMatch,omitempty"`
	// Namespace is the explicit namespace URI for exact matching.
	Namespace string `json:"namespace,omitempty"`
	// Field is the explicit wrapper/scalar root field.
	Field *Field `json:"field,omitempty"`
	// Reason explains an unsupported root without inferred replacement.
	Reason string `json:"reason,omitempty"`
	// Source identifies the registry assignment.
	Source Coordinate `json:"source"`
}

// Inventory contains deterministic facts from complete verified files.
type Inventory struct {
	// SchemaVersion identifies this facts format; it is independent of product IR.
	SchemaVersion int `json:"schemaVersion"`
	// RegistrySHA256 identifies the complete verified registry input.
	RegistrySHA256 string `json:"registrySHA256"`
	// ModelsSHA256 identifies the complete verified model input.
	ModelsSHA256 string `json:"modelsSHA256"`
	// Models is sorted by native identifier.
	Models []Model `json:"models"`
	// Roots is sorted by exact action name.
	Roots []Root `json:"roots"`
}

func xmlName(name string) bool {
	if name == "" || !utf8.ValidString(name) {
		return false
	}
	for i, r := range name {
		if r == '_' || unicode.IsLetter(r) || (i > 0 && (unicode.IsDigit(r) || unicode.IsMark(r) || r == '-' || r == '.')) {
			continue
		}
		return false
	}
	return true
}

func verified(source Source) bool {
	digest := sha256.Sum256(source.Bytes)
	return source.Name != "" && len(source.Bytes) > 0 && len(source.Bytes) <= 8<<20 && hex.EncodeToString(digest[:]) == source.SHA256
}

func coordinate(set *token.FileSet, pos token.Pos) Coordinate {
	position := set.Position(pos)
	return Coordinate{position.Filename, position.Line, position.Column}
}

func identifier(expr ast.Expr, name string) bool {
	id, ok := expr.(*ast.Ident)
	return ok && id.Name == name
}

func nativeType(expr ast.Expr, models map[string]*ast.StructType) Type {
	switch value := expr.(type) {
	case *ast.StarExpr:
		if _, nested := value.X.(*ast.StarExpr); nested {
			return Type{Kind: "unsupported"}
		}
		result := nativeType(value.X, models)
		result.Optional = true
		return result
	case *ast.ArrayType:
		if value.Len == nil {
			item := nativeType(value.Elt, models)
			return Type{Kind: "array", Items: &item}
		}
	case *ast.Ident:
		if _, known := models[value.Name]; known {
			return Type{Kind: "model", Name: value.Name}
		}
		switch value.Name {
		case "string", "bool", "int", "int8", "int16", "int32", "int64", "uint", "uint8", "uint16", "uint32", "uint64", "float32", "float64":
			return Type{Kind: "scalar", Name: value.Name}
		}
	}
	return Type{Kind: "unsupported"}
}

func fields(ctx context.Context, set *token.FileSet, model *ast.StructType, all map[string]*ast.StructType) ([]Field, error) {
	var fields []Field
	seen := map[string]bool{}
	for _, declaration := range model.Fields.List {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if len(declaration.Names) != 1 {
			return nil, ErrInvalid
		}
		name := declaration.Names[0].Name
		if seen[name] {
			return nil, ErrInvalid
		}
		seen[name] = true
		field := Field{GoName: name, Type: nativeType(declaration.Type, all), Source: coordinate(set, declaration.Pos())}
		if declaration.Tag != nil {
			text, err := strconv.Unquote(declaration.Tag.Value)
			if err != nil {
				return nil, ErrInvalid
			}
			tag := reflect.StructTag(text)
			field.JSONName = strings.Split(tag.Get("json"), ",")[0]
			xmlTag := strings.Split(tag.Get("xml"), ",")
			field.XMLName = xmlTag[0]
			if len(xmlTag) > 1 {
				field.XMLOptions = xmlTag[1:]
			}
			if namespace, local, found := strings.Cut(field.XMLName, " "); found {
				field.Namespace, field.XMLName = namespace, local
			}
		}
		fields = append(fields, field)
	}
	return fields, nil
}

func rootCandidate(action, model string, source Coordinate, all map[string]Model) Root {
	root := Root{Action: action, Model: model, Kind: "unsupported", Source: source}
	definition, known := all[model]
	if !known {
		root.Reason = "registered model missing"
		return root
	}
	if len(definition.Fields) != 1 {
		root.Reason = "root requires one explicit wrapper/scalar field"
		return root
	}
	field := definition.Fields[0]
	for _, option := range field.XMLOptions {
		if option != "omitempty" {
			root.Reason = "root XML tag option unsupported"
			return root
		}
	}
	if !xmlName(field.XMLName) {
		root.Reason = "root tag missing or unsupported path/attribute"
		return root
	}
	if field.Namespace != "" {
		uri, err := url.Parse(field.Namespace)
		if err != nil || !uri.IsAbs() {
			root.Reason = "root namespace is not an absolute URI"
			return root
		}
	}
	root.Name, root.Namespace, root.Field = field.XMLName, field.Namespace, &field
	root.NamespaceMatch = "local"
	if field.Namespace != "" {
		root.NamespaceMatch = "exact"
	}
	switch field.Type.Kind {
	case "scalar":
		root.Kind = "scalar"
	case "model":
		root.Kind = "structured"
	default:
		root.Reason = "root field type unsupported"
	}
	return root
}

// ParseXML verifies complete source hashes, reads Go AST data and returns sorted
// explicit root facts. registryName selects a reviewed registry symbol, never an
// action-name convention. Dynamic/duplicate writes or aliases fail without partial
// output. Unsupported root shapes retain reasons. Cancellation is preserved;
// source bytes are never mutated/retained and independent calls are concurrency safe.
func ParseXML(ctx context.Context, registry, models Source, registryName string) (Inventory, error) {
	if err := ctx.Err(); err != nil {
		return Inventory{}, err
	}
	if !verified(registry) || !verified(models) || registryName == "" {
		return Inventory{}, ErrInvalid
	}
	set := token.NewFileSet()
	modelFile, err := parser.ParseFile(set, models.Name, models.Bytes, 0)
	if err != nil {
		return Inventory{}, ErrInvalid
	}
	registryFile, err := parser.ParseFile(set, registry.Name, registry.Bytes, 0)
	if err != nil {
		return Inventory{}, ErrInvalid
	}
	if modelFile.Name.Name != registryFile.Name.Name {
		return Inventory{}, ErrInvalid
	}
	modelRegistryReference := false
	ast.Inspect(modelFile, func(node ast.Node) bool {
		if id, ok := node.(*ast.Ident); ok && id.Name == registryName && (id.Obj == nil || id.Obj.Kind == ast.Var) {
			modelRegistryReference = true
		}
		return !modelRegistryReference
	})
	if modelRegistryReference {
		return Inventory{}, ErrInvalid
	}
	registryObject := registryFile.Scope.Lookup(registryName)
	if registryObject == nil || registryObject.Kind != ast.Var {
		return Inventory{}, ErrInvalid
	}
	reflectAliases := map[string]bool{}
	for _, imported := range registryFile.Imports {
		path, err := strconv.Unquote(imported.Path.Value)
		if err != nil {
			return Inventory{}, ErrInvalid
		}
		if path == "reflect" {
			name := "reflect"
			if imported.Name != nil {
				name = imported.Name.Name
			}
			if name != "." && name != "_" {
				reflectAliases[name] = true
			}
		}
	}
	// The bounded profile accepts an empty, explicitly typed map initializer.
	// Every other reference must be an indexed lookup/write, so aliases, whole-map
	// replacement, address-taking and indirect mutations cannot escape discovery.
	declaration, ok := registryObject.Decl.(*ast.ValueSpec)
	if !ok || len(declaration.Names) != 1 || len(declaration.Values) != 1 {
		return Inventory{}, ErrInvalid
	}
	initializer, ok := declaration.Values[0].(*ast.CallExpr)
	if !ok || len(initializer.Args) != 1 || initializer.Ellipsis.IsValid() || !identifier(initializer.Fun, "make") || initializer.Fun.(*ast.Ident).Obj != nil {
		return Inventory{}, ErrInvalid
	}
	mapType, ok := initializer.Args[0].(*ast.MapType)
	if !ok || !identifier(mapType.Key, "string") {
		return Inventory{}, ErrInvalid
	}
	mapValue, ok := mapType.Value.(*ast.SelectorExpr)
	if !ok || mapValue.Sel.Name != "Type" {
		return Inventory{}, ErrInvalid
	}
	mapReflect, ok := mapValue.X.(*ast.Ident)
	if !ok || mapReflect.Obj != nil || !reflectAliases[mapReflect.Name] {
		return Inventory{}, ErrInvalid
	}
	allowedReferences := map[*ast.Ident]bool{declaration.Names[0]: true}
	ast.Inspect(registryFile, func(node ast.Node) bool {
		if index, ok := node.(*ast.IndexExpr); ok {
			if id, ok := index.X.(*ast.Ident); ok && id.Obj == registryObject {
				allowedReferences[id] = true
			}
		}
		return true
	})
	ambiguous := false
	ast.Inspect(registryFile, func(node ast.Node) bool {
		if id, ok := node.(*ast.Ident); ok && id.Obj == registryObject && !allowedReferences[id] {
			ambiguous = true
		}
		return !ambiguous
	})
	if ambiguous {
		return Inventory{}, ErrInvalid
	}
	allowedAssignments := map[*ast.AssignStmt]bool{}
	for _, declaration := range registryFile.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if !ok || function.Recv != nil || function.Name.Name != "init" || function.Body == nil {
			continue
		}
		for _, statement := range function.Body.List {
			if assignment, ok := statement.(*ast.AssignStmt); ok {
				allowedAssignments[assignment] = true
			}
		}
	}
	declarations := map[string]*ast.StructType{}
	for _, node := range modelFile.Decls {
		group, ok := node.(*ast.GenDecl)
		if !ok || group.Tok != token.TYPE {
			continue
		}
		for _, spec := range group.Specs {
			definition, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}
			body, ok := definition.Type.(*ast.StructType)
			if !ok {
				continue
			}
			if _, exists := declarations[definition.Name.Name]; exists {
				return Inventory{}, ErrInvalid
			}
			declarations[definition.Name.Name] = body
		}
	}
	result := Inventory{SchemaVersion: 1, RegistrySHA256: registry.SHA256, ModelsSHA256: models.SHA256, Models: []Model{}, Roots: []Root{}}
	all := map[string]Model{}
	for name, body := range declarations {
		items, err := fields(ctx, set, body, declarations)
		if err != nil {
			return Inventory{}, err
		}
		model := Model{Name: name, Fields: items}
		all[name] = model
		result.Models = append(result.Models, model)
	}
	seen := map[string]bool{}
	invalid := false
	ast.Inspect(registryFile, func(node ast.Node) bool {
		if invalid || ctx.Err() != nil {
			return false
		}
		switch statement := node.(type) {
		case *ast.CallExpr:
			for _, arg := range statement.Args {
				if identifier(arg, registryName) {
					invalid = true
					return false
				}
			}
		case *ast.AssignStmt:
			for _, expr := range statement.Rhs {
				if identifier(expr, registryName) {
					invalid = true
					return false
				}
			}
			for _, expr := range statement.Lhs {
				index, ok := expr.(*ast.IndexExpr)
				if !ok || !identifier(index.X, registryName) {
					continue
				}
				registryID := index.X.(*ast.Ident)
				if registryID.Obj != registryObject || !allowedAssignments[statement] {
					invalid = true
					return false
				}
				if statement.Tok != token.ASSIGN || len(statement.Lhs) != 1 || len(statement.Rhs) != 1 {
					invalid = true
					return false
				}
				key, ok := index.Index.(*ast.BasicLit)
				if !ok || key.Kind != token.STRING {
					invalid = true
					return false
				}
				action, err := strconv.Unquote(key.Value)
				if err != nil || action == "" || seen[action] {
					invalid = true
					return false
				}
				call, ok := statement.Rhs[0].(*ast.CallExpr)
				if !ok || len(call.Args) != 1 || call.Ellipsis.IsValid() {
					invalid = true
					return false
				}
				selector, ok := call.Fun.(*ast.SelectorExpr)
				if !ok || selector.Sel.Name != "TypeOf" {
					invalid = true
					return false
				}
				reflectID, ok := selector.X.(*ast.Ident)
				if !ok || reflectID.Obj != nil || !reflectAliases[reflectID.Name] {
					invalid = true
					return false
				}
				value, ok := call.Args[0].(*ast.CompositeLit)
				if !ok || len(value.Elts) != 0 {
					invalid = true
					return false
				}
				model, ok := value.Type.(*ast.Ident)
				if !ok || model.Obj != nil {
					invalid = true
					return false
				}
				seen[action] = true
				result.Roots = append(result.Roots, rootCandidate(action, model.Name, coordinate(set, statement.Pos()), all))
			}
		}
		return true
	})
	if err := ctx.Err(); err != nil {
		return Inventory{}, err
	}
	if invalid || len(result.Roots) == 0 {
		return Inventory{}, ErrInvalid
	}
	sort.Slice(result.Models, func(i, j int) bool { return result.Models[i].Name < result.Models[j].Name })
	sort.Slice(result.Roots, func(i, j int) bool { return result.Roots[i].Action < result.Roots[j].Action })
	return result, nil
}
