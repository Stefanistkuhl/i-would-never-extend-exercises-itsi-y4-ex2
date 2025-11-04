package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
)

var exportCmd = &cobra.Command{
	Use:   "export [output]",
	Short: "Export the entire store to a gzip archive",
	Long:  `Exports the database and all capture files to a tar.gz archive. If output path is not specified, defaults to pcapstore-export-YYYYMMDD-HHMMSS.tar.gz in the current directory.`,
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		outputPath := ""
		if len(args) > 0 {
			outputPath = args[0]
		} else {
			cwd, _ := os.Getwd()
			filename := fmt.Sprintf("pcapstore-export-%s.tar.gz", time.Now().Format("20060102-150405"))
			outputPath = filepath.Join(cwd, filename)
		}

		c, err := getClient()
		if err != nil {
			return err
		}

		fmt.Printf("Exporting store to %s...\n", outputPath)
		if err := c.ExportStore(outputPath); err != nil {
			return fmt.Errorf("failed to export store: %w", err)
		}

		stat, err := os.Stat(outputPath)
		if err == nil {
			sizeMB := float64(stat.Size()) / (1024 * 1024)
			fmt.Printf("Export completed successfully: %s (%.2f MB)\n", outputPath, sizeMB)
		} else {
			fmt.Printf("Export completed successfully: %s\n", outputPath)
		}

		return nil
	},
}
