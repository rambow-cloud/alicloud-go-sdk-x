package codegen

import (
	"bytes"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unicode"
)

func TestGuidesUseSeparateLanguagesWithSharedEvidence(t *testing.T) {
	p := readPolicyProduct(t, "sts")
	product, err := renderProduct(p)
	if err != nil {
		t.Fatal(err)
	}
	ecs, err := renderProduct(readPolicyProduct(t, "ecs"))
	if err != nil {
		t.Fatal(err)
	}
	for _, guides := range []struct {
		name string
		en   []byte
		zh   []byte
	}{
		{"sts", product["docs/products/sts.md"], product["docs/products/sts.zh-CN.md"]},
		{"ecs", ecs["docs/products/ecs.md"], ecs["docs/products/ecs.zh-CN.md"]},
	} {
		en, zh := string(guides.en), string(guides.zh)
		englishLink := "[中文](" + guides.name + ".zh-CN.md)"
		if !strings.Contains(en, englishLink) || !strings.Contains(zh, "[English]("+guides.name+".md)") {
			t.Fatal("missing language navigation", guides.name)
		}
		if strings.Contains(en, "## English") || strings.Contains(zh, "## 中文") {
			t.Fatal("mixed guide format", guides.name)
		}
		if strings.ContainsFunc(strings.ReplaceAll(en, englishLink, ""), func(r rune) bool { return unicode.Is(unicode.Han, r) }) {
			t.Fatal("Chinese prose in English guide", guides.name)
		}
		if !strings.ContainsFunc(zh, func(r rune) bool { return unicode.Is(unicode.Han, r) }) {
			t.Fatal("missing Chinese guidance", guides.name)
		}
		if guides.name == "sts" {
			inTable := false
			rows := 0
			for _, line := range strings.Split(en, "\n") {
				if !strings.HasPrefix(line, "| ") {
					inTable = false
					continue
				}
				if strings.HasPrefix(line, "| ---") {
					inTable = true
					continue
				}
				if inTable {
					rows++
					if !strings.Contains(zh, line) {
						t.Fatal("Chinese guide lost source or policy evidence", guides.name, line)
					}
				}
			}
			if rows != len(p.Operations)+len(p.Policy.Operations) {
				t.Fatal("source/policy rows not checked", rows)
			}
		}
	}
}

func TestDocumentationProseSafetyAndReadableFormatting(t *testing.T) {
	text := normalizeProse("### Title\n**Strong** `Value` [safe](https://help.aliyun.com/page) [unsafe](javascript:alert) <script>discard()</script> <b>word</b>\x00\u202e\n\ngo:linkname hijack runtime.fatal\nDeprecated: source-only\ndata:text/plain,hidden")
	if !strings.Contains(text, "Title\nStrong Value safe (https://help.aliyun.com/page)") || strings.Contains(text, "javascript:") || strings.Contains(text, "discard()") || strings.ContainsAny(text, "\x00\u202e") {
		t.Fatal("unsafe or unreadable prose", text)
	}
	if !strings.Contains(text, "Upstream text: go:linkname") || !strings.Contains(text, "Upstream text: Deprecated:") {
		t.Fatal("directive/deprecation injection", text)
	}
	r, err := newProductRenderer(readProductIR(t, "sts"))
	if err != nil {
		t.Fatal(err)
	}
	source := productSource{File: "products/sts/main.tea", Line: 1, Column: 1, EndLine: 2, EndColumn: 1}
	var b bytes.Buffer
	b.WriteString("package safe\n// Value is documented.\n")
	r.appendProse(&b, []productDocument{{Attribute: "description", Text: "go:build exclude\n//go:linkname hijack\n*/\npackage injected\nfunc malicious() {}", Source: source}})
	b.WriteString("type Value struct{}\n")
	formatted, err := productFormat(&b)
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := parser.ParseFile(token.NewFileSet(), "safe.go", formatted, parser.ParseComments)
	if err != nil {
		t.Fatal(err)
	}
	if parsed.Name.Name != "safe" || len(parsed.Decls) != 1 {
		t.Fatal("source escaped comments")
	}
	if _, ok := parsed.Decls[0].(*ast.GenDecl); !ok {
		t.Fatal("unexpected declaration")
	}
	for _, group := range parsed.Comments {
		for _, comment := range group.List {
			if strings.HasPrefix(comment.Text, "//go:") || strings.HasPrefix(comment.Text, "//line ") {
				t.Fatal("active compiler directive")
			}
		}
	}
}

