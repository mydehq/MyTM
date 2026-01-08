// Package repo handles the generation of repository metadata (index.json, versions.json).
package repo

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/mydehq/mytm/internal/utils"
)

// Index represents the main repository entry point 'index.json'.
// This file lists all available themes and their latest versions.
type Index struct {
	SchemaVer   int              `json:"schema_ver"`   // Schema version for compatibility
	RepoName    string           `json:"repo_name"`    // Human readable repo name
	RepoURL     string           `json:"repo_url"`     // The URL where this repo is hosted
	ReleaseTime int64            `json:"release"`      // Unix timestamp of generation
	MaxVersions int              `json:"max_versions"` // Max versions kept per theme
	Mirrors     []string         `json:"mirrors"`      // List of mirror URLs
	Themes      map[string]Theme `json:"themes"`       // Map of theme_name -> theme_info
}

// Theme represents the simplified theme info in index.json.
type Theme struct {
	Latest string `json:"latest"` // Latest version tag (e.g. "1.2.3")
}

// VersionEntry represents a single version in 'versions.json'.
type VersionEntry struct {
	Ver  string `json:"ver"`  // The semantic version string
	Hash Hash   `json:"hash"` // Checksum of the .tar.gz
}

// Hash holds the checksum value and algorithm.
type Hash struct {
	Value string `json:"value"`
	Algo  string `json:"algo"`
}

// UpdateVersionsFile reads versions.json, adds/updates the given version, and writes it back.
// It also handles truncating the list if it exceeds maxVersions.
// Returns 'true' if the file was modified.
func UpdateVersionsFile(versionsFile, version, hashStr, algo string, maxVersions int) (bool, error) {
	var versions []VersionEntry

	// 1. Read existing file
	if utils.FileExists(versionsFile) {
		data, err := os.ReadFile(versionsFile)
		if err != nil {
			return false, err
		}
		if len(data) > 0 {
			if err := json.Unmarshal(data, &versions); err != nil {
				return false, fmt.Errorf("invalid versions.json: %w", err)
			}
		}
	}

	updated := false
	found := false

	// 2. Check if this version already exists. If so, update its hash.
	for i, v := range versions {
		if v.Ver == version {
			found = true
			if v.Hash.Value != hashStr {
				versions[i].Hash.Value = hashStr
				versions[i].Hash.Algo = algo
				updated = true
			}
			break
		}
	}

	// 3. If new version, Prepend it (Add to top of list)
	if !found {
		versions = append([]VersionEntry{{
			Ver:  version,
			Hash: Hash{Value: hashStr, Algo: algo},
		}}, versions...)
		updated = true
	}

	// 4. Truncate list if too long (keep only newest N versions)
	if len(versions) > maxVersions {
		versions = versions[:maxVersions]
	}

	// 5. Write back to disk
	if updated {
		// MarshalIndent creates pretty-printed JSON
		data, err := json.MarshalIndent(versions, "", "  ")
		if err != nil {
			return false, err
		}
		if err := os.WriteFile(versionsFile, data, 0644); err != nil {
			return false, err
		}
	}

	return updated, nil
}

// PruneOldVersions removes .tar.gz files that are no longer listed in versions.json.
// This happens when UpdateVersionsFile truncates the list.
func PruneOldVersions(versionsFile string, maxVersions int, archiveDir string) ([]string, error) {
	if !utils.FileExists(versionsFile) {
		return nil, nil
	}

	data, err := os.ReadFile(versionsFile)
	if err != nil {
		return nil, err
	}
	var versions []VersionEntry
	if err := json.Unmarshal(data, &versions); err != nil {
		return nil, err
	}

	var removed []string
	if len(versions) > maxVersions {
		// Logic here repeats truncation to identify *what* was removed.
		// ideally UpdateVersionsFile could return this, but keeping functions simpler for now.
		toRemove := versions[maxVersions:] // The tail of the list
		versions = versions[:maxVersions]  // The kept head

		// Write the truncated list back (This is technically duplicate write but safe)
		newData, err := json.MarshalIndent(versions, "", "  ")
		if err != nil {
			return nil, err
		}
		if err := os.WriteFile(versionsFile, newData, 0644); err != nil {
			return nil, err
		}

		for _, v := range toRemove {
			removed = append(removed, v.Ver)
		}
	}

	return removed, nil
}

// GenerateIndex builds the master 'index.json' by scanning all theme directories.
func GenerateIndex(outputDir string, repoName, repoURL string, mirrors []string, maxVersions int) error {
	indexFile := filepath.Join(outputDir, "index.json")

	themes := make(map[string]Theme)

	entries, err := os.ReadDir(outputDir)
	if err != nil {
		return err
	}

	// Iterate through "dist" directory to find themes
	for _, entry := range entries {
		if entry.IsDir() {
			themeName := entry.Name()
			versionsFile := filepath.Join(outputDir, themeName, "versions.json")
			
			// If a theme has a versions.json, we consider it valid.
			if utils.FileExists(versionsFile) {
				data, err := os.ReadFile(versionsFile)
				if err == nil {
					var v []VersionEntry
					// We assume versions.json is sorted (newest first) because UpdateVersionsFile does that.
					if json.Unmarshal(data, &v) == nil && len(v) > 0 {
						themes[themeName] = Theme{Latest: v[0].Ver}
					}
				}
			}
		}
	}

	idx := Index{
		SchemaVer:   2,
		RepoName:    repoName,
		RepoURL:     repoURL,
		ReleaseTime: time.Now().Unix(),
		MaxVersions: maxVersions,
		Mirrors:     mirrors,
		Themes:      themes,
	}

	data, err := json.MarshalIndent(idx, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(indexFile, data, 0644)
}
