// movie_rename.go — mahin movie rename
package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/mahin/mahin-cli-v1/cleaner"
	"github.com/mahin/mahin-cli-v1/db"
)

var movieRenameCmd = &cobra.Command{
	Use:   "rename",
	Short: "Rename files to clean names",
	Long: `Automatically renames messy filenames to clean format.
Example: Scream.2022.1080p.WEBRip.x264-RARBG.mkv → Scream (2022).mkv`,
	Run: runMovieRename,
}

func runMovieRename(cmd *cobra.Command, args []string) {
	database, err := db.Open()
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Database error: %v\n", err)
		os.Exit(1)
	}
	defer database.Close()

	media, err := database.ListMedia(0, 10000)
	if err != nil || len(media) == 0 {
		fmt.Println("📭 No media found.")
		return
	}

	// Find files that need renaming
	type renameItem struct {
		media   db.Media
		oldPath string
		newPath string
		oldName string
		newName string
	}

	var items []renameItem
	for _, m := range media {
		if m.CurrentFilePath == "" {
			continue
		}
		dir := filepath.Dir(m.CurrentFilePath)
		oldName := filepath.Base(m.CurrentFilePath)
		newName := cleaner.ToCleanFileName(m.CleanTitle, m.Year, m.FileExtension)

		if oldName != newName {
			items = append(items, renameItem{
				media:   m,
				oldPath: m.CurrentFilePath,
				newPath: filepath.Join(dir, newName),
				oldName: oldName,
				newName: newName,
			})
		}
	}

	if len(items) == 0 {
		fmt.Println("✅ All files already have clean names!")
		return
	}

	fmt.Printf("📝 Found %d files to rename:\n\n", len(items))
	for i, item := range items {
		fmt.Printf("  %d. %s\n", i+1, item.oldName)
		fmt.Printf("     → %s\n\n", item.newName)
	}

	fmt.Print("Rename all? [y/N]: ")
	scanner := bufio.NewScanner(os.Stdin)
	if !scanner.Scan() {
		return
	}
	confirm := strings.ToLower(strings.TrimSpace(scanner.Text()))
	if confirm != "y" && confirm != "yes" {
		fmt.Println("❌ Cancelled.")
		return
	}

	success := 0
	for _, item := range items {
		if err := os.Rename(item.oldPath, item.newPath); err != nil {
			fmt.Fprintf(os.Stderr, "  ❌ Failed: %s → %v\n", item.oldName, err)
			continue
		}
		database.UpdateMediaPath(item.media.ID, item.newPath)
		database.InsertMoveHistory(item.media.ID, item.oldPath, item.newPath,
			item.oldName, item.newName)
		fmt.Printf("  ✅ %s → %s\n", item.oldName, item.newName)
		success++
	}

	fmt.Printf("\n✅ Renamed %d/%d files.\n", success, len(items))
}
