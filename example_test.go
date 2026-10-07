package alicloud_test

import (
	"errors"
	"fmt"

	alicloud "github.com/rambow-cloud/alicloud-go-sdk-x"
)

func ExampleAPIError() {
	// A future service call will return this error through its error result.
	err := fmt.Errorf("describe instances: %w", &alicloud.APIError{
		Code: "Throttling", RequestID: "example-request", HTTPStatusCode: 429,
	})
	var apiErr *alicloud.APIError
	if errors.As(err, &apiErr) {
		fmt.Println(apiErr.Code, apiErr.RequestID, apiErr.HTTPStatusCode)
	}
	// Output: Throttling example-request 429
}
