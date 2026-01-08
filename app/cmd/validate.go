package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mydehq/mytm/internal/theme"
	"github.com/mydehq/mytm/internal/utils"
	"github.com/spf13/cobra"
)

var (
	validateThemeName string
)

// validateCmd represents the validate command
var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate theme directories",
	Run: func(cmd *cobra.Command, args []string) {
		// If inputDir is not set via flag, try to find "themes" in current dir
		if inputDir == "" {
			inputDir = "themes"
		}
        
        targetDir := inputDir
        if validateThemeName != "" {
            targetDir = filepath.Join(inputDir, validateThemeName)
             validateSingle(targetDir)
        } else {
             validateAll(targetDir)
        }
	},
}

func validateSingle(path string) {
    if !utils.DirExists(path) {
        fmt.Printf("Error: Theme directory '%s' not found.\n", path)
        os.Exit(1)
    }
    
    fmt.Printf("Validating %s... ", filepath.Base(path))
    if err := theme.ValidateThemeDir(path); err != nil {
        fmt.Printf("FAILED\nReason: %v\n", err)
        os.Exit(1)
    }
    fmt.Println("OK")
}

func validateAll(dir string) {
    if !utils.DirExists(dir) {
        fmt.Printf("Error: Input directory '%s' not found.\n", dir)
        os.Exit(1)
    }
    
    entries, err := os.ReadDir(dir)
    if err != nil {
        fmt.Printf("Error reading dir: %v\n", err)
        os.Exit(1)
    }
    
    failed := false
    for _, entry := range entries {
        if entry.IsDir() {
            path := filepath.Join(dir, entry.Name())
             fmt.Printf("Validating %s... ", entry.Name())
            if err := theme.ValidateThemeDir(path); err != nil {
                fmt.Printf("FAILED\n  -> %v\n", err)
                failed = true
            } else {
                fmt.Println("OK")
            }
        }
    }
    
    if failed {
        os.Exit(1)
    }
}

func init() {
	rootCmd.AddCommand(validateCmd)

	validateCmd.Flags().StringVarP(&inputDir, "input-dir", "i", "", "Input directory containing themes")
    validateCmd.Flags().StringVarP(&validateThemeName, "theme", "t", "", "Specific theme to validate")
}
