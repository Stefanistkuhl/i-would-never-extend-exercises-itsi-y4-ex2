package cmd

import (
	"fmt"

	"github.com/stefanistkuhl/i-would-never-extend-exercises-itsi-y4-ex2/cmd/sortercmd"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:           "pcap-sorter",
	Short:         "Quick utility to manage caputre files.",
	Long:          `Quick utility to manage caputre files for itsi lab ex2.`,
	SilenceErrors: true,
	SilenceUsage:  true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
	},
}

func init() {
	cobra.OnFinalize()
	rootCmd.AddCommand(sortercmd.ServeCmd)

}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Printf("%v\n", err)
	}
}
