// Package config handles loading and parsing the application configuration.
package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

// Config represents the top-level structure of 'config.yml'.
// We use "Struct Tags" (the text in backticks like `yaml:"..."`) to tell the YAML parser
// how to map the YAML keys to our Go struct fields.
type Config struct {
	Packaging PackagingConfig `yaml:"packaging"`
}

// PackagingConfig holds settings related to theme packaging.
type PackagingConfig struct {
	InputDir    string     `yaml:"input-dir"`
	OutputDir   string     `yaml:"output-dir"`
	MaxVersions int        `yaml:"max-versions"` // How many versions to keep (e.g., 5)
	Meta        MetaConfig `yaml:"meta"`
    Repo        RepoConfig `yaml:"repo"`
}

// MetaConfig controls which meta files are generated.
type MetaConfig struct {
	Readme    bool `yaml:"readme"`     // Should we generate a README.md?
	IndexHTML bool `yaml:"index-html"` // Should we generate a web index?
}

// RepoConfig holds metadata about the repository itself.
type RepoConfig struct {
    Name    string   `yaml:"name"`    // Display name of the repo (e.g. "My Themes")
    URL     string   `yaml:"url"`     // The base URL where this repo is hosted
    Branch  string   `yaml:"branch"`  // Git branch (optional)
    Mirrors []string `yaml:"mirrors"` // List of mirror URLs
}

// LoadConfig reads and parses the config file at 'path'.
// It returns a pointer to the Config struct and an error if it fails.
func LoadConfig(path string) (*Config, error) {
	// os.ReadFile reads the entire file into memory as a byte slice.
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err // Return the error up the stack
	}

	var cfg Config
	// yaml.Unmarshal parses the byte slice ('data') into the struct pointer ('&cfg').
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
