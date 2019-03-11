package cache

import (
	"github.com/go-redis/redis"
	"github.com/spf13/viper"
)

type RedisClient struct {
	redisClient *redis.Client
}

func (cc *RedisClient) Init(conf *viper.Viper) {
	options := map[string]string{
		"host": conf.GetString("cache.host"),
		"port": conf.GetString("cache.port"),
	}

	cc.redisClient = redis.NewClient(&redis.Options{
		Addr: options["host"] + ":" + options["port"],
		DB:   0,
	})
}
