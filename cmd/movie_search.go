// movie_search.go — mahin movie search <name>
package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/mahin/mahin-cli-v1/db"
)

var movieSearchCmd = &cobra.Command{
	Use:   "search [name]",
	Short: "Search for a movie or TV show",
	Long: `Searches the local database for movies/TV shows matching the query.
Uses fuzzy matching (LIKE) so partial names work.`,
	Args: cobra.MinimumNArgs(1),
	Run:  runMovieSearch,
}

func runMovieSearch(cmd *cobra.Command, args []string) {
	database, err := db.Open()
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Database error: %v\n", err)
		os.Exit(1)
	}
	defer database.Close()

	query := strings.Join(args, " ")
	fmt.Printf("🔎 Searching for: %s\n\n", query)

	results, err := database.SearchMedia(query)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Search error: %v\n", err)
		os.Exit(1)
	}

	if len(results) == 0 {
		fmt.Println("📭 No results found in your library.")
		fmt.Println("   Try: mahin movie scan <folder> first.")
		return
	}

	if len(results) == 1 {
		printMediaDetail(&results[0])
		return
	}

	// Multiple results
	fmt.Printf("Found %d results:\n\n", len(results))
	for i, m := range results {
		yearStr := ""
		if m.Year > 0 {
			yearStr = fmt.Sprintf("(%d)", m.Year)
		}

		rating := "N/A"
		if m.TmdbRating > 0 {
			rating = fmt.Sprintf("%.1f", m.TmdbRating)
		}

		typeIcon := "🎬"
		if m.Type == "tv" {
			typeIcon = "📺"
		}

		fmt.Printf("  %d. %-35s %-6s  ⭐ %-4s  %s\n",
			i+1, m.CleanTitle, yearStr, rating, typeIcon)
	}

	fmt.Println()
	fmt.Println("Use 'mahin movie info <id>' for full details.")
}

// printMediaDetail for search results (reuses from movie_ls.go via pointer)
func printMediaDetailFromSearch(m *db.Media) {
	printMediaDetail(m)
}
