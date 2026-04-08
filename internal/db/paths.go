package db

import (
	"os"
	"path/filepath"
	"strings"
)

func ResolvePath(explicitFlag string) (string, error) {
	if p := strings.TrimSpace(explicitFlag); p != "" {
		return p, nil
	}
	if p := strings.TrimSpace(os.Getenv("DOGEAR_DB")); p != "" {
		return p, nil
	}
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "dogear", "dogear.db"), nil
}
