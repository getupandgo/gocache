package impl

import "github.com/spf13/viper"

func (cc *RedisClient) cleanCache(requiredSize int) error {
	memFreed, err := cc.RemoveExpiredRecords()
	if err != nil {
		return err
	}

	recordsOverflow, sizeOverflow, err := cc.checkOverflow(requiredSize)

	if sizeOverflow {
		memToEvict := requiredSize - memFreed

		err = cc.evictBySize(memToEvict)
		if err != nil {
			return err
		}
	}

	if recordsOverflow {
		err = cc.evictByCapacity()

		if err != nil {
			return err
		}
	}

	return nil
}

func (cc *RedisClient) evictBySize(sizeRequired int) error {
	lowHitPages, err := cc.ZRange("hits", 0, -1).Result()
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
	lowHitPages, err := cc.ZRange("hits", 0, -1).Result()
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
	memst, err := cc.Do("MEMORY", "STATS").Result()
	if err != nil {
		return nil, err
	}

	res, err := resToMap(memst)
	if err != nil {
		return nil, err
	}

	return res, nil
}
