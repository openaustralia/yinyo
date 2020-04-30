package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/openaustralia/yinyo/pkg/apiserver"
	"github.com/openaustralia/yinyo/pkg/commands"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/api/resource"
)

func getMandatoryEnv(name string) string {
	host, ok := os.LookupEnv(name)
	if !ok {
		log.Fatalf("environment variable %v was not set", name)
	}
	return host
}

func durationStringToSeconds(durationString string) int64 {
	duration, err := time.ParseDuration(durationString)
	if err != nil {
		log.Fatal(err)
	}
	return int64(duration.Seconds())
}

func memoryStringToBytes(memoryString string) int64 {
	memory, err := resource.ParseQuantity(memoryString)
	if err != nil {
		log.Fatal(err)
	}
	return memory.Value()
}

func main() {
	// Show the source of the error with the standard logger. Don't show date & time
	log.SetFlags(log.Lshortfile)

	var defaultMaxRunTimeString, maxRunTimeString, defaultMemoryString, maxMemoryString string

	var rootCmd = &cobra.Command{
		Use:   "server",
		Short: "Serves the Yinyo API",
		Run: func(cmd *cobra.Command, args []string) {
			server := apiserver.Server{}
			minioOptions := commands.MinioOptions{
				Host:      getMandatoryEnv("STORE_HOST"),
				Bucket:    getMandatoryEnv("STORE_BUCKET"),
				AccessKey: getMandatoryEnv("STORE_ACCESS_KEY"),
				SecretKey: getMandatoryEnv("STORE_SECRET_KEY"),
			}
			var tls bool
			if os.Getenv("REDIS_TLS") == "true" {
				tls = true
			}
			redisOptions := commands.RedisOptions{
				Address:  getMandatoryEnv("REDIS_HOST"),
				Password: getMandatoryEnv("REDIS_PASSWORD"),
				TLS:      tls,
			}
			authenticationURL := os.Getenv("AUTHENTICATION_URL")
			usageURL := os.Getenv("USAGE_URL")
			runDockerImage := getMandatoryEnv("RUN_DOCKER_IMAGE")
			options := commands.StartupOptions{Minio: minioOptions, Redis: redisOptions, AuthenticationURL: authenticationURL, UsageURL: usageURL}
			err := server.Initialise(
				&options,
				durationStringToSeconds(defaultMaxRunTimeString),
				durationStringToSeconds(maxRunTimeString),
				memoryStringToBytes(defaultMemoryString),
				memoryStringToBytes(maxMemoryString),
				runDockerImage,
			)
			if err != nil {
				log.Fatal(err)
			}
			server.Run(":8080")
		},
	}

	rootCmd.Flags().StringVar(&defaultMaxRunTimeString, "defaultmaxruntime", "1h", "Set the default maximum run time if the user doesn't say")
	rootCmd.Flags().StringVar(&maxRunTimeString, "maxruntime", "24h", "Set the global maximum run time that all runs can not exceed")
	rootCmd.Flags().StringVar(&defaultMemoryString, "defaultmemory", "1Gi", "Set the default memory that a run allocates if the user doesn't say")
	rootCmd.Flags().StringVar(&maxMemoryString, "maxmemory", "1.5Gi", "Set the maximum memory that a run can allocate")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
