package cli

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
)

var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Statistics operations",
	Long:  `Commands for viewing statistics`,
}

var statsByIdCmd = &cobra.Command{
	Use:   "by-id <id>",
	Short: "Get file statistics by ID",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid file ID: %w", err)
		}

		c, err := getClient()
		if err != nil {
			return err
		}

		stats, err := c.GetFileStats(id)
		if err != nil {
			return fmt.Errorf("failed to get file stats: %w", err)
		}

		return outputJSON(stats)
	},
}

var statsSummaryCmd = &cobra.Command{
	Use:   "summary",
	Short: "Get summary statistics",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient()
		if err != nil {
			return err
		}

		summary, err := c.GetSummary()
		if err != nil {
			return fmt.Errorf("failed to get summary: %w", err)
		}

		return outputJSON(summary)
	},
}

var statsByHostnameCmd = &cobra.Command{
	Use:   "by-hostname",
	Short: "Get statistics grouped by hostname",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient()
		if err != nil {
			return err
		}

		stats, err := c.GetStatsByHostname()
		if err != nil {
			return fmt.Errorf("failed to get stats by hostname: %w", err)
		}

		return outputJSON(stats)
	},
}

var statsByScenarioCmd = &cobra.Command{
	Use:   "by-scenario",
	Short: "Get statistics grouped by scenario",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient()
		if err != nil {
			return err
		}

		stats, err := c.GetStatsByScenario()
		if err != nil {
			return fmt.Errorf("failed to get stats by scenario: %w", err)
		}

		return outputJSON(stats)
	},
}
