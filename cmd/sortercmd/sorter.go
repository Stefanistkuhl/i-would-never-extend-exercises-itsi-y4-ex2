package sortercmd

import (
	"fmt"

	"github.com/stefanistkuhl/i-would-never-extend-exercises-itsi-y4-ex2/pkg/sorter"

	"github.com/spf13/cobra"
)

var ServeCmd = &cobra.Command{
	Use:   "serve",
	Short: "Starts the pcap sorter server",
	Long:  `Starts the pcap sorter server`,
	Run: func(cmd *cobra.Command, args []string) {
		sorter.StartSorter()
		fmt.Println("Server started now do other stuff")
	},
}
