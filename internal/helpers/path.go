package helpers

import (
	"os"
	"path/filepath"
	"strings"
)

// ExpandPath takes a path and if it has a "~/" prefix, it will expand
// it to os.UserHomeDir()
func ExpandPath(path string) (string, error) {
	if !strings.HasPrefix(path, "~/") {
		return path, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(home, path[2:]), nil
}
