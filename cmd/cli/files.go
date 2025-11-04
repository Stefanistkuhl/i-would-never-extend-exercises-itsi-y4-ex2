package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/spf13/cobra"
)

var filesCmd = &cobra.Command{
	Use:   "files",
	Short: "File operations",
	Long:  `Commands for managing capture files`,
}

var filesListCmd = &cobra.Command{
	Use:     "list",
	Short:   "List all files",
	Aliases: []string{"ls"},
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient()
		if err != nil {
			return err
		}

		files, err := c.ListFiles()
		if err != nil {
			return fmt.Errorf("failed to list files: %w", err)
		}

		return outputJSON(files)
	},
}

var filesGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get file details",
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

		file, err := c.GetFile(id)
		if err != nil {
			return fmt.Errorf("failed to get file: %w", err)
		}

		return outputJSON(file)
	},
}

var filesDownloadCmd = &cobra.Command{
	Use:   "download <id> [output]",
	Short: "Download a file",
	Args:  cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid file ID: %w", err)
		}

		outputPath := args[0] + ".pcap"
		if len(args) > 1 {
			outputPath = args[1]
		} else {
			cwd, _ := os.Getwd()
			outputPath = filepath.Join(cwd, fmt.Sprintf("file-%d.pcap", id))
		}

		c, err := getClient()
		if err != nil {
			return err
		}

		if err := c.DownloadFile(id, outputPath); err != nil {
			return fmt.Errorf("failed to download file: %w", err)
		}

		fmt.Printf("Downloaded file to: %s\n", outputPath)
		return nil
	},
}

var filesDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a file",
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

		if err := c.DeleteFile(id); err != nil {
			return fmt.Errorf("failed to delete file: %w", err)
		}

		fmt.Printf("File %d deleted successfully\n", id)
		return nil
	},
}

var filesStatsCmd = &cobra.Command{
	Use:   "stats <id>",
	Short: "Get file statistics",
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

var filesByHostnameCmd = &cobra.Command{
	Use:   "by-hostname <hostname>",
	Short: "List files by hostname",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient()
		if err != nil {
			return err
		}

		files, err := c.GetFilesByHostname(args[0])
		if err != nil {
			return fmt.Errorf("failed to get files by hostname: %w", err)
		}

		return outputJSON(files)
	},
}

var filesByScenarioCmd = &cobra.Command{
	Use:   "by-scenario <scenario>",
	Short: "List files by scenario",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient()
		if err != nil {
			return err
		}

		files, err := c.GetFilesByScenario(args[0])
		if err != nil {
			return fmt.Errorf("failed to get files by scenario: %w", err)
		}

		return outputJSON(files)
	},
}
