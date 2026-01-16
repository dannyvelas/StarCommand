package helpers

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var FallbackFile = "config/all.yml"

func GetConfigPath(hostAlias string) string {
	return fmt.Sprintf("config/%s.yml", hostAlias)
}

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
