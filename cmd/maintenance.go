package cmd

import (
	"fmt"

	"github.com/asciimoo/hister/client"

	"github.com/spf13/cobra"
)

var reindexCmd = &cobra.Command{
	Use:   "reindex",
	Short: "Rebuild the search index",
	Long:  `Rebuild the search index from all stored documents using the running server.`,
	Run: func(cmd *cobra.Command, args []string) {
		skipSensitive := false
		if b, err := cmd.Flags().GetBool("exclude-sensitive"); err == nil {
			skipSensitive = b
		}
		c := newClient(client.WithTimeout(0))
		if err := c.Reindex(skipSensitive, cfg.Indexer.DetectLanguages); err != nil {
			msg := "Reindex error: " + err.Error()
			if isConnectionError(err) {
				msg += "\n  Make sure the Hister server is running before executing reindex."
			}
			exit(1, msg)
		}
	},
}

var cleanupCmd = &cobra.Command{
	Use:   "cleanup",
	Short: "Remove stale local documents and orphaned data",
	Long:  `Remove indexed local documents that no longer match configured directories, and remove HTML and favicon files that are no longer referenced by any document in the index`,
	Run: func(_ *cobra.Command, _ []string) {
		c := newClient(client.WithTimeout(0))
		result, err := c.Cleanup()
		if err != nil {
			msg := "Cleanup error: " + err.Error()
			if isConnectionError(err) {
				msg += "\n  Make sure the Hister server is running before executing cleanup."
			}
			exit(1, msg)
		}
		fmt.Printf("Checked %d indexed local document(s)\n", result.LocalDocumentsChecked)
		fmt.Printf("Skipped %d indexed local document(s)\n", result.LocalDocumentsSkipped)
		fmt.Printf("Removed %d stale local document(s)\n", result.LocalDocumentsRemoved)
		fmt.Printf("Removed %d orphaned HTML file(s)\n", result.HTMLRemoved)
		fmt.Printf("Removed %d orphaned favicon file(s)\n", result.FaviconRemoved)
	},
}
