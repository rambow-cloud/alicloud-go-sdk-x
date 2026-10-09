package roamodel

import (
	"context"
	"encoding/json/v2"
	"errors"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"unicode/utf8"

	alicloud "github.com/rambow-cloud/alicloud-go-sdk-x"
)

// Request encodes reviewed generated ROA facades without modifying input.
// Tags select exact path/query/header/JSON body locations; unknown shapes fail.
func Request(ctx context.Context, input any, method, template string) (alicloud.Request, error) {
	if err := ctx.Err(); err != nil {
		return alicloud.Request{}, err
	}
	invalid := errors.New("roamodel: invalid request binding or value")
	value := reflect.ValueOf(input)
	if value.Kind() != reflect.Pointer || value.IsNil() || value.Elem().Kind() != reflect.Struct {
		return alicloud.Request{}, invalid
	}
	value = value.Elem()
	parameters := map[string]string{}
	query := url.Values{}
	headers := http.Header{}
	members := map[string]any{}
	var body any
	bodyRoot, bodyFields := false, false
	var binaryBody []byte
	binaryPresent := false
	for index := 0; index < value.NumField(); index++ {
		if err := ctx.Err(); err != nil {
			return alicloud.Request{}, err
		}
		field, definition := value.Field(index), value.Type().Field(index)
		location := definition.Tag.Get("roa")
		wire := strings.Split(definition.Tag.Get("json"), ",")[0]
		if !field.CanInterface() || wire == "" || wire == "-" {
			return alicloud.Request{}, invalid
		}
		switch location {
		case "path", "headers", "header-model", "query", "body", "body-member", "binary-body":
		default:
			return alicloud.Request{}, invalid
		}
		absent := false
		switch field.Kind() {
		case reflect.Pointer, reflect.Map, reflect.Slice, reflect.Interface:
			absent = field.IsNil()
		}
		if location == "body-member" {
			bodyFields = true
		}
		if absent {
			continue
		}
		for field.Kind() == reflect.Pointer || field.Kind() == reflect.Interface {
			if field.IsNil() {
				absent = true
				break
			}
			field = field.Elem()
		}
		if absent {
			continue
		}
		switch location {
		case "path":
			if field.Kind() != reflect.String || parameters[wire] != "" {
				return alicloud.Request{}, invalid
			}
			parameters[wire] = field.String()
		case "headers":
			if err := appendHeaderMap(headers, field); err != nil {
				return alicloud.Request{}, err
			}
		case "header-model":
			if err := appendHeaderModel(headers, field); err != nil {
				return alicloud.Request{}, err
			}
		case "binary-body":
			if bodyRoot || bodyFields || binaryPresent || field.Type() != reflect.TypeFor[[]byte]() {
				return alicloud.Request{}, invalid
			}
			if field.Len() > 8<<20 {
				return alicloud.Request{}, errors.New("roamodel: request body exceeds byte limit")
			}
			binaryBody = append([]byte{}, field.Bytes()...)
			binaryPresent = true
		case "query":
			if definition.Tag.Get("rpc") != "" && definition.Tag.Get("rpc") != "json" {
				return alicloud.Request{}, invalid
			}
			if definition.Tag.Get("rpc") == "json" {
				bytes, err := json.Marshal(field.Interface())
				if err != nil {
					return alicloud.Request{}, err
				}
				query.Set(wire, string(bytes))
			} else {
				text, ok := queryScalar(field)
				if !ok {
					return alicloud.Request{}, invalid
				}
				query.Set(wire, text)
			}
		case "body":
			if binaryPresent || bodyRoot || bodyFields || field.Kind() != reflect.Struct && field.Kind() != reflect.Map {
				return alicloud.Request{}, invalid
			}
			bodyRoot = true
			body = field.Interface()
		case "body-member":
			if binaryPresent || bodyRoot {
				return alicloud.Request{}, invalid
			}
			members[wire] = field.Interface()
		default:
			return alicloud.Request{}, invalid
		}
	}
	decoded, escaped, err := Path(ctx, template, parameters)
	if err != nil {
		return alicloud.Request{}, err
	}
	request := alicloud.Request{Method: method, Path: decoded, RawPath: escaped, Query: query, Header: headers}
	if binaryPresent {
		request.Body = binaryBody
		headers.Set("Content-Type", "application/octet-stream")
	}
	if bodyFields {
		body = members
	}
	if bodyRoot || bodyFields {
		bytes, err := json.Marshal(body)
		if err != nil {
			return alicloud.Request{}, err
		}
		if len(bytes) > 8<<20 {
			return alicloud.Request{}, errors.New("roamodel: request body exceeds byte limit")
		}
		request.Body = bytes
		headers.Set("Content-Type", "application/json")
	}
	if err := ctx.Err(); err != nil {
		return alicloud.Request{}, err
	}
	return request, nil
}

