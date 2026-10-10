package alicloud

import (
	"context"
	"encoding/xml"
	"errors"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/rambow-cloud/alicloud-go-sdk-x/internal/xmlmodel"
)

var ossBucketPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{1,61}[a-z0-9]$`)
var ossScopePattern = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)
var ossHostLabelPattern = regexp.MustCompile(`^[A-Za-z0-9](?:[A-Za-z0-9-]{0,61}[A-Za-z0-9])?$`)

func validateOSSIdentity(op Operation, bucket, region string) error {
	if op.Authentication != AuthenticationOSS4 {
		if bucket != "" {
			return errors.New("alicloud: bucket identity requires OSS4")
		}
		return nil
	}
	if !ossScopePattern.MatchString(region) || (bucket != "" && !ossBucketPattern.MatchString(bucket)) {
		return errors.New("alicloud: invalid OSS region or bucket")
	}
	return nil
}

func prepareOSSOrigin(u *url.URL, bucket string) error {
	host := u.Hostname()
	if net.ParseIP(host) != nil || len(host) > 253 || strings.Contains(host, ":") || (bucket != "" && strings.HasPrefix(strings.ToLower(host), bucket+".")) {
		return errors.New("alicloud: unsupported OSS origin")
	}
	for _, label := range strings.Split(host, ".") {
		if !ossHostLabelPattern.MatchString(label) {
			return errors.New("alicloud: invalid OSS origin")
		}
	}
	if bucket != "" {
		if len(bucket)+1+len(host) > 253 {
			return errors.New("alicloud: invalid OSS origin")
		}
		u.Host = bucket + "." + u.Host
	}
	return nil
}

func responseRequestID(op Operation, header http.Header) string {
	if op.Authentication == AuthenticationOSS4 {
		return header.Get("X-Oss-Request-Id")
	}
	return header.Get("X-Acs-Request-Id")
}

func setOSSManagedHeader(header http.Header, name, value string) {
	for key := range header {
		if strings.EqualFold(key, name) {
			delete(header, key)
		}
	}
	header.Set(name, value)
}

func decodeOSSError(ctx context.Context, data []byte, status int) (serviceEnvelope, error) {
	if err := ctx.Err(); err != nil {
		return serviceEnvelope{}, err
	}
	if len(data) == 0 {
		return serviceEnvelope{Code: strconv.Itoa(status)}, nil
	}
	var wire struct {
		Code      string `json:"Code"`
		Message   string `json:"Message"`
		RequestID string `json:"RequestId"`
		ECCode    string `json:"EC"`
	}
	if err := xmlmodel.Decode(ctx, data, xmlmodel.Root{Name: xml.Name{Local: "Error"}}, &wire); err != nil {
		return serviceEnvelope{}, err
	}
	return serviceEnvelope{Code: wire.Code, Message: wire.Message, RequestID: wire.RequestID, ECCode: wire.ECCode}, nil
}
