package app

import (
	"os"
	"path/filepath"
	"strings"
)

// ExpandHome expands a leading "~/" to the current user's home directory.
// If home is not available, the path is returned unchanged.
func ExpandHome(path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return path
	}
	if path == "~" {
		if home, err := os.UserHomeDir(); err == nil {
			return home
		}
		return path
	}
	if strings.HasPrefix(path, "~/") {
		if home, err := os.UserHomeDir(); err == nil {
			return filepath.Join(home, strings.TrimPrefix(path, "~/"))
		}
	}
	return path
}
