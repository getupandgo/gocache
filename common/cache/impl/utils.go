package impl

import (
	"errors"

	"github.com/getupandgo/gocache/common/structs"

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

func parseZ(z *[]redis.Z) []structs.ScoredPage {
	zPages := *z
	hitRate := make([]structs.ScoredPage, len(zPages))

	for i, rtn := range zPages {
		formattedPageRec := structs.ScoredPage{rtn.Member.(string), int(rtn.Score)}
		hitRate[i] = formattedPageRec
	}

	return hitRate
}
