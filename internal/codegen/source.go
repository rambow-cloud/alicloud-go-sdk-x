package codegen

import (
	"encoding/json/v2"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

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
