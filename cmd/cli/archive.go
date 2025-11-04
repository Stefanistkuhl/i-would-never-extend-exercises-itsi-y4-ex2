package cli

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
)

var archiveCmd = &cobra.Command{
	Use:   "archive",
	Short: "Archive operations",
	Long:  `Commands for managing file archiving`,
}

var archiveFileCmd = &cobra.Command{
	Use:   "file <id>",
	Short: "Archive a file",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid file ID: %w", err)
		}

		client, err := getClient()
		if err != nil {
			return err
		}

		if err := client.ArchiveFile(id); err != nil {
			return fmt.Errorf("failed to archive file: %w", err)
		}

		fmt.Printf("File %d archived successfully\n", id)
		return nil
	},
}

var archiveListCmd = &cobra.Command{
	Use:   "list",
	Short: "List archive status",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient()
		if err != nil {
			return err
		}

		archive, err := c.GetArchive()
		if err != nil {
			return fmt.Errorf("failed to get archive: %w", err)
		}

		return outputJSON(archive)
	},
}

var archiveStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Get archive status",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient()
		if err != nil {
			return err
		}

		status, err := c.ArchiveStatus()
		if err != nil {
			return fmt.Errorf("failed to get archive status: %w", err)
		}

		return outputJSON(status)
	},
}
