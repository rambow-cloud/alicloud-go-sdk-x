package signing

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/rambow-cloud/alicloud-go-sdk-x/credentials"
)

const oss4Algorithm = "OSS4-HMAC-SHA256"

var errOSS4 = errors.New("signing: invalid OSS V4 inputs")

// OSS4Options supplies reviewed bucket scope, region, time and optional headers.
// No region/clock/credential discovery occurs in the signer.
type OSS4Options struct {
	// Bucket is the virtual-host bucket name, or empty for service-level calls.
	// Request.URL.Path excludes this bucket; the canonical resource includes it.
	Bucket string
	// Region is the exact OSS region identifier; it is required.
	Region string
	// Time is the nonzero signing instant; UTC is used without changing this value.
	Time time.Time
	// AdditionalHeaders explicitly selects optional signing headers. Names are
	// case-insensitive and must be unique, present and outside the required set.
	// content-length requires a positive known Request.ContentLength and body;
	// host uses Request.Host or URL.Host. The slice is never mutated/retained.
	AdditionalHeaders []string
}

func scopePart(value string) bool {
	if value == "" {
		return false
	}
	for _, r := range value {
		if (r < 'a' || r > 'z') && (r < 'A' || r > 'Z') && (r < '0' || r > '9') && r != '-' && r != '_' {
			return false
		}
	}
	return true
}

// credentialID accepts the period used by temporary STS identifiers. Scope and
// authorization separators remain invalid; regions use the stricter scopePart.
func credentialID(value string) bool {
	if value == "" {
		return false
	}
	for _, r := range value {
		if (r < 'a' || r > 'z') && (r < 'A' || r > 'Z') && (r < '0' || r > '9') && r != '-' && r != '_' && r != '.' {
			return false
		}
	}
	return true
}

func bucketName(value string) bool {
	if value == "" {
		return true
	}
	if len(value) < 3 || len(value) > 63 || value[0] == '-' || value[len(value)-1] == '-' {
		return false
	}
	for _, c := range value {
		if (c < 'a' || c > 'z') && (c < '0' || c > '9') && c != '-' {
			return false
		}
	}
	return true
}

func headerName(value string) bool {
	if value == "" {
		return false
	}
	for _, r := range value {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || strings.ContainsRune("!#$%&'*+-.^_`|~", r) {
			continue
		}
		return false
	}
	return true
}

func headerValue(value string) bool {
	for i := 0; i < len(value); i++ {
		if (value[i] < 32 && value[i] != '\t') || value[i] == 127 {
			return false
		}
	}
	return true
}

func requiredOSS4Header(name string) bool {
	return name == "content-type" || name == "content-md5" || strings.HasPrefix(name, "x-oss-")
}

func oss4Query(raw string) (string, error) {
	query, err := url.ParseQuery(raw)
	if err != nil {
		return "", errOSS4
	}
	pairs := make(map[string]string, len(query))
	var keys []string
	for name, values := range query {
		if name == "" || !utf8.ValidString(name) || len(values) != 1 || !utf8.ValidString(values[0]) {
			return "", errOSS4
		}
		switch strings.ToLower(name) {
		case "signature", "ossaccesskeyid", "expires", "security-token", "x-oss-signature", "x-oss-credential", "x-oss-date", "x-oss-security-token", "x-oss-expires", "x-oss-signature-version":
			return "", errOSS4
		}
		encoded := escape(name)
		keys = append(keys, encoded)
		pairs[encoded] = escape(values[0])
	}
	sort.Strings(keys)
	for i, key := range keys {
		if pairs[key] != "" {
			keys[i] = key + "=" + pairs[key]
		}
	}
	return strings.Join(keys, "&"), nil
}

func oss4Path(path string) string {
	return strings.ReplaceAll(escape(path), "%2F", "/")
}

func oss4Key(secret, date, region string) []byte {
	key := []byte("aliyun_v4" + secret)
	for _, value := range []string{date, region, "oss", "aliyun_v4_request"} {
		mac := hmac.New(sha256.New, key)
		mac.Write([]byte(value))
		key = mac.Sum(nil)
	}
	return key
}

func oss4Signature(key []byte, value string) string {
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(value))
	return hex.EncodeToString(mac.Sum(nil))
}

