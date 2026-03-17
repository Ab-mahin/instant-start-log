// movie_scan.go — mahin movie scan <folder>
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

var movieScanCmd = &cobra.Command{
	Use:   "scan [folder]",
	Short: "Scan a folder for movies and TV shows",
	Long: `Scans a folder for video files, cleans filenames, fetches metadata
from TMDb, downloads thumbnails, and stores everything in the database.`,
	Args: cobra.MaximumNArgs(1),
	Run:  runMovieScan,
}

func runMovieScan(cmd *cobra.Command, args []string) {
	database, err := db.Open()
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Database error: %v\n", err)
		os.Exit(1)
	}
	defer database.Close()

	// Determine scan folder
	scanDir := ""
	if len(args) > 0 {
		scanDir = args[0]
	} else {
		scanDir, _ = database.GetConfig("scan_dir")
		if scanDir == "" {
			fmt.Fprintln(os.Stderr, "❌ No folder specified. Use: mahin movie scan <folder>")
			os.Exit(1)
		}
	}

	// Expand ~ to home
	if strings.HasPrefix(scanDir, "~") {
		home, _ := os.UserHomeDir()
		scanDir = filepath.Join(home, scanDir[1:])
	}

	// Check folder exists
	info, err := os.Stat(scanDir)
	if err != nil || !info.IsDir() {
		fmt.Fprintf(os.Stderr, "❌ Folder not found: %s\n", scanDir)
		os.Exit(1)
	}

	// Get TMDb API key
	apiKey, _ := database.GetConfig("tmdb_api_key")
	if apiKey == "" {
		apiKey = os.Getenv("TMDB_API_KEY")
	}
	if apiKey == "" {
		fmt.Fprintln(os.Stderr, "⚠️  No TMDb API key configured.")
		fmt.Fprintln(os.Stderr, "   Set it with: mahin movie config set tmdb_api_key YOUR_KEY")
		fmt.Fprintln(os.Stderr, "   Or set TMDB_API_KEY environment variable.")
		fmt.Fprintln(os.Stderr, "   Scanning will proceed without metadata fetching.")
		fmt.Println()
	}

	client := tmdb.NewClient(apiKey)

	fmt.Printf("🔍 Scanning: %s\n\n", scanDir)

	var totalFiles, movieCount, tvCount, skipped int

	entries, err := os.ReadDir(scanDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Cannot read folder: %v\n", err)
		os.Exit(1)
	}

	for _, entry := range entries {
		name := entry.Name()
		fullPath := filepath.Join(scanDir, name)

		// Handle both files and directories
		if entry.IsDir() {
			// For directories, look for video files inside
			subEntries, err := os.ReadDir(fullPath)
			if err != nil {
				continue
			}
			foundVideo := false
			for _, sub := range subEntries {
				if !sub.IsDir() && cleaner.IsVideoFile(sub.Name()) {
					foundVideo = true
					name = entry.Name() // use directory name for cleaning
					fullPath = filepath.Join(fullPath, sub.Name())
					break
				}
			}
			if !foundVideo {
				continue
			}
		} else if !cleaner.IsVideoFile(name) {
			continue
		}

		totalFiles++

		// Clean the filename
		result := cleaner.Clean(name)
		fmt.Printf("  📄 %s\n", name)
		fmt.Printf("     → %s", result.CleanTitle)
		if result.Year > 0 {
			fmt.Printf(" (%d)", result.Year)
		}
		fmt.Printf(" [%s]\n", result.Type)

		// Check if already in DB by path
		existing, _ := database.SearchMedia(result.CleanTitle)
		alreadyExists := false
		for _, e := range existing {
			if e.OriginalFilePath == fullPath {
				alreadyExists = true
				fmt.Println("     ⏩ Already in database, skipping")
				break
			}
		}
		if alreadyExists {
			skipped++
			if result.Type == "movie" {
				movieCount++
			} else {
				tvCount++
			}
			continue
		}

		// Build media record
		fi, _ := os.Stat(fullPath)
		m := &db.Media{
			Title:            result.CleanTitle,
			CleanTitle:       result.CleanTitle,
			Year:             result.Year,
			Type:             result.Type,
			OriginalFileName: name,
			OriginalFilePath: fullPath,
			CurrentFilePath:  fullPath,
			FileExtension:    result.Extension,
		}
		if fi != nil {
			m.FileSize = fi.Size()
		}

		// Fetch metadata from TMDb
		if apiKey != "" {
			searchQuery := result.CleanTitle
			if result.Year > 0 {
				searchQuery += " " + strconv.Itoa(result.Year)
			}

			results, err := client.SearchMulti(searchQuery)
			if err == nil && len(results) > 0 {
				best := results[0]
				m.TmdbID = best.ID
				m.TmdbRating = best.VoteAvg
				m.Popularity = best.Popularity
				m.Description = best.Overview
				m.Genre = tmdb.GenreNames(best.GenreIDs)

			if best.MediaType == "movie" || best.MediaType == "" {
				m.Type = "movie"
				fetchMovieDetails(client, best.ID, m)
			} else if best.MediaType == "tv" {
				m.Type = "tv"
				fetchTVDetails(client, best.ID, m)
			}

				// Download thumbnail
				if best.PosterPath != "" {
					slug := cleaner.ToSlug(m.CleanTitle)
					if m.Year > 0 {
						slug += "-" + strconv.Itoa(m.Year)
					}
					thumbDir := filepath.Join(database.BasePath, "thumbnails", slug)
					os.MkdirAll(thumbDir, 0755)
					thumbPath := filepath.Join(thumbDir, slug+".jpg")
					if err := client.DownloadPoster(best.PosterPath, thumbPath); err == nil {
						m.ThumbnailPath = thumbPath
						fmt.Println("     🖼️  Thumbnail saved")
					}
				}

				fmt.Printf("     ✅ TMDb: %s (⭐ %.1f)\n", m.Title, m.TmdbRating)
			} else {
				fmt.Println("     ⚠️  No TMDb match found")
			}
		}

		// Insert into database
		_, err = database.InsertMedia(m)
		if err != nil {
			// Try update if duplicate tmdb_id
			if m.TmdbID > 0 {
				database.UpdateMediaByTmdbID(m)
			} else {
				fmt.Fprintf(os.Stderr, "     ❌ DB error: %v\n", err)
			}
		}

		if m.Type == "movie" {
			movieCount++
		} else {
			tvCount++
		}
		fmt.Println()
	}

	// Log scan history
	database.InsertScanHistory(scanDir, totalFiles, movieCount, tvCount)

	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("📊 Scan Complete!\n")
	fmt.Printf("   Total files: %d\n", totalFiles)
	fmt.Printf("   Movies:      %d\n", movieCount)
	fmt.Printf("   TV Shows:    %d\n", tvCount)
	if skipped > 0 {
		fmt.Printf("   Skipped:     %d (already in DB)\n", skipped)
	}
}
