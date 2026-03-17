// movie_move.go — mahin movie move
package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/mahin/mahin-cli-v1/cleaner"
	"github.com/mahin/mahin-cli-v1/db"
)

var movieMoveCmd = &cobra.Command{
	Use:   "move [directory]",
	Short: "Browse a local directory and move a movie/TV show file",
	Long: `Browse a local directory for video files, select one, and move it
to a configured destination (Movies, TV Shows, Archive, or custom path).
The move is logged for undo support.

If no directory is given, you'll be prompted to choose one.`,
	Args: cobra.MaximumNArgs(1),
	Run:  runMovieMove,
}

func runMovieMove(cmd *cobra.Command, args []string) {
	database, err := db.Open()
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Database error: %v\n", err)
		os.Exit(1)
	}
	defer database.Close()

	scanner := bufio.NewScanner(os.Stdin)
	home, _ := os.UserHomeDir()

	// Step 1: Determine the source directory
	sourceDir := ""
	if len(args) > 0 {
		sourceDir = expandHome(args[0], home)
	} else {
		sourceDir = promptSourceDirectory(scanner, database, home)
		if sourceDir == "" {
			return
		}
	}

	// Validate directory
	info, err := os.Stat(sourceDir)
	if err != nil || !info.IsDir() {
		fmt.Fprintf(os.Stderr, "❌ Directory not found: %s\n", sourceDir)
		return
	}

	// Step 2: List video files in the directory
	files := listVideoFiles(sourceDir)
	if len(files) == 0 {
		fmt.Printf("📭 No video files found in: %s\n", sourceDir)
		return
	}

	fmt.Printf("\n🎬 Video files in: %s\n\n", sourceDir)
	for i, f := range files {
		result := cleaner.Clean(f.Name())
		typeIcon := "🎬"
		if result.Type == "tv" {
			typeIcon = "📺"
		}
		yearStr := ""
		if result.Year > 0 {
			yearStr = fmt.Sprintf("(%d)", result.Year)
		}
		fmt.Printf("  %2d. %s %s %s  [%s]\n", i+1, typeIcon, result.CleanTitle, yearStr, humanSize(f.Size()))
	}

	// Step 3: Select a file
	fmt.Println()
	fmt.Print("  Select file [number]: ")
	if !scanner.Scan() {
		return
	}
	choice, err := strconv.Atoi(strings.TrimSpace(scanner.Text()))
	if err != nil || choice < 1 || choice > len(files) {
		fmt.Println("❌ Invalid selection")
		return
	}

	selectedFile := files[choice-1]
	selectedPath := filepath.Join(sourceDir, selectedFile.Name())
	result := cleaner.Clean(selectedFile.Name())

	fmt.Printf("\n  Selected: %s\n", result.CleanTitle)
	if result.Year > 0 {
		fmt.Printf("  Year:     %d\n", result.Year)
	}
	fmt.Printf("  Type:     %s\n", result.Type)

	// Step 4: Choose destination
	destDir := promptDestination(scanner, database, home)
	if destDir == "" {
		return
	}

	// Step 5: Build clean filename and confirm
	cleanName := cleaner.ToCleanFileName(result.CleanTitle, result.Year, result.Extension)
	destPath := filepath.Join(destDir, cleanName)

	fmt.Println()
	fmt.Println("  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("  📄 From: %s\n", selectedPath)
	fmt.Printf("  📁 To:   %s\n", destPath)
	fmt.Println("  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()
	fmt.Print("  Are you sure? [y/N]: ")

	if !scanner.Scan() {
		return
	}
	confirm := strings.ToLower(strings.TrimSpace(scanner.Text()))
	if confirm != "y" && confirm != "yes" {
		fmt.Println("  ❌ Move cancelled.")
		return
	}

	// Step 6: Create destination directory
	if err := os.MkdirAll(destDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "  ❌ Cannot create directory: %v\n", err)
		return
	}

	// Step 7: Move the file
	if err := os.Rename(selectedPath, destPath); err != nil {
		fmt.Fprintf(os.Stderr, "  ❌ Move failed: %v\n", err)
		return
	}

	// Step 8: Track history for undo
	// Check if this file exists in the database
	var mediaID int64
	existing, _ := database.SearchMedia(result.CleanTitle)
	for _, e := range existing {
		if e.CurrentFilePath == selectedPath || e.OriginalFilePath == selectedPath {
			mediaID = e.ID
			break
		}
	}

	// If not in DB, insert it
	if mediaID == 0 {
		m := &db.Media{
			Title:            result.CleanTitle,
			CleanTitle:       result.CleanTitle,
			Year:             result.Year,
			Type:             result.Type,
			OriginalFileName: selectedFile.Name(),
			OriginalFilePath: selectedPath,
			CurrentFilePath:  destPath,
			FileExtension:    result.Extension,
			FileSize:         selectedFile.Size(),
		}
		mediaID, _ = database.InsertMedia(m)
	} else {
		database.UpdateMediaPath(mediaID, destPath)
	}

	// Log move history for undo
	if mediaID > 0 {
		database.InsertMoveHistory(mediaID, selectedPath, destPath,
			selectedFile.Name(), cleanName)
	}

	// Save JSON history
	saveHistoryLog(database.BasePath, result.CleanTitle, result.Year,
		selectedPath, destPath)

	fmt.Println()
	fmt.Println("  ✅ Moved successfully!")
	fmt.Printf("     %s\n", selectedPath)
	fmt.Printf("     → %s\n", destPath)
}

