// Package codegen implements the repository's offline RPC generation profile.
package codegen

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json/jsontext"
	"encoding/json/v2"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

const maxMetadataBytes = 16 << 20

var component = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9-]*$`)

// Source records the immutable provenance of one protocol extraction.
type Source struct {
	Operation   string `json:"operation"`
	File        string `json:"file"`
	URL         string `json:"url"`
	RetrievedAt string `json:"retrievedAt"`
	RawSHA256   string `json:"rawSHA256"`
	SHA256      string `json:"sha256"`
}

// Manifest identifies a pinned metadata product and its supported protocol.
type Manifest struct {
	SchemaVersion int        `json:"schemaVersion"`
	Product       string     `json:"product"`
	Version       string     `json:"version"`
	Style         string     `json:"style"`
	Package       string     `json:"package"`
	Service       string     `json:"service"`
	Sources       []Source   `json:"sources"`
	DSL           *DSLSource `json:"dsl,omitempty"`
}

// Schema contains supported wire structure and evidence of unsupported shapes.
type Schema struct {
	Type                 string                    `json:"type"`
	Format               string                    `json:"format"`
	Required             bool                      `json:"required"`
	Properties           map[string]*Schema        `json:"properties"`
	Items                *Schema                   `json:"items"`
	Ref                  string                    `json:"$ref"`
	OneOf                []jsontext.Value          `json:"oneOf"`
	AllOf                []jsontext.Value          `json:"allOf"`
	AnyOf                []jsontext.Value          `json:"anyOf"`
	AdditionalProperties jsontext.Value            `json:"additionalProperties"`
	Minimum              jsontext.Value            `json:"minimum"`
	Maximum              jsontext.Value            `json:"maximum"`
	MaxItems             int                       `json:"maxItems"`
	Extra                map[string]jsontext.Value `json:",unknown"`
}

// Parameter describes a metadata query parameter.
type Parameter struct {
	Name   string  `json:"name"`
	In     string  `json:"in"`
	Style  string  `json:"style"`
	Schema *Schema `json:"schema"`
}

// Snapshot is the protocol-only subset of an official operation document.
type Snapshot struct {
	Methods    []string    `json:"methods"`
	Schemes    []string    `json:"schemes"`
	Path       string      `json:"path"`
	Parameters []Parameter `json:"parameters"`
	Responses  map[string]struct {
		Schema *Schema `json:"schema"`
	} `json:"responses"`
	Components struct {
		Schemas map[string]*Schema `json:"schemas"`
	} `json:"components"`
}

// MetadataURL returns the official unauthenticated operation metadata URL.
func MetadataURL(product, version, operation string) (string, error) {
	for _, value := range []string{product, version, operation} {
		if !component.MatchString(value) {
			return "", errors.New("codegen: invalid metadata URL component")
		}
	}
	return fmt.Sprintf("https://api.aliyun.com/meta/v1/products/%s/versions/%s/apis/%s/api.json?language=EN_US", product, version, operation), nil
}

// Extract strips documentation and examples, retaining wire facts only. Unknown
// structural keys are preserved so future profiles can inspect the same snapshot.
func Extract(raw []byte) ([]byte, error) {
	if len(raw) > maxMetadataBytes {
		return nil, errors.New("codegen: metadata too large")
	}
	var document map[string]any
	if err := json.Unmarshal(raw, &document); err != nil {
		return nil, fmt.Errorf("decode metadata: %w", err)
	}
	if document == nil {
		return nil, errors.New("codegen: metadata must be an object")
	}
	selected := map[string]any{}
	for _, key := range []string{"methods", "schemes", "path", "parameters", "responses"} {
		if value, ok := document[key]; ok {
			selected[key] = value
		}
	}
	// Empty components are excluded to preserve existing snapshot bytes.
	if components, ok := document["components"].(map[string]any); ok {
		if schemas, ok := components["schemas"].(map[string]any); ok && len(schemas) > 0 {
			selected["components"] = map[string]any{"schemas": schemas}
		}
	}
	// Only schema annotations are removed. Wire property names may themselves be
	// "Description", "example" or "title" and must remain intact.
	var pruneSchema func(map[string]any)
	pruneSchema = func(s map[string]any) {
		for _, key := range []string{"description", "title", "example", "examples", "enumValueTitles", "docRequired"} {
			delete(s, key)
		}
		if props, ok := s["properties"].(map[string]any); ok {
			for _, value := range props {
				if child, ok := value.(map[string]any); ok {
					pruneSchema(child)
				}
			}
		}
		for _, key := range []string{"items", "additionalProperties"} {
			if child, ok := s[key].(map[string]any); ok {
				pruneSchema(child)
			}
		}
		for _, key := range []string{"oneOf", "allOf", "anyOf"} {
			if children, ok := s[key].([]any); ok {
				for _, value := range children {
					if child, ok := value.(map[string]any); ok {
						pruneSchema(child)
					}
				}
			}
		}
	}
	if parameters, ok := selected["parameters"].([]any); ok {
		for _, value := range parameters {
			if p, ok := value.(map[string]any); ok {
				for _, key := range []string{"description", "title", "example", "examples"} {
					delete(p, key)
				}
				if s, ok := p["schema"].(map[string]any); ok {
					pruneSchema(s)
				}
			}
		}
	}
	if responses, ok := selected["responses"].(map[string]any); ok {
		for _, value := range responses {
			if response, ok := value.(map[string]any); ok {
				delete(response, "description")
				delete(response, "headers")
				if s, ok := response["schema"].(map[string]any); ok {
					pruneSchema(s)
				}
			}
		}
	}
	if components, ok := selected["components"].(map[string]any); ok {
		if schemas, ok := components["schemas"].(map[string]any); ok {
			for _, value := range schemas {
				if schema, ok := value.(map[string]any); ok {
					pruneSchema(schema)
				}
			}
		}
	}
	data, err := json.Marshal(selected, json.Deterministic(true))
	if err != nil {
		return nil, err
	}
	return append(data, '\n'), nil
}

// Fetch performs one bounded metadata GET, without retries or credentials. The
// caller owns cancellation; redirects are rejected even by injected transports.
func Fetch(ctx context.Context, client *http.Client, product, version, operation string, now time.Time) (Source, []byte, error) {
	uri, err := MetadataURL(product, version, operation)
	if err != nil {
		return Source{}, nil, err
	}
	if client == nil {
		client = &http.Client{}
	}
	copyClient := *client
	copyClient.Timeout = 30 * time.Second
	copyClient.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return Source{}, nil, err
	}
	response, err := copyClient.Do(request)
	if err != nil {
		return Source{}, nil, fmt.Errorf("fetch metadata: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return Source{}, nil, fmt.Errorf("codegen: metadata HTTP status %d", response.StatusCode)
	}
	raw, err := io.ReadAll(io.LimitReader(response.Body, maxMetadataBytes+1))
	if err != nil {
		return Source{}, nil, err
	}
	data, err := Extract(raw)
	if err != nil {
		return Source{}, nil, err
	}
	var snapshot Snapshot
	if err := json.Unmarshal(data, &snapshot); err != nil {
		return Source{}, nil, err
	}
	if err := checkProtocol(snapshot); err != nil {
		return Source{}, nil, err
	}
	return Source{Operation: operation, File: operation + ".json", URL: uri, RetrievedAt: now.UTC().Format(time.RFC3339), RawSHA256: digest(raw), SHA256: digest(data)}, data, nil
}

func digest(data []byte) string { sum := sha256.Sum256(data); return hex.EncodeToString(sum[:]) }

func readJSON(path string, output any, strict bool) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if len(data) > maxMetadataBytes {
		return fmt.Errorf("codegen: input exceeds limit: %s", filepath.Base(path))
	}
	if strict {
		err = json.Unmarshal(data, output, json.RejectUnknownMembers(true))
	} else {
		err = json.Unmarshal(data, output)
	}
	if err != nil {
		return fmt.Errorf("decode %s: %w", filepath.Base(path), err)
	}
	return nil
}

func validDigest(value string) bool {
	data, err := hex.DecodeString(value)
	return err == nil && len(data) == sha256.Size && value == strings.ToLower(value)
}

func checkProtocol(s Snapshot) error {
	contains := func(values []string, target string) bool {
		for _, value := range values {
			if strings.ToLower(value) == target {
				return true
			}
		}
		return false
	}
	if !contains(s.Methods, "post") || !contains(s.Schemes, "https") || (s.Path != "" && s.Path != "/") {
		return errors.New("codegen: unsupported RPC method, scheme or path")
	}
	root, err := s.resolve(s.Responses["200"].Schema)
	if err != nil {
		return err
	}
	if root.Type != "object" {
		return errors.New("codegen: JSON 200 object response required")
	}
	return nil
}

func verifySource(m Manifest, source Source, data []byte) error {
	expected, err := MetadataURL(m.Product, m.Version, source.Operation)
	if err != nil {
		return err
	}
	if source.URL != expected {
		return errors.New("codegen: source URL does not match pinned product/version/operation")
	}
	if _, err := url.ParseRequestURI(source.URL); err != nil {
		return err
	}
	if _, err := time.Parse(time.RFC3339, source.RetrievedAt); err != nil {
		return errors.New("codegen: invalid retrieval time")
	}
	if !validDigest(source.RawSHA256) || !validDigest(source.SHA256) || source.SHA256 != digest(data) {
		return errors.New("codegen: snapshot checksum or provenance mismatch")
	}
	if !bytes.HasSuffix(data, []byte("\n")) {
		return errors.New("codegen: snapshot must end with newline")
	}
	return nil
}