func setHeader(headers http.Header, name, text string, override bool) error {
	canonical := http.CanonicalHeaderKey(name)
	if !headerName(name) || !override && headers[canonical] != nil || managedHeader(strings.ToLower(name)) || !utf8.ValidString(text) || strings.ContainsFunc(text, func(r rune) bool { return r < 32 && r != '\t' || r == 127 }) {
		return errors.New("roamodel: invalid request header")
	}
	headers[canonical] = []string{text}
	return nil
}

func appendHeaderMap(headers http.Header, field reflect.Value) error {
	if field.Kind() != reflect.Map || field.Type().Key().Kind() != reflect.String || field.Type().Elem().Kind() != reflect.String {
		return errors.New("roamodel: invalid header map")
	}
	iter := field.MapRange()
	for iter.Next() {
		if err := setHeader(headers, iter.Key().String(), iter.Value().String(), false); err != nil {
			return err
		}
	}
	return nil
}

func appendHeaderModel(headers http.Header, value reflect.Value) error {
	if value.Kind() != reflect.Struct {
		return errors.New("roamodel: invalid header model")
	}
	common := 0
	for _, location := range []string{"headers", "header"} {
		for i := 0; i < value.NumField(); i++ {
			field, definition := value.Field(i), value.Type().Field(i)
			tag := definition.Tag.Get("roa")
			if !field.CanInterface() || tag != "headers" && tag != "header" {
				return errors.New("roamodel: invalid header model field")
			}
			if tag != location {
				continue
			}
			if tag == "headers" {
				common++
				if err := appendHeaderMap(headers, field); err != nil {
					return err
				}
				continue
			}
			if field.Kind() != reflect.Pointer || field.Type().Elem().Kind() != reflect.String {
				return errors.New("roamodel: invalid string header field")
			}
			if field.IsNil() {
				continue
			}
			wire := strings.Split(definition.Tag.Get("json"), ",")[0]
			if err := setHeader(headers, wire, field.Elem().String(), true); err != nil {
				return err
			}
		}
	}
	if common != 1 {
		return errors.New("roamodel: incomplete header model")
	}
	return nil
}

func queryScalar(v reflect.Value) (string, bool) {
	switch v.Kind() {
	case reflect.String:
		if !utf8.ValidString(v.String()) {
			return "", false
		}
		return v.String(), true
	case reflect.Bool:
		return strconv.FormatBool(v.Bool()), true
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(v.Int(), 10), true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(v.Uint(), 10), true
	case reflect.Float32, reflect.Float64:
		return strconv.FormatFloat(v.Float(), 'g', -1, v.Type().Bits()), true
	}
	return "", false
}

func headerName(name string) bool {
	if name == "" {
		return false
	}
	for _, c := range name {
		if !(c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || strings.ContainsRune("!#$%&'*+-.^_`|~", c)) {
			return false
		}
	}
	return true
}

func managedHeader(name string) bool {
	switch name {
	case "authorization", "host", "content-type", "content-length", "transfer-encoding", "x-acs-action", "x-acs-version", "x-acs-date", "x-acs-signature-nonce", "x-acs-security-token", "x-acs-content-sha256":
		return true
	}
	return false
}
