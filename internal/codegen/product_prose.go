package codegen

import (
	"bytes"
	"encoding/json/v2"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"unicode"
)

type metadataProse struct {
	Operation    string        `json:"operation"`
	File         string        `json:"file"`
	SourceSHA256 string        `json:"sourceSHA256"`
	FieldSource  productSource `json:"fieldSource"`
	Pointer      string        `json:"pointer"`
	English      string        `json:"english"`
	TextSHA256   string        `json:"textSHA256"`
}
type productProseFile struct {
	Generator            string                   `json:"generator"`
	SchemaVersion        int                      `json:"schemaVersion"`
	Product              string                   `json:"product"`
	Version              string                   `json:"version"`
	IRSHA256             string                   `json:"irSHA256"`
	CorpusManifestSHA256 string                   `json:"corpusManifestSHA256"`
	Repository           string                   `json:"repository"`
	Revision             string                   `json:"revision"`
	License              string                   `json:"license"`
	MatchedOperations    int                      `json:"matchedOperations"`
	Fields               map[string]metadataProse `json:"fields"`
	Reasons              []map[string]any         `json:"reasons"`
}
type proseSourceManifest struct {
	SchemaVersion       int              `json:"schemaVersion"`
	Repository          string           `json:"repository"`
	Revision            string           `json:"revision"`
	License             string           `json:"license"`
	AdapterVersion      int              `json:"adapterVersion"`
	RequestedOperations int              `json:"requestedOperations"`
	MissingOperations   []map[string]any `json:"missingOperations"`
	Files               []struct {
		File   string `json:"file"`
		URL    string `json:"url"`
		SHA256 string `json:"sha256"`
	} `json:"files"`
}

