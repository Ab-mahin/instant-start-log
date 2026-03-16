// root.go — defines the root cobra command and wires all subcommands together.
// The only logic here is registering child commands and calling Execute().
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "mahin",
	Short: "mahin-cli-v1 — a self-updating CLI",
	Long: `mahin-cli-v1 is a small CLI with hello, version, and self-update.

self-update pulls the latest files from the cloned git repository
using git pull --ff-only.`,
}

func init() {
	// Keep the CLI surface focused on project commands only.
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	rootCmd.AddCommand(helloCmd, versionCmd, selfUpdateCmd, movieCmd)
}

// Execute is called by main.go. It is the single public entry point.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
