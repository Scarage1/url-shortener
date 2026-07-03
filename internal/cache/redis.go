package cache

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

var Ctx = context.Background()

func ConnectRedis(addr string) *redis.Client {

	client := redis.NewClient(
		&redis.Options{
			Addr: addr,
		},
	)

	err := client.Ping(Ctx).Err()

	if err != nil {
		panic(err)
	}

	fmt.Println("Connected to Redis")

	return client
}
