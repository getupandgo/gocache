package impl

import (
	"time"

	"github.com/getupandgo/gocache/common/utils"

	"github.com/rs/zerolog/log"
)

const serviceFieldsCount = 2 // "ttl" and "hits" records in sorted set

func (db *RedisClient) evictIfFull(recordSize int) error {
	recordsOverflow, sizeOverflow, err := db.isOverflowed(recordSize)
	if err != nil {
		return err
	}

	isFull := recordsOverflow || sizeOverflow
	if !isFull {
		return nil
	}

	_, err = db.Expire()
	if err != nil {
		return err
	}

	recordsOverflow, sizeOverflow, err = db.isOverflowed(recordSize)

	if sizeOverflow {
		maxCacheSize := int(db.options.MaxSize)

		_, currCacheSize, err := db.getMemStats()
		if err != nil {
			return err
		}

		newCacheSize := recordSize + int(currCacheSize)

		err = db.evictBySize(newCacheSize, maxCacheSize)
		if err != nil {
			return err
		}
	}

	if recordsOverflow {
		err = db.evictByCapacity()
		if err != nil {
			return err
		}
	}

	return err
}

func (db *RedisClient) evictBySize(newCacheSize int, maxCacheSize int) error {
	lowHitPages, err := db.ZRange("hits", 0, -1).Result()
	if err != nil {
		return err
	}

	for _, page := range lowHitPages {
		sizeFreed, err := db.Remove(page)
		if err != nil {
			log.Warn().
				Err(err).
				Msg("Eviction by size error")

			continue
		}

		newCacheSize -= sizeFreed

		if newCacheSize < maxCacheSize {
			return nil
		}
	}

	return err
}

func (db *RedisClient) evictByCapacity() error {
	lowHitPages, err := db.ZRange("hits", 0, -1).Result()
	if err != nil {
		return err
	}

	pURL := lowHitPages[0]

	_, err = db.Remove(pURL)

	return err
}

func (db *RedisClient) isOverflowed(requiredSize int) (bool, bool, error) {
	currentRecCount, currSize, err := db.getMemStats()
	if err != nil {
		return false, false, err
	}

	maxRecords := db.options.MaxCount + serviceFieldsCount
	newSize := currSize + int64(requiredSize)

	recordsOverflow := currentRecCount >= maxRecords
	sizeOverflow := newSize >= db.options.MaxSize

	return recordsOverflow, sizeOverflow, nil
}

func (db *RedisClient) getMemStats() (int64, int64, error) {
	memst, err := db.Do("MEMORY", "STATS").Result()
	if err != nil {
		return 0, 0, err
	}

	res, err := resToMap(memst)
	if err != nil {
		return 0, 0, err
	}

	currRecCount := res["keys.count"].(int64)
	currSize := res["total.allocated"].(int64)

	return currRecCount, currSize, nil
}

func (db *RedisClient) WatchExpiredRecords() error {
	interval := db.options.ExpirationInterval
	pollDuration, err := time.ParseDuration(interval)
	if err != nil {
		return err
	}

	ticker := time.NewTicker(pollDuration)

	go func() {
		for t := range ticker.C {
			_ = t

			if _, err := db.Expire(); err != nil {
				log.Info().
					Err(err).
					Msg("Failed to delete expired records")
			}
		}
	}()

	return nil
}

func (db *RedisClient) isExpired(url string) (bool, error) {
	recordTTL, err := db.ZScore("ttl", url).Result()
	if err != nil {
		return false, err
	}

	timeNow, err := utils.CalculateTTLFromNow("0")
	if err != nil {
		return false, err
	}

	return int64(recordTTL) < timeNow, nil
}
