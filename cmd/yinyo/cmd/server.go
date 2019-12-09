package cmd

import (
	"github.com/openaustralia/yinyo/pkg/server"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(serverCmd)
}

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Serves the Yinyo API",
	Run: func(cmd *cobra.Command, args []string) {
		server.Run()
	},
}
