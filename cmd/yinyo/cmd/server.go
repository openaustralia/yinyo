package cmd

import (
	"log"

	"github.com/openaustralia/yinyo/pkg/apiserver"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(serverCmd)
}

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Serves the Yinyo API",
	Run: func(cmd *cobra.Command, args []string) {
		server := apiserver.Server{}
		err := server.Initialise()
		if err != nil {
			log.Fatal(err)
		}
		server.Run(":8080")
	},
}
