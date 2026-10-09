package profilecreds

import (
	"context"
	"time"

	"github.com/rambow-cloud/alicloud-go-sdk-x/credentials"
	"github.com/rambow-cloud/alicloud-go-sdk-x/feature/externalcreds"
	"github.com/rambow-cloud/alicloud-go-sdk-x/feature/stscreds"
	"github.com/rambow-cloud/alicloud-go-sdk-x/internal/sharedconfig"
)

func cloudSSOSource(selected sharedconfig.Profile, o Options) (credentials.Provider, error) {
	now := o.CacheOptions.Now
	if now == nil {
		now = time.Now
	}
	load := func(ctx context.Context) (sharedconfig.Profile, error) {
		file, err := sharedconfig.Load(ctx, o.Filename)
		if err != nil {
			return sharedconfig.Profile{}, err
		}
		p, err := file.Select(selected.Name)
		if err != nil {
			return sharedconfig.Profile{}, err
		}
		if p.Mode != selected.Mode || p.CloudSSOSignInURL != selected.CloudSSOSignInURL || p.CloudSSOAccountID != selected.CloudSSOAccountID || p.CloudSSOConfigurationID != selected.CloudSSOConfigurationID {
			return sharedconfig.Profile{}, ErrConfigurationChanged
		}
		return p, nil
	}
	token := stscreds.TokenProviderFunc(func(ctx context.Context) (string, error) {
		p, err := load(ctx)
		if err != nil {
			return "", err
		}
		if p.CloudSSOAccessToken == "" || !now().Before(time.Unix(p.CloudSSOAccessExpires, 0)) {
			return "", ErrLoginRequired
		}
		return p.CloudSSOAccessToken, nil
	})
	portal, err := externalcreds.NewCloudSSOProvider(externalcreds.CloudSSOOptions{SignInURL: selected.CloudSSOSignInURL, AccountID: selected.CloudSSOAccountID, AccessConfigurationID: selected.CloudSSOConfigurationID, AccessToken: token, Retrieval: externalcreds.Options{HTTPClient: o.HTTPClient}})
	if err != nil {
		return nil, ErrInvalidConfiguration
	}
	return credentials.ProviderFunc(func(ctx context.Context) (credentials.Credentials, error) {
		p, err := load(ctx)
		if err != nil {
			return credentials.Credentials{}, err
		}
		c := snapshot(p)
		if c.AccessKeyID != "" && c.AccessKeySecret != "" && c.SecurityToken != "" && now().Add(refreshWindow(o.CacheOptions)).Before(c.ExpiresAt) {
			c.Source = "Profile.CloudSSO"
			return c, nil
		}
		return portal.Retrieve(ctx)
	}), nil
}