// SignOSS4 signs owned HTTPS virtual-host/service requests with UNSIGNED-PAYLOAD.
// It replaces managed signing headers on cloned headers/URL only after validation
// and context checks succeed. Bodies, GetBody and other request fields are not read,
// closed, replaced or retained. Invalid/multi-value inputs return a redacted error;
// expiry/cancellation remain inspectable. It supports no V1/V2, presigned URLs,
// CloudBox scope, checksum calculation or endpoint construction. Callers must own
// the request during signing; separate requests may be signed concurrently.
func SignOSS4(ctx context.Context, request *http.Request, c credentials.Credentials, options OSS4Options) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if request == nil || request.URL == nil || request.URL.Scheme != "https" || request.URL.Host == "" || request.URL.User != nil || request.URL.Opaque != "" || request.URL.Fragment != "" || request.URL.RawFragment != "" || !bucketName(options.Bucket) || !scopePart(options.Region) || !credentialID(c.AccessKeyID) || c.AccessKeySecret == "" || options.Time.IsZero() || !headerValue(c.SecurityToken) {
		return errOSS4
	}
	if !c.ExpiresAt.IsZero() && !options.Time.Before(c.ExpiresAt) {
		return credentials.ErrExpired
	}
	switch request.Method {
	case "GET", "PUT", "POST", "DELETE", "HEAD", "OPTIONS":
	default:
		return errOSS4
	}
	path := request.URL.Path
	if path == "" {
		path = "/"
	}
	if !strings.HasPrefix(path, "/") || !utf8.ValidString(path) {
		return errOSS4
	}
	if request.URL.RawPath != "" {
		decoded, err := url.PathUnescape(request.URL.RawPath)
		if err != nil || decoded != request.URL.Path {
			return errOSS4
		}
	}
	query, err := oss4Query(request.URL.RawQuery)
	if err != nil {
		return err
	}
	headers := make(http.Header, len(request.Header)+4)
	seen := map[string]bool{}
	for name, values := range request.Header {
		lower := strings.ToLower(name)
		if !headerName(name) || seen[lower] || len(values) != 1 || !headerValue(values[0]) || strings.HasPrefix(lower, "x-acs-") || lower == "host" || lower == "x-oss-meta-*" {
			return errOSS4
		}
		seen[lower] = true
		headers.Set(lower, values[0])
	}
	dateTime := options.Time.UTC().Format("20060102T150405Z")
	date := options.Time.UTC().Format("20060102")
	headers.Set("x-oss-date", dateTime)
	headers.Set("x-oss-content-sha256", "UNSIGNED-PAYLOAD")
	headers.Del("x-oss-security-token")
	if c.SecurityToken != "" {
		headers.Set("x-oss-security-token", c.SecurityToken)
	}
	additional := make([]string, 0, len(options.AdditionalHeaders))
	selected := map[string]string{}
	for name, values := range headers {
		lower := strings.ToLower(name)
		if requiredOSS4Header(lower) {
			selected[lower] = strings.TrimSpace(values[0])
		}
	}
	for _, name := range options.AdditionalHeaders {
		lower := strings.ToLower(name)
		if !headerName(name) || requiredOSS4Header(lower) || lower == "authorization" {
			return errOSS4
		}
		switch lower {
		case "connection", "transfer-encoding", "trailer", "upgrade", "te", "proxy-authorization", "proxy-authenticate":
			return errOSS4
		}
		if _, exists := selected[lower]; exists {
			return errOSS4
		}
		var value string
		switch lower {
		case "host":
			value = request.Host
			if value == "" {
				value = request.URL.Host
			}
		case "content-length":
			if request.ContentLength <= 0 || request.Body == nil || request.Body == http.NoBody || len(request.TransferEncoding) != 0 {
				return errOSS4
			}
			value = strconv.FormatInt(request.ContentLength, 10)
			if values, exists := headers[http.CanonicalHeaderKey(lower)]; exists && values[0] != value {
				return errOSS4
			}
		default:
			values, exists := headers[http.CanonicalHeaderKey(lower)]
			if !exists {
				return errOSS4
			}
			value = values[0]
		}
		if !headerValue(value) || strings.TrimSpace(value) == "" {
			return errOSS4
		}
		additional = append(additional, lower)
		selected[lower] = strings.TrimSpace(value)
	}
	sort.Strings(additional)
	names := make([]string, 0, len(selected))
	for name := range selected {
		names = append(names, name)
	}
	sort.Strings(names)
	var canonicalHeaders strings.Builder
	for _, name := range names {
		canonicalHeaders.WriteString(name + ":" + selected[name] + "\n")
	}
	resource := path
	if options.Bucket != "" {
		resource = "/" + options.Bucket + path
	}
	canonical := request.Method + "\n" + oss4Path(resource) + "\n" + query + "\n" + canonicalHeaders.String() + "\n" + strings.Join(additional, ";") + "\nUNSIGNED-PAYLOAD"
	scope := date + "/" + options.Region + "/oss/aliyun_v4_request"
	stringToSign := oss4Algorithm + "\n" + dateTime + "\n" + scope + "\n" + digest([]byte(canonical))
	key := oss4Key(c.AccessKeySecret, date, options.Region)
	signature := oss4Signature(key, stringToSign)
	authorization := oss4Algorithm + " Credential=" + c.AccessKeyID + "/" + scope
	if len(additional) != 0 {
		authorization += ", AdditionalHeaders=" + strings.Join(additional, ";")
	}
	headers.Set("authorization", authorization+", Signature="+signature)
	if err := ctx.Err(); err != nil {
		return err
	}
	location := *request.URL
	location.Path, location.RawPath, location.RawQuery = path, oss4Path(path), query
	request.URL, request.Header = &location, headers
	return nil
}
