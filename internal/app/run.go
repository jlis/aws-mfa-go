package app

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"
	"time"

	"github.com/jlis/aws-mfa-go/internal/awssts"
	"github.com/jlis/aws-mfa-go/internal/credentials"
)

type STSFactory func(ctx context.Context, region, accessKeyID, secretAccessKey string) (awssts.Client, error)

type Deps struct {
	Now        func() time.Time
	Env        Env
	STSFactory STSFactory

	Stdout io.Writer
	Stderr io.Writer
	Stdin  io.Reader
}

func DefaultDeps() Deps {
	return Deps{
		Now: func() time.Time { return time.Now().UTC() },
		Env: OSEnv{},
		STSFactory: func(ctx context.Context, region, accessKeyID, secretAccessKey string) (awssts.Client, error) {
			return awssts.NewRealClient(ctx, region, accessKeyID, secretAccessKey)
		},
	}
}

var token6Digits = regexp.MustCompile(`^\d{6}$`)

type RunInputs struct {
	Inputs
	// Region is optional; if empty, we will fall back to env/default.
	Region string
}

// Run is the main entry point for the business logic.
//
// It loads the credentials store, resolves config precedence, decides whether to refresh,
// optionally prompts for an MFA token, calls STS, and writes short-term credentials.
func Run(ctx context.Context, in RunInputs, deps Deps) error {
	if deps.Now == nil || deps.Env == nil || deps.STSFactory == nil {
		return errors.New("missing required dependencies")
	}
	if deps.Stdout == nil {
		deps.Stdout = io.Discard
	}
	if deps.Stderr == nil {
		deps.Stderr = io.Discard
	}
	if deps.Stdin == nil {
		deps.Stdin = strings.NewReader("")
	}

	credsPath := ExpandHome(in.CredentialsFile)
	store, err := credentials.Load(credsPath)
	if err != nil {
		return err
	}

	resolved, err := Resolve(ctx, in.Inputs, deps.Env, store)
	if err != nil {
		return err
	}

	_, _ = fmt.Fprintf(deps.Stdout, "INFO - Using profile: %s\n", resolved.ShortTermSection)

	ltKeyID, err := store.MustGet(resolved.LongTermSection, "aws_access_key_id")
	if err != nil {
		return fmt.Errorf("long-term section [%s] missing aws_access_key_id", resolved.LongTermSection)
	}
	ltSecret, err := store.MustGet(resolved.LongTermSection, "aws_secret_access_key")
	if err != nil {
		return fmt.Errorf("long-term section [%s] missing aws_secret_access_key", resolved.LongTermSection)
	}

	now := deps.Now().UTC()
	dec := DecideRefresh(now, store, resolved.ShortTermSection, resolved.Force)
	if !dec.ShouldRefresh {
		// Match upstream-ish wording.
		if dec.Remaining != nil && dec.ExpiresAt != nil {
			_, _ = fmt.Fprintf(deps.Stdout,
				"INFO - Your credentials are still valid for %.0f seconds they will expire at %s\n",
				dec.Remaining.Seconds(),
				dec.ExpiresAt.Format(expirationLayout),
			)
			return nil
		}
		_, _ = fmt.Fprintln(deps.Stdout, "INFO - Your credentials are still valid.")
		return nil
	}

	switch dec.Reason {
	case "expired":
		_, _ = fmt.Fprintln(deps.Stdout, "INFO - Your credentials have expired, renewing.")
	case "short-term section missing":
		_, _ = fmt.Fprintf(deps.Stdout, "INFO - Short term credentials section [%s] is missing, obtaining new credentials.\n", resolved.ShortTermSection)
	default:
		_, _ = fmt.Fprintln(deps.Stdout, "INFO - Obtaining new credentials.")
	}

	token := strings.TrimSpace(resolved.Token)
	if token == "" {
		token, err = promptToken(deps.Stdout, deps.Stdin, resolved.Device, resolved.DurationSeconds)
		if err != nil {
			return err
		}
	}
	if !token6Digits.MatchString(token) {
		return errors.New("token must be six digits")
	}

	region := strings.TrimSpace(in.Region)
	if region == "" {
		region = strings.TrimSpace(deps.Env.Get("AWS_REGION"))
	}
	if region == "" {
		region = strings.TrimSpace(deps.Env.Get("AWS_DEFAULT_REGION"))
	}
	if region == "" {
		region = "us-east-1"
	}

	stsClient, err := deps.STSFactory(ctx, region, ltKeyID, ltSecret)
	if err != nil {
		return err
	}

	out, err := stsClient.GetSessionToken(ctx, awssts.GetSessionTokenInput{
		SerialNumber:    resolved.Device,
		TokenCode:       token,
		DurationSeconds: resolved.DurationSeconds,
	})
	if err != nil {
		return err
	}

	// Ensure section exists and write keys required by AWS SDKs.
	sec := resolved.ShortTermSection
	store.Section(sec)

	// Keep close to upstream: provide both session/security token keys.
	store.Set(sec, "aws_access_key_id", out.AccessKeyID)
	store.Set(sec, "aws_secret_access_key", out.SecretAccessKey)
	store.Set(sec, "aws_session_token", out.SessionToken)
	store.Set(sec, "aws_security_token", out.SessionToken)
	store.Set(sec, "expiration", out.Expiration.UTC().Format(expirationLayout))

	// Future-proofing: upstream writes this; we keep it explicit even in v1.
	store.Set(sec, "assumed_role", "False")
	store.DeleteKey(sec, "assumed_role_arn")

	if err := store.SaveAtomic(); err != nil {
		return err
	}

	_, _ = fmt.Fprintf(deps.Stdout, "INFO - Success! Your credentials will expire in %d seconds at: %s\n",
		resolved.DurationSeconds,
		out.Expiration.UTC().Format(time.RFC3339),
	)
	return nil
}

func promptToken(stdout io.Writer, stdin io.Reader, device string, duration int32) (string, error) {
	_, _ = fmt.Fprintf(stdout, "Enter AWS MFA code for device [%s] (renewing for %d seconds):", device, duration)
	r := bufio.NewReader(stdin)
	line, err := r.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return "", fmt.Errorf("read token: %w", err)
	}
	return strings.TrimSpace(line), nil
}
