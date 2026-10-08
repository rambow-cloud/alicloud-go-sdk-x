package signing

import (
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestAnonymousRPCIndependentWireAndSourceIsolation(t *testing.T) {
	r, _ := http.NewRequest("POST", "https://example.invalid/?OIDCToken=a%2Bb%26c%3D&securityTOKEN=source&Signature=source&AccessKeyId=source&Action=wrong", nil)
	r.Header["authorization"] = []string{"source"}
	r.Header["x-acs-security-token"] = []string{"source"}
	r.Header.Set("Traceparent", "keep")
	now := time.Date(2026, 10, 8, 12, 30, 0, 0, time.FixedZone("offset", 8*3600))
	if err := PrepareAnonymousRPC(r, "AssumeRoleWithOIDC", "2015-04-01", now, "fixed-nonce"); err != nil {
		t.Fatal(err)
	}
	want := "Action=AssumeRoleWithOIDC&Format=json&OIDCToken=a%2Bb%26c%3D&SignatureNonce=fixed-nonce&Timestamp=2026-10-08T04%3A30%3A00Z&Version=2015-04-01"
	if r.URL.RawQuery != want || len(r.Header) != 3 || r.Header.Get("Traceparent") != "keep" || r.Header.Get("X-Acs-Action") != "AssumeRoleWithOIDC" || r.Header.Get("X-Acs-Version") != "2015-04-01" {
		t.Fatal("anonymous framing mismatch")
	}
}

func TestAnonymousRPCRejectsUnsupportedRequest(t *testing.T) {
	if PrepareAnonymousRPC(nil, "a", "v", time.Now(), "n") == nil {
		t.Fatal("nil accepted")
	}
	for _, address := range []string{"http://example.invalid/", "https://u:p@example.invalid/", "https://example.invalid/other", "https://example.invalid/?OIDCToken=%zz-sensitive"} {
		r, _ := http.NewRequest("POST", address, nil)
		err := PrepareAnonymousRPC(r, "a", "v", time.Now(), "n")
		if err == nil || strings.Contains(err.Error(), "sensitive") {
			t.Fatal("invalid request accepted or exposed")
		}
	}
	r, _ := http.NewRequest("GET", "https://example.invalid/", nil)
	if PrepareAnonymousRPC(r, "a", "v", time.Now(), "n") == nil {
		t.Fatal("method accepted")
	}
}
