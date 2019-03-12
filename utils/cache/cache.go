package cache

import (
	"github.com/go-redis/redis"
	"github.com/spf13/viper"
)

type (
	CacheClient interface {
		UpsertPage(pg *PageMsg) error
	}

	RedisClient struct {
		redisClient *redis.Client
	}

	PageMsg struct {
		URL         string
		PageContent string
	}
)

func Init(conf *viper.Viper) CacheClient {
	opts := map[string]string{
		"host": conf.GetString("redis.host"),
		"port": conf.GetString("redis.port"),
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr: opts["host"] + ":" + opts["port"],
		DB:   0,
	})

	return &RedisClient{redisClient}
}

func (cc *RedisClient) UpsertPage(pg *PageMsg) error {
	return cc.redisClient.Watch(func(tx *redis.Tx) error {
		_, err := tx.Pipelined(
			func(pipe redis.Pipeliner) error {
				pipe.HSet(pg.URL, "content", pg.PageContent)

				pipe.ZIncr("hits", redis.Z{
					Score:  1,
					Member: pg.URL,
				})

				pipe.ZIncr("ttl", redis.Z{
					Score:  1,
					Member: pg.URL,
				})

				return nil
			})
		return err
	}, pg.URL)
}
