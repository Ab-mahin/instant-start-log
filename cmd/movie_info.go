// movie_info.go — mahin movie info <id-or-title>
// Accepts a numeric ID (from library) or a title string.
// Checks local DB first; if not found by title, falls back to TMDb API,
// fetches full details, stores in DB, then displays.
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

var movieInfoCmd = &cobra.Command{
	Use:   "info [id or title]",
	Short: "Show detailed info for a movie or TV show",
	Long: `Display full metadata for a media item.

If a numeric ID is given, it looks up the item from your local library.
If a title is given, it first searches the local database. If not found,
it queries the TMDb API, saves the result, and then displays it.`,
	Args: cobra.MinimumNArgs(1),
	Run:  runMovieInfo,
}

func runMovieInfo(cmd *cobra.Command, args []string) {
	database, err := db.Open()
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Database error: %v\n", err)
		os.Exit(1)
	}
	defer database.Close()

	query := strings.Join(args, " ")

	// 1) Try local DB first (by ID or title)
	m, err := resolveMediaByQuery(database, query)
	if err == nil {
		fmt.Println("📚 Found in local library:")
		fmt.Println()
		printMediaDetail(m)
		return
	}

	// 3) Not in DB — fall back to TMDb API
	fmt.Printf("🔎 Not found locally. Searching TMDb for: %s\n\n", query)

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
	tmdbResults, err := client.SearchMulti(query)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ TMDb search error: %v\n", err)
		os.Exit(1)
	}
	if len(tmdbResults) == 0 {
		fmt.Println("📭 No results found on TMDb either.")
		return
	}

	// Pick the first (most relevant) result
	selected := tmdbResults[0]
	title := selected.GetDisplayTitle()
	year := selected.GetYear()
	yearInt := 0
	if year != "" {
		yearInt, _ = strconv.Atoi(year)
	}

	fmt.Printf("⏳ Fetching details for: %s (%s)...\n", title, year)

	// Check if this TMDb ID already exists in DB (avoid duplicates)
	existing, _ := database.GetMediaByTmdbID(selected.ID)
	if existing != nil {
		fmt.Println("📚 Already in your library:")
		fmt.Println()
		printMediaDetail(existing)
		return
	}

	// Build media record with full details
	m := &db.Media{
		Title:       title,
		CleanTitle:  title,
		Year:        yearInt,
		TmdbID:      selected.ID,
		TmdbRating:  selected.VoteAvg,
		Popularity:  selected.Popularity,
		Description: selected.Overview,
		Genre:       tmdb.GenreNames(selected.GenreIDs),
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
		}
	}

	// Save to DB
	_, err = database.InsertMedia(m)
	if err != nil {
		if m.TmdbID > 0 {
			err = database.UpdateMediaByTmdbID(m)
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "❌ DB error: %v\n", err)
			return
		}
	}

	fmt.Println()
	fmt.Println("✅ Saved to your library!")
	fmt.Println()
	printMediaDetail(m)
}


// fetchMovieDetails populates a Media record with TMDb movie details + credits.
func fetchMovieDetails(client *tmdb.Client, tmdbID int, m *db.Media) {
	details, err := client.GetMovieDetails(tmdbID)
	if err == nil {
		m.ImdbID = details.ImdbID
		m.Title = details.Title
		genres := make([]string, len(details.Genres))
		for i, g := range details.Genres {
			genres[i] = g.Name
		}
		m.Genre = strings.Join(genres, ", ")
	}

	credits, err := client.GetMovieCredits(tmdbID)
	if err == nil {
		var directors, castNames []string
		for _, c := range credits.Crew {
			if c.Job == "Director" {
				directors = append(directors, c.Name)
			}
		}
		m.Director = strings.Join(directors, ", ")

		for i, c := range credits.Cast {
			if i >= 10 {
				break
			}
			castNames = append(castNames, c.Name)
		}
		m.CastList = strings.Join(castNames, ", ")
	}
}

// fetchTVDetails populates a Media record with TMDb TV details + credits.
func fetchTVDetails(client *tmdb.Client, tmdbID int, m *db.Media) {
	details, err := client.GetTVDetails(tmdbID)
	if err == nil {
		m.Title = details.Name
		genres := make([]string, len(details.Genres))
		for i, g := range details.Genres {
			genres[i] = g.Name
		}
		m.Genre = strings.Join(genres, ", ")
	}

	credits, err := client.GetTVCredits(tmdbID)
	if err == nil {
		var directors, castNames []string
		for _, c := range credits.Crew {
			if c.Job == "Director" || c.Job == "Executive Producer" {
				directors = append(directors, c.Name)
			}
		}
		if len(directors) > 5 {
			directors = directors[:5]
		}
		m.Director = strings.Join(directors, ", ")

		for i, c := range credits.Cast {
			if i >= 10 {
				break
			}
			castNames = append(castNames, c.Name)
		}
		m.CastList = strings.Join(castNames, ", ")
	}
}
