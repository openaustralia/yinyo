package cmd

import (
	"log"

	"github.com/openaustralia/yinyo/pkg/wrapper"
	"github.com/spf13/cobra"
)

var appPath, importPath, cachePath, envPath, runOutput, serverURL, buildCommand, runCommand string
var wrapperEnvironment map[string]string

func init() {
	wrapperCmd.Flags().StringVar(&appPath, "apppath", "/app", "herokuish app path")
	wrapperCmd.Flags().StringVar(&importPath, "importpath", "/tmp/app", "herokuish import path")
	wrapperCmd.Flags().StringVar(&cachePath, "cachepath", "/tmp/cache", "herokuish cache path")
	wrapperCmd.Flags().StringVar(&envPath, "envpath", "/tmp/env", "herokuish env path")
	wrapperCmd.Flags().StringVar(&runOutput, "output", "", "relative path to output file")
	wrapperCmd.Flags().StringVar(&serverURL, "server", "http://yinyo-server.yinyo-system:8080", "override yinyo server URL")
	wrapperCmd.Flags().StringVar(&buildCommand, "buildcommand", "/bin/herokuish buildpack build", "override the herokuish build command (for testing)")
	wrapperCmd.Flags().StringVar(&runCommand, "runcommand", "/bin/herokuish procfile start scraper", "override the herokuish run command (for testing)")
	wrapperCmd.Flags().StringToStringVar(&wrapperEnvironment, "env", map[string]string{}, "Set one or more environment variables (e.g. --env foo=twiddle,bar=blah)")
	rootCmd.AddCommand(wrapperCmd)
}

var wrapperCmd = &cobra.Command{
	Use:   "wrapper run_name run_token",
	Short: "Manages the building and running of a scraper",
	Long:  "Manages the building and running of a scraper inside a container. Used internally by the system.",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		err := wrapper.Run(args[0], args[1], serverURL, wrapper.Options{
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
