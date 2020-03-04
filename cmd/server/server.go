package main

import (
	"fmt"
	"log"
	"os"

	"github.com/openaustralia/yinyo/pkg/apiserver"
	"github.com/openaustralia/yinyo/pkg/commands"
	"github.com/spf13/cobra"
)

func main() {
	// Show the source of the error with the standard logger. Don't show date & time
	log.SetFlags(log.Lshortfile)

	var maxRunTime int64

	var rootCmd = &cobra.Command{
		Use:   "server",
		Short: "Serves the Yinyo API",
		Run: func(cmd *cobra.Command, args []string) {
			server := apiserver.Server{}
			minioOptions := commands.MinioOptions{
				Host:      os.Getenv("STORE_HOST"),
				Bucket:    os.Getenv("STORE_BUCKET"),
				AccessKey: os.Getenv("STORE_ACCESS_KEY"),
				SecretKey: os.Getenv("STORE_SECRET_KEY"),
			}
			redisOptions := commands.RedisOptions{
				Address:  "redis:6379",
				Password: os.Getenv("REDIS_PASSWORD"),
			}
			authenticationURL := os.Getenv("AUTHENTICATION_URL")
			usageURL := os.Getenv("USAGE_URL")
			options := commands.StartupOptions{Minio: minioOptions, Redis: redisOptions, AuthenticationURL: authenticationURL, UsageURL: usageURL}
			err := server.Initialise(&options, maxRunTime)
			if err != nil {
				log.Fatal(err)
			}
			server.Run(":8080")
		},
	}

	rootCmd.Flags().Int64Var(&maxRunTime, "maxruntime", 86400, "Set the global maximum run time in seconds that all runs can not exceed")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
