package config

import (
	"github.com/go-redis/redis/v8"
)

type RedisClient struct {
	*redis.Client
}

func NewRedis(addr string) (*RedisClient, error) {
	redis := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: "",
		DB:       0,
	})

	_, err := redis.Ping(redis.Context()).Result()
	if err != nil {
		return nil, err
	}

	return &RedisClient{redis}, nil
}
