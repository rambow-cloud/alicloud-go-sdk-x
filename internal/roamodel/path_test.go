package roamodel

import (
	"context"
	"errors"
	"testing"
)

func TestPathEncodingAndExactParameterSet(t *testing.T) {
	path, raw, err := Path(context.Background(), "/functions/{name}/{name}", map[string]string{"name": "a/b?c#d% e+中文"})
	if err != nil || path != "/functions/a/b?c#d% e+中文/a/b?c#d% e+中文" || raw != "/functions/a%2Fb%3Fc%23d%25%20e%2B%E4%B8%AD%E6%96%87/a%2Fb%3Fc%23d%25%20e%2B%E4%B8%AD%E6%96%87" {
		t.Fatal(path, raw, err)
	}
	for _, test := range []struct {
		template string
		values   map[string]string
	}{
		{"relative/{name}", map[string]string{"name": "a"}},
		{"/functions/{name", map[string]string{"name": "a"}},
		{"/functions/name}", nil},
		{"/functions/{{name}", map[string]string{"name": "a"}},
		{"/functions/{name}", nil},
		{"/functions/{name}", map[string]string{"name": ""}},
		{"/functions/{name}", map[string]string{"name": "a", "extra": "b"}},
		{"/functions/{name}", map[string]string{"name": "\xff"}},
		{"/functions/{name}?query=literal", map[string]string{"name": "a"}},
		{"/functions/%ZZ", nil},
	} {
		if _, _, err := Path(context.Background(), test.template, test.values); err == nil {
			t.Fatal("accepted invalid path/parameters")
		}
	}
}

func TestPathCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, _, err := Path(ctx, "/functions/{name}", map[string]string{"name": "a"}); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
}
