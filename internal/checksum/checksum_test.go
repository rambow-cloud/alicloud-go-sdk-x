package checksum_test

import (
	"bytes"
	"context"
	"errors"
	"math"
	"strconv"
	"strings"
	"testing"

	"github.com/rambow-cloud/alicloud-go-sdk-x/internal/checksum"
)

func TestOfficialCRCVectorsAndContinuation(t *testing.T) {
	for _, vector := range []struct {
		data string
		want uint64
	}{
		{"123456789", 0x995dc9bbdf1939fa},
		{"This is a test of the emergency broadcast system.", 0x27db187fc15bbc72},
	} {
		input := []byte(vector.data)
		before := bytes.Clone(input)
		value, err := checksum.CRC64(context.Background(), 0, input)
		if err != nil || value != vector.want || !bytes.Equal(input, before) {
			t.Fatalf("vector/ownership failed: %x %v", value, err)
		}
		for split := 0; split <= len(input); split++ {
			first, err := checksum.CRC64(context.Background(), 0, input[:split])
			if err != nil {
				t.Fatal(err)
			}
			second, err := checksum.CRC64(context.Background(), first, input[split:])
			if err != nil || second != vector.want {
				t.Fatalf("split %d: %x %v", split, second, err)
			}
		}
	}
	for _, seed := range []uint64{0, 1, math.MaxUint64} {
		got, err := checksum.CRC64(context.Background(), seed, nil)
		if err != nil || got != seed {
			t.Fatal("empty input changed seed", got, err)
		}
	}
}

func TestMD5KnownVectors(t *testing.T) {
	for _, vector := range []struct{ data, want string }{
		{"", "1B2M2Y8AsgTpgAmY7PhCfg=="},
		{"abc", "kAFQmDzST7DWlj99KOF/cg=="},
	} {
		input := []byte(vector.data)
		before := bytes.Clone(input)
		got, err := checksum.ContentMD5(context.Background(), input)
		if err != nil || got != vector.want || !bytes.Equal(input, before) {
			t.Fatalf("MD5/ownership failed: %s %v", got, err)
		}
	}
}

func TestServerCRCParsingAndRedaction(t *testing.T) {
	for _, local := range []uint64{0, 1, 0x995dc9bbdf1939fa, math.MaxUint64} {
		if err := checksum.VerifyCRC64(local, strconv.FormatUint(local, 10)); err != nil {
			t.Fatal(err)
		}
	}
	if err := checksum.VerifyCRC64(1, "00001"); err != nil {
		t.Fatal("leading zeros rejected", err)
	}
	for _, invalid := range []string{"", " ", "1 ", "+1", "-1", "0x1", "1.0", "１２", "18446744073709551616", strings.Repeat("0", 21), "server-payload-secret"} {
		err := checksum.VerifyCRC64(0, invalid)
		if !errors.Is(err, checksum.ErrInvalid) || (len(invalid) > 6 && strings.Contains(err.Error(), invalid)) {
			t.Fatalf("invalid header leaked/accepted: %v", err)
		}
	}
	if err := checksum.VerifyCRC64(1, "2"); !errors.Is(err, checksum.ErrMismatch) || strings.Contains(err.Error(), "1") || strings.Contains(err.Error(), "2") {
		t.Fatal("mismatch contract failed", err)
	}
}

func TestCancellationReturnsNoPartialChecksum(t *testing.T) {
	data := bytes.Repeat([]byte{'x'}, 3<<15)
	for _, ctx := range []context.Context{canceled(), &cancelInChunks{Context: context.Background()}} {
		value, err := checksum.CRC64(ctx, 77, data)
		if value != 0 || !errors.Is(err, context.Canceled) {
			t.Fatal("partial CRC/cancellation lost", value, err)
		}
	}
	for _, ctx := range []context.Context{canceled(), &cancelInChunks{Context: context.Background()}} {
		value, err := checksum.ContentMD5(ctx, data)
		if value != "" || !errors.Is(err, context.Canceled) {
			t.Fatal("partial MD5/cancellation lost", value, err)
		}
	}
	ctx := canceled()
	if _, err := checksum.CRC64(ctx, 1, nil); !errors.Is(err, context.Canceled) {
		t.Fatal("empty input cancellation lost")
	}
}

func canceled() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	return ctx
}

type cancelInChunks struct {
	context.Context
	checks int
}

func (c *cancelInChunks) Err() error {
	c.checks++
	if c.checks >= 2 {
		return context.Canceled
	}
	return nil
}
