// Command doccheck checks public Go documentation, runnable external examples,
// and the project's JSON import policy. It runs from the module root.
package main

import (
	"bytes"
	"encoding/json/jsontext"
	"encoding/json/v2"
	"fmt"
	"go/ast"
	"go/doc"
	"go/parser"
	"go/token"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

type packageInfo struct {
	Dir          string
	ImportPath   string
	Name         string
	GoFiles      []string
	CgoFiles     []string
	TestGoFiles  []string
	XTestGoFiles []string
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	cmd := exec.Command("go", "list", "-json", "./...")
	cmd.Stderr = os.Stderr
	data, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("list packages: %w", err)
	}
	decoder := jsontext.NewDecoder(bytes.NewReader(data))
	var failures []string
	count := 0
	for {
		var pkg packageInfo
		if err := json.UnmarshalDecode(decoder, &pkg); err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("decode go list: %w", err)
		}
		problems, err := checkPackage(pkg)
		if err != nil {
			return err
		}
		failures = append(failures, problems...)
		if isPublic(pkg) {
			count++
		}
	}
	if len(failures) > 0 {
		return fmt.Errorf("documentation policy failed:\n%s", strings.Join(failures, "\n"))
	}
	fmt.Printf("Documentation policy passed for %d public packages; JSON imports use v2.\n", count)
	return nil
}

func isPublic(pkg packageInfo) bool {
	return pkg.Name != "main" && !strings.Contains("/"+pkg.ImportPath+"/", "/internal/")
}

func checkPackage(pkg packageInfo) ([]string, error) {
	fset := token.NewFileSet()
	var failures []string
	report := func(pos token.Pos, message string) {
		failures = append(failures, fmt.Sprintf("%s: %s", fset.Position(pos), message))
	}
	var external []*ast.File
	hasOverview := false
	files := append(append(append(append([]string{}, pkg.GoFiles...), pkg.CgoFiles...), pkg.TestGoFiles...), pkg.XTestGoFiles...)
	for _, name := range files {
		file, err := parser.ParseFile(fset, filepath.Join(pkg.Dir, name), nil, parser.ParseComments)
		if err != nil {
			return nil, err
		}
		for _, imp := range file.Imports {
			path, err := strconv.Unquote(imp.Path.Value)
			if err != nil {
				return nil, err
			}
			if path == "encoding/json" || strings.HasPrefix(path, "github.com/go-json-experiment/json") {
				report(imp.Pos(), "use encoding/json/v2 directly (Go 1.27)")
			}
		}
		if !isPublic(pkg) {
			continue
		}
		if strings.HasSuffix(name, "_test.go") {
			if file.Name.Name == pkg.Name+"_test" {
				external = append(external, file)
			}
			continue
		}
		if name == "doc.go" && startsWithName(file.Doc, "Package "+pkg.Name) {
			hasOverview = true
		}
		for _, decl := range file.Decls {
			switch decl := decl.(type) {
			case *ast.FuncDecl:
				if !decl.Name.IsExported() {
					continue
				}
				if decl.Recv != nil && !ast.IsExported(typeName(decl.Recv.List[0].Type)) {
					continue
				}
				if !startsWithName(decl.Doc, decl.Name.Name) {
					report(decl.Pos(), decl.Name.Name+" needs a doc comment beginning with its name")
				}
			case *ast.GenDecl:
				for _, spec := range decl.Specs {
					switch spec := spec.(type) {
					case *ast.TypeSpec:
						if !spec.Name.IsExported() {
							continue
						}
						comment := spec.Doc
						if comment == nil && len(decl.Specs) == 1 {
							comment = decl.Doc
						}
						if !startsWithName(comment, spec.Name.Name) {
							report(spec.Pos(), spec.Name.Name+" needs a doc comment beginning with its name")
						}
						ast.Inspect(spec.Type, func(node ast.Node) bool {
							switch node := node.(type) {
							case *ast.StructType:
								checkFields(node.Fields, report)
							case *ast.InterfaceType:
								checkFields(node.Methods, report)
							}
							return true
						})
					case *ast.ValueSpec:
						comment := spec.Doc
						if comment == nil && len(decl.Specs) == 1 {
							comment = decl.Doc
						}
						for _, name := range spec.Names {
							if name.IsExported() && !startsWithName(comment, name.Name) {
								report(name.Pos(), name.Name+" needs a doc comment beginning with its name")
							}
						}
					}
				}
			}
		}
	}
	if !isPublic(pkg) {
		return failures, nil
	}
	if !hasOverview {
		failures = append(failures, pkg.ImportPath+": missing doc.go package overview")
	}
	hasExample := false
	for _, example := range doc.Examples(external...) {
		if example.Output != "" || example.EmptyOutput {
			hasExample = true
		}
	}
	if !hasExample {
		failures = append(failures, pkg.ImportPath+": missing external Example with Output")
	}
	return failures, nil
}

func startsWithName(comment *ast.CommentGroup, name string) bool {
	if comment == nil {
		return false
	}
	text := strings.TrimSpace(comment.Text())
	return text == name || strings.HasPrefix(text, name+" ") || strings.HasPrefix(text, name+"\n")
}

func typeName(expr ast.Expr) string {
	switch expr := expr.(type) {
	case *ast.Ident:
		return expr.Name
	case *ast.SelectorExpr:
		return expr.Sel.Name
	case *ast.StarExpr:
		return typeName(expr.X)
	case *ast.IndexExpr:
		return typeName(expr.X)
	case *ast.IndexListExpr:
		return typeName(expr.X)
	}
	return ""
}

func checkFields(list *ast.FieldList, report func(token.Pos, string)) {
	if list == nil {
		return
	}
	for _, field := range list.List {
		var names []string
		if len(field.Names) == 0 {
			names = append(names, typeName(field.Type))
		}
		for _, name := range field.Names {
			names = append(names, name.Name)
		}
		for _, name := range names {
			if ast.IsExported(name) && !startsWithName(field.Doc, name) && !startsWithName(field.Comment, name) {
				report(field.Pos(), name+" field or interface method needs a doc comment beginning with its name")
			}
		}
	}
}
