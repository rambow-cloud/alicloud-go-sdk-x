// Package checksum computes protocol checksums for already owned bytes.
// It does not select operations, consume streams or define integrity policy.
package checksum

import (
	"context"
	"crypto/md5"
	"encoding/base64"
	"errors"
	"hash/crc64"
	"strconv"
)

// ErrInvalid indicates a missing, malformed or overflowing server CRC header.
// It never exposes the supplied value.
var ErrInvalid = errors.New("checksum: invalid CRC64 header")

// ErrMismatch indicates different local and server CRC64 values.
// It never exposes checksum values or payload data.
var ErrMismatch = errors.New("checksum: CRC64 mismatch")

var ecma = crc64.MakeTable(crc64.ECMA)

func chunks(ctx context.Context, data []byte, consume func([]byte)) error {
	const size = 32 << 10
	for len(data) > 0 {
		if err := ctx.Err(); err != nil {
			return err
		}
		n := min(len(data), size)
		consume(data[:n])
		data = data[n:]
	}
	return ctx.Err()
}

// CRC64 computes the OSS-compatible CRC using Go's reflected ECMA table and
// finalized seed convention. seed is zero for a new object, or the finalized CRC
// of preceding bytes for append/continuation. Empty data preserves seed.
// Cancellation returns zero with the original context error, never a partial
// checksum. data is read without retention/mutation; callers must not mutate it
// during calculation. Independent calls are concurrency safe.
func CRC64(ctx context.Context, seed uint64, data []byte) (uint64, error) {
	value := seed
	if err := chunks(ctx, data, func(part []byte) { value = crc64.Update(value, ecma, part) }); err != nil {
		return 0, err
	}
	return value, nil
}

// ContentMD5 returns the base64 MD5 of the exact payload bytes, including empty
// input, for the protocol Content-MD5 header. It is an integrity checksum, not an
// authentication primitive. Cancellation returns an empty string with the
// original context error. data is read without retention/mutation; callers must
// not mutate it during calculation. Independent calls are concurrency safe.
func ContentMD5(ctx context.Context, data []byte) (string, error) {
	digest := md5.New()
	if err := chunks(ctx, data, func(part []byte) { _, _ = digest.Write(part) }); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(digest.Sum(nil)), nil
}

// VerifyCRC64 compares local with an unsigned decimal server CRC64 header.
// server must contain 1-20 ASCII digits; leading zeros are accepted. Missing,
// signed, whitespace, hexadecimal or overflowing values return ErrInvalid;
// different valid values return ErrMismatch. Errors contain no supplied values.
// A caller must decide whether a particular operation requires this header;
// absence is never treated as verification success here. Calls are concurrency safe.
func VerifyCRC64(local uint64, server string) error {
	if len(server) == 0 || len(server) > 20 {
		return ErrInvalid
	}
	for i := range len(server) {
		if server[i] < '0' || server[i] > '9' {
			return ErrInvalid
		}
	}
	value, err := strconv.ParseUint(server, 10, 64)
	if err != nil {
		return ErrInvalid
	}
	if value != local {
		return ErrMismatch
	}
	return nil
}
