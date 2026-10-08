package sts

import "github.com/rambow-cloud/alicloud-go-sdk-x/internal/stsrole"

// Validate checks required fields, documented session syntax and policy JSON;
// the role's maximum session duration and authorization remain service decisions.
func (in AssumeRoleInput) Validate() error {
	return (stsrole.Input{
		RoleARN: in.RoleARN, RoleSessionName: in.RoleSessionName,
		DurationSeconds: in.DurationSeconds, Policy: in.Policy,
		ExternalID: in.ExternalID, SourceIdentity: in.SourceIdentity,
	}).Validate()
}

func validateAssumeRole(in AssumeRoleInput) error { return in.Validate() }
