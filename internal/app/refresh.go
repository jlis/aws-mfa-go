package app

import (
	"fmt"
	"time"

	"aws-mfa-go/internal/credentials"
)

const expirationLayout = "2006-01-02 15:04:05"

var shortTermRequiredKeys = []string{
	"aws_access_key_id",
	"aws_secret_access_key",
	"aws_session_token",
	"aws_security_token",
	"expiration",
}

// ParseExpiration parses the stored `expiration` value.
//
// Upstream aws-mfa stores UTC time with layout `YYYY-MM-DD HH:MM:SS`.
func ParseExpiration(exp string) (time.Time, error) {
	t, err := time.ParseInLocation(expirationLayout, exp, time.UTC)
	if err != nil {
		return time.Time{}, fmt.Errorf("parse expiration %q: %w", exp, err)
	}
	return t.UTC(), nil
}

type RefreshDecision struct {
	ShouldRefresh bool
	// ExpiresAt is only set when a valid expiration was present.
	ExpiresAt *time.Time
	// Remaining is only set when a valid expiration was present.
	Remaining *time.Duration
	Reason    string
}

// DecideRefresh determines whether we need to fetch new short-term credentials.
//
// The goal is to keep behavior close to upstream aws-mfa, but minimal:
// - If `--force`, always refresh
// - If short-term section missing, refresh
// - If required keys missing/empty/invalid, refresh
// - Otherwise, refresh only when expired
func DecideRefresh(now time.Time, store *credentials.Store, shortTermSection string, force bool) RefreshDecision {
	if force {
		return RefreshDecision{ShouldRefresh: true, Reason: "forced refresh"}
	}

	if !store.HasSection(shortTermSection) {
		return RefreshDecision{ShouldRefresh: true, Reason: "short-term section missing"}
	}

	for _, k := range shortTermRequiredKeys {
		v, ok := store.Get(shortTermSection, k)
		if !ok || v == "" {
			return RefreshDecision{ShouldRefresh: true, Reason: "short-term section missing required keys"}
		}
	}

	expStr, _ := store.Get(shortTermSection, "expiration")
	exp, err := ParseExpiration(expStr)
	if err != nil {
		return RefreshDecision{ShouldRefresh: true, Reason: "invalid expiration"}
	}

	remaining := exp.Sub(now.UTC())
	if remaining <= 0 {
		return RefreshDecision{ShouldRefresh: true, ExpiresAt: &exp, Remaining: &remaining, Reason: "expired"}
	}

	return RefreshDecision{ShouldRefresh: false, ExpiresAt: &exp, Remaining: &remaining, Reason: "still valid"}
}
