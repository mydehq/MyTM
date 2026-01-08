// Package theme handles validating theme directories and parsing theme manifests.
package theme

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"gopkg.in/yaml.v3"
)

// ThemeManifest represents the contents of 'theme.yml'.
type ThemeManifest struct {
	Version string                 `yaml:"version"`
	Author  string                 `yaml:"author"`
	URL     string                 `yaml:"url"`
	Config  map[string]interface{} `yaml:"config"` // 'interface{}' means any type, similar to 'any' in TS or Object in Java.
}

// ValidateThemeDir checks if a directory is a valid theme package.
// It checks for existence, directory bit, and a valid 'theme.yml'.
func ValidateThemeDir(themeDir string) error {
	// os.Stat returns file info.
	info, err := os.Stat(themeDir)
	if err != nil {
		// fmt.Errorf with %w wraps the error, preserving the original error chain.
		return fmt.Errorf("theme dir does not exist: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("path is not a directory: %s", themeDir)
	}

	// filepath.Join handles creating a safe path (e.g. themeDir/theme.yml)
	themeYAML := filepath.Join(themeDir, "theme.yml")
	if _, err := os.Stat(themeYAML); os.IsNotExist(err) {
		return fmt.Errorf("missing theme.yml in %s", themeDir)
	}

	manifest, err := parseThemeYAML(themeYAML)
	if err != nil {
		return fmt.Errorf("invalid theme.yml: %w", err)
	}

	return validateManifest(manifest)
}

// parseThemeYAML is a helper to read and unmarshal the YAML file.
func parseThemeYAML(path string) (*ThemeManifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var manifest ThemeManifest
	if err := yaml.Unmarshal(data, &manifest); err != nil {
		return nil, err
	}
	return &manifest, nil
}

// validateManifest checks if the fields in theme.yml logic are correct.
func validateManifest(m *ThemeManifest) error {
	// 1. Check Version (using Regex for Semantic Versioning X.Y.Z)
	if m.Version == "" {
		return fmt.Errorf("missing 'version'")
	}
	if matched, _ := regexp.MatchString(`^[0-9]+\.[0-9]+\.[0-9]+$`, m.Version); !matched {
		return fmt.Errorf("invalid version '%s': must be X.Y.Z", m.Version)
	}

	// 2. Check Author
	if m.Author == "" {
		return fmt.Errorf("missing 'author'")
	}

	// 3. Check URL (Simple regex start check)
	if m.URL == "" {
		return fmt.Errorf("missing 'url'")
	}
	if matched, _ := regexp.MatchString(`^https?://`, m.URL); !matched {
		return fmt.Errorf("invalid url '%s': must start with http:// or https://", m.URL)
	}

	// 4. Check Config (Must exist)
	if m.Config == nil {
		return fmt.Errorf("missing 'config' section")
	}

	return nil
}

// GetVersion is a utility to quickly extract just the version from a theme directory.
func GetVersion(themeDir string) (string, error) {
    themeYAML := filepath.Join(themeDir, "theme.yml")
    manifest, err := parseThemeYAML(themeYAML)
    if err != nil {
        return "", err
    }
    return manifest.Version, nil
}
