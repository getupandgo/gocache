package impl

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

const SERVICE_FIELDS_COUNT = 2 // "ttl" and "hits" records in sorted set

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
		maxCacheSize := viper.GetInt("limits.max_size")

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

	if len(lowHitPages) == 0 {
		log.Warn().
			Err(err).
			Msg("Eviction by size - no items to evict")

		return nil
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
	currRecCount, currSize, err := db.getMemStats()
	if err != nil {
		return false, false, err
	}

	maxRecCount := viper.GetInt64("limits.record.max_number")
	maxSize := viper.GetInt64("limits.max_size")

	recordsOverflow := currRecCount >= maxRecCount+SERVICE_FIELDS_COUNT
	sizeOverflow := currSize+int64(requiredSize) >= maxSize

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
