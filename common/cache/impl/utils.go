package impl

import (
	"errors"
	"github.com/go-redis/redis"
)

func resToMap(res interface{}) (map[string]interface{}, error) {
	resSlice, ok := res.([]interface{})
	if !ok {
		return nil, errors.New("unsupportable value")
	}

	resMap := make(map[string]interface{})

	for i := 0; i < len(resSlice); i += 2 {
		k := resSlice[i].(string)

		resMap[k] = resSlice[i+1]
	}

	return resMap, nil
}

func ztoMap(z *[]redis.Z) map[int64]string { //todo fixme
	zPages := *z
	hitRate := make(map[int64]string, len(zPages))

	for _, rtn := range zPages {
		url := rtn.Member.(string)

		hitRate[int64(rtn.Score)] = url
	}

	return hitRate
}