func loadProductProse(root string, products []productIR) error {
	const corpus = "sources/openapi-meta/prose/"
	manifestPath := corpus + "manifest.json"
	if err := safeOutputPath(root, manifestPath); err != nil {
		return err
	}
	manifestBytes, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(manifestPath)))
	if errors.Is(err, fs.ErrNotExist) {
		for _, p := range products {
			if _, e := os.Stat(filepath.Join(root, "metadata", "prose", p.Product+".json")); !errors.Is(e, fs.ErrNotExist) {
				return errors.New("prose: projection requires source corpus")
			}
		}
		return nil
	}
	if err != nil {
		return err
	}
	var manifest proseSourceManifest
	if err := readJSON(filepath.Join(root, filepath.FromSlash(manifestPath)), &manifest, true); err != nil {
		return err
	}
	if manifest.SchemaVersion != 1 || manifest.AdapterVersion != 1 || manifest.Repository != "https://github.com/aliyun/aliyun-openapi-meta" || manifest.License != "Apache-2.0" || len(manifest.Revision) != 40 {
		return errors.New("prose: invalid source manifest")
	}
	sources := map[string][]byte{}
	pins := map[string]string{}
	for _, pin := range manifest.Files {
		if (pin.File != "LICENSE" && !strings.HasPrefix(pin.File, "canonical/")) || pins[pin.File] != "" || pin.URL != "https://raw.githubusercontent.com/aliyun/aliyun-openapi-meta/"+manifest.Revision+"/"+pin.File {
			return errors.New("prose: unsafe or duplicate source pin")
		}
		if err := safeOutputPath(root, corpus+pin.File); err != nil {
			return err
		}
		b, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(corpus+pin.File)))
		if err != nil {
			return err
		}
		if digest(b) != pin.SHA256 {
			return errors.New("prose: source checksum differs")
		}
		sources[pin.File] = b
		pins[pin.File] = pin.SHA256
	}
	if pins["LICENSE"] == "" {
		return errors.New("prose: missing license")
	}
	if err := filepath.WalkDir(filepath.Join(root, filepath.FromSlash(corpus+"canonical")), func(file string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return errors.New("prose: symlink source")
		}
		if entry.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(filepath.Join(root, filepath.FromSlash(corpus)), file)
		if err != nil {
			return err
		}
		if pins[filepath.ToSlash(rel)] == "" {
			return errors.New("prose: unlisted source")
		}
		return nil
	}); err != nil {
		return err
	}
	for i := range products {
		p := &products[i]
		file := "metadata/prose/" + p.Product + ".json"
		if err := safeOutputPath(root, file); err != nil {
			return err
		}
		var projection productProseFile
		if err := readJSON(filepath.Join(root, filepath.FromSlash(file)), &projection, true); err != nil {
			return err
		}
		ir, err := os.ReadFile(filepath.Join(root, "models", p.Product, "ir.json"))
		if err != nil {
			return err
		}
		if projection.Generator != "darabonba prose" || projection.SchemaVersion != 1 || projection.Product != p.Product || projection.Version != p.Version || projection.IRSHA256 != digest(ir) || projection.CorpusManifestSHA256 != digest(manifestBytes) || projection.Repository != manifest.Repository || projection.Revision != manifest.Revision || projection.License != manifest.License {
			return errors.New("prose: projection provenance differs")
		}
		operations := map[string]productOperation{}
		fields := map[string]productField{}
		for _, op := range p.Operations {
			operations[op.Name] = op
		}
		for _, model := range p.Models {
			for _, field := range append(append([]productField{}, model.InheritedFields...), model.Fields...) {
				fields[model.ID+"#"+field.DSLName] = field
			}
		}
		decoded := map[string]any{}
		for key, entry := range projection.Fields {
			field, ok := fields[key]
			op, opOK := operations[entry.Operation]
			expected := "canonical/" + p.Product + "/" + p.Version + "/" + entry.Operation + ".json"
			if !ok || !opOK || entry.File != expected || entry.SourceSHA256 != pins[entry.File] || entry.FieldSource != field.Source || documentCoverage(field.Documentation).Status == "emitted" || strings.TrimSpace(entry.English) == "" || !englishProse(normalizeProse(entry.English)) || strings.ContainsFunc(entry.English, func(r rune) bool { return unicode.Is(unicode.Han, r) }) || digest([]byte(entry.English)) != entry.TextSHA256 {
				return errors.New("prose: invalid field or text binding")
			}
			reachable := false
			for _, id := range op.ReachableModels {
				reachable = reachable || strings.HasPrefix(key, id+"#")
			}
			if !reachable {
				return errors.New("prose: field not reachable from operation")
			}
			if !strings.HasSuffix(entry.Pointer, "/description_en") && !strings.HasSuffix(entry.Pointer, "/help_en") {
				return errors.New("prose: unsupported source attribute")
			}
			value, loaded := decoded[entry.File]
			if !loaded {
				if err := json.Unmarshal(sources[entry.File], &value); err != nil {
					return err
				}
				decoded[entry.File] = value
			}
			text, err := prosePointer(value, entry.Pointer)
			if err != nil || text != entry.English {
				return errors.New("prose: source text differs")
			}
		}
		p.MetadataProse = projection.Fields
		p.MetadataProseRevision = projection.Revision
	}
	return nil
}

func prosePointer(value any, pointer string) (string, error) {
	if !strings.HasPrefix(pointer, "/") {
		return "", errors.New("prose: invalid pointer")
	}
	for _, part := range strings.Split(pointer[1:], "/") {
		part = strings.ReplaceAll(strings.ReplaceAll(part, "~1", "/"), "~0", "~")
		switch node := value.(type) {
		case map[string]any:
			value = node[part]
		case []any:
			index, err := strconv.Atoi(part)
			if err != nil || index < 0 || index >= len(node) {
				return "", errors.New("prose: invalid pointer index")
			}
			value = node[index]
		default:
			return "", errors.New("prose: invalid pointer node")
		}
	}
	text, ok := value.(string)
	if !ok {
		return "", errors.New("prose: pointer is not text")
	}
	return text, nil
}

func (r *productRenderer) appendMetadataProse(b *bytes.Buffer, key string) bool {
	entry, ok := r.p.MetadataProse[key]
	if !ok {
		return false
	}
	b.WriteString("//\n// Optional official CLI metadata prose (Apache-2.0; informational, not SDK validation):\n//\n")
	for _, line := range strings.Split(normalizeProse(entry.English), "\n") {
		fmt.Fprintf(b, "// %s\n", line)
	}
	fmt.Fprintf(b, "//\n// Source: https://github.com/aliyun/aliyun-openapi-meta/blob/%s/%s\n// JSON pointer: %s\n", r.p.MetadataProseRevision, entry.File, entry.Pointer)
	return true
}
