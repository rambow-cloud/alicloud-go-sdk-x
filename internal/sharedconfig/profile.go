// Package sharedconfig reads bounded native Alibaba Cloud CLI profile snapshots.
package sharedconfig

import (
	"context"
	"crypto/sha256"
	"encoding/json/v2"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/rambow-cloud/alicloud-go-sdk-x/credentials"
)

// ErrInvalid identifies invalid configuration without exposing its contents.
var ErrInvalid = errors.New("profilecreds: invalid profile configuration")

// ErrFileMissing distinguishes an absent file from an absent selected profile.
// Both errors remain inspectable as credentials.ErrNotFound.
var ErrFileMissing = errors.Join(credentials.ErrNotFound, errors.New("profilecreds: configuration file absent"))

// Profile contains native CLI fields, kept internal because they contain secrets.
type Profile struct {
	Name                    string `json:"name"`
	Mode                    string `json:"mode"`
	Region                  string `json:"region_id"`
	AccessKeyID             string `json:"access_key_id"`
	AccessKeySecret         string `json:"access_key_secret"`
	SecurityToken           string `json:"sts_token"`
	STSExpiration           int64  `json:"sts_expiration"`
	RoleARN                 string `json:"ram_role_arn"`
	Session                 string `json:"ram_session_name"`
	Duration                int64  `json:"expired_seconds"`
	ExternalID              string `json:"external_id"`
	SourceProfile           string `json:"source_profile"`
	STSRegion               string `json:"sts_region"`
	STSEndpoint             string `json:"sts_endpoint"`
	OAuthAccessToken        string `json:"oauth_access_token"`
	OAuthRefreshToken       string `json:"oauth_refresh_token"`
	OAuthAccessExpires      int64  `json:"oauth_access_token_expire"`
	OAuthRefreshExpires     int64  `json:"oauth_refresh_token_expire"`
	OAuthSite               string `json:"oauth_site_type"`
	MetadataRole            string `json:"ram_role_name"`
	CredentialsURI          string `json:"credentials_uri"`
	ProcessCommand          string `json:"process_command"`
	OIDCProviderARN         string `json:"oidc_provider_arn"`
	OIDCTokenFile           string `json:"oidc_token_file"`
	CloudSSOSignInURL       string `json:"cloud_sso_sign_in_url"`
	CloudSSOAccessToken     string `json:"access_token"`
	CloudSSOAccessExpires   int64  `json:"cloud_sso_access_token_expire"`
	CloudSSOAccountID       string `json:"cloud_sso_account_id"`
	CloudSSOConfigurationID string `json:"cloud_sso_access_config"`
}

// File is an immutable configuration snapshot after Load returns.
type File struct {
	Current  string    `json:"current"`
	Profiles []Profile `json:"profiles"`
	Digest   [32]byte  `json:"-"`
}

// DefaultFilename resolves the native CLI configuration location.
func DefaultFilename() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return "", credentials.ErrNotFound
	}
	return filepath.Join(home, ".aliyun", "config.json"), nil
}

// Load reads at most one MiB with strict JSON v2 and sanitized errors.
func Load(ctx context.Context, filename string) (File, error) {
	if err := ctx.Err(); err != nil {
		return File{}, err
	}
	if filename == "" {
		var err error
		filename, err = DefaultFilename()
		if err != nil {
			return File{}, err
		}
	}
	f, err := os.Open(filename)
	if errors.Is(err, os.ErrNotExist) {
		return File{}, ErrFileMissing
	}
	if err != nil {
		return File{}, ErrInvalid
	}
	defer f.Close()
	const limit = 1 << 20
	b, err := io.ReadAll(io.LimitReader(f, limit+1))
	if ctx.Err() != nil {
		return File{}, ctx.Err()
	}
	if err != nil || len(b) > limit {
		return File{}, ErrInvalid
	}
	var value File
	if json.Unmarshal(b, &value) != nil || len(value.Profiles) == 0 {
		return File{}, ErrInvalid
	}
	seen := make(map[string]bool)
	for _, p := range value.Profiles {
		if strings.TrimSpace(p.Name) == "" || seen[p.Name] {
			return File{}, ErrInvalid
		}
		seen[p.Name] = true
	}
	value.Digest = sha256.Sum256(b)
	return value, nil
}

// Select resolves a named profile, falling back only when no name was selected.
func (f File) Select(name string) (Profile, error) {
	if name == "" {
		name = f.Current
	}
	if name == "" {
		name = "default"
	}
	for _, p := range f.Profiles {
		if p.Name == name {
			return p, nil
		}
	}
	return Profile{}, credentials.ErrNotFound
}
