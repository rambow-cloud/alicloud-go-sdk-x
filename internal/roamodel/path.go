// Package roamodel encodes reviewed ROA request components for generated clients.
package roamodel

import (
	"context"
	"errors"
	"net/url"
	"strings"
	"unicode/utf8"
)

// Path builds decoded and escaped absolute paths from a reviewed template.
// Parameters are single segments. It never modifies parameters or cleans paths.
func Path(ctx context.Context, template string, parameters map[string]string) (string, string, error) {
	if err := ctx.Err(); err != nil {
		return "", "", err
	}
	invalid := errors.New("roamodel: invalid path template or parameters")
	if !strings.HasPrefix(template, "/") || strings.ContainsAny(template, "?#\r\n\t") || !utf8.ValidString(template) {
		return "", "", invalid
	}
	used := map[string]bool{}
	var escaped strings.Builder
	for rest := template; rest != ""; {
		if err := ctx.Err(); err != nil {
			return "", "", err
		}
		start := strings.IndexAny(rest, "{}")
		if start == -1 {
			escaped.WriteString(rest)
			break
		}
		if rest[start] != '{' {
			return "", "", invalid
		}
		end := strings.IndexByte(rest[start+1:], '}')
		if end == -1 {
			return "", "", invalid
		}
		end += start + 1
		name := rest[start+1 : end]
		if name == "" || strings.ContainsAny(name, "{}/") {
			return "", "", invalid
		}
		value, ok := parameters[name]
		if !ok || value == "" || !utf8.ValidString(value) {
			return "", "", invalid
		}
		used[name] = true
		escaped.WriteString(rest[:start])
		escaped.WriteString(strings.ReplaceAll(url.QueryEscape(value), "+", "%20"))
		rest = rest[end+1:]
	}
	if len(used) != len(parameters) {
		return "", "", invalid
	}
	raw := escaped.String()
	decoded, err := url.PathUnescape(raw)
	if err != nil {
		return "", "", invalid
	}
	if err := ctx.Err(); err != nil {
		return "", "", err
	}
	return decoded, raw, nil
}