// promptSourceDirectory shows configured directories and a custom option.
func promptSourceDirectory(scanner *bufio.Scanner, database *db.DB, home string) string {
	scanDir, _ := database.GetConfig("scan_dir")
	scanDir = expandHome(scanDir, home)

	fmt.Println("📂 Choose a directory to browse:")
	fmt.Println()
	fmt.Printf("  1. 📥 Downloads   (%s)\n", expandHome("~/Downloads", home))
	if scanDir != "" && scanDir != expandHome("~/Downloads", home) {
		fmt.Printf("  2. 🔍 Scan Dir    (%s)\n", scanDir)
		fmt.Println("  3. 📁 Custom path")
		fmt.Println()
		fmt.Print("  Choose [1-3]: ")
	} else {
		fmt.Println("  2. 📁 Custom path")
		fmt.Println()
		fmt.Print("  Choose [1-2]: ")
	}

	if !scanner.Scan() {
		return ""
	}

	input := strings.TrimSpace(scanner.Text())
	hasScanDir := scanDir != "" && scanDir != expandHome("~/Downloads", home)

	switch input {
	case "1":
		return expandHome("~/Downloads", home)
	case "2":
		if hasScanDir {
			return scanDir
		}
		return promptCustomPath(scanner, home)
	case "3":
		if hasScanDir {
			return promptCustomPath(scanner, home)
		}
		fmt.Println("❌ Invalid choice")
		return ""
	default:
		fmt.Println("❌ Invalid choice")
		return ""
	}
}

func promptCustomPath(scanner *bufio.Scanner, home string) string {
	fmt.Print("  Enter path: ")
	if !scanner.Scan() {
		return ""
	}
	return expandHome(strings.TrimSpace(scanner.Text()), home)
}

// promptDestination shows destination options.
func promptDestination(scanner *bufio.Scanner, database *db.DB, home string) string {
	moviesDir, _ := database.GetConfig("movies_dir")
	tvDir, _ := database.GetConfig("tv_dir")
	archiveDir, _ := database.GetConfig("archive_dir")

	moviesDir = expandHome(moviesDir, home)
	tvDir = expandHome(tvDir, home)
	archiveDir = expandHome(archiveDir, home)

	fmt.Println()
	fmt.Println("  Destination:")
	fmt.Printf("  1. 🎬 Movies     (%s)\n", moviesDir)
	fmt.Printf("  2. 📺 TV Shows   (%s)\n", tvDir)
	fmt.Printf("  3. 📦 Archive    (%s)\n", archiveDir)
	fmt.Println("  4. 📁 Custom path")
	fmt.Println()
	fmt.Print("  Choose [1-4]: ")

	if !scanner.Scan() {
		return ""
	}

	switch strings.TrimSpace(scanner.Text()) {
	case "1":
		return moviesDir
	case "2":
		return tvDir
	case "3":
		return archiveDir
	case "4":
		return promptCustomPath(scanner, home)
	default:
		fmt.Println("❌ Invalid choice")
		return ""
	}
}

// listVideoFiles returns video files in a directory sorted by name.
func listVideoFiles(dir string) []os.FileInfo {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}

	var files []os.FileInfo
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if cleaner.IsVideoFile(entry.Name()) {
			info, err := entry.Info()
			if err == nil {
				files = append(files, info)
			}
		}
	}
	return files
}

// humanSize formats bytes into human-readable size.
func humanSize(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)
	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.1f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.1f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.1f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

func expandHome(path, home string) string {
	if strings.HasPrefix(path, "~/") {
		return filepath.Join(home, path[2:])
	}
	if path == "~" {
		return home
	}
	return path
}

func saveHistoryLog(basePath, title string, year int, from, to string) {
	slug := cleaner.ToSlug(title)
	if year > 0 {
		slug += "-" + strconv.Itoa(year)
	}
	histDir := filepath.Join(basePath, "json", "history", slug)
	os.MkdirAll(histDir, 0755)

	logFile := filepath.Join(histDir, "move-log.json")

	f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer f.Close()

	entry := fmt.Sprintf(`{"from":"%s","to":"%s","timestamp":"%s"}`+"\n",
		from, to, "now")
	f.WriteString(entry)
}
