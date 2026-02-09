package credentials

import (
	"os"
	"path/filepath"
	"testing"
)

func TestStore_LoadMissing_SaveAtomic(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".aws", "credentials")

	s, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	// Populate a minimal set of keys.
	s.Set("default-long-term", "aws_access_key_id", "AKIA_TEST")
	s.Set("default-long-term", "aws_secret_access_key", "SECRET_TEST")
	s.Set("default", "aws_access_key_id", "ASIA_TEST")

	if err := s.SaveAtomic(); err != nil {
		t.Fatalf("SaveAtomic: %v", err)
	}

	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected credentials file to exist: %v", err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load (after save): %v", err)
	}

	if v, err := loaded.MustGet("default-long-term", "aws_access_key_id"); err != nil || v != "AKIA_TEST" {
		t.Fatalf("MustGet: got %q err=%v", v, err)
	}
	if v, ok := loaded.Get("default", "aws_access_key_id"); !ok || v != "ASIA_TEST" {
		t.Fatalf("Get: expected %q ok=true, got %q ok=%v", "ASIA_TEST", v, ok)
	}
}
