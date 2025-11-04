package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var healthCmd = &cobra.Command{
	Use:   "health",
	Short: "Check server health",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient()
		if err != nil {
			return err
		}

		health, err := c.Health()
		if err != nil {
			return fmt.Errorf("failed to check health: %w", err)
		}

		return outputJSON(health)
	},
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Get server status",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient()
		if err != nil {
			return err
		}

		status, err := c.Status()
		if err != nil {
			return fmt.Errorf("failed to get status: %w", err)
		}

		return outputJSON(status)
	},
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Get server version",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient()
		if err != nil {
			return err
		}

		version, err := c.Version()
		if err != nil {
			return fmt.Errorf("failed to get version: %w", err)
		}

		return outputJSON(version)
	},
}
