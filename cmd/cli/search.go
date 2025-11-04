package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var searchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search files",
	Long:  `Search for files with optional query string`,
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient()
		if err != nil {
			return err
		}

		query := ""
		if len(args) > 0 {
			query = args[0]
		}

		results, err := c.Search(query)
		if err != nil {
			return fmt.Errorf("failed to search: %w", err)
		}

		return outputJSON(results)
	},
}
