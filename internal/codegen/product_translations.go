package codegen

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
	"unicode"
)

type productTranslation struct {
	Attribute  string        `json:"attribute"`
	Source     productSource `json:"source"`
	TextSHA256 string        `json:"sourceTextSHA256"`
	English    string        `json:"english"`
	Chinese    string        `json:"chinese"`
}

type productTranslationFile struct {
	SchemaVersion        int                           `json:"schemaVersion"`
	Product              string                        `json:"product"`
	SourceManifestSHA256 string                        `json:"sourceManifestSHA256"`
	Operations           map[string]productTranslation `json:"operations"`
}

func loadProductTranslations(root string, products []productIR) error {
	for i := range products {
		p := &products[i]
		file := "translations/" + p.Product + ".json"
		if err := safeOutputPath(root, file); err != nil {
			return err
		}
		var translation productTranslationFile
		if err := readJSON(filepath.Join(root, filepath.FromSlash(file)), &translation, true); err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				continue
			}
			return fmt.Errorf("translation %s: %w", p.Product, err)
		}
		if translation.SchemaVersion != 1 || translation.Product != p.Product || translation.SourceManifestSHA256 != p.Provenance.SourceManifestSHA256 {
			return errors.New("translation: schema/product/source mismatch")
		}
		p.Translations = translation.Operations
		if err := validateProductTranslations(*p); err != nil {
			return err
		}
	}
	return nil
}

func validateProductTranslations(p productIR) error {
	for name, translated := range p.Translations {
		if translated.Attribute != "summary" || strings.TrimSpace(translated.English) == "" || !englishProse(normalizeProse(translated.English)) || strings.ContainsFunc(translated.English, func(r rune) bool { return unicode.Is(unicode.Han, r) }) || !strings.ContainsFunc(translated.Chinese, func(r rune) bool { return unicode.Is(unicode.Han, r) }) {
			return errors.New("translation: invalid attribute or language")
		}
		bound := false
		for _, op := range p.Operations {
			if op.Name != name {
				continue
			}
			if documentCoverage(op.Documentation).Status == "emitted" {
				return errors.New("translation: existing English prose cannot be replaced")
			}
			for _, doc := range op.Documentation {
				if doc.Attribute == translated.Attribute && doc.Source == translated.Source && digest([]byte(doc.Text)) == translated.TextSHA256 && doc.Text == translated.Chinese {
					bound = true
				}
			}
		}
		if !bound {
			return fmt.Errorf("translation: unbound source for %s", name)
		}
	}
	return nil
}

func (r *productRenderer) appendOperationProse(b *bytes.Buffer, op productOperation) {
	// Source annotations remain in IR; translation never changes runtime discovery.
	if translated, ok := r.p.Translations[op.Name]; ok {
		b.WriteString("//\n// Reviewed English translation of upstream documentation (Apache-2.0):\n//\n")
		for _, line := range strings.Split(normalizeProse(translated.English), "\n") {
			b.WriteString("// " + line + "\n")
		}
		b.WriteString("//\n// Source: " + r.sourceURL(translated.Source) + "\n")
		return
	}
	r.appendProse(b, op.Documentation)
	if documentCoverage(op.Documentation).Status != "emitted" {
		fmt.Fprintf(b, "// Source: %s\n", r.sourceURL(op.Source))
	}
}
