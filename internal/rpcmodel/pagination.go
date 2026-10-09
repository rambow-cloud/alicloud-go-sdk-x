package rpcmodel

import (
	"errors"
	"strconv"
)

// PaginationInteger decodes an explicitly reviewed native numeric cursor field.
// Only int32, int64 and decimal strings are accepted; errors omit raw values.
// Callers must validate operation-specific bounds before narrowing conversions.
func PaginationInteger(value any) (int64, error) {
	switch v := value.(type) {
	case int32:
		return int64(v), nil
	case int64:
		return v, nil
	case string:
		if v == "" {
			return 0, errors.New("invalid native pagination integer")
		}
		for i, c := range v {
			if c < '0' || c > '9' {
				if i == 0 && c == '-' {
					continue
				}
				return 0, errors.New("invalid native pagination integer")
			}
		}
		n, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return 0, errors.New("invalid native pagination integer")
		}
		return n, nil
	default:
		return 0, errors.New("unsupported native pagination integer type")
	}
}
