package credentials_test

import (
	"context"
	"errors"
	"github.com/rambow-cloud/alicloud-go-sdk-x/credentials"
	"github.com/rambow-cloud/alicloud-go-sdk-x/sdktest"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestCacheCoalescesAndIsolatesCancellation(t *testing.T) {
	started := make(chan struct{})
	release := make(chan struct{})
	var calls atomic.Int32
	p := credentials.ProviderFunc(func(ctx context.Context) (credentials.Credentials, error) {
		calls.Add(1)
		close(started)
		select {
		case <-release:
			return credentials.Credentials{AccessKeyID: "test", AccessKeySecret: "secret"}, nil
		case <-ctx.Done():
			return credentials.Credentials{}, ctx.Err()
		}
	})
	c, _ := credentials.NewCache(p, credentials.CacheOptions{})
	ctx, cancel := context.WithCancel(context.Background())
	first := make(chan error, 1)
	go func() { _, err := c.Retrieve(ctx); first <- err }()
	<-started
	var wg sync.WaitGroup
	for range 20 {
		wg.Go(func() {
			v, err := c.Retrieve(context.Background())
			if err != nil || v.AccessKeyID != "test" {
				t.Error("shared refresh failed", err)
			}
		})
	}
	cancel()
	if err := <-first; !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	close(release)
	wg.Wait()
	if calls.Load() != 1 {
		t.Fatal("refresh not coalesced", calls.Load())
	}
}
func TestCacheExpirationAndRefreshFailure(t *testing.T) {
	clock := sdktest.NewClock(time.Unix(1000, 0))
	var calls atomic.Int32
	p := credentials.ProviderFunc(func(context.Context) (credentials.Credentials, error) {
		if calls.Add(1) > 1 {
			return credentials.Credentials{}, errors.New("refresh failed")
		}
		return credentials.Credentials{AccessKeyID: "test", AccessKeySecret: "secret", ExpiresAt: clock.Now().Add(2 * time.Minute)}, nil
	})
	c, _ := credentials.NewCache(p, credentials.CacheOptions{Now: clock.Now})
	if _, err := c.Retrieve(context.Background()); err != nil {
		t.Fatal(err)
	}
	clock.Advance(2 * time.Minute)
	v, err := c.Retrieve(context.Background())
	if err == nil || v != (credentials.Credentials{}) {
		t.Fatal("expired value served", err)
	}
	expired, _ := credentials.NewCache(credentials.ProviderFunc(func(context.Context) (credentials.Credentials, error) {
		return credentials.Credentials{AccessKeyID: "x", AccessKeySecret: "y", ExpiresAt: clock.Now()}, nil
	}), credentials.CacheOptions{Now: clock.Now})
	_, err = expired.Retrieve(context.Background())
	if !errors.Is(err, credentials.ErrExpired) {
		t.Fatal(err)
	}
}
func TestCacheEarlyRefreshAndInvalidation(t *testing.T) {
	clock := sdktest.NewClock(time.Unix(1000, 0))
	started := make(chan struct{})
	release := make(chan struct{})
	var calls atomic.Int32
	p := credentials.ProviderFunc(func(ctx context.Context) (credentials.Credentials, error) {
		n := calls.Add(1)
		if n == 2 {
			close(started)
			select {
			case <-release:
			case <-ctx.Done():
				return credentials.Credentials{}, ctx.Err()
			}
		}
		return credentials.Credentials{AccessKeyID: string(rune('0' + n)), AccessKeySecret: "secret", ExpiresAt: clock.Now().Add(2 * time.Minute)}, nil
	})
	c, _ := credentials.NewCache(p, credentials.CacheOptions{Now: clock.Now})
	v, _ := c.Retrieve(context.Background())
	if v.AccessKeyID != "1" {
		t.Fatal(v)
	}
	clock.Advance(time.Minute)
	v, err := c.Retrieve(context.Background())
	if err != nil || v.AccessKeyID != "1" {
		t.Fatal("early refresh blocked", err)
	}
	<-started
	c.Invalidate()
	close(release)
	v, err = c.Retrieve(context.Background())
	if err != nil || v.AccessKeyID != "3" {
		t.Fatal("invalidated refresh published", v, err)
	}
}
func TestCacheBoundsRefresh(t *testing.T) {
	c, _ := credentials.NewCache(credentials.ProviderFunc(func(ctx context.Context) (credentials.Credentials, error) {
		<-ctx.Done()
		return credentials.Credentials{}, ctx.Err()
	}), credentials.CacheOptions{RefreshTimeout: time.Millisecond})
	_, err := c.Retrieve(context.Background())
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatal(err)
	}
}
