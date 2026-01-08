/*
Copyright © 2026 soymadip
*/
// Package cmd handles the command-line interface logic.
// It uses the 'Cobra' library, which is a standard for Go CLIs (used by Kubernetes, Docker, etc.).
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
// This is what runs when you type just 'mytm'.
var rootCmd = &cobra.Command{
	Use:   "mytm",
	Short: "MyTM Theme Packager and Repository Manager",
	Long: `MyTM is a CLI tool for packaging, validating, and managing MyDE themes.
It handles wrapping themes into .tar.gz archives, verifying their structure,
and generating repository indices (index.json) for distribution.`,
	// Run: func(cmd *cobra.Command, args []string) { }, // We don't have a default action, so we leave this commented out.
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	// Execute() parses the command line flags and runs the appropriate function.
	err := rootCmd.Execute()
	if err != nil {
		// If there is an error (like a wrong flag), we exit with status 1 (failure).
		os.Exit(1)
	}
}

// init() is a special Go function that runs AUTOMATICALLY before main().
// It's used to set up things like flags (options).
func init() {
	// Here we define flags that are valid for the *entire* application (PersistentFlags).
	// For example, if we had a '--verbose' flag, we'd put it here.
	
	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.mytm.yaml)")

	// Local flags only apply to this specific command.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}


