package redis

import (
	"github.com/go-redis/redis/v8"
)

type Client struct {
	*redis.Client
}

func New(host string, port string) (*Client, error) {
	redis := redis.NewClient(&redis.Options{
		Addr:     host + ":" + port,
		Password: "",
		DB:       0,
	})

	_, err := redis.Ping(redis.Context()).Result()
	if err != nil {
		return nil, err
	}

	return &Client{redis}, nil
}
