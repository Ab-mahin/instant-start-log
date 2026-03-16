// movie_stats.go — mahin movie stats
package cmd

import (
	"fmt"
	"os"
	"sort"

	"github.com/spf13/cobra"

	"github.com/mahin/mahin-cli-v1/db"
)

var movieStatsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Show library statistics",
	Long:  `Display total counts, top genres, and average ratings.`,
	Run:   runMovieStats,
}

func runMovieStats(cmd *cobra.Command, args []string) {
	database, err := db.Open()
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Database error: %v\n", err)
		os.Exit(1)
	}
	defer database.Close()

	totalMovies, _ := database.CountMedia("movie")
	totalTV, _ := database.CountMedia("tv")
	total, _ := database.CountMedia("")

	if total == 0 {
		fmt.Println("📭 No media in library. Run 'mahin movie scan <folder>' first.")
		return
	}

	fmt.Println("📊 Library Statistics")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("  🎬 Total Movies:    %d\n", totalMovies)
	fmt.Printf("  📺 Total TV Shows:  %d\n", totalTV)
	fmt.Printf("  📁 Total:           %d\n", total)
	fmt.Println()

	// Top genres
	genres, err := database.TopGenres(10)
	if err == nil && len(genres) > 0 {
		type gc struct {
			name  string
			count int
		}
		var sorted []gc
		for n, c := range genres {
			sorted = append(sorted, gc{n, c})
		}
		sort.Slice(sorted, func(i, j int) bool {
			return sorted[i].count > sorted[j].count
		})

		fmt.Println("  🎭 Top Genres:")
		for i, g := range sorted {
			if i >= 10 {
				break
			}
			bar := ""
			for j := 0; j < g.count && j < 30; j++ {
				bar += "█"
			}
			fmt.Printf("     %-20s %s %d\n", g.name, bar, g.count)
		}
		fmt.Println()
	}

	// Average ratings
	var avgImdb, avgTmdb float64
	var imdbCount, tmdbCount int

	allMedia, _ := database.ListMedia(0, 10000)
	for _, m := range allMedia {
		if m.ImdbRating > 0 {
			avgImdb += m.ImdbRating
			imdbCount++
		}
		if m.TmdbRating > 0 {
			avgTmdb += m.TmdbRating
			tmdbCount++
		}
	}

	if imdbCount > 0 {
		fmt.Printf("  ⭐ Avg IMDb Rating: %.1f\n", avgImdb/float64(imdbCount))
	}
	if tmdbCount > 0 {
		fmt.Printf("  ⭐ Avg TMDb Rating: %.1f\n", avgTmdb/float64(tmdbCount))
	}
}
