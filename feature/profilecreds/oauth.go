package profilecreds

import (
	"context"
	"encoding/json/v2"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	alicloud "github.com/rambow-cloud/alicloud-go-sdk-x"
	"github.com/rambow-cloud/alicloud-go-sdk-x/credentials"
	"github.com/rambow-cloud/alicloud-go-sdk-x/internal/sharedconfig"
)

// OAuthError describes a bounded OAuth failure without URL, tokens or response text.
// Inspect StatusCode/Stage and errors.Is; do not log raw credential responses.
type OAuthError struct {
	// Stage is the SDK-controlled refresh or exchange operation name.
	Stage string
	// StatusCode is the HTTP status, or zero for a transport/decoding failure.
	StatusCode int
	// Err is an inspectable cause; explicit application access may expose transport details.
	Err error
}

// Error omits cause text because it may contain sensitive request information.
func (e *OAuthError) Error() string {
	return fmt.Sprintf("profilecreds: OAuth %s failed (status=%d)", e.Stage, e.StatusCode)
}

// Unwrap preserves cancellation and ErrLoginRequired for errors.Is.
func (e *OAuthError) Unwrap() error { return e.Err }

// String returns the safe default error representation.
func (e *OAuthError) String() string { return e.Error() }

// GoString omits Err details for %#v formatting.
func (e *OAuthError) GoString() string { return e.Error() }

// oauthSource is called only by one bounded Cache refresh at a time.
type oauthSource struct {
	profile     sharedconfig.Profile
	filename    string
	client      alicloud.HTTPClient
	now         func() time.Time
	window      time.Duration
	seedChecked bool
}

func (s *oauthSource) Retrieve(ctx context.Context) (credentials.Credentials, error) {
	if err := ctx.Err(); err != nil {
		return credentials.Credentials{}, err
	}
	now := s.now()
	if !s.seedChecked {
		s.seedChecked = true
		seed := snapshot(s.profile)
		if validTemporary(seed, now.Add(s.window)) {
			seed.Source = "Profile.OAuth"
			return seed, nil
		}
	}
	release, err := sharedconfig.LockSession(ctx, s.filename)
	if err != nil {
		return credentials.Credentials{}, err
	}
	defer release()
	file, err := sharedconfig.Load(ctx, s.filename)
	if err != nil {
		return credentials.Credentials{}, err
	}
	current, err := file.Select(s.profile.Name)
	if err != nil || !strings.EqualFold(current.Mode, "OAuth") || current.OAuthSite != s.profile.OAuthSite {
		return credentials.Credentials{}, ErrConfigurationChanged
	}
	previous := s.profile
	fingerprint := file.Digest
	s.profile = current
	// Another SDK process may already have renewed the session while this caller waited.
	if (current.STSExpiration != previous.STSExpiration || current.AccessKeyID != previous.AccessKeyID) && validTemporary(snapshot(current), s.now().Add(s.window)) {
		value := snapshot(current)
		value.Source = "Profile.OAuth"
		return value, nil
	}
	base, clientID := "https://oauth.aliyun.com", "4038181954557748008"
	if s.profile.OAuthSite == "INTL" {
		base, clientID = "https://oauth.alibabacloud.com", "4103531455503354461"
	}
	if strings.TrimSpace(s.profile.OAuthAccessToken) == "" || !now.Add(30*time.Second).Before(time.Unix(s.profile.OAuthAccessExpires, 0)) {
		if strings.TrimSpace(s.profile.OAuthRefreshToken) == "" || (s.profile.OAuthRefreshExpires > 0 && !now.Before(time.Unix(s.profile.OAuthRefreshExpires, 0))) {
			return credentials.Credentials{}, ErrLoginRequired
		}
		form := url.Values{"grant_type": {"refresh_token"}, "refresh_token": {s.profile.OAuthRefreshToken}, "client_id": {clientID}}
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, base+"/v1/token", strings.NewReader(form.Encode()))
		if err != nil {
			return credentials.Credentials{}, ErrInvalidConfiguration
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		var response struct {
			AccessToken  string `json:"access_token"`
			RefreshToken string `json:"refresh_token"`
			ExpiresIn    int64  `json:"expires_in"`
			TokenType    string `json:"token_type"`
		}
		if err := s.do(req, "refresh", &response); err != nil {
			return credentials.Credentials{}, err
		}
		if strings.TrimSpace(response.AccessToken) == "" || response.ExpiresIn <= 0 || response.ExpiresIn > int64((365*24*time.Hour)/time.Second) || (response.TokenType != "" && !strings.EqualFold(response.TokenType, "Bearer")) {
			return credentials.Credentials{}, &OAuthError{Stage: "refresh", StatusCode: 200, Err: ErrInvalidConfiguration}
		}
		// Retain the last refresh token when the server omits an optional replacement.
		// Commit token rotation before exchange so a failed exchange cannot lose it.
		s.profile.OAuthAccessToken = response.AccessToken
		if response.RefreshToken != "" {
			s.profile.OAuthRefreshToken = response.RefreshToken
		}
		s.profile.OAuthAccessExpires = s.now().Add(time.Duration(response.ExpiresIn) * time.Second).Unix()
		fingerprint, err = sharedconfig.PersistSession(ctx, s.filename, s.profile, fingerprint)
		if err != nil {
			return credentials.Credentials{}, err
		}
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, base+"/v1/exchange", nil)
	if err != nil {
		return credentials.Credentials{}, ErrInvalidConfiguration
	}
	req.Header.Set("Authorization", "Bearer "+s.profile.OAuthAccessToken)
	req.Header.Set("Content-Type", "application/json")
	var wire struct {
		NativeID     *string `json:"AccessKeyId"`
		NativeSecret *string `json:"AccessKeySecret"`
		NativeToken  *string `json:"SecurityToken"`
		NativeExpiry *string `json:"Expiration"`
		LegacyID     *string `json:"accessKeyId"`
		LegacySecret *string `json:"accessKeySecret"`
		LegacyToken  *string `json:"securityToken"`
		LegacyExpiry *string `json:"expiration"`
	}
	if err := s.do(req, "exchange", &wire); err != nil {
		return credentials.Credentials{}, err
	}
	// Live CN responses use PascalCase; the pinned CLI uses camelCase tags with
	// its legacy decoder's case folding. Accept only these complete wire variants.
	legacy := wire.LegacyID != nil || wire.LegacySecret != nil || wire.LegacyToken != nil || wire.LegacyExpiry != nil
	native := wire.NativeID != nil || wire.NativeSecret != nil || wire.NativeToken != nil || wire.NativeExpiry != nil
	if legacy && native {
		return credentials.Credentials{}, &OAuthError{Stage: "exchange", StatusCode: 200, Err: ErrInvalidConfiguration}
	}
	id, secret, token, expiry := wire.NativeID, wire.NativeSecret, wire.NativeToken, wire.NativeExpiry
	if legacy {
		id, secret, token, expiry = wire.LegacyID, wire.LegacySecret, wire.LegacyToken, wire.LegacyExpiry
	}
	if id == nil || secret == nil || token == nil || expiry == nil {
		return credentials.Credentials{}, &OAuthError{Stage: "exchange", StatusCode: 200, Err: credentials.ErrMissingCredentials}
	}
	result := struct{ AccessKeyID, AccessKeySecret, SecurityToken, Expiration string }{*id, *secret, *token, *expiry}
	expires, err := time.Parse(time.RFC3339, result.Expiration)
	if err != nil {
		return credentials.Credentials{}, &OAuthError{Stage: "exchange", StatusCode: 200, Err: ErrInvalidConfiguration}
	}
	v := credentials.Credentials{AccessKeyID: result.AccessKeyID, AccessKeySecret: result.AccessKeySecret, SecurityToken: result.SecurityToken, ExpiresAt: expires.UTC(), Source: "Profile.OAuth"}
	if strings.TrimSpace(v.AccessKeyID) == "" || strings.TrimSpace(v.AccessKeySecret) == "" || strings.TrimSpace(v.SecurityToken) == "" {
		return credentials.Credentials{}, &OAuthError{Stage: "exchange", StatusCode: 200, Err: credentials.ErrMissingCredentials}
	}
	if !validTemporary(v, s.now()) {
		return credentials.Credentials{}, &OAuthError{Stage: "exchange", StatusCode: 200, Err: credentials.ErrExpired}
	}
	s.profile.AccessKeyID = v.AccessKeyID
	s.profile.AccessKeySecret = v.AccessKeySecret
	s.profile.SecurityToken = v.SecurityToken
	s.profile.STSExpiration = v.ExpiresAt.Unix()
	if _, err := sharedconfig.PersistSession(ctx, s.filename, s.profile, fingerprint); err != nil {
		return credentials.Credentials{}, err
	}
	return v, nil
}

