package main

import (
	"context"
	"errors"
	alicloud "github.com/rambow-cloud/alicloud-go-sdk-x"
	"github.com/rambow-cloud/alicloud-go-sdk-x/feature/stscreds"
	"github.com/rambow-cloud/alicloud-go-sdk-x/service/sts"
	"testing"
)

type roleMock struct{ calls int }

func (m *roleMock) AssumeRole(ctx context.Context, _ *sts.AssumeRoleInput, _ ...func(*sts.Options)) (*sts.AssumeRoleOutput, error) {
	m.calls++
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return nil, &alicloud.APIError{Code: "Forbidden", HTTPStatusCode: 403}
}

func TestExternalGeneratedWorkload(t *testing.T) {
	if err := ours(context.Background()); err != nil {
		t.Fatal(err)
	}
}
func TestPinnedOfficialWorkload(t *testing.T) {
	if err := official(); err != nil {
		t.Fatal(err)
	}
}
func TestExternalSmallMockStructuredErrorsAndCancellation(t *testing.T) {
	mock := new(roleMock)
	var api sts.AssumeRoleAPI = mock
	role := roleFixture
	session := "acceptance"
	p, err := stscreds.NewAssumeRoleProviderFromClient(api, sts.AssumeRoleInput{RoleARN: &role, RoleSessionName: &session})
	if err != nil {
		t.Fatal(err)
	}
	_, err = p.Retrieve(context.Background())
	var ae *alicloud.APIError
	if !errors.As(err, &ae) || ae.Code != "Forbidden" || mock.calls != 1 {
		t.Fatal("structured mock error lost")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err = p.Retrieve(ctx); !errors.Is(err, context.Canceled) || mock.calls != 1 {
		t.Fatal("cancellation did not stop mock call")
	}
}
