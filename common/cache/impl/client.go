package impl

import (
	"strconv"
	"time"

	"github.com/getupandgo/gocache/common/utils"

	"github.com/getupandgo/gocache/common/structs"
	"github.com/go-redis/redis"
	"github.com/rs/zerolog/log"
)

type RedisClient struct {
	*redis.Client
	options utils.DBOptions
}

func Init(opts utils.DBOptions) (*RedisClient, error) {
	rc := &RedisClient{}
	rc.Client = redis.NewClient(&redis.Options{
		Addr: opts.Connection,
		DB:   0,
	})

	rc.options = opts

	err := rc.WatchExpiredRecords()
	if err != nil {
		return nil, err
	}
	_, err = rc.Ping().Result()
	if err != nil {
		return nil, err
	}

	return rc, nil

}

func (db *RedisClient) Get(url string) ([]byte, error) {
	var getCommand *redis.StringCmd
	var getPage func(string) error

	getPage = func(url string) error {
		getTx := func(tx *redis.Tx) error {
			recordExpired, err := db.isExpired(url)
			if err != nil {
				return err
			}

			if recordExpired {
				_, err := db.Remove(url)
				if err != nil {
					return err
				}
			} else {
				_, err = tx.Pipelined(
					func(pipe redis.Pipeliner) error {
						pipe.ZIncr("hits", redis.Z{
							Score:  1,
							Member: url,
						})

						getCommand = db.HGet(url, "content")

						return nil
					})
			}

			return err
		}

		err := db.Watch(getTx, url)

		if err == redis.TxFailedErr {
			log.Warn().
				Err(err).
				Msg("Failed to insert page with url " + url + ", retry")

			return getPage(url)
		}

		return err
	}

	err := getPage(url)
	if err != nil {
		return nil, err
	}

	if getCommand == nil {
		return nil, nil
	}

	pageContent, err := getCommand.Bytes()
	if err != nil {
		return nil, err
	}

	return pageContent, nil
}

func (db *RedisClient) Upsert(pg *structs.Page) (bool, error) {
	err := db.evictIfFull(pg.TotalSize)

	var setCommand *redis.BoolCmd
	var upsertPage func(string) error

	upsertPage = func(url string) error {
		upsertTx := func(tx *redis.Tx) error {
			_, err = tx.Pipelined(
				func(pipe redis.Pipeliner) error {
					setCommand = pipe.HSet(pg.URL, "content", pg.Content)

					pipe.ZIncr("hits", redis.Z{
						Score:  1,
						Member: pg.URL,
					})

					pipe.ZAdd("ttl", redis.Z{
						Score:  float64(pg.TTL),
						Member: pg.URL,
					})

					return err
				})

			return err
		}

		err := db.Watch(upsertTx, url)

		if err == redis.TxFailedErr {
			log.Warn().
				Err(err).
				Msg("Failed to insert page with url " + url + ", retry")

			return upsertPage(url)
		}

		return err
	}

	err = upsertPage(pg.URL)
	if err != nil {
		return false, err
	}

	return setCommand.Result()
}

func (db *RedisClient) Top() ([]structs.ScoredPage, error) {
	topPagesNum := db.options.TopRecordsCount

	topPages, err := db.ZRevRangeWithScores("hits", 0, topPagesNum-1).Result()
	if err != nil {
		return nil, err
	}

	return parseZ(&topPages), nil
}

func (db *RedisClient) Remove(url string) (int, error) {
	var memoryUsageCommand *redis.IntCmd
	var removePage func(string) error

	removePage = func(url string) error {
		removeTx := func(tx *redis.Tx) error {
			_, err := tx.Pipelined(
				func(pipe redis.Pipeliner) error {
					memoryUsageCommand = pipe.MemoryUsage(url)

					pipe.HDel(url, "content")
					pipe.ZRem("hits", url)
					pipe.ZRem("ttl", url)

					return nil

				})

			return err
		}

		err := db.Watch(removeTx, url)

		if err == redis.TxFailedErr {
			log.Warn().
				Err(err).
				Msg("Failed to remove page with url " + url + ", retry")

			return removePage(url)
		}

		return err
	}

	err := removePage(url)
	if err != nil {
		return 0, err
	}

	bytesFreed, err := memoryUsageCommand.Result()

	return int(bytesFreed), err
}

func (db *RedisClient) Expire() (int, error) {
	nowUnix := time.Now().Unix()

	sPages, err := db.ZRangeByScore("ttl", redis.ZRangeBy{
		Min: "0",
		Max: strconv.Itoa(int(nowUnix)),
	}).Result()

	if err != nil {
		return 0, err
	}

	var freedTotal int

	for _, sPage := range sPages {
		sizeFreed, err := db.Remove(sPage)
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
