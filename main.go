package main

import (
	"os"

	"github.com/coder/flog"
	"github.com/go-redis/redis/v8"
	"github.com/spf13/cobra"
)

func main() {
	cmd := &cobra.Command{
		Use:     "github-enricher",
		Short:   "Enrich GitHub data",
		Example: "github-enricher < input.csv > output.csv",
		Run: func(cmd *cobra.Command, args []string) {
			const (
				RedisAddrEnv     = "REDIS_ADDR"
				RedisPasswordEnv = "REDIS_PASSWORD"
			)
			redisAddr, ok := os.LookupEnv(RedisAddrEnv)
			if !ok {
				flog.Fatal("%q must be provided", RedisAddrEnv)
			}

			redisPassword := os.Getenv(RedisPasswordEnv)

			e := engine{
				Log: flog.New(),
				Redis: redis.NewClient(&redis.Options{
					Addr:     redisAddr,
					Password: redisPassword,
				}),
			}
			err := e.Run(os.Stdout, os.Stdin)
			if err != nil {
				flog.Fatal("%v+", err)
			}
		},
	}
	err := cmd.Execute()
	if err != nil {
		flog.Fatal(err.Error())
	}
}
