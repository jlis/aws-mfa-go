package app

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/jlis/aws-mfa-go/internal/credentials"
)

func TestParseExpirationUTC(t *testing.T) {
	got, err := ParseExpiration("2026-02-09 12:34:56")
	if err != nil {
		t.Fatalf("ParseExpiration: %v", err)
	}
	if got.Location() != time.UTC {
		t.Fatalf("expected UTC, got %v", got.Location())
	}
	if got.Year() != 2026 || got.Month() != time.February || got.Day() != 9 || got.Hour() != 12 {
		t.Fatalf("unexpected parsed time: %v", got)
	}
}

func TestDecideRefresh_MissingSection(t *testing.T) {
	dir := t.TempDir()
	store, err := credentials.Load(filepath.Join(dir, "credentials"))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	dec := DecideRefresh(time.Now().UTC(), store, "default", false)
	if !dec.ShouldRefresh {
		t.Fatalf("expected refresh when section missing")
	}
}

func TestDecideRefresh_StillValid(t *testing.T) {
	dir := t.TempDir()
	store, err := credentials.Load(filepath.Join(dir, "credentials"))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	now := time.Date(2026, 2, 9, 10, 0, 0, 0, time.UTC)
	exp := time.Date(2026, 2, 9, 11, 0, 0, 0, time.UTC)

	sec := "default"
	store.Set(sec, "aws_access_key_id", "ASIA_TEST")
	store.Set(sec, "aws_secret_access_key", "SECRET")
	store.Set(sec, "aws_session_token", "TOKEN")
	store.Set(sec, "aws_security_token", "TOKEN")
	store.Set(sec, "expiration", exp.Format(expirationLayout))

	dec := DecideRefresh(now, store, sec, false)
	if dec.ShouldRefresh {
		t.Fatalf("expected no refresh when still valid, got reason=%q", dec.Reason)
	}
	if dec.ExpiresAt == nil || !dec.ExpiresAt.Equal(exp) {
		t.Fatalf("expected expiresAt=%v, got %v", exp, dec.ExpiresAt)
	}
	if dec.Remaining == nil || *dec.Remaining != (exp.Sub(now)) {
		t.Fatalf("expected remaining=%v, got %v", exp.Sub(now), dec.Remaining)
	}
}

func TestDecideRefresh_Expired(t *testing.T) {
	dir := t.TempDir()
	store, err := credentials.Load(filepath.Join(dir, "credentials"))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	now := time.Date(2026, 2, 9, 12, 0, 0, 0, time.UTC)
	exp := time.Date(2026, 2, 9, 11, 0, 0, 0, time.UTC)

	sec := "default"
	store.Set(sec, "aws_access_key_id", "ASIA_TEST")
	store.Set(sec, "aws_secret_access_key", "SECRET")
	store.Set(sec, "aws_session_token", "TOKEN")
	store.Set(sec, "aws_security_token", "TOKEN")
	store.Set(sec, "expiration", exp.Format(expirationLayout))

	dec := DecideRefresh(now, store, sec, false)
	if !dec.ShouldRefresh {
		t.Fatalf("expected refresh when expired")
	}
}

func TestDecideRefresh_Force(t *testing.T) {
	dir := t.TempDir()
	store, err := credentials.Load(filepath.Join(dir, "credentials"))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	dec := DecideRefresh(time.Now().UTC(), store, "default", true)
	if !dec.ShouldRefresh {
		t.Fatalf("expected refresh when forced")
	}
}
