package cache

import (
	"github.com/go-redis/redis"
	"github.com/spf13/viper"
)

type (
	CacheClient interface {
		UpsertPage(pg *PageMsg) error
		GetTopPages() (error, map[string]int64)
		RemovePage(pageUrl string) error
	}

	RedisClient struct {
		redisClient   *redis.Client
		top_items_num int64
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

	return &RedisClient{redisClient, conf.GetInt64("app.top_items_num")}
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

func (cc *RedisClient) GetTopPages() (error, map[string]int64) {
	res, err := cc.redisClient.ZRangeWithScores("hits", 0, cc.top_items_num).Result()
	if err != nil {
		return err, nil
	}

	return nil, ztoMap(&res)
}

func (cc *RedisClient) RemovePage(pageUrl string) error {
	return cc.redisClient.Watch(func(tx *redis.Tx) error {
		_, err := tx.Pipelined(
			func(pipe redis.Pipeliner) error {
				pipe.HDel(pageUrl, "content")

				pipe.ZRem("hits", pageUrl)

				pipe.ZRem("ttl", pageUrl)

				return nil
			})
		return err
	}, pageUrl)
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
