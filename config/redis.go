package config

import "github.com/go-redis/redis/v8"

var Redis *redis.Client

func NewRedis() {
	opt, err := redis.ParseURL("redis://localhost:6379/0")
	if err != nil {
		panic(err)
	}

	redis := redis.NewClient(opt)
	Redis = redis
}