func TestRealDocumentationCoverageAttributionAndMissingFallbacks(t *testing.T) {
	p := readPolicyProduct(t, "sts")
	r, err := newProductRenderer(p)
	if err != nil {
		t.Fatal(err)
	}
	report := r.documentationCoverage()
	if report.EnglishOperations != len(r.operations) || report.EnglishFields < 10 || report.TotalOperations != len(r.operations) || report.ChineseSemanticTranslation != "not-available-in-pinned-input" {
		t.Fatal("incorrect docs coverage", report)
	}
	files, err := renderProduct(p)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(files["service/sts/types.gen.go"], []byte("Minimum value: 900")) || !bytes.Contains(files["service/sts/operations.gen.go"], []byte("Obtains a Security Token Service")) {
		t.Fatal("licensed descriptions missing")
	}
	if !bytes.Contains(files["service/sts/NOTICE"], []byte(p.Provenance.Revision)) || !bytes.Contains(files["service/sts/LICENSE"], []byte("END OF TERMS AND CONDITIONS")) {
		t.Fatal("missing redistribution notices")
	}
	pinned, err := os.ReadFile(filepath.Join("..", "..", "sources", "darabonba", "licenses", "alibabacloud-openapi-paginator.NOTICE"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(productApacheLicense, pinned) {
		t.Fatal("standard license terms changed")
	}
	if files["docs/products/sts.documentation.json"] == nil {
		t.Fatal("missing coverage artifact")
	}
	for _, test := range []struct {
		docs   []productDocument
		status string
	}{{nil, "missing"}, {[]productDocument{{Attribute: "description", Text: " "}}, "empty"}, {[]productDocument{{Attribute: "description", Text: "仅有中文"}}, "no-english-prose"}, {[]productDocument{{Attribute: "description", Text: "English text"}}, "emitted"}} {
		if got := documentCoverage(test.docs).Status; got != test.status {
			t.Fatal(got, test.status)
		}
	}
}

func TestDocumentationRejectsUnboundSourceAndPromotedExamples(t *testing.T) {
	for _, mutate := range []func(*productIR){
		func(p *productIR) {
			p.Models[0].Fields[0].Documentation = []productDocument{{Attribute: "description", Text: "unsafe", Source: productSource{File: "../secret", Line: 1, Column: 1, EndLine: 1, EndColumn: 2}}}
		},
		func(p *productIR) {
			p.Operations[0].Documentation = append(p.Operations[0].Documentation, p.Operations[0].Documentation[0])
		},
		func(p *productIR) {
			p.Models[0].Fields[0].Documentation = []productDocument{{Attribute: "example", Text: "actual-key"}}
		},
	} {
		p := readProductIR(t, "sts")
		mutate(&p)
		if _, err := renderProduct(p); err == nil {
			t.Fatal("invalid docs rendered")
		}
	}
}

func TestDescriptionEditsLeaveExecutableExamplesAndNativeModelsUnchanged(t *testing.T) {
	p := readPolicyProduct(t, "sts")
	before, err := renderProduct(p)
	if err != nil {
		t.Fatal(err)
	}
	p.Operations[0].Documentation[0].Text = "Replacement service summary."
	after, err := renderProduct(p)
	if err != nil {
		t.Fatal(err)
	}
	for name, data := range before {
		if strings.Contains(name, "examples") || strings.HasSuffix(name, "client.gen.go") || strings.HasSuffix(name, "policy.gen.go") {
			if !bytes.Equal(data, after[name]) {
				t.Fatal("prose changed executable behavior", name)
			}
		}
	}
}
