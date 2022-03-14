package main

import (
	"context"
	"os"

	"github.com/coder/flog"
	"github.com/go-redis/redis/v8"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
)

func main() {
	var enrichers []string
	cmd := &cobra.Command{
		Use:     "github-enricher",
		Short:   "Enrich GitHub data (https://github.com/ammario/github-enricher)",
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
			rd := redis.NewClient(&redis.Options{
				Addr:     redisAddr,
				Password: redisPassword,
			})

			err := rd.Ping(context.Background()).Err()
			if err != nil {
				flog.Fatal("redis ping: %+v", err)
			}

			allEnrichers, err := setupEnrichers()
			if err != nil {
				flog.Fatal("setup enrichers: %+v", err)
			}

			e := engine{
				Log: flog.New(),
				Enrichers: lo.Map(enrichers, func(name string, _ int) enricher {
					enricher, _ := lo.Find(allEnrichers, func(e enricher) bool {
						return e.FieldName == name
					})
					return cachedEnricher(flog.New().WithPrefix("cache: "), rd, enricher)
				}),
			}
			err = e.Run(os.Stdout, os.Stdin)
			if err != nil {
				flog.Fatal("%+v", err)
			}
		},
	}
	cmd.Flags().StringSliceVarP(
		&enrichers,
		"enrichers",
		"e",
		[]string{"email", "name", "gender"},
		"The list of enabled enrichers",
	)
	err := cmd.Execute()
	if err != nil {
		flog.Fatal(err.Error())
	}
}
