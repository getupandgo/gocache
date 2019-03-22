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
	conn *redis.Client

	sync.Mutex
}

func Init() (*RedisClient, error) {
	opts := map[string]string{
		"host": viper.GetString("redis.host"),
		"port": viper.GetString("redis.port"),
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr: opts["host"] + ":" + opts["port"],
		DB:   0,
	})

	_, err := redisClient.Ping().Result()
	if err != nil {
		return nil, err
	}

	rc := &RedisClient{}
	rc.conn = redisClient

	return rc, nil

}

func (cc *RedisClient) GetPage(url string) ([]byte, error) {
	pipe := cc.conn.TxPipeline()

	pipe.ZIncr("hits", redis.Z{
		Score:  1,
		Member: url,
	})

	content, err := cc.conn.HGet(url, "content").Bytes()
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
	newRecSize := int(unsafe.Sizeof(&pg))

	if !cc.enoughCapacityForRecord(newRecSize) {
		if err := cc.evictRecords(newRecSize); err != nil {
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
	_, err = cc.conn.Pipelined(
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

	res, err := cc.conn.ZRevRangeWithScores("hits", 0, topPagesNum).Result()
	if err != nil {
		return nil, err
	}

	return parseZ(&res), nil
}

func (cc *RedisClient) RemovePage(url string) (int, error) {
	var memUsageRes *redis.IntCmd

	_, err := cc.conn.TxPipelined(
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

	sPages, err := cc.conn.ZRange("ttl", 0, nowFromEpoch).Result()
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

func (cc *RedisClient) evictRecords(sizeRequired int) error {
	_, err := cc.RemoveExpiredRecords()
	if err != nil {
		return err
	}

	if !cc.enoughCapacityForRecord(sizeRequired) {
		for !cc.enoughCapacityForRecord(sizeRequired) {
			pipe := cc.conn.TxPipeline()

			res, err := pipe.ZRangeWithScores("hits", 0, 1).Result()
			if err != nil {
				return err
			}

			page := res[2].Member.(string)

			pipe.HDel(page, "content")
			pipe.ZRem("hits", page)
			pipe.ZRem("ttl", page)

			_, err = pipe.Exec()
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (cc *RedisClient) evictBySize(sizeRequired int) error {
	lowHitPages, err := cc.conn.ZRange("hits", 0, -1).Result()
	if err != nil {
		return err
	}

	for i := 0; sizeRequired > 0; i++ {
		pURL := lowHitPages[i]

		sizeFreed, err := cc.RemovePage(pURL)
		if err != nil {
			return err
		}

		sizeRequired -= sizeFreed
	}

	return err
}

func (cc *RedisClient) evictByCapacity() error {
	lowHitPages, err := cc.conn.ZRange("hits", 0, -1).Result()
	if err != nil {
		return err
	}

	pURL := lowHitPages[0]

	_, err = cc.RemovePage(pURL)

	return err
}

func (cc *RedisClient) isOverflowed(requiredSize int) (bool, error) {
	sizeOverflow, recordsOverflow, err := cc.checkOverflow(requiredSize)

	return sizeOverflow || recordsOverflow, err
}

func (cc *RedisClient) checkOverflow(requiredSize int) (bool, bool, error) {
	res, err := cc.getRedisMemStats()
	if err != nil {
		return false, false, err
	}

	currRecCount := res["keys.count"].(int64)
	currSize := res["total.allocated"].(int64)

	maxRecCount := viper.GetInt64("limits.record.max_number")
	maxSize := viper.GetInt64("limits.max_size")

	recordsOverflow := currRecCount+1 > maxRecCount+2
	sizeOverflow := currSize+int64(requiredSize) > maxSize

	return recordsOverflow, sizeOverflow, nil
}

func (cc *RedisClient) getRedisMemStats() (map[string]interface{}, error) {
	memst, err := cc.conn.Do("MEMORY", "STATS").Result()
	if err != nil {
		return nil, err
	}

	res, err := resToMap(memst)
	if err != nil {
		return nil, err
	}

	return res, nil
}
