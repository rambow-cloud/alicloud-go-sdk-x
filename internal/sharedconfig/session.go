package sharedconfig

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json/jsontext"
	"encoding/json/v2"
	"errors"
	"io"
	"os"
	"path/filepath"
	"time"
)

// ErrPersistence identifies bounded session-write failures without file contents.
var ErrPersistence = errors.New("profilecreds: OAuth session persistence failed")

// ErrChanged means a concurrent configuration edit invalidated a session write.
var ErrChanged = errors.New("profilecreds: profile configuration changed during OAuth renewal")

// LockSession serializes SDK renewal against the same canonical CLI file.
// Existing locks are never stolen. Context bounds waiting; release removes only
// the file created by this call. CLI reconfiguration must not run concurrently.
func LockSession(ctx context.Context, filename string) (func(), error) {
	lockname := filename + ".alicloud-go-sdk-x.oauth.lock"
	for {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		f, err := os.OpenFile(lockname, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
		if err == nil {
			if f.Close() != nil {
				_ = os.Remove(lockname)
				return nil, ErrPersistence
			}
			return func() { _ = os.Remove(lockname) }, nil
		}
		if !errors.Is(err, os.ErrExist) {
			return nil, ErrPersistence
		}
		timer := time.NewTimer(20 * time.Millisecond)
		select {
		case <-ctx.Done():
			timer.Stop()
			return nil, ctx.Err()
		case <-timer.C:
		}
	}
}

// PersistSession merges only native OAuth/STS fields, preserving every other
// JSON value. Caller holds LockSession. Atomic replacement uses a private same-dir
// temporary file; detected external edits leave the current file intact.
func PersistSession(ctx context.Context, filename string, profile Profile, expected [32]byte) (digest [32]byte, failure error) {
	if ctx.Err() != nil {
		return digest, ctx.Err()
	}
	read := func() ([]byte, error) {
		f, err := os.Open(filename)
		if err != nil {
			return nil, ErrPersistence
		}
		defer f.Close()
		b, err := io.ReadAll(io.LimitReader(f, (1<<20)+1))
		if err != nil || len(b) > 1<<20 {
			return nil, ErrPersistence
		}
		return b, nil
	}
	original, err := read()
	if err != nil {
		return digest, err
	}
	if sha256.Sum256(original) != expected {
		return digest, ErrChanged
	}
	var root map[string]jsontext.Value
	if json.Unmarshal(original, &root) != nil {
		return digest, ErrPersistence
	}
	var profiles []map[string]jsontext.Value
	if json.Unmarshal(root["profiles"], &profiles) != nil {
		return digest, ErrPersistence
	}
	found := false
	for _, p := range profiles {
		var name string
		if json.Unmarshal(p["name"], &name) != nil {
			return digest, ErrPersistence
		}
		if name != profile.Name {
			continue
		}
		found = true
		for key, value := range map[string]any{"access_key_id": profile.AccessKeyID, "access_key_secret": profile.AccessKeySecret, "sts_token": profile.SecurityToken, "sts_expiration": profile.STSExpiration, "oauth_access_token": profile.OAuthAccessToken, "oauth_refresh_token": profile.OAuthRefreshToken, "oauth_access_token_expire": profile.OAuthAccessExpires} {
			encoded, err := json.Marshal(value)
			if err != nil {
				return digest, ErrPersistence
			}
			p[key] = encoded
		}
	}
	if !found {
		return digest, ErrChanged
	}
	b, err := json.Marshal(profiles)
	if err != nil {
		return digest, ErrPersistence
	}
	root["profiles"] = b
	b, err = json.Marshal(root, jsontext.WithIndent("  "))
	if err != nil {
		return digest, ErrPersistence
	}
	b = append(b, '\n')
	if len(b) > 1<<20 {
		return digest, ErrPersistence
	}
	f, err := os.CreateTemp(filepath.Dir(filename), ".alicloud-oauth-*.tmp")
	if err != nil {
		return digest, ErrPersistence
	}
	temp := f.Name()
	defer os.Remove(temp)
	if _, err = f.Write(b); err != nil {
		f.Close()
		return digest, ErrPersistence
	}
	if err = f.Sync(); err != nil {
		f.Close()
		return digest, ErrPersistence
	}
	if err = f.Close(); err != nil {
		return digest, ErrPersistence
	}
	if ctx.Err() != nil {
		return digest, ctx.Err()
	}
	current, err := read()
	if err != nil {
		return digest, err
	}
	if !bytes.Equal(original, current) {
		return digest, ErrChanged
	}
	if os.Rename(temp, filename) != nil {
		return digest, ErrPersistence
	}
	return sha256.Sum256(b), nil
}
