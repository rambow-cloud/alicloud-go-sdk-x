package codegen

import (
	"bytes"
	_ "embed"
	"errors"
	"fmt"
	"html"
	"net/url"
	"regexp"
	"strings"
	"unicode"
)

// Full standard terms are copied byte-for-byte from the pinned Paginator license.
// No Paginator implementation is emitted or required at runtime.
//
//go:embed apache-license.txt
var productApacheLicense []byte

var proseLinks = regexp.MustCompile(`!?\[([^\]\n]*)\]\(([^\s)]+)\)`)
var proseBlocks = regexp.MustCompile(`(?is)<(script|style)\b[^>]*>.*?</(script|style)\s*>`)
var proseHTML = regexp.MustCompile(`(?i)</?(a|b|i|p|div|span|br|ul|ol|li|pre|code|strong|em|table|tr|td|th|details|summary|img|iframe|svg|object|embed|form|input|h[1-6])\b[^>]*>`)
var proseURLs = regexp.MustCompile(`(?i)(?:[a-z][a-z0-9+.-]*://|javascript:|data:)\S+`)
var proseStrong = regexp.MustCompile(`\*\*([^*\n]+)\*\*`)
var proseHeading = regexp.MustCompile(`(?m)^\s*#{1,6}\s+`)

func safeProseURL(value string) bool {
	u, err := url.Parse(value)
	return err == nil && u.Scheme == "https" && u.Hostname() != "" && u.User == nil && !strings.ContainsAny(value, "\r\n\t<>\"'\\")
}

// Plain paragraphs keep source text inside comments and never acquire SDK semantics.
func normalizeProse(text string) string {
	text = html.UnescapeString(text)
	text = strings.ReplaceAll(strings.ReplaceAll(text, "\r\n", "\n"), "\r", "\n")
	text = strings.Map(func(r rune) rune {
		if r == '\n' {
			return r
		}
		if r == '\t' {
			return ' '
		}
		if unicode.IsControl(r) || unicode.Is(unicode.Cf, r) {
			return -1
		}
		return r
	}, text)
	text = proseBlocks.ReplaceAllString(text, "")
	text = proseHTML.ReplaceAllString(text, "")
	text = proseLinks.ReplaceAllStringFunc(text, func(match string) string {
		parts := proseLinks.FindStringSubmatch(match)
		if safeProseURL(parts[2]) {
			return parts[1] + " (" + parts[2] + ")"
		}
		return parts[1] + " (link omitted)"
	})
	text = proseURLs.ReplaceAllStringFunc(text, func(link string) string {
		// Closing prose punctuation is outside the URL, including Markdown delimiters.
		trimmed := strings.TrimRight(link, ").,;]>`")
		if safeProseURL(trimmed) {
			return link
		}
		return "[link omitted]"
	})
	text = proseHeading.ReplaceAllString(text, "")
	text = proseStrong.ReplaceAllString(text, "$1")
	text = strings.ReplaceAll(text, "`", "")
	var lines []string
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "go:") || strings.HasPrefix(line, "line ") || strings.HasPrefix(line, "#cgo") || strings.HasPrefix(line, "Deprecated:") {
			line = "Upstream text: " + line
		}
		if line == "" && (len(lines) == 0 || lines[len(lines)-1] == "") {
			continue
		}
		lines = append(lines, line)
	}
	return strings.TrimSpace(strings.Join(lines, "\n"))
}

func englishProse(text string) bool {
	for _, r := range text {
		if unicode.Is(unicode.Latin, r) && unicode.IsLetter(r) {
			return true
		}
	}
	return false
}

func (r *productRenderer) sourceURL(source productSource) string {
	identifier := r.p.Product + "-" + strings.ReplaceAll(r.p.Version, "-", "")
	return fmt.Sprintf("%s/blob/%s/%s/main.tea#L%d", r.p.Provenance.Repository, r.p.Provenance.Revision, identifier, source.Line)
}
func (r *productRenderer) validateDocuments(docs []productDocument) error {
	seen := map[string]bool{}
	for _, doc := range docs {
		if doc.Attribute != "summary" && doc.Attribute != "description" && doc.Attribute != "example" {
			return errors.New("documentation: unsupported attribute")
		}
		if seen[doc.Attribute] {
			return errors.New("documentation: duplicate attribute")
		}
		seen[doc.Attribute] = true
		if doc.Attribute == "example" && doc.Text != "" {
			return errors.New("documentation: source examples must remain coordinates only")
		}
		s := doc.Source
		if s.File != "products/"+r.p.Product+"/main.tea" || s.Line < 1 || s.Column < 1 || s.EndLine < s.Line || s.EndColumn < 1 {
			return errors.New("documentation: invalid source coordinate")
		}
	}
	return nil
}
func (r *productRenderer) appendProse(b *bytes.Buffer, docs []productDocument) {
	for _, doc := range docs {
		if doc.Attribute == "example" {
			continue
		}
		text := normalizeProse(doc.Text)
		if text == "" || !englishProse(text) {
			continue
		}
		b.WriteString("//\n// Upstream service documentation (Apache-2.0; informational, not SDK validation):\n//\n")
		for _, line := range strings.Split(text, "\n") {
			if line == "" {
				b.WriteString("//\n")
			} else {
				fmt.Fprintf(b, "// %s\n", line)
			}
		}
		fmt.Fprintf(b, "//\n// Source: %s\n", r.sourceURL(doc.Source))
	}
}

