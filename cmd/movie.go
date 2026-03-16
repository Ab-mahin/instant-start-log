// movie.go — parent command: mahin movie
package cmd

import (
	"github.com/spf13/cobra"
)

var movieCmd = &cobra.Command{
	Use:   "movie",
	Short: "Movie & TV show manager",
	Long:  `Scan, search, organize, and manage your movie and TV show library.`,
}

func init() {
	movieCmd.AddCommand(
		movieScanCmd,
		movieLsCmd,
		movieSearchCmd,
		movieSuggestCmd,
		movieMoveCmd,
		movieUndoCmd,
		movieInfoCmd,
		moviePlayCmd,
		movieStatsCmd,
		movieRenameCmd,
		movieConfigCmd,
	)
}
