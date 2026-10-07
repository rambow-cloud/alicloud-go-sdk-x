package signing

import (
	"github.com/rambow-cloud/alicloud-go-sdk-x/credentials"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestOfficialFixedVector(t *testing.T) {
	r, _ := http.NewRequest("POST", "https://ecs.cn-shanghai.aliyuncs.com/?RegionId=cn-shanghai&ImageId=win2019_1809_x64_dtc_zh-cn_40G_alibase_20230811.vhd", nil)
	now, _ := time.Parse(time.RFC3339, "2023-10-26T10:22:32Z")
	err := Sign(r, nil, credentials.Credentials{AccessKeyID: "YourAccessKeyId", AccessKeySecret: "YourAccessKeySecret"}, "RunInstances", "2014-05-26", now, "3156853299f313e23d1673dc12e1703d")
	if err != nil {
		t.Fatal(err)
	}
	expected := "ACS3-HMAC-SHA256 Credential=YourAccessKeyId,SignedHeaders=host;x-acs-action;x-acs-content-sha256;x-acs-date;x-acs-signature-nonce;x-acs-version,Signature=06563a9e1b43f5dfe96b81484da74bceab24a1d853912eee15083a6f0f3283c0"
	if r.Header.Get("Authorization") != expected {
		t.Fatalf("official vector mismatch: %s", r.Header.Get("Authorization"))
	}
}
func TestEscapingBodyAndToken(t *testing.T) {
	r, _ := http.NewRequest("POST", "https://example.invalid/api/a%2Fb/%E4%B8%AD%20%E6%96%87?empty&x=a+b&x=%2B", nil)
	r.Header.Set("Content-Type", " application/json ")
	r.Header.Set("X-Acs-Custom", " custom ")
	c := credentials.Credentials{AccessKeyID: "test", AccessKeySecret: "secret", SecurityToken: "token"}
	err := Sign(r, []byte(`{"x":1}`), c, "Action", "Version", time.Unix(0, 0), "nonce")
	if err != nil {
		t.Fatal(err)
	}
	if r.URL.EscapedPath() != "/api/a%2Fb/%E4%B8%AD%20%E6%96%87" || r.URL.RawQuery != "empty=&x=%2B&x=a%20b" {
		t.Fatal(r.URL)
	}
	auth := r.Header.Get("Authorization")
	if !strings.Contains(auth, "content-type;host") || !strings.Contains(auth, "x-acs-custom") || !strings.Contains(auth, "x-acs-security-token") {
		t.Fatal(auth)
	}
	before := auth
	c.SecurityToken = ""
	_ = Sign(r, []byte(`{"x":2}`), c, "Action", "Version", time.Unix(0, 0), "nonce")
	if r.Header.Get("X-Acs-Security-Token") != "" || r.Header.Get("Authorization") == before {
		t.Fatal("body/token not signed")
	}
	if got := CanonicalQuery(url.Values{"a": {"~ *+"}, "中": {"文"}}); got != "a=~%20%2A%2B&%E4%B8%AD=%E6%96%87" {
		t.Fatal(got)
	}
}