type descriptionCoverage struct {
	Status  string          `json:"status"`
	Sources []productSource `json:"sources"`
}

func documentCoverage(docs []productDocument) descriptionCoverage {
	result := descriptionCoverage{Status: "missing", Sources: []productSource{}}
	for _, doc := range docs {
		if doc.Attribute == "example" {
			continue
		}
		result.Sources = append(result.Sources, doc.Source)
		text := normalizeProse(doc.Text)
		if englishProse(text) {
			result.Status = "emitted"
		} else if result.Status != "emitted" {
			if text == "" {
				result.Status = "empty"
			} else {
				result.Status = "no-english-prose"
			}
		}
	}
	return result
}

type productDocCoverage struct {
	Generator                  string                         `json:"generator"`
	SchemaVersion              int                            `json:"schemaVersion"`
	Product                    string                         `json:"product"`
	License                    string                         `json:"license"`
	SourceManifestSHA256       string                         `json:"sourceManifestSHA256"`
	EnglishOperations          int                            `json:"englishOperations"`
	TotalOperations            int                            `json:"totalOperations"`
	EnglishFields              int                            `json:"englishFields"`
	TotalFields                int                            `json:"totalFields"`
	ChineseSemanticTranslation string                         `json:"chineseSemanticTranslation"`
	PairedGuides               string                         `json:"pairedGuides"`
	Examples                   string                         `json:"examples"`
	Operations                 map[string]descriptionCoverage `json:"operations"`
	Fields                     map[string]descriptionCoverage `json:"fields"`
}

func (r *productRenderer) documentationCoverage() productDocCoverage {
	report := productDocCoverage{Generator: "sdkgen product", SchemaVersion: 1, Product: r.p.Product, License: "Apache-2.0", SourceManifestSHA256: r.p.Provenance.SourceManifestSHA256, ChineseSemanticTranslation: "not-available-in-pinned-input", PairedGuides: "usage-contracts-and-source-index", Examples: "offline-invocation-and-reviewed-capability-examples", Operations: map[string]descriptionCoverage{}, Fields: map[string]descriptionCoverage{}}
	for _, op := range r.operations {
		coverage := documentCoverage(op.Documentation)
		report.Operations[op.Name] = coverage
		report.TotalOperations++
		if coverage.Status == "emitted" {
			report.EnglishOperations++
		}
	}
	for id, m := range r.models {
		for _, f := range m.Fields {
			coverage := documentCoverage(f.Documentation)
			report.Fields[id+"#"+f.DSLName] = coverage
			report.TotalFields++
			if coverage.Status == "emitted" {
				report.EnglishFields++
			}
		}
	}
	return report
}
func (r *productRenderer) appendDocumentationGuide(b *bytes.Buffer) {
	report := r.documentationCoverage()
	fmt.Fprintf(b, "## Documentation sources / 文档来源\n\n### English\n\nEnglish Go comments reuse licensed parser descriptions and summaries: %d/%d operations\nand %d/%d fields have emitted prose. Missing/empty/non-English descriptions are retained\nin `%s.documentation.json`; comments do not create validators or requiredness. This\npaired guide covers usage/contracts and the same source index in both languages.\nChinese semantic translations are not available in the pinned input; none are invented.\nAll executable Examples use offline scripted responses, not upstream example values.\nSee [documentation rules](../product-documentation.md), package LICENSE and NOTICE.\n\n### 中文\n\n英文 Go 注释复用授权 parser 说明/摘要：%d/%d 操作、%d/%d 字段有已输出说明。\n缺失/空/纯非英文说明保留在 `%s.documentation.json`，说明不自动变为校验或必填。\n本指南中英文同步使用/契约及相同来源索引；固定输入没有中文语义翻译，不编造。\n所有可执行 Example 使用离线脚本响应，不使用上游 example 值。规则见\n[文档规格](../product-documentation.md)，许可/归属见包内 LICENSE 和 NOTICE。\n\n", report.EnglishOperations, report.TotalOperations, report.EnglishFields, report.TotalFields, r.p.Product, report.EnglishOperations, report.TotalOperations, report.EnglishFields, report.TotalFields, r.p.Product)
	b.WriteString("| Operation / 操作 | English prose / 英文说明 | Pinned source / 固定来源 |\n| --- | --- | --- |\n")
	for _, op := range r.operations {
		fmt.Fprintf(b, "| %s | %s | [DSL](%s) |\n", op.Name, report.Operations[op.Name].Status, r.sourceURL(op.Source))
	}
	b.WriteString("\n")
}
