package credentials

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/ini.v1"
)

// Store represents an AWS-style shared credentials file (INI).
//
// We purposely keep this package “dumb”: it loads/saves INI and provides helpers.
// Higher-level logic (profile selection, refresh decisions) lives elsewhere.
type Store struct {
	path string
	ini  *ini.File
}

// Load reads the credentials file at path. If the file does not exist, an empty
// store is returned (SaveAtomic will create the file).
func Load(path string) (*Store, error) {
	if path == "" {
		return nil, errors.New("credentials file path is empty")
	}

	opts := ini.LoadOptions{
		// AWS files are typically lowercase keys; keep the same behavior.
		Insensitive:         true,
		IgnoreInlineComment: false,
	}

	cfg := ini.Empty(opts)

	// If it doesn't exist, we still return an empty file so the caller can create it.
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return &Store{path: path, ini: cfg}, nil
		}
		return nil, fmt.Errorf("stat credentials file: %w", err)
	}

	loaded, err := ini.LoadSources(opts, path)
	if err != nil {
		return nil, fmt.Errorf("read/parse credentials file: %w", err)
	}
	return &Store{path: path, ini: loaded}, nil
}

func (s *Store) Path() string { return s.path }

func (s *Store) HasSection(name string) bool {
	_, err := s.ini.GetSection(name)
	return err == nil
}

// Section returns the named section, creating it if needed.
func (s *Store) Section(name string) *ini.Section {
	// ini.v1 creates sections on-demand.
	return s.ini.Section(name)
}

func (s *Store) Get(section, key string) (string, bool) {
	sec := s.ini.Section(section)
	if !sec.HasKey(key) {
		return "", false
	}
	return strings.TrimSpace(sec.Key(key).String()), true
}

func (s *Store) MustGet(section, key string) (string, error) {
	v, ok := s.Get(section, key)
	if !ok || v == "" {
		return "", fmt.Errorf("missing %q in [%s]", key, section)
	}
	return v, nil
}

func (s *Store) Set(section, key, value string) {
	s.ini.Section(section).Key(key).SetValue(value)
}

func (s *Store) DeleteKey(section, key string) {
	s.ini.Section(section).DeleteKey(key)
}

// WriteTo writes the INI to the provided writer.
func (s *Store) WriteTo(w io.Writer) (int64, error) {
	n, err := s.ini.WriteTo(w)
	if err != nil {
		return n, fmt.Errorf("write ini: %w", err)
	}
	return n, nil
}

// SaveAtomic writes the credentials file to disk using an atomic rename.
// This reduces the chance of leaving a partially-written credentials file.
func (s *Store) SaveAtomic() error {
	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("ensure credentials dir: %w", err)
	}

	perm := os.FileMode(0o600)
	if st, err := os.Stat(s.path); err == nil {
		perm = st.Mode().Perm()
	}

	tmp, err := os.CreateTemp(dir, "aws-mfa-go-credentials-*.tmp")
	if err != nil {
		return fmt.Errorf("create temp credentials file: %w", err)
	}
	tmpName := tmp.Name()
	defer func() {
		_ = os.Remove(tmpName)
	}()

	if err := tmp.Chmod(perm); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("chmod temp credentials file: %w", err)
	}

	if _, err := s.WriteTo(tmp); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close temp credentials file: %w", err)
	}

	// Atomic on POSIX when in same directory.
	if err := os.Rename(tmpName, s.path); err != nil {
		return fmt.Errorf("replace credentials file: %w", err)
	}
	return nil
}
