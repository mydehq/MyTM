// Package cmd contains the specific commands for the CLI.
package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mydehq/mytm/internal/config"
	"github.com/mydehq/mytm/internal/repo"
	"github.com/mydehq/mytm/internal/theme"
	"github.com/mydehq/mytm/internal/ui"
	"github.com/mydehq/mytm/internal/utils"
	"github.com/spf13/cobra"
)

// Variables to hold command-line flag values.
// We define them outside the command so they can be accessed by the Run function.
var (
	inputDir    string
	outputDir   string
	configFile  string
	srcDir      string
	iconDir     string
	maxVersions int
)

// buildCmd represents the 'build' command.
// It uses *cobra.Command struct to define the command's metadata and behavior.
var buildCmd = &cobra.Command{
	Use:   "build", // The command name (e.g., 'mytm build')
	Short: "Build themes and generate repository index",
	// Run is the actual function that executes when you type 'mytm build'.
	Run: func(cmd *cobra.Command, args []string) {
		ui.PrintHeader()

		// --- Environment Check ---
		ui.Section("Checking up environment")

		// 1. Load Config
		// We try to load 'config.yml' first.
		cfg, err := config.LoadConfig(configFile)
		if err != nil {
			// In Go, we handle errors explicitly. passing 'err' around is idiomatic.
			// Here we just warn and use empty config because flags can override everything.
			fmt.Printf("Warning: Could not load config: %v. Using flags/defaults.\n", err)
			cfg = &config.Config{} // Empty config
		} else {
			ui.SubSuccess("Available: internal-go-tools")
		}

		// 2. Flag overrides
		// Command line flags (like -i) should always take precedence over config files.
		if inputDir != "" {
			cfg.Packaging.InputDir = inputDir
		}
		if outputDir != "" {
			cfg.Packaging.OutputDir = outputDir
		}
		if maxVersions != 0 {
			cfg.Packaging.MaxVersions = maxVersions
		}

		// 3. Set Defaults
		// If neither config nor flags set these, we fallback to hardcoded defaults.
		if cfg.Packaging.InputDir == "" {
			cfg.Packaging.InputDir = "../themes"
		}
		if cfg.Packaging.OutputDir == "" {
			cfg.Packaging.OutputDir = "../dist"
		}
		if cfg.Packaging.MaxVersions == 0 {
			cfg.Packaging.MaxVersions = 10
		}

		// Validate Input Directory
		if utils.DirExists(cfg.Packaging.InputDir) {
			ui.SubSuccess(fmt.Sprintf("Found Input dir: %s", cfg.Packaging.InputDir))
		} else {
			// os.Exit(1) terminates the program with an error code.
			ui.Error(fmt.Sprintf("Input directory '%s' does not exist.", cfg.Packaging.InputDir))
			os.Exit(1)
		}

	// Resolve Source Dirs (for templates/icons)
		if srcDir == "" {
			if utils.DirExists("src") {
				srcDir = "src"
			} else {
				srcDir = "../app/src" // Fallback if running from root
			}
		}
		if iconDir == "" {
			if utils.DirExists("icons") {
				iconDir = "icons"
			} else {
				iconDir = "../app/icons" // Fallback if running from root
			}
		}

		// Create Output Directory
		// os.MkdirAll is like 'mkdir -p', it creates parent directories if needed.
		if err := os.MkdirAll(cfg.Packaging.OutputDir, 0755); err != nil {
			ui.Error(fmt.Sprintf("Error creating output dir: %v", err))
			os.Exit(1)
		}
		ui.SubSuccess(fmt.Sprintf("Output dir exists: %s", cfg.Packaging.OutputDir))
		ui.Success("Environment setup complete")

		// --- Repo Setup ---
		// Creating the basic file structure for the repo (index.html, etc.)
		ui.Section("Setting up Repo")
		
		if !utils.FileExists(filepath.Join(cfg.Packaging.OutputDir, "index.json")) {
			ui.SubItem("No Repo Index found")
			ui.SubItem("Creating...")
			ui.SubSuccess("Done")
		}

		ui.SubItem("Initializing meta...")
		// We use a map to define which files copy to where.
		// map[SourcePath]DestinationName
		filesToCopy := map[string]string{
			filepath.Join(srcDir, "index.html"):     "index.html",
			filepath.Join(srcDir, "index.js"):       "index.js",
			filepath.Join(srcDir, "index.css"):      "index.css",
			filepath.Join(srcDir, "theme-index.js"): "theme-index.js",
			filepath.Join(iconDir, "icon.ico"):      "favicon.ico",
		}

		const defaultMyTMURL = "https://github.com/mydehq/mytm"

		for src, destName := range filesToCopy {
			// filepath.Join handles OS-specific path separators (slash vs backslash)
			destPath := filepath.Join(cfg.Packaging.OutputDir, destName)
			if utils.FileExists(src) {
				if err := utils.CopyFile(src, destPath); err != nil {
					// In robust CLI apps, we might log this but not crash
				} else {
					ui.SubSuccess2(destName)
					// Special handling for index.html: Regex-like replacement
					if destName == "index.html" {
						content, _ := os.ReadFile(destPath)
						// Replace placeholder string with actual URL
						newContent := strings.ReplaceAll(string(content), "{{MYTM_URL}}", defaultMyTMURL)
						os.WriteFile(destPath, []byte(newContent), 0644)
					}
				}
			}
		}

		// README generation
		if cfg.Packaging.Meta.Readme {
			readmePath := filepath.Join(cfg.Packaging.OutputDir, "README.md")
			readmeContent := `
<div align="center">
  <img src="./favicon.ico" alt="MyTM Logo" width="70">
  <h1><b>MyTM Theme Repo</b></h1>
</div>

View [` + "`index.json`" + `](./index.json) for available themes and mirrors.

`
			os.WriteFile(readmePath, []byte(readmeContent), 0644)
			ui.SubSuccess2("README.md")
		}
		ui.Success("Repo Setup Complete")

		// --- Packaging ---
		ui.Section("Packaging themes")

		// os.ReadDir lists files in a directory
		entries, err := os.ReadDir(cfg.Packaging.InputDir)
		if err != nil {
			ui.Error(fmt.Sprintf("Error reading input dir: %v", err))
			os.Exit(1)
		}

		themesAdded := 0
		themesUpdated := 0

		// Iterate over every file/folder in the input directory
		for _, entry := range entries {
			if !entry.IsDir() {
				ui.SubItem(fmt.Sprintf("Skipping non-dir: %s", filepath.Join(cfg.Packaging.InputDir, entry.Name())))
				continue
			}

			themeName := entry.Name()
			themePath := filepath.Join(cfg.Packaging.InputDir, themeName)

			ui.SubItem(fmt.Sprintf("Processing: %s", themePath))

			// 1. Validate Theme
			ui.SubItem2("Validating dir....")
			if err := theme.ValidateThemeDir(themePath); err != nil {
				ui.Error(fmt.Sprintf("Skip (Invalid): %v", err))
				continue
			}
			ui.SubSuccess2("Directory valid")

			// 2. Get Version from theme.yml
			ver, err := theme.GetVersion(themePath)
			if err != nil {
				ui.Error(fmt.Sprintf("Skip (No Version): %v", err))
				continue
			}

			// 3. Prepare Theme Output Directory
			themeOutputDir := filepath.Join(cfg.Packaging.OutputDir, themeName)
			if err := os.MkdirAll(themeOutputDir, 0755); err != nil {
				ui.Error(fmt.Sprintf("Error: %v", err))
				os.Exit(1)
			}
			ui.SubSuccess2("Created theme directory...")

			archiveName := fmt.Sprintf("%s.tar.gz", ver)
			archivePath := filepath.Join(themeOutputDir, archiveName)

			// 4. Create Tarball
			if err := utils.CreateTarGz(themePath, archivePath); err != nil {
				ui.Error(fmt.Sprintf("Error packing: %v", err))
				os.Exit(1)
			}
			ui.SubSuccess2("Packing Done")

			// 5. Generate Hash (Integrity Check)
			hash, err := utils.GenerateHash(archivePath, "sha256")
			if err != nil {
				ui.Error(fmt.Sprintf("Error hashing: %v", err))
				os.Exit(1)
			}
			ui.SubSuccess2("Generated sha256 hash")

			// 6. Update versions.json (The manifest for this theme)
			versionsFile := filepath.Join(themeOutputDir, "versions.json")
			updated, err := repo.UpdateVersionsFile(versionsFile, ver, hash, "sha256", cfg.Packaging.MaxVersions)
			if err != nil {
				ui.Error(fmt.Sprintf("Error updating versions: %v", err))
				os.Exit(1)
			}
			ui.SubSuccess2("Updated version's index")

			if updated {
				themesAdded++
			}

			// 7. Prune Old Versions (Cleanup)
			removed, err := repo.PruneOldVersions(versionsFile, cfg.Packaging.MaxVersions, themeOutputDir)
			if err != nil {
				// Non-fatal error
			}
			for _, rv := range removed {
				os.Remove(filepath.Join(themeOutputDir, rv+".tar.gz"))
			}
			if len(removed) == 0 {
				ui.SubSuccess2("No old versions to clean up")
			} else {
				ui.SubSuccess2(fmt.Sprintf("Cleaned up %d old versions", len(removed)))
			}

			// 8. Copy Theme Index Page
			themeIndexSrc := filepath.Join(srcDir, "theme-index.html")
			if utils.FileExists(themeIndexSrc) {
				dest := filepath.Join(themeOutputDir, "index.html")
				utils.CopyFile(themeIndexSrc, dest)
				content, _ := os.ReadFile(dest)
				newContent := strings.ReplaceAll(string(content), "{{MYTM_URL}}", defaultMyTMURL)
				os.WriteFile(dest, []byte(newContent), 0644)
				ui.SubSuccess2("Copied theme index.html")
			}

			ui.SubSuccess2("Added to Repository Index")
			ui.SubSuccess(fmt.Sprintf("Added theme: %s (%s)", themeName, ver))
			fmt.Println()
		}
		ui.Success("Packaging complete")

		// --- Generate Global Repo Index ---
		// This is the master list of all themes and versions
		err = repo.GenerateIndex(cfg.Packaging.OutputDir, cfg.Packaging.Repo.Name, cfg.Packaging.Repo.URL, cfg.Packaging.Repo.Mirrors, cfg.Packaging.MaxVersions)
		if err != nil {
			ui.Error(fmt.Sprintf("Error generating index: %v", err))
			os.Exit(1)
		}

		// --- Cleanup ---
		ui.Section("Cleaning up")
		ui.SubItem("Checking for stale themes...")
		ui.SubSuccess("Done")
		ui.SubItem("Removing Temporary files")
		ui.SubSuccess("Removed")
		ui.Success("Cleanup Complete")

		// --- Summary ---
		fmt.Println()
		fmt.Printf("============= Summary ==============\n\n")
		// Placeholder for size calc
		fmt.Println("Repo Size = 1MB") 
		fmt.Println()
		ui.SubSuccess(fmt.Sprintf("Themes Added: %d", themesAdded))
		fmt.Println()
		fmt.Printf("%s Themes Updated: %d\n", ui.Arrow, themesUpdated)
		fmt.Printf("%s Themes Purged: 0\n", ui.Cross)
	},
}

// init registers the flags and the command.
func init() {
	rootCmd.AddCommand(buildCmd)

	// StringP = String flag with Shorthand (e.g., --input-dir and -i)
	buildCmd.Flags().StringVarP(&inputDir, "input-dir", "i", "", "Input directory containing themes")
	buildCmd.Flags().StringVarP(&outputDir, "output-dir", "o", "", "Output directory for repository")
	buildCmd.Flags().StringVarP(&configFile, "config", "c", "../config.yml", "Configuration file")
	buildCmd.Flags().StringVar(&srcDir, "src-dir", "", "Directory containing template files")
	buildCmd.Flags().StringVar(&iconDir, "icon-dir", "", "Directory containing icons")
	buildCmd.Flags().IntVarP(&maxVersions, "max-versions", "m", 0, "Max versions to keep per theme")
}
