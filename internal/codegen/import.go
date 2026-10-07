package codegen

import (
	"context"
	"encoding/json/jsontext"
	"encoding/json/v2"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// Import writes protocol snapshots and a manifest after collecting all sources.
// An optional rawDir imports already downloaded documents without network access;
// otherwise Fetch is used. Existing metadata is replaced only by this explicit command.
func Import(ctx context.Context, client *http.Client, m Manifest, operations []string, rawDir, out string) error {
	if m.SchemaVersion != 1 || m.Style != "RPC" || !packageName.MatchString(m.Package) || m.Service != m.Package || len(operations) == 0 {
		return errors.New("codegen: invalid import profile")
	}
	files := map[string][]byte{}
	m.Sources = nil
	for _, operation := range operations {
		if err := ctx.Err(); err != nil {
			return err
		}
		uri, err := MetadataURL(m.Product, m.Version, operation)
		if err != nil {
			return err
		}
		if !exportedName.MatchString(operation) {
			return errors.New("codegen: invalid operation name")
		}
		if _, exists := files[operation+".json"]; exists {
			return errors.New("codegen: duplicate import operation")
		}
		var source Source
		var data []byte
		if rawDir != "" {
			path := filepath.Join(rawDir, operation+".json")
			raw, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			data, err = Extract(raw)
			if err != nil {
				return err
			}
			info, err := os.Stat(path)
			if err != nil {
				return err
			}
			source = Source{Operation: operation, File: operation + ".json", URL: uri, RetrievedAt: info.ModTime().UTC().Format(time.RFC3339), RawSHA256: digest(raw), SHA256: digest(data)}
		} else {
			source, data, err = Fetch(ctx, client, m.Product, m.Version, operation, time.Now())
			if err != nil {
				return err
			}
		}
		var snapshot Snapshot
		if err := json.Unmarshal(data, &snapshot); err != nil {
			return err
		}
		if err := checkProtocol(snapshot); err != nil {
			return err
		}
		files[source.File] = data
		m.Sources = append(m.Sources, source)
	}
	sort.Slice(m.Sources, func(i, j int) bool { return m.Sources[i].Operation < m.Sources[j].Operation })
	data, err := json.Marshal(m, json.Deterministic(true), jsontext.WithIndent("  "))
	if err != nil {
		return err
	}
	files["manifest.json"] = append(data, '\n')
	if err := os.MkdirAll(out, 0755); err != nil {
		return err
	}
	for _, name := range sortedKeys(files) {
		if err := os.WriteFile(filepath.Join(out, name), files[name], 0644); err != nil {
			return fmt.Errorf("write metadata: %w", err)
		}
	}
	return nil
}

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
