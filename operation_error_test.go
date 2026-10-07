package alicloud_test

import (
	"context"
	"errors"
	"fmt"
	alicloud "github.com/rambow-cloud/alicloud-go-sdk-x"
	"strings"
	"testing"
)

func TestOperationErrorInspectionAndRedaction(t *testing.T) {
	api := &alicloud.APIError{Code: "InvalidParameter", Message: "sensitive-query", RequestID: "request", HTTPStatusCode: 400}
	wrapped := &alicloud.OperationError{Service: "ecs", Operation: "DescribeInstances", Err: api, Metadata: alicloud.Metadata{Attempts: 2}}
	var got *alicloud.APIError
	if !errors.As(wrapped, &got) || got != api {
		t.Fatal("cause lost")
	}
	for _, err := range []error{wrapped, api, &alicloud.OperationError{Err: fmt.Errorf("sensitive-url: %w", context.Canceled)}} {
		if strings.Contains(err.Error(), "sensitive") {
			t.Fatal("sensitive default formatting")
		}
	}
	if !errors.Is(&alicloud.OperationError{Err: context.Canceled}, context.Canceled) {
		t.Fatal("cancellation lost")
	}
}
