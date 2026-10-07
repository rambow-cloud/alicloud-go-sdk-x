// Package rpcmodel implements private snapshots and the lowered RPC query profile.
package rpcmodel

import (
	"context"
	"errors"
	"fmt"
	"math"
	"net/url"
	"reflect"
	"sort"
	"strconv"
	"strings"
)

// Snapshot copies a generated model graph, preserving nil, widths and explicit zeros.
// Callers must not mutate the source concurrently. Cyclic inputs are rejected.
func Snapshot[T any](ctx context.Context, input *T) (*T, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if input == nil {
		return new(T), nil
	}
	v, err := clone(ctx, reflect.ValueOf(input), map[visit]bool{})
	if err != nil {
		return nil, err
	}
	return v.Interface().(*T), nil
}

type visit struct {
	typ     reflect.Type
	pointer uintptr
}

func clone(ctx context.Context, v reflect.Value, active map[visit]bool) (reflect.Value, error) {
	if err := ctx.Err(); err != nil {
		return reflect.Value{}, err
	}
	result := reflect.New(v.Type()).Elem()
	if (v.Kind() == reflect.Pointer || v.Kind() == reflect.Slice || v.Kind() == reflect.Map) && v.IsNil() {
		return result, nil
	}
	if v.Kind() == reflect.Pointer || v.Kind() == reflect.Slice || v.Kind() == reflect.Map {
		pointer := v.Pointer
		if v.Kind() == reflect.Map {
			pointer = func() uintptr { return uintptr(v.UnsafePointer()) }
		}
		key := visit{v.Type(), pointer()}
		if active[key] {
			return reflect.Value{}, errors.New("rpcmodel: cyclic input")
		}
		active[key] = true
		defer delete(active, key)
	}
	switch v.Kind() {
	case reflect.Pointer:
		child, err := clone(ctx, v.Elem(), active)
		if err != nil {
			return reflect.Value{}, err
		}
		result.Set(reflect.New(v.Type().Elem()))
		result.Elem().Set(child)
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			if !result.Field(i).CanSet() {
				return reflect.Value{}, errors.New("rpcmodel: private field")
			}
			child, err := clone(ctx, v.Field(i), active)
			if err != nil {
				return reflect.Value{}, err
			}
			result.Field(i).Set(child)
		}
	case reflect.Slice:
		result.Set(reflect.MakeSlice(v.Type(), v.Len(), v.Len()))
		for i := 0; i < v.Len(); i++ {
			child, err := clone(ctx, v.Index(i), active)
			if err != nil {
				return reflect.Value{}, err
			}
			result.Index(i).Set(child)
		}
	case reflect.Map:
		result.Set(reflect.MakeMapWithSize(v.Type(), v.Len()))
		iter := v.MapRange()
		for iter.Next() {
			child, err := clone(ctx, iter.Value(), active)
			if err != nil {
				return reflect.Value{}, err
			}
			result.SetMapIndex(iter.Key(), child)
		}
	case reflect.String, reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Float32, reflect.Float64:
		result.Set(v)
	default:
		return reflect.Value{}, fmt.Errorf("rpcmodel: unsupported kind %s", v.Kind())
	}
	return result, nil
}

// Query flattens exact JSON member names and one-based repeated indexes.
// It omits nil pointers, never rewrites string payloads, and rejects collisions.
func Query(ctx context.Context, model any) (url.Values, error) {
	result := url.Values{}
	if err := flatten(ctx, reflect.ValueOf(model), "", result, 0); err != nil {
		return nil, err
	}
	return result, nil
}

func flatten(ctx context.Context, v reflect.Value, prefix string, result url.Values, depth int) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if depth > 128 {
		return errors.New("rpcmodel: query nesting exceeds limit")
	}
	if !v.IsValid() {
		return nil
	}
	if v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return nil
		}
		return flatten(ctx, v.Elem(), prefix, result, depth+1)
	}
	join := func(name string) string {
		if prefix == "" {
			return name
		}
		return prefix + "." + name
	}
	switch v.Kind() {
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			name := strings.Split(v.Type().Field(i).Tag.Get("json"), ",")[0]
			if name == "-" {
				continue
			}
			if name == "" {
				return errors.New("rpcmodel: missing wire name")
			}
			if err := flatten(ctx, v.Field(i), join(name), result, depth+1); err != nil {
				return err
			}
		}
		return nil
	case reflect.Slice:
		for i := 0; i < v.Len(); i++ {
			if err := flatten(ctx, v.Index(i), join(strconv.Itoa(i+1)), result, depth+1); err != nil {
				return err
			}
		}
		return nil
	case reflect.Map:
		if v.Type().Key().Kind() != reflect.String {
			return errors.New("rpcmodel: non-string query map key")
		}
		keys := v.MapKeys()
		sort.Slice(keys, func(i, j int) bool { return keys[i].String() < keys[j].String() })
		for _, key := range keys {
			if err := flatten(ctx, v.MapIndex(key), join(key.String()), result, depth+1); err != nil {
				return err
			}
		}
		return nil
	}
	var value string
	switch v.Kind() {
	case reflect.String:
		value = v.String()
	case reflect.Bool:
		value = strconv.FormatBool(v.Bool())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		value = strconv.FormatInt(v.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		value = strconv.FormatUint(v.Uint(), 10)
	case reflect.Float32, reflect.Float64:
		if math.IsNaN(v.Float()) || math.IsInf(v.Float(), 0) {
			return errors.New("rpcmodel: non-finite number")
		}
		value = strconv.FormatFloat(v.Float(), 'g', -1, v.Type().Bits())
	default:
		return fmt.Errorf("rpcmodel: unsupported query kind %s", v.Kind())
	}
	if prefix == "" {
		return errors.New("rpcmodel: scalar query root")
	}
	if _, exists := result[prefix]; exists {
		return errors.New("rpcmodel: duplicate query key")
	}
	result.Set(prefix, value)
	return nil
}
