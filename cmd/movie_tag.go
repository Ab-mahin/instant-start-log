// movie_tag.go — mahin movie tag (add, remove, list)
package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/mahin/mahin-cli-v1/db"
	"github.com/spf13/cobra"
)

var movieTagCmd = &cobra.Command{
	Use:   "tag",
	Short: "Manage tags on media items",
	Long:  `Add, remove, or list tags on movies and TV shows in your library.`,
}

var movieTagAddCmd = &cobra.Command{
	Use:   "add <id|title> <tag>",
	Short: "Add a tag to a media item",
	Args:  cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		database, err := db.Open()
		if err != nil {
			fmt.Fprintf(os.Stderr, "❌ Database error: %v\n", err)
			return
		}
		defer database.Close()

		query := strings.Join(args[:len(args)-1], " ")
		tag := strings.TrimSpace(args[len(args)-1])

		m, err := resolveMediaByQuery(database, query)
		if err != nil {
			fmt.Fprintf(os.Stderr, "❌ %v\n", err)
			return
		}

		if err := database.AddTag(m.ID, tag); err != nil {
			if strings.Contains(err.Error(), "UNIQUE constraint") {
				fmt.Fprintf(os.Stderr, "❌ Tag %q already exists on %s\n", tag, m.CleanTitle)
				return
			}
			fmt.Fprintf(os.Stderr, "❌ Failed to add tag: %v\n", err)
			return
		}
		fmt.Printf("✅ Added tag %q to %s (#%d)\n", tag, m.CleanTitle, m.ID)
	},
}

var movieTagRemoveCmd = &cobra.Command{
	Use:   "remove <id|title> <tag>",
	Short: "Remove a tag from a media item",
	Args:  cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		database, err := db.Open()
		if err != nil {
			fmt.Fprintf(os.Stderr, "❌ Database error: %v\n", err)
			return
		}
		defer database.Close()

		query := strings.Join(args[:len(args)-1], " ")
		tag := strings.TrimSpace(args[len(args)-1])

		m, err := resolveMediaByQuery(database, query)
		if err != nil {
			fmt.Fprintf(os.Stderr, "❌ %v\n", err)
			return
		}

		if err := database.RemoveTag(m.ID, tag); err != nil {
			if err == db.ErrTagNotFound {
				fmt.Fprintf(os.Stderr, "❌ Tag %q not found on %s\n", tag, m.CleanTitle)
				return
			}
			fmt.Fprintf(os.Stderr, "❌ Failed to remove tag: %v\n", err)
			return
		}
		fmt.Printf("✅ Removed tag %q from %s (#%d)\n", tag, m.CleanTitle, m.ID)
	},
}

var movieTagListCmd = &cobra.Command{
	Use:   "list [id|title]",
	Short: "List tags — for a media item, or all tags",
	Run: func(cmd *cobra.Command, args []string) {
		database, err := db.Open()
		if err != nil {
			fmt.Fprintf(os.Stderr, "❌ Database error: %v\n", err)
			return
		}
		defer database.Close()

		if len(args) > 0 {
			listTagsForMedia(database, strings.Join(args, " "))
		} else {
			listAllTags(database)
		}
	},
}

func listTagsForMedia(database *db.DB, query string) {
	m, err := resolveMediaByQuery(database, query)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ %v\n", err)
		return
	}

	tags, err := database.ListTagsByMedia(m.ID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to list tags: %v\n", err)
		return
	}

	if len(tags) == 0 {
		fmt.Printf("📭 No tags on %s (#%d)\n", m.CleanTitle, m.ID)
		return
	}

	fmt.Printf("🏷️  Tags for %s (#%d):\n", m.CleanTitle, m.ID)
	for _, t := range tags {
		fmt.Printf("  • %s\n", t.Tag)
	}
}

func listAllTags(database *db.DB) {
	tagCounts, err := database.ListAllTags()
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to list tags: %v\n", err)
		return
	}

	if len(tagCounts) == 0 {
		fmt.Println("📭 No tags found in your library")
		return
	}

	fmt.Println("🏷️  All tags:")
	for tag, count := range tagCounts {
		fmt.Printf("  • %-20s (%d media)\n", tag, count)
	}
}

func init() {
	movieTagCmd.AddCommand(movieTagAddCmd, movieTagRemoveCmd, movieTagListCmd)
}
