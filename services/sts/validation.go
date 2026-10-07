package sts

import (
	"encoding/json/v2"
	"errors"
	"regexp"
	"strings"
)

// These reviewed constraints are prose-only and remain handwritten extensions.
var sessionPattern = regexp.MustCompile(`^[A-Za-z0-9.@_-]{2,64}$`)
var externalPattern = regexp.MustCompile(`^[A-Za-z0-9_+=,.@:/-]+$`)
var sourcePattern = regexp.MustCompile(`^[A-Za-z0-9_+=,.@-]{2,64}$`)

// Validate checks required fields, documented session syntax and policy JSON;
// the role's maximum session duration and authorization remain service decisions.
func (in AssumeRoleInput) Validate() error {
	if !strings.HasPrefix(in.RoleARN, "acs:ram::") || !strings.Contains(in.RoleARN, ":role/") || !sessionPattern.MatchString(in.RoleSessionName) {
		return errors.New("sts: invalid role ARN or session name")
	}
	if in.DurationSeconds != 0 && in.DurationSeconds < 900 {
		return errors.New("sts: duration must be at least 900 seconds")
	}
	if len(in.Policy) > 2048 {
		return errors.New("sts: policy exceeds 2048 bytes")
	}
	if in.Policy != "" {
		var policy map[string]any
		if err := json.Unmarshal([]byte(in.Policy), &policy); err != nil || policy == nil {
			return errors.New("sts: invalid policy JSON object")
		}
	}
	if in.ExternalID != "" && (len(in.ExternalID) < 2 || len(in.ExternalID) > 1224 || !externalPattern.MatchString(in.ExternalID)) {
		return errors.New("sts: invalid external ID")
	}
	if in.SourceIdentity != "" {
		if !sourcePattern.MatchString(in.SourceIdentity) || strings.HasPrefix(in.SourceIdentity, "acs:") || strings.HasPrefix(in.SourceIdentity, "aliyun:") || strings.HasPrefix(in.SourceIdentity, "alibabacloud:") {
			return errors.New("sts: invalid source identity")
		}
	}
	return nil
}

func validateAssumeRole(in AssumeRoleInput) error { return in.Validate() }
