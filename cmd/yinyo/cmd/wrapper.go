package cmd

import (
	"github.com/openaustralia/yinyo/pkg/wrapper"
	"github.com/spf13/cobra"
)

var appPath, importPath, cachePath, runOutput, serverURL, buildCommand, runCommand string
var wrapperEnvironment map[string]string

func init() {
	wrapperCmd.Flags().StringVar(&appPath, "app", "/app", "herokuish app path")
	wrapperCmd.Flags().StringVar(&importPath, "import", "/tmp/app", "herokuish import path")
	wrapperCmd.Flags().StringVar(&cachePath, "cache", "/tmp/cache", "herokuish cache path")
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
		wrapper.Run(wrapper.Options{
			RunName:      args[0],
			RunToken:     args[1],
			ServerURL:    serverURL,
			ImportPath:   importPath,
			CachePath:    cachePath,
			AppPath:      appPath,
			Environment:  wrapperEnvironment,
			BuildCommand: buildCommand,
			RunCommand:   runCommand,
			RunOutput:    runOutput,
		})
	},
}
