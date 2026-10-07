// Package sts provides a handwritten AssumeRole reference client for API version
// 2015-04-01. It uses source credentials supplied to alicloud.Config; returned
// temporary credentials do not replace that source. The client is concurrency
// safe and exposes a small operation interface for fakes. Standard retry treats
// token issuance as non-idempotent. No other STS operations are implemented.
package sts
