package cli

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
)

var compressionCmd = &cobra.Command{
	Use:   "compression",
	Short: "Compression operations",
	Long:  `Commands for managing file compression`,
}

var compressionFileCmd = &cobra.Command{
	Use:   "file <id>",
	Short: "Compress a file",
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

		if err := c.CompressFile(id); err != nil {
			return fmt.Errorf("failed to compress file: %w", err)
		}

		fmt.Printf("File %d compression initiated\n", id)
		return nil
	},
}

var compressionTriggerCmd = &cobra.Command{
	Use:   "trigger",
	Short: "Trigger compression for pending files",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient()
		if err != nil {
			return err
		}

		result, err := c.CompressTrigger()
		if err != nil {
			return fmt.Errorf("failed to trigger compression: %w", err)
		}

		return outputJSON(result)
	},
}
