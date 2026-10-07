package alicloud_test

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	alicloud "github.com/rambow-cloud/alicloud-go-sdk-x"
)

func TestAPIErrorWrappedAndRedacted(t *testing.T) {
	want := &alicloud.APIError{Code: "InvalidParameter", Message: "sensitive-input", RequestID: "req-123", HTTPStatusCode: 400}
	err := fmt.Errorf("operation failed: %w", want)
	var got *alicloud.APIError
	if !errors.As(err, &got) || got != want {
		t.Fatal("wrapped service error cannot be extracted")
	}
	if strings.Contains(err.Error(), want.Message) {
		t.Fatal("default error string exposes service message")
	}
	for _, field := range []string{want.Code, want.RequestID, "400"} {
		if !strings.Contains(err.Error(), field) {
			t.Errorf("default error string omits %q", field)
		}
	}
}
