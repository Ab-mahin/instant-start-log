// movie_config.go — mahin movie config
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/mahin/mahin-cli-v1/db"
)

var movieConfigCmd = &cobra.Command{
	Use:   "config [get|set] [key] [value]",
	Short: "Manage movie CLI configuration",
	Long: `View or update configuration settings.

Keys:
  movies_dir     - Default movies directory
  tv_dir         - Default TV shows directory
  archive_dir    - Default archive directory
  scan_dir       - Default scan directory
  tmdb_api_key   - TMDb API key
  page_size      - Items per page in list view

Examples:
  mahin movie config                           # Show all
  mahin movie config get movies_dir            # Get one
  mahin movie config set movies_dir ~/Movies   # Set one
  mahin movie config set tmdb_api_key abc123   # Set API key`,
	Run: runMovieConfig,
}

func runMovieConfig(cmd *cobra.Command, args []string) {
	database, err := db.Open()
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Database error: %v\n", err)
		os.Exit(1)
	}
	defer database.Close()

	if len(args) == 0 {
		// Show all config
		showAllConfig(database)
		return
	}

	action := args[0]

	switch action {
	case "get":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "❌ Usage: mahin movie config get <key>")
			os.Exit(1)
		}
		val, err := database.GetConfig(args[1])
		if err != nil {
			fmt.Printf("  %s = (not set)\n", args[1])
		} else {
			fmt.Printf("  %s = %s\n", args[1], val)
		}

	case "set":
		if len(args) < 3 {
			fmt.Fprintln(os.Stderr, "❌ Usage: mahin movie config set <key> <value>")
			os.Exit(1)
		}
		key, value := args[1], args[2]
		if err := database.SetConfig(key, value); err != nil {
			fmt.Fprintf(os.Stderr, "❌ Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("  ✅ %s = %s\n", key, value)

	default:
		fmt.Fprintf(os.Stderr, "❌ Unknown action: %s. Use 'get' or 'set'.\n", action)
		os.Exit(1)
	}
}

func showAllConfig(database *db.DB) {
	fmt.Println("⚙️  Configuration:")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	keys := []string{"movies_dir", "tv_dir", "archive_dir", "scan_dir", "tmdb_api_key", "page_size"}
	for _, key := range keys {
		val, err := database.GetConfig(key)
		if err != nil {
			val = "(not set)"
		}
		// Mask API key
		if key == "tmdb_api_key" && len(val) > 8 {
			val = val[:4] + "..." + val[len(val)-4:]
		}
		fmt.Printf("  %-15s = %s\n", key, val)
	}
}
