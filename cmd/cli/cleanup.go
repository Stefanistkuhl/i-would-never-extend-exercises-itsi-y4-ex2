package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var cleanupCmd = &cobra.Command{
	Use:   "cleanup",
	Short: "Cleanup operations",
	Long:  `Commands for managing cleanup operations`,
}

var cleanupCandidatesCmd = &cobra.Command{
	Use:   "candidates",
	Short: "Get cleanup candidates",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient()
		if err != nil {
			return err
		}

		candidates, err := c.GetCleanupCandidates()
		if err != nil {
			return fmt.Errorf("failed to get cleanup candidates: %w", err)
		}

		return outputJSON(candidates)
	},
}

var cleanupExecuteCmd = &cobra.Command{
	Use:   "execute",
	Short: "Execute cleanup",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient()
		if err != nil {
			return err
		}

		result, err := c.CleanupExecute()
		if err != nil {
			return fmt.Errorf("failed to execute cleanup: %w", err)
		}

		return outputJSON(result)
	},
}
