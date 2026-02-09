package app

import (
	"context"
	"path/filepath"
	"testing"

	"aws-mfa-go/internal/credentials"
)

type mapEnv map[string]string

func (m mapEnv) Get(key string) string { return m[key] }

func TestResolve_DevicePrecedence(t *testing.T) {
	dir := t.TempDir()
	store, err := credentials.Load(filepath.Join(dir, "credentials"))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	store.Set("default-long-term", "aws_mfa_device", "from-store")

	in := Inputs{
		Profile:        "default",
		ProfileChanged: true,
		Device:         "from-flag",
		DeviceChanged:  true,
		LongTermSuffix: "long-term",
	}

	env := mapEnv{"MFA_DEVICE": "from-env"}

	got, err := Resolve(context.Background(), in, env, store)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if got.Device != "from-flag" {
		t.Fatalf("expected flag to win, got %q", got.Device)
	}
}

func TestResolve_DeviceEnvThenStore(t *testing.T) {
	dir := t.TempDir()
	store, err := credentials.Load(filepath.Join(dir, "credentials"))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	store.Set("default-long-term", "aws_mfa_device", "from-store")

	in := Inputs{
		Profile:        "default",
		ProfileChanged: true,
		Device:         "",
		DeviceChanged:  false,
		LongTermSuffix: "long-term",
	}

	got, err := Resolve(context.Background(), in, mapEnv{"MFA_DEVICE": "from-env"}, store)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if got.Device != "from-env" {
		t.Fatalf("expected env to win, got %q", got.Device)
	}

	got2, err := Resolve(context.Background(), in, mapEnv{}, store)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if got2.Device != "from-store" {
		t.Fatalf("expected store to be used, got %q", got2.Device)
	}
}

func TestResolve_DurationDefaultAndEnv(t *testing.T) {
	dir := t.TempDir()
	store, err := credentials.Load(filepath.Join(dir, "credentials"))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	// Need device to resolve.
	store.Set("default-long-term", "aws_mfa_device", "device")

	in := Inputs{
		Profile:        "default",
		ProfileChanged: true,
		LongTermSuffix: "long-term",
	}

	got, err := Resolve(context.Background(), in, mapEnv{}, store)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if got.DurationSeconds != 43200 {
		t.Fatalf("expected default duration 43200, got %d", got.DurationSeconds)
	}

	got2, err := Resolve(context.Background(), in, mapEnv{"MFA_STS_DURATION": "1800"}, store)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if got2.DurationSeconds != 1800 {
		t.Fatalf("expected env duration 1800, got %d", got2.DurationSeconds)
	}
}
