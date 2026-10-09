package codegen

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json/v2"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
)

const module = "github.com/rambow-cloud/alicloud-go-sdk-x"
const maxMetadataBytes = 16 << 20

var exportedName = regexp.MustCompile("^[A-Z][A-Za-z0-9]*$")
var packageName = regexp.MustCompile("^[a-z][a-z0-9]*$")

func quote(s string) string     { return strconv.Quote(s) }
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

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
