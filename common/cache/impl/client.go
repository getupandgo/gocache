package impl

import (
	"github.com/getupandgo/gocache/common/config"
	"github.com/getupandgo/gocache/common/structs"
	"github.com/go-redis/redis"
	"github.com/spf13/viper"
	"sync"
	"time"
	"unsafe"
)

type RedisClient struct {
	conn   *redis.Client
	limits map[string]int

	recordsNum  int64
	overallSize int64

	cpMtx *sync.Mutex
}

func Init(conf *viper.Viper) (*RedisClient, error) {
	opts := map[string]string{
		"host": conf.GetString("redis.host"),
		"port": conf.GetString("redis.port"),
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr: opts["host"] + ":" + opts["port"],
		DB:   0,
	})

	limits, err := config.GetMapStringInt(conf, "limits")
	if err != nil {
		return nil, err
	}

	rc := &RedisClient{redisClient, limits, 0, 0, &sync.Mutex{}}

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

	ttl := time.Now().Add(time.Second * time.Duration(cc.limits["ttl"]))

	return cc.conn.Watch(func(tx *redis.Tx) error {
		_, err := tx.Pipelined(
			func(pipe redis.Pipeliner) error {
				pipe.HSet(pg.Url, "content", pg.Content)

				pipe.ZIncr("hits", redis.Z{
					Score:  1,
					Member: pg.Url,
				})

				pipe.ZIncr("ttl", redis.Z{
					Score:  float64(ttl.Unix()),
					Member: pg.Url,
				})

				return nil
			})
		return err
	}, pg.Url)
}

func (cc *RedisClient) GetTopPages() (map[int64]string, error) {
	topPagesNum := int64(cc.limits["top_records_num"])

	res, err := cc.conn.ZRevRangeWithScores("hits", 0, topPagesNum).Result()
	if err != nil {
		return nil, err
	}

	return ztoMap(&res), nil
}

func (cc *RedisClient) RemovePage(url string) (int64, error) {
	pipe := cc.conn.TxPipeline()

	pipe.HDel(url, "content")
	pipe.ZRem("hits", url)
	pipe.ZRem("ttl", url)

	_, err := pipe.Exec()
	if err != nil {
		return 0, err
	}

	if err := cc.syncCapacity(); err != nil {
		return 0, err
	}

	return 1, nil
}

func (cc *RedisClient) syncCapacity() error {
	res, err := cc.getMemStats()
	if err != nil {
		return err
	}

	cc.recordsNum = res["keys.count"].(int64)
	cc.overallSize = res["total.allocated"].(int64)

	return nil
}

func (cc *RedisClient) RemoveExpiredRecords() (int64, error) {
	nowFromEpoch := time.Now().Unix()

	pipe := cc.conn.TxPipeline()

	sPages, err := pipe.ZRange("ttl", 0, nowFromEpoch).Result()
	if err != nil {
		return 0, err
	}

	for _, sPage := range sPages {
		pipe.HDel(sPage, "content").Result()
		pipe.ZRem("hits", sPage)
		pipe.ZRem("ttl", sPage)
	}

	_, err = pipe.Exec()
	if err != nil {
		return 0, err
	}

	err = cc.syncCapacity()
	if err != nil {
		return 0, err
	}

	return int64(len(sPages)), err
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

func (cc *RedisClient) getMemStats() (map[string]interface{}, error) {
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

func (cc *RedisClient) enoughCapacityForRecord(requiredSize int) bool {
	maxRecords := cc.recordsNum+1 < int64(cc.limits["max_record_num"])
	maxSize := cc.overallSize+1 < int64(cc.limits["max_cache_size"]+requiredSize)

	return maxSize || maxRecords
}
