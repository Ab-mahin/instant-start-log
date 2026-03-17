// movie_search.go — mahin movie search <name>
// Searches TMDb API, fetches full details, and saves to local database.
package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/mahin/mahin-cli-v1/cleaner"
	"github.com/mahin/mahin-cli-v1/db"
	"github.com/mahin/mahin-cli-v1/tmdb"
)

var movieSearchCmd = &cobra.Command{
	Use:   "search [name]",
	Short: "Search TMDb for a movie or TV show and save to database",
	Long: `Searches the TMDb API for movies/TV shows matching the query.
Fetches full metadata (rating, genres, cast, crew, poster) and saves
to the local database. Categorizes as Movie or TV Show automatically.
Does NOT require the file to exist in your library.`,
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

	// Get TMDb API key
	apiKey, _ := database.GetConfig("tmdb_api_key")
	if apiKey == "" {
		apiKey = os.Getenv("TMDB_API_KEY")
	}
	if apiKey == "" {
		fmt.Fprintln(os.Stderr, "❌ No TMDb API key configured.")
		fmt.Fprintln(os.Stderr, "   Set it with: mahin movie config set tmdb_api_key YOUR_KEY")
		os.Exit(1)
	}

	client := tmdb.NewClient(apiKey)
	query := strings.Join(args, " ")
	fmt.Printf("🔎 Searching TMDb for: %s\n\n", query)

	// Search TMDb API
	results, err := client.SearchMulti(query)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ TMDb search error: %v\n", err)
		os.Exit(1)
	}

	if len(results) == 0 {
		fmt.Println("📭 No results found on TMDb.")
		return
	}

	// Show results and let user pick
	fmt.Printf("Found %d results:\n\n", len(results))
	for i, r := range results {
		if i >= 15 {
			break
		}
		title := r.GetDisplayTitle()
		year := r.GetYear()
		typeIcon := "🎬"
		typeLabel := "Movie"
		if r.MediaType == "tv" {
			typeIcon = "📺"
			typeLabel = "TV Show"
		}

		rating := "N/A"
		if r.VoteAvg > 0 {
			rating = fmt.Sprintf("%.1f", r.VoteAvg)
		}

		yearStr := ""
		if year != "" {
			yearStr = fmt.Sprintf("(%s)", year)
		}

		fmt.Printf("  %d. %s %-35s %-6s  ⭐ %-4s  [%s]\n",
			i+1, typeIcon, title, yearStr, rating, typeLabel)
	}

	fmt.Println()
	fmt.Print("Enter number to save (0 to cancel): ")

	var choice int
	_, err = fmt.Scan(&choice)
	if err != nil || choice < 1 || choice > len(results) || choice > 15 {
		fmt.Println("❌ Cancelled.")
		return
	}

	selected := results[choice-1]
	title := selected.GetDisplayTitle()
	year := selected.GetYear()
	yearInt := 0
	if year != "" {
		yearInt, _ = strconv.Atoi(year)
	}

	fmt.Printf("\n⏳ Fetching full details for: %s...\n", title)

	// Build media record
	m := &db.Media{
		Title:      title,
		CleanTitle: title,
		Year:       yearInt,
		TmdbID:     selected.ID,
		TmdbRating: selected.VoteAvg,
		Popularity: selected.Popularity,
		Description: selected.Overview,
		Genre:      tmdb.GenreNames(selected.GenreIDs),
	}

	if selected.MediaType == "movie" || selected.MediaType == "" {
		m.Type = "movie"
		fetchMovieDetails(client, selected.ID, m)
	} else if selected.MediaType == "tv" {
		m.Type = "tv"
		fetchTVDetails(client, selected.ID, m)
	}

	// Download thumbnail
	if selected.PosterPath != "" {
		slug := cleaner.ToSlug(m.CleanTitle)
		if m.Year > 0 {
			slug += "-" + strconv.Itoa(m.Year)
		}
		thumbDir := filepath.Join(database.BasePath, "thumbnails", slug)
		os.MkdirAll(thumbDir, 0755)
		thumbPath := filepath.Join(thumbDir, slug+".jpg")
		if err := client.DownloadPoster(selected.PosterPath, thumbPath); err == nil {
			m.ThumbnailPath = thumbPath
			fmt.Println("🖼️  Thumbnail saved")
		}
	}

	// Save JSON to movie or tv folder based on type
	jsonDir := filepath.Join(database.BasePath, "json", m.Type)
	os.MkdirAll(jsonDir, 0755)

	// Insert into database (or update if already exists by tmdb_id)
	_, err = database.InsertMedia(m)
	if err != nil {
		if m.TmdbID > 0 {
			err = database.UpdateMediaByTmdbID(m)
			if err == nil {
				fmt.Printf("🔄 Updated existing record for: %s\n", m.Title)
			} else {
				fmt.Fprintf(os.Stderr, "❌ DB error: %v\n", err)
				return
			}
		} else {
			fmt.Fprintf(os.Stderr, "❌ DB error: %v\n", err)
			return
		}
	}

	// Print saved details
	typeIcon := "🎬"
	typeLabel := "Movie"
	folder := "movie"
	if m.Type == "tv" {
		typeIcon = "📺"
		typeLabel = "TV Show"
		folder = "tv"
	}

	fmt.Println()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("✅ Saved to database!\n\n")
	fmt.Printf("  %s  %s (%s)\n", typeIcon, m.Title, typeLabel)
	fmt.Printf("  📅  Year: %d\n", m.Year)
	fmt.Printf("  ⭐  Rating: %.1f\n", m.TmdbRating)
	fmt.Printf("  🎭  Genre: %s\n", m.Genre)
	if m.Director != "" {
		fmt.Printf("  🎬  Director: %s\n", m.Director)
	}
	if m.CastList != "" {
		fmt.Printf("  👥  Cast: %s\n", m.CastList)
	}
	if m.Description != "" {
		desc := m.Description
		if len(desc) > 150 {
			desc = desc[:147] + "..."
		}
		fmt.Printf("  📝  %s\n", desc)
	}
	fmt.Printf("  📁  Stored in: %s/ folder\n", folder)
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
}
