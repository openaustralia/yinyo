package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "wrapper",
	Short: "Manages the building and running of a scraper",
	Long:  "Manages the building and running of a scraper inside a container. Used internally by the system.",
}

// Execute makes it all happen
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
