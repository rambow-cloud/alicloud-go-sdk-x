// Package signing implements ACS3-HMAC-SHA256 and reviewed internal OSS V4
// signing. Callers own requests; protocol selection remains source-reviewed.
package signing

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"github.com/rambow-cloud/alicloud-go-sdk-x/credentials"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

const algorithm = "ACS3-HMAC-SHA256"

func digest(data []byte) string { sum := sha256.Sum256(data); return hex.EncodeToString(sum[:]) }
func escape(s string) string    { return strings.ReplaceAll(url.QueryEscape(s), "+", "%20") }

// CanonicalQuery sorts original parameter names before RFC3986 encoding, as
// required by ACS3. Repeated values are ordered by their encoded form.
func CanonicalQuery(q url.Values) string {
	var pairs []string
	keys := make([]string, 0, len(q))
	for k := range q {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		values := q[k]
		if len(values) == 0 {
			values = []string{""}
		}
		encoded := make([]string, len(values))
		for i, v := range values {
			encoded[i] = escape(v)
		}
		sort.Strings(encoded)
		for _, v := range encoded {
			pairs = append(pairs, escape(k)+"="+v)
		}
	}
	return strings.Join(pairs, "&")
}

// Sign replaces signing headers and canonicalizes URL bytes in-place. Nonce and
// time are explicit to support deterministic official-vector verification.
func Sign(r *http.Request, body []byte, c credentials.Credentials, action, version string, now time.Time, nonce string) error {
	if r == nil || r.URL == nil || c.AccessKeyID == "" || c.AccessKeySecret == "" || action == "" || version == "" || nonce == "" {
		return errors.New("signing: incomplete signing inputs")
	}
	q, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		return errors.New("signing: invalid query encoding")
	}
	r.URL.RawQuery = CanonicalQuery(q)
	path := r.URL.EscapedPath()
	if path == "" {
		path = "/"
	}
	segments := strings.Split(path, "/")
	for i, s := range segments {
		decoded, err := url.PathUnescape(s)
		if err != nil {
			return errors.New("signing: invalid path encoding")
		}
		segments[i] = escape(decoded)
	}
	canonicalPath := strings.Join(segments, "/")
	r.URL.RawPath = canonicalPath
	if r.Header == nil {
		r.Header = make(http.Header)
	}
	// Normalize names before setting managed headers, including direct map assignments.
	normalized := make(http.Header)
	for k, values := range r.Header {
		for _, v := range values {
			normalized.Add(k, v)
		}
	}
	r.Header = normalized
	r.Header.Set("x-acs-action", action)
	r.Header.Set("x-acs-version", version)
	r.Header.Set("x-acs-date", now.UTC().Format("2006-01-02T15:04:05Z"))
	r.Header.Set("x-acs-signature-nonce", nonce)
	r.Header.Set("x-acs-content-sha256", digest(body))
	r.Header.Del("x-acs-security-token")
	if c.SecurityToken != "" {
		r.Header.Set("x-acs-security-token", c.SecurityToken)
	}
	host := r.Host
	if host == "" {
		host = r.URL.Host
	}
	headers := map[string]string{"host": strings.TrimSpace(host)}
	for k, values := range r.Header {
		lower := strings.ToLower(k)
		if lower == "content-type" || strings.HasPrefix(lower, "x-acs-") {
			clean := make([]string, len(values))
			for i, v := range values {
				clean[i] = strings.TrimSpace(v)
			}
			headers[lower] = strings.Join(clean, ",")
		}
	}
	keys := make([]string, 0, len(headers))
	for k := range headers {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var canonicalHeaders strings.Builder
	for _, k := range keys {
		canonicalHeaders.WriteString(k + ":" + headers[k] + "\n")
	}
	signed := strings.Join(keys, ";")
	canonical := strings.Join([]string{r.Method, canonicalPath, r.URL.RawQuery, canonicalHeaders.String(), signed, digest(body)}, "\n")
	mac := hmac.New(sha256.New, []byte(c.AccessKeySecret))
	mac.Write([]byte(algorithm + "\n" + digest([]byte(canonical))))
	r.Header.Set("Authorization", algorithm+" Credential="+c.AccessKeyID+",SignedHeaders="+signed+",Signature="+hex.EncodeToString(mac.Sum(nil)))
	return nil
}
