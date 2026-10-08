package signing

import (
	"errors"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// PrepareAnonymousRPC applies the pinned OpenApi doRPCRequest anonymous framing.
// It mutates an owned HTTPS POST request and never adds credentials or a signature.
func PrepareAnonymousRPC(r *http.Request, action, version string, now time.Time, nonce string) error {
	if r == nil || r.URL == nil || r.URL.Scheme != "https" || r.URL.Hostname() == "" || r.URL.User != nil || r.Method != "POST" || r.URL.Path != "/" || action == "" || version == "" || nonce == "" {
		return errors.New("anonymous RPC: invalid request")
	}
	q, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		return errors.New("anonymous RPC: invalid query")
	}
	if r.Header == nil {
		r.Header = make(http.Header)
	}
	for key := range r.Header {
		switch strings.ToLower(key) {
		case "authorization", "x-acs-security-token", "x-acs-credentials-provider", "x-acs-zero-trust-idtoken", "x-acs-signature-nonce", "x-acs-date", "x-acs-content-sha256", "x-acs-action", "x-acs-version":
			delete(r.Header, key)
		}
	}
	for key := range q {
		switch strings.ToLower(key) {
		case "accesskeyid", "accesskeysecret", "securitytoken", "signature", "signaturemethod", "signatureversion", "signaturetype", "bearertoken", "action", "version", "format", "timestamp", "signaturenonce":
			q.Del(key)
		}
	}
	q.Set("Action", action)
	q.Set("Version", version)
	q.Set("Format", "json")
	q.Set("Timestamp", now.UTC().Format("2006-01-02T15:04:05Z"))
	q.Set("SignatureNonce", nonce)
	r.URL.RawQuery = q.Encode()
	r.Header.Set("X-Acs-Action", action)
	r.Header.Set("X-Acs-Version", version)
	return nil
}
