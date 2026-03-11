package db

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

var Redis *redis.Client

func InitRedis(addr, password string) error {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
	})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		return fmt.Errorf("connect redis: %w", err)
	}
	Redis = rdb
	return nil
}
