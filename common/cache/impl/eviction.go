package impl

import "github.com/spf13/viper"

func (db *RedisClient) evict(requiredSize int) error {
	memFreed, err := db.Expire()
	if err != nil {
		return err
	}

	recordsOverflow, sizeOverflow, err := db.checkOverflow(requiredSize)

	if sizeOverflow {
		memToEvict := requiredSize - memFreed

		err = db.evictBySize(memToEvict)
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

	return nil
}

func (db *RedisClient) evictBySize(sizeRequired int) error {
	lowHitPages, err := db.ZRange("hits", 0, -1).Result()
	if err != nil {
		return err
	}

	for i := 0; sizeRequired > 0; i++ {
		pURL := lowHitPages[i]

		sizeFreed, err := db.Remove(pURL)
		if err != nil {
			return err
		}

		sizeRequired -= sizeFreed
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

func (db *RedisClient) isOverflowed(requiredSize int) (bool, error) {
	sizeOverflow, recordsOverflow, err := db.checkOverflow(requiredSize)

	return sizeOverflow || recordsOverflow, err
}

func (db *RedisClient) checkOverflow(requiredSize int) (bool, bool, error) {
	res, err := db.getTotalMemStats()
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

func (db *RedisClient) getTotalMemStats() (map[string]interface{}, error) {
	memst, err := db.Do("MEMORY", "STATS").Result()
	if err != nil {
		return nil, err
	}

	res, err := resToMap(memst)
	if err != nil {
		return nil, err
	}

	return res, nil
}
