package roamodel

import (
	"context"
	"errors"
	"io"
	"net/http"
	"reflect"
	"strings"
)

// DecodeStream binds a reviewed binary response facade without reading its body.
// Native headers use lowercase names and the first value, matching the reviewed
// Tea scalar map representation. Retained maps are copied; caller owns the body.
func DecodeStream(ctx context.Context, response *http.Response, output any) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	invalid := errors.New("roamodel: invalid binary output binding")
	value := reflect.ValueOf(output)
	if response == nil || response.Body == nil || value.Kind() != reflect.Pointer || value.IsNil() || value.Elem().Kind() != reflect.Struct {
		return invalid
	}
	value = value.Elem()
	seen := map[string]bool{}
	for i := 0; i < value.NumField(); i++ {
		field, definition := value.Field(i), value.Type().Field(i)
		wire := strings.Split(definition.Tag.Get("json"), ",")[0]
		if wire == "-" {
			continue
		}
		if !field.CanSet() || seen[wire] {
			return invalid
		}
		seen[wire] = true
		switch wire {
		case "body":
			if field.Type() != reflect.TypeFor[io.ReadCloser]() {
				return invalid
			}
			field.Set(reflect.ValueOf(response.Body))
		case "headers":
			if field.Type() != reflect.TypeFor[map[string]string]() {
				return invalid
			}
			headers := map[string]string{}
			for name, values := range response.Header {
				if len(values) > 0 {
					headers[strings.ToLower(name)] = values[0]
				}
			}
			field.Set(reflect.ValueOf(headers))
		case "statusCode":
			if field.Type() != reflect.TypeFor[*int32]() {
				return invalid
			}
			status := int32(response.StatusCode)
			field.Set(reflect.ValueOf(&status))
		default:
			return invalid
		}
	}
	if len(seen) != 3 {
		return invalid
	}
	return ctx.Err()
}
