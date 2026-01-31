package utils

import (
	"os"
	"strings"
)

// GetVersion reads the version from a file, returns empty string if file doesn't exist
func GetVersion(versionFile string) string {
	data, err := os.ReadFile(versionFile)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

// ShouldUpdate returns true if the current version differs from the latest version
func ShouldUpdate(versionFile string, latestVersion string) bool {
	currentVersion := GetVersion(versionFile)
	return currentVersion != latestVersion
}

// WriteVersion writes the version string to a file
func WriteVersion(versionFile string, version string) error {
	return os.WriteFile(versionFile, []byte(version), 0644)
}
