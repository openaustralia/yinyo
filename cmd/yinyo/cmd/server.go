package cmd

import (
	"log"
	"os"

	"github.com/openaustralia/yinyo/pkg/apiserver"
	"github.com/openaustralia/yinyo/pkg/commands"
	"github.com/spf13/cobra"
)

var maxRunTime int64

func init() {
	serverCmd.Flags().Int64Var(&maxRunTime, "maxruntime", 86400, "Set the global maximum run time in seconds that all runs can not exceed")
	rootCmd.AddCommand(serverCmd)
}

var serverCmd = &cobra.Command{
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
		err := server.Initialise(commands.StartupOptions{Minio: minioOptions, Redis: redisOptions}, maxRunTime)
		if err != nil {
			log.Fatal(err)
		}
		server.Run(":8080")
	},
}
