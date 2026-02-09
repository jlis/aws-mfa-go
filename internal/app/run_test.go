package app

import (
	"bytes"
	"context"
	"path/filepath"
	"testing"
	"time"

	"aws-mfa-go/internal/awssts"
	"aws-mfa-go/internal/credentials"
)

type fakeSTS struct {
	calls int
	out   awssts.GetSessionTokenOutput
	err   error
}

func (f *fakeSTS) GetSessionToken(ctx context.Context, in awssts.GetSessionTokenInput) (awssts.GetSessionTokenOutput, error) {
	f.calls++
	return f.out, f.err
}

func TestRun_RefreshWritesCredentials(t *testing.T) {
	dir := t.TempDir()
	credsPath := filepath.Join(dir, "credentials")

	store, err := credentials.Load(credsPath)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	// Long-term config + device in long-term section.
	store.Set("default-long-term", "aws_access_key_id", "AKIA_LT")
	store.Set("default-long-term", "aws_secret_access_key", "SECRET_LT")
	store.Set("default-long-term", "aws_mfa_device", "arn:aws:iam::123456789012:mfa/me")
	if err := store.SaveAtomic(); err != nil {
		t.Fatalf("SaveAtomic: %v", err)
	}

	exp := time.Date(2026, 2, 9, 12, 0, 0, 0, time.UTC)
	fake := &fakeSTS{
		out: awssts.GetSessionTokenOutput{
			AccessKeyID:     "ASIA_ST",
			SecretAccessKey: "SECRET_ST",
			SessionToken:    "TOKEN_ST",
			Expiration:      exp,
		},
	}

	var stdout bytes.Buffer
	deps := DefaultDeps()
	deps.Env = mapEnv{"AWS_REGION": "us-east-1"} // deterministic
	deps.Now = func() time.Time { return time.Date(2026, 2, 9, 11, 0, 0, 0, time.UTC) }
	deps.Stdout = &stdout
	deps.Stderr = &stdout
	deps.STSFactory = func(ctx context.Context, region, accessKeyID, secretAccessKey string) (awssts.Client, error) {
		return fake, nil
	}

	err = Run(context.Background(), RunInputs{
		Inputs: Inputs{
			Profile:         "default",
			ProfileChanged:  true,
			LongTermSuffix:  "long-term",
			ShortTermSuffix: "none",
			CredentialsFile: credsPath,
			Token:           "123456",
			TokenChanged:    true,
			Force:           true,
		},
	}, deps)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if fake.calls != 1 {
		t.Fatalf("expected STS call once, got %d", fake.calls)
	}

	updated, err := credentials.Load(credsPath)
	if err != nil {
		t.Fatalf("Load updated: %v", err)
	}
	if v, _ := updated.Get("default", "aws_access_key_id"); v != "ASIA_ST" {
		t.Fatalf("expected short-term key id, got %q", v)
	}
	if v, _ := updated.Get("default", "aws_security_token"); v != "TOKEN_ST" {
		t.Fatalf("expected security token, got %q", v)
	}
	if v, _ := updated.Get("default", "expiration"); v != exp.Format(expirationLayout) {
		t.Fatalf("expected expiration %q, got %q", exp.Format(expirationLayout), v)
	}
}

func TestRun_SkipsRefreshWhenStillValid(t *testing.T) {
	dir := t.TempDir()
	credsPath := filepath.Join(dir, "credentials")

	store, err := credentials.Load(credsPath)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	store.Set("default-long-term", "aws_access_key_id", "AKIA_LT")
	store.Set("default-long-term", "aws_secret_access_key", "SECRET_LT")
	store.Set("default-long-term", "aws_mfa_device", "device")

	now := time.Date(2026, 2, 9, 10, 0, 0, 0, time.UTC)
	exp := time.Date(2026, 2, 9, 11, 0, 0, 0, time.UTC)

	store.Set("default", "aws_access_key_id", "ASIA_OLD")
	store.Set("default", "aws_secret_access_key", "SECRET_OLD")
	store.Set("default", "aws_session_token", "TOKEN_OLD")
	store.Set("default", "aws_security_token", "TOKEN_OLD")
	store.Set("default", "expiration", exp.Format(expirationLayout))

	if err := store.SaveAtomic(); err != nil {
		t.Fatalf("SaveAtomic: %v", err)
	}

	fake := &fakeSTS{}
	deps := DefaultDeps()
	deps.Env = mapEnv{"AWS_REGION": "us-east-1"}
	deps.Now = func() time.Time { return now }
	deps.Stdout = ioDiscard{}
	deps.Stderr = ioDiscard{}
	deps.STSFactory = func(ctx context.Context, region, accessKeyID, secretAccessKey string) (awssts.Client, error) {
		return fake, nil
	}

	err = Run(context.Background(), RunInputs{
		Inputs: Inputs{
			Profile:         "default",
			ProfileChanged:  true,
			LongTermSuffix:  "long-term",
			ShortTermSuffix: "none",
			CredentialsFile: credsPath,
			Force:           false,
		},
	}, deps)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if fake.calls != 0 {
		t.Fatalf("expected no STS calls, got %d", fake.calls)
	}
}

// ioDiscard is a tiny io.Writer used to avoid importing io in this test file.
type ioDiscard struct{}

func (ioDiscard) Write(p []byte) (int, error) { return len(p), nil }
