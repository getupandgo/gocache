package cache

import (
	"github.com/getupandgo/gocache/utils/structs"
	"github.com/go-redis/redis"
	"github.com/spf13/viper"
)

type (
	CacheClient interface {
		GetPage(url string) ([]byte, error)
		UpsertPage(pg *structs.Page) error
		RemovePage(url string) error
		GetTopPages() (map[string]int64, error)
	}

	RedisClient struct {
		redisClient   *redis.Client
		top_items_num int64
		ttl           int64
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

	limits := map[string]int64{
		"top_items": conf.GetInt64("app.top_items_num"),
		"ttl":       conf.GetInt64("app.ttl"),
	}

	return &RedisClient{redisClient, limits["top_items"], limits["ttl"]}
}

func (cc *RedisClient) GetPage(url string) ([]byte, error) {
	content, err := cc.redisClient.HGet(url, "content").Bytes()
	if err != nil {
		return nil, err
	}

	return content, nil
}

func (cc *RedisClient) UpsertPage(pg *structs.Page) error {
	return cc.redisClient.Watch(func(tx *redis.Tx) error {
		_, err := tx.Pipelined(
			func(pipe redis.Pipeliner) error {
				pipe.HSet(pg.Url, "content", pg.Content)

				pipe.ZIncr("hits", redis.Z{
					Score:  1,
					Member: pg.Url,
				})

				pipe.ZIncr("ttl", redis.Z{
					Score:  float64(cc.ttl),
					Member: pg.Url,
				})

				return nil
			})
		return err
	}, pg.Url)
}

func (cc *RedisClient) GetTopPages() (map[string]int64, error) {
	res, err := cc.redisClient.ZRangeWithScores("hits", 0, cc.top_items_num).Result()
	if err != nil {
		return nil, err
	}

	return ztoMap(&res), nil
}

func (cc *RedisClient) RemovePage(url string) error {
	return cc.redisClient.Watch(func(tx *redis.Tx) error {
		_, err := tx.Pipelined(
			func(pipe redis.Pipeliner) error {
				pipe.HDel(url, "content")

				pipe.ZRem("hits", url)

				pipe.ZRem("ttl", url)

				return nil
			})
		return err
	}, url)
}

func ztoMap(z *[]redis.Z) map[string]int64 {
	zPages := *z
	hitRate := make(map[string]int64, len(zPages))

	for _, rtn := range zPages {
		url := rtn.Member.(string)

		hitRate[url] = int64(rtn.Score)
	}

	return hitRate
}
