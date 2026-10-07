package sts_test

import (
	"encoding/json/v2"
	"fmt"
	"strings"
	"testing"

	"github.com/rambow-cloud/alicloud-go-sdk-x/service/sts"
)

func TestSensitiveFormattingDoesNotChangeExplicitJSON(t *testing.T) {
	secret := "secret-not-for-fmt"
	token := "token-not-for-fmt"
	credentials := sts.AssumeRoleOutputCredentials{AccessKeySecret: &secret, SecurityToken: &token}
	out := sts.AssumeRoleOutput{Credentials: &credentials}
	response := sts.AssumeRoleResponse{Body: &out}
	for _, value := range []any{credentials, &credentials, out, &out, response, &response, sts.AssumeRoleInput{Policy: &secret}} {
		for _, format := range []string{"%v", "%+v", "%#v"} {
			text := fmt.Sprintf(format, value)
			if strings.Contains(text, secret) || strings.Contains(text, token) || !strings.Contains(text, "redacted") {
				t.Fatal("sensitive fmt leaked", text)
			}
		}
	}
	data, err := json.Marshal(out)
	if err != nil || !strings.Contains(string(data), secret) || !strings.Contains(string(data), token) {
		t.Fatal("explicit JSON was modified", err)
	}
}