func validTemporary(v credentials.Credentials, now time.Time) bool {
	return strings.TrimSpace(v.AccessKeyID) != "" && strings.TrimSpace(v.AccessKeySecret) != "" && strings.TrimSpace(v.SecurityToken) != "" && !v.ExpiresAt.IsZero() && now.Before(v.ExpiresAt)
}

func (s *oauthSource) do(req *http.Request, stage string, output any) error {
	resp, err := s.client.Do(req)
	if err != nil {
		return &OAuthError{Stage: stage, Err: err}
	}
	if resp == nil || resp.Body == nil {
		return &OAuthError{Stage: stage, Err: ErrInvalidConfiguration}
	}
	defer resp.Body.Close()
	const limit = 1 << 20
	body, err := io.ReadAll(io.LimitReader(resp.Body, limit+1))
	if req.Context().Err() != nil {
		return req.Context().Err()
	}
	if err != nil {
		return &OAuthError{Stage: stage, StatusCode: resp.StatusCode, Err: err}
	}
	if len(body) > limit {
		return &OAuthError{Stage: stage, StatusCode: resp.StatusCode, Err: ErrInvalidConfiguration}
	}
	if resp.StatusCode != http.StatusOK {
		cause := errors.New("profilecreds: OAuth endpoint rejected request")
		var failure struct {
			Error string `json:"error"`
			Code  string `json:"code"`
		}
		_ = json.Unmarshal(body, &failure)
		if resp.StatusCode == 401 || failure.Error == "invalid_grant" || failure.Error == "invalid_token" || failure.Code == "invalid_grant" || failure.Code == "invalid_token" {
			cause = ErrLoginRequired
		}
		return &OAuthError{Stage: stage, StatusCode: resp.StatusCode, Err: cause}
	}
	if json.Unmarshal(body, output) != nil {
		return &OAuthError{Stage: stage, StatusCode: resp.StatusCode, Err: ErrInvalidConfiguration}
	}
	return nil
}
