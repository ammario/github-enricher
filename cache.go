package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/coder/flog"
	"github.com/go-redis/redis/v8"
)

func cachedEnricher(log *flog.Logger, rd *redis.Client, e enricher) enricher {
	ce := e
	ce.Run = func(ctx context.Context, row map[string]string) (string, error) {
		start := time.Now()
		// Enrichment should be deterministic on dependencies, so use that as key.
		var key strings.Builder
		fmt.Fprintf(&key, "%v_", e.FieldName)
		for _, depKey := range ce.CacheDeps {
			fmt.Fprintf(&key, "[%v:%v]", depKey, row[depKey])
		}

		cachedValue, err := rd.Get(ctx, key.String()).Result()
		if err == redis.Nil {
			newValue, err := e.Run(ctx, row)
			if err != nil {
				return "", err
			}
			err = rd.Set(ctx, key.String(), newValue, 0).Err()
			if err != nil {
				return "", fmt.Errorf("redis set: %w", err)
			}
			log.Info("MISS: [%0.3fs] on %s = %s", time.Since(start).Seconds(), key.String(), newValue)
			return newValue, nil
		} else if err != nil {
			return "", fmt.Errorf("redis get: %w", err)
		}
		return cachedValue, nil
	}
	return ce
}
