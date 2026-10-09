package externalcreds

import (
	"bytes"
	"context"
	"encoding/json/v2"
	"errors"
	"io"
	"os/exec"
	"strings"
	"time"

	"github.com/rambow-cloud/alicloud-go-sdk-x/credentials"
)

// ProcessOptions configures an explicitly registered credential command.
type ProcessOptions struct {
	// Retrieval controls total timeout and stdout bound; HTTPClient is unused.
	Retrieval Options
	// AllowLongLived explicitly permits AK process results; false requires renewable STS.
	AllowLongLived bool
}

// ProcessProvider runs an explicit argv command without invoking a shell.
// Construct with NewProcessProvider; concurrent calls launch separate commands.
// Wrap in credentials.Cache to coalesce renewal. Stdout is bounded and stderr is
// discarded; stdin is disconnected. Configure a trusted credential executable.
type ProcessProvider struct {
	command        []string
	settings       settings
	allowLongLived bool
}

// NewProcessProvider copies nonempty argv and validates limits without execution.
// Results use the native CLI External profile schema: mode, access_key_id,
// access_key_secret, sts_token and sts_expiration (Unix seconds). No recursive
// profile, process or network discovery is allowed from a process response.
func NewProcessProvider(command []string, options ProcessOptions) (*ProcessProvider, error) {
	if len(command) == 0 || strings.TrimSpace(command[0]) == "" {
		return nil, errors.New("externalcreds: credential executable required")
	}
	for _, arg := range command {
		if strings.ContainsRune(arg, 0) {
			return nil, errors.New("externalcreds: invalid process argument")
		}
	}
	s, err := configure(options.Retrieval, false)
	if err != nil {
		return nil, err
	}
	return &ProcessProvider{command: append([]string{}, command...), settings: s, allowLongLived: options.AllowLongLived}, nil
}

// String returns a redacted representation without executable paths or arguments.
func (*ProcessProvider) String() string { return "ProcessProvider(<redacted>)" }

// GoString returns a redacted representation for Go-syntax formatting.
func (p *ProcessProvider) GoString() string { return p.String() }

type limitedOutput struct {
	bytes.Buffer
	limit int64
}

func (w *limitedOutput) Write(b []byte) (int, error) {
	if int64(w.Len())+int64(len(b)) > w.limit {
		return 0, errors.New("externalcreds: process output exceeds bound")
	}
	return w.Buffer.Write(b)
}

// Retrieve runs one bounded credential process and validates its temporary result.
// Process errors omit command, stdout and stderr. Cancellation/deadline identity
// is preserved. The direct process is terminated on timeout; WaitDelay also bounds
// inherited output pipes. No interactive input, cache or retry occurs here.
func (p *ProcessProvider) Retrieve(ctx context.Context) (credentials.Credentials, error) {
	if err := ctx.Err(); err != nil {
		return credentials.Credentials{}, err
	}
	if p == nil || len(p.command) == 0 {
		return credentials.Credentials{}, errors.New("externalcreds: unconfigured process provider")
	}
	ctx, cancel := context.WithTimeout(ctx, p.settings.timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, p.command[0], p.command[1:]...)
	hideProcessWindow(cmd)
	cmd.WaitDelay = 250 * time.Millisecond
	stdout := &limitedOutput{limit: p.settings.limit}
	cmd.Stdout = stdout
	cmd.Stderr = io.Discard
	err := cmd.Run()
	if e := ctx.Err(); e != nil {
		return credentials.Credentials{}, e
	}
	if err != nil {
		return credentials.Credentials{}, &processError{cause: err}
	}
	var out struct {
		Mode            string `json:"mode"`
		AccessKeyID     string `json:"access_key_id"`
		AccessKeySecret string `json:"access_key_secret"`
		Token           string `json:"sts_token"`
		Expiration      int64  `json:"sts_expiration"`
	}
	if json.Unmarshal(stdout.Bytes(), &out) != nil {
		return credentials.Credentials{}, errors.New("externalcreds: invalid process credential JSON")
	}
	if strings.EqualFold(out.Mode, "AK") && p.allowLongLived {
		source, err := credentials.NewStaticProvider(credentials.Credentials{AccessKeyID: out.AccessKeyID, AccessKeySecret: out.AccessKeySecret, Source: "External.AK"})
		if err != nil {
			return credentials.Credentials{}, err
		}
		return source.Retrieve(ctx)
	}
	if !strings.EqualFold(out.Mode, "StsToken") {
		return credentials.Credentials{}, errors.New("externalcreds: process must return explicitly allowed AK or StsToken mode")
	}
	return (wireCredentials{AccessKeyID: out.AccessKeyID, AccessKeySecret: out.AccessKeySecret, SecurityToken: out.Token, ExpirationInt64: out.Expiration}).snapshot("External.Process")
}

type processError struct{ cause error }

func (*processError) Error() string   { return "externalcreds: credential process failed" }
func (e *processError) Unwrap() error { return e.cause }

// ParseCommand splits a native process_command into argv without shell evaluation.
// It supports whitespace, single/double quotes and escaped quotes/backslashes
// inside double quotes. Windows path backslashes remain literal otherwise.
// Empty arguments are preserved. Unterminated quotes and an empty executable fail.
// Returned argv is caller-owned; variables and command substitutions stay literal.
func ParseCommand(command string) ([]string, error) {
	var args []string
	var b strings.Builder
	var quote rune
	started := false
	runes := []rune(command)
	for i := 0; i < len(runes); i++ {
		c := runes[i]
		if quote != 0 {
			if c == quote {
				quote = 0
			} else if c == '\\' && quote == '"' && i+1 < len(runes) && (runes[i+1] == '"' || runes[i+1] == '\\') {
				i++
				b.WriteRune(runes[i])
			} else {
				b.WriteRune(c)
			}
			continue
		}
		switch c {
		case '\'', '"':
			quote = c
			started = true
		case ' ', '\t', '\r', '\n':
			if started {
				args = append(args, b.String())
				b.Reset()
				started = false
			}
		default:
			b.WriteRune(c)
			started = true
		}
	}
	if quote != 0 {
		return nil, errors.New("externalcreds: unterminated process argument")
	}
	if started {
		args = append(args, b.String())
	}
	if len(args) == 0 || args[0] == "" {
		return nil, errors.New("externalcreds: credential executable required")
	}
	return args, nil
}
