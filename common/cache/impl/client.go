package impl

import (
	"sync"
	"time"

	"github.com/getupandgo/gocache/common/structs"
	"github.com/go-redis/redis"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

type RedisClient struct {
	*redis.Client

	sync.Mutex
}

func Init() (*RedisClient, error) {
	opts := map[string]string{
		"host": viper.GetString("redis.host"),
		"port": viper.GetString("redis.port"),
	}

	rc := &RedisClient{}
	rc.Client = redis.NewClient(&redis.Options{
		Addr: opts["host"] + ":" + opts["port"],
		DB:   0,
	})

	_, err := rc.Ping().Result()
	if err != nil {
		return nil, err
	}

	return rc, nil

}

func (cc *RedisClient) GetPage(url string) ([]byte, error) {
	pipe := cc.TxPipeline()

	pipe.ZIncr("hits", redis.Z{
		Score:  1,
		Member: url,
	})

	content, err := cc.HGet(url, "content").Bytes()
	if err != nil {
		return nil, err
	}

	_, err = pipe.Exec()
	if err != nil {
		return nil, err
	}

	return content, nil
}

func (cc *RedisClient) UpsertPage(pg *structs.Page) error {
	isOverflowed, err := cc.isOverflowed(pg.TotalSize)
	if err != nil {
		return err
	}

	if isOverflowed {
		err = cc.cleanCache(pg.TotalSize)
		if err != nil {
			return err
		}
	}

	defaultTTL := viper.GetInt("limits.record.ttl")

	TTLFromNow := time.
		Now().
		Add(time.Second * time.Duration(defaultTTL)).
		Unix()

		//return cc.conn.Watch(func(tx *redis.Tx) error {
		//	_, err := tx.Pipelined(
	_, err = cc.Pipelined(
		func(pipe redis.Pipeliner) error {
			pipe.HSet(pg.URL, "content", pg.Content)

			pipe.ZIncr("hits", redis.Z{
				Score:  1,
				Member: pg.URL,
			})

			pipe.ZIncr("ttl", redis.Z{
				Score:  float64(TTLFromNow),
				Member: pg.URL,
			})

			return nil
		})
	return err
	//}, pg.URL)
}

func (cc *RedisClient) GetTopPages() ([]structs.ScoredPage, error) {
	topPagesNum := viper.GetInt64("limits.top_records_number")

	res, err := cc.ZRevRangeWithScores("hits", 0, topPagesNum).Result()
	if err != nil {
		return nil, err
	}

	return parseZ(&res), nil
}

func (cc *RedisClient) RemovePage(url string) (int, error) {
	var memUsageRes *redis.IntCmd

	_, err := cc.TxPipelined(
		func(pipe redis.Pipeliner) error {
			memUsageRes = pipe.MemoryUsage(url)

			pipe.HDel(url, "content")
			pipe.ZRem("hits", url)
			pipe.ZRem("ttl", url)

			return nil
		})

	if err != nil {
		return 0, err
	}

	bytesFreed, err := memUsageRes.Result()

	return int(bytesFreed), err
}

func (cc *RedisClient) RemoveExpiredRecords() (int, error) {
	nowFromEpoch := time.Now().Unix()

	sPages, err := cc.ZRange("ttl", 0, nowFromEpoch).Result()
	if err != nil {
		return 0, err
	}

	var freedTotal int

	for _, sPage := range sPages {
		sizeFreed, err := cc.RemovePage(sPage)
		if err != nil {
			log.Warn().
				Err(err).
				Msg("Failed to remove expired items")

			continue
		}

		freedTotal += sizeFreed
	}

	return freedTotal, err
}
