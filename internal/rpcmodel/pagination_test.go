package rpcmodel

import (
	"strings"
	"testing"
)

func TestPaginationIntegerExactNativeWidthsAndSafeErrors(t *testing.T) {
	for _, v := range []any{int32(20), int64(20), "20", "020"} {
		n, err := PaginationInteger(v)
		if err != nil || n != 20 {
			t.Fatalf("unexpected integer: %d %v", n, err)
		}
	}
	for _, v := range []any{"", "private-cursor", "1.0", "+20", " 20", "9223372036854775808", float64(20), nil} {
		_, err := PaginationInteger(v)
		if err == nil || strings.Contains(err.Error(), "private-cursor") {
			t.Fatal("invalid value accepted or leaked")
		}
	}
}
