// movie_info.go — mahin movie info <id>
package cmd

import (
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/mahin/mahin-cli-v1/db"
)

var movieInfoCmd = &cobra.Command{
	Use:   "info [id]",
	Short: "Show detailed info for a movie or TV show",
	Long:  `Display full metadata for a media item by its ID number.`,
	Args:  cobra.ExactArgs(1),
	Run:   runMovieInfo,
}

func runMovieInfo(cmd *cobra.Command, args []string) {
	database, err := db.Open()
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Database error: %v\n", err)
		os.Exit(1)
	}
	defer database.Close()

	id, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		fmt.Fprintln(os.Stderr, "❌ Invalid ID. Use a number from 'mahin movie ls'.")
		os.Exit(1)
	}

	m, err := database.GetMediaByID(id)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Media not found: %v\n", err)
		os.Exit(1)
	}

	printMediaDetail(m)
}
