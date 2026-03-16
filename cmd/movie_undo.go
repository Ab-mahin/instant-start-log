// movie_undo.go — mahin movie undo
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/mahin/mahin-cli-v1/db"
)

var movieUndoCmd = &cobra.Command{
	Use:   "undo",
	Short: "Undo the last move operation",
	Long:  `Reverts the most recent movie move operation.`,
	Run:   runMovieUndo,
}

func runMovieUndo(cmd *cobra.Command, args []string) {
	database, err := db.Open()
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Database error: %v\n", err)
		os.Exit(1)
	}
	defer database.Close()

	lastMove, err := database.GetLastMove()
	if err != nil {
		fmt.Println("📭 No move operations to undo.")
		return
	}

	fmt.Println("⏪ Undoing last move...")
	fmt.Println()
	fmt.Printf("  📁 %s\n", lastMove.ToPath)
	fmt.Printf("  → %s\n", lastMove.FromPath)
	fmt.Println()

	// Check source exists
	if _, err := os.Stat(lastMove.ToPath); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "❌ File not found at: %s\n", lastMove.ToPath)
		fmt.Fprintln(os.Stderr, "   It may have been moved or deleted manually.")
		os.Exit(1)
	}

	// Move back
	if err := os.Rename(lastMove.ToPath, lastMove.FromPath); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Undo failed: %v\n", err)
		os.Exit(1)
	}

	// Mark as undone in DB
	database.MarkMoveUndone(lastMove.ID)

	// Update media path back
	database.UpdateMediaPath(lastMove.MediaID, lastMove.FromPath)

	fmt.Println("✅ Undo successful!")
	fmt.Printf("   %s\n", lastMove.ToPath)
	fmt.Printf("   → %s\n", lastMove.FromPath)
}
