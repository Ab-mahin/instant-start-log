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
	var mediaID int64
	existing, _ := database.SearchMedia(result.CleanTitle)
	for _, e := range existing {
		if e.CurrentFilePath == selectedPath || e.OriginalFilePath == selectedPath {
			mediaID = e.ID
			break
		}
	}

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

	if mediaID > 0 {
		database.InsertMoveHistory(mediaID, selectedPath, destPath,
			selectedFile.Name(), cleanName)
	}

	saveHistoryLog(database.BasePath, result.CleanTitle, result.Year,
		selectedPath, destPath)

	fmt.Println()
	fmt.Println("  ✅ Moved successfully!")
	fmt.Printf("     %s\n", selectedPath)
	fmt.Printf("     → %s\n", destPath)
}
