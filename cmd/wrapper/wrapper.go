package main

import (
	"fmt"
	"log"
	"os"

	"github.com/openaustralia/yinyo/pkg/apiclient"
	"github.com/openaustralia/yinyo/pkg/protocol"
	"github.com/openaustralia/yinyo/pkg/wrapper"
	"github.com/spf13/cobra"
)

func main() {
	// Show the source of the error with the standard logger. Don't show date & time
	log.SetFlags(log.Lshortfile)

	var appPath, importPath, cachePath, envPath, runOutput, serverURL, buildCommand, runCommand string
	var wrapperEnvironment map[string]string

	var rootCmd = &cobra.Command{
		Use:   "wrapper run_name",
		Short: "Manages the building and running of a scraper",
		Long:  "Manages the building and running of a scraper inside a container. Used internally by the system.",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			run := &apiclient.Run{
				Run:    protocol.Run{Name: args[0]},
				Client: apiclient.New(serverURL),
			}
			err := wrapper.Run(run, &wrapper.Options{
				ImportPath:   importPath,
				CachePath:    cachePath,
				AppPath:      appPath,
				EnvPath:      envPath,
				Environment:  wrapperEnvironment,
				BuildCommand: buildCommand,
				RunCommand:   runCommand,
				RunOutput:    runOutput,
			})
			if err != nil {
				log.Fatal(err)
			}
		},
	}

	rootCmd.Flags().StringVar(&appPath, "apppath", "/app", "herokuish app path")
	rootCmd.Flags().StringVar(&importPath, "importpath", "/tmp/app", "herokuish import path")
	rootCmd.Flags().StringVar(&cachePath, "cachepath", "/tmp/cache", "herokuish cache path")
	rootCmd.Flags().StringVar(&envPath, "envpath", "/tmp/env", "herokuish env path")
	rootCmd.Flags().StringVar(&runOutput, "output", "", "relative path to output file")
	rootCmd.Flags().StringVar(&serverURL, "server", "http://yinyo-server.yinyo-system:8080", "override yinyo server URL")
	rootCmd.Flags().StringVar(&buildCommand, "buildcommand", "/bin/herokuish buildpack build", "override the herokuish build command (for testing)")
	rootCmd.Flags().StringVar(&runCommand, "runcommand", "/bin/herokuish procfile start scraper", "override the herokuish run command (for testing)")
	rootCmd.Flags().StringToStringVar(&wrapperEnvironment, "env", map[string]string{}, "Set one or more environment variables (e.g. --env foo=twiddle,bar=blah)")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
