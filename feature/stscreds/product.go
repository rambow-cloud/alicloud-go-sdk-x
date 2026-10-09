package stscreds

import (
	"context"
	"errors"
	"time"

	"github.com/rambow-cloud/alicloud-go-sdk-x/credentials"
	"github.com/rambow-cloud/alicloud-go-sdk-x/internal/rpcmodel"
	"github.com/rambow-cloud/alicloud-go-sdk-x/internal/stsrole"
	productsts "github.com/rambow-cloud/alicloud-go-sdk-x/service/sts"
)

// NewAssumeRoleProvider adapts the generated service/sts AssumeRoleAPI
// into renewable credentials.
// It validates the reviewed role/session, optional identity/policy and minimum
// duration rules; authorization and maximum role lifetime remain service decisions.
// Nil duration is omitted, while an explicit duration below 900 is rejected.
//
// Input pointers and option registrations are copied during construction and for
// every retrieval. Callers must not mutate input during construction. The API and
// option callbacks remain shared and must be concurrency safe; callbacks must not
// retain inputs/options. Wrap the result in credentials.Cache for synchronized
// refresh, and configure the STS client with separate source credentials.
// This constructor adds no Profile discovery, automatic cache or retry behavior.
func NewAssumeRoleProvider(api productsts.AssumeRoleAPI, input productsts.AssumeRoleInput, options ...func(*productsts.Options)) (*AssumeRoleProvider, error) {
	if rpcmodel.IsNil(api) {
		return nil, errors.New("stscreds: STS API required")
	}
	owned, err := rpcmodel.Snapshot(context.Background(), &input)
	if err != nil {
		return nil, err
	}
	if owned.DurationSeconds != nil && *owned.DurationSeconds < 900 {
		return nil, errors.New("stscreds: duration must be at least 900 seconds")
	}
	// Shared helper policy does not transform the native wire model.
	validation := stsrole.Input{
		RoleARN: value(owned.RoleARN), RoleSessionName: value(owned.RoleSessionName),
		DurationSeconds: value(owned.DurationSeconds), Policy: value(owned.Policy),
		ExternalID: value(owned.ExternalID), SourceIdentity: value(owned.SourceIdentity),
	}
	if err := validation.Validate(); err != nil {
		return nil, err
	}
	for _, option := range options {
		if option == nil {
			return nil, errors.New("stscreds: nil option")
		}
	}
	registrations := append([]func(*productsts.Options){}, options...)
	return &AssumeRoleProvider{retrieve: func(ctx context.Context) (credentials.Credentials, error) {
		request, err := rpcmodel.Snapshot(ctx, owned)
		if err != nil {
			return credentials.Credentials{}, err
		}
		out, err := api.AssumeRole(ctx, request, append([]func(*productsts.Options){}, registrations...)...)
		if err != nil {
			return credentials.Credentials{}, err
		}
		if out == nil || out.Credentials == nil {
			return credentials.Credentials{}, errors.New("stscreds: missing response credentials")
		}
		c := out.Credentials
		if c.Expiration == nil || *c.Expiration == "" {
			return credentials.Credentials{}, credentials.ErrExpired
		}
		expires, err := time.Parse(time.RFC3339, *c.Expiration)
		if err != nil {
			// time.Parse embeds the raw value in its error. Keep this diagnostic bounded.
			return credentials.Credentials{}, errors.New("stscreds: invalid credential expiration")
		}
		return credentials.Credentials{AccessKeyID: value(c.AccessKeyID), AccessKeySecret: value(c.AccessKeySecret), SecurityToken: value(c.SecurityToken), ExpiresAt: expires.UTC()}, nil
	}}, nil
}

func value[T any](pointer *T) (v T) {
	if pointer != nil {
		return *pointer
	}
	return v
}

// NewAssumeRoleProviderFromClient forwards to NewAssumeRoleProvider.
// It has the same validation, ownership, cancellation and concurrency contracts.
func NewAssumeRoleProviderFromClient(api productsts.AssumeRoleAPI, input productsts.AssumeRoleInput, options ...func(*productsts.Options)) (*AssumeRoleProvider, error) {
	return NewAssumeRoleProvider(api, input, options...)
}
