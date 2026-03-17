// movie_move_helpers.go — helper functions for movie move command
package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/mahin/mahin-cli-v1/cleaner"
	"github.com/mahin/mahin-cli-v1/db"
)

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
		from, to, time.Now().Format(time.RFC3339))
	f.WriteString(entry)
}
