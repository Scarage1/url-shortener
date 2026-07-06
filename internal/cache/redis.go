package cache

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

var Ctx = context.Background()

func ConnectRedis(addr string) (*redis.Client, error) {

	client := redis.NewClient(
		&redis.Options{
			Addr: addr,
		},
	)

	if err := client.Ping(Ctx).Err(); err != nil {
		return nil, fmt.Errorf("connect to redis: %w", err)
	}

	return client, nil
}
