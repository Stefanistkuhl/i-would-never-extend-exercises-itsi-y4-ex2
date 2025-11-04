package cli

import (
	"github.com/stefanistkuhl/i-would-never-extend-exercises-itsi-y4-ex2/pkg/client"

	"github.com/spf13/cobra"
)

var (
	ServerFlag   string
	PasswordFlag string
	PortFlag     int
	SocketFlag   string
	RawFlag      bool
)

func getClient() (*client.Client, error) {
	return client.NewClient(ServerFlag, PasswordFlag, PortFlag, SocketFlag)
}

func AddAllCommands(rootCmd *cobra.Command) {
	// Files group
	filesCmd.AddCommand(filesListCmd)
	filesCmd.AddCommand(filesGetCmd)
	filesCmd.AddCommand(filesDownloadCmd)
	filesCmd.AddCommand(filesDeleteCmd)
	filesCmd.AddCommand(filesStatsCmd)
	filesCmd.AddCommand(filesByHostnameCmd)
	filesCmd.AddCommand(filesByScenarioCmd)
	rootCmd.AddCommand(filesCmd)

	// Stats group
	statsCmd.AddCommand(statsByIdCmd)
	statsCmd.AddCommand(statsSummaryCmd)
	statsCmd.AddCommand(statsByHostnameCmd)
	statsCmd.AddCommand(statsByScenarioCmd)
	rootCmd.AddCommand(statsCmd)

	// Archive group
	archiveCmd.AddCommand(archiveFileCmd)
	archiveCmd.AddCommand(archiveListCmd)
	archiveCmd.AddCommand(archiveStatusCmd)
	rootCmd.AddCommand(archiveCmd)

	// Config group
	configCmd.AddCommand(configGetCmd)
	configCmd.AddCommand(configUpdateCmd)
	rootCmd.AddCommand(configCmd)

	// Compression group
	compressionCmd.AddCommand(compressionFileCmd)
	compressionCmd.AddCommand(compressionTriggerCmd)
	rootCmd.AddCommand(compressionCmd)

	// Cleanup group
	cleanupCmd.AddCommand(cleanupCandidatesCmd)
	cleanupCmd.AddCommand(cleanupExecuteCmd)
	rootCmd.AddCommand(cleanupCmd)

	// Standalone commands
	rootCmd.AddCommand(searchCmd)
	rootCmd.AddCommand(exportCmd)
	rootCmd.AddCommand(healthCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(versionCmd)
}
