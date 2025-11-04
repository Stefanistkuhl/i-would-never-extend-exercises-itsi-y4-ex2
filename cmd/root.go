package cmd

import (
	"fmt"

	"github.com/stefanistkuhl/i-would-never-extend-exercises-itsi-y4-ex2/cmd/cli"
	"github.com/stefanistkuhl/i-would-never-extend-exercises-itsi-y4-ex2/cmd/sortercmd"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:           "pcapstore",
	Short:         "Quick utility to manage capture files",
	Long:          `Quick utility to manage capture files for itsi lab ex2.`,
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

	rootCmd.PersistentFlags().StringVarP(&cli.ServerFlag, "server", "s", "", "Server URL (overrides config and env)")
	rootCmd.PersistentFlags().StringVarP(&cli.PasswordFlag, "password", "p", "", "Password (overrides config and env)")
	rootCmd.PersistentFlags().IntVar(&cli.PortFlag, "port", 0, "Server port (overrides config and env)")
	rootCmd.PersistentFlags().StringVar(&cli.SocketFlag, "socket", "", "Unix socket path (overrides config and env)")
	rootCmd.PersistentFlags().BoolVar(&cli.RawFlag, "raw", false, "Output raw JSON (for piping to jq)")

	rootCmd.AddCommand(sortercmd.ServeCmd)

	cli.AddAllCommands(rootCmd)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Printf("%v\n", err)
	}
}
