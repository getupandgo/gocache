package utils

import (
	"time"
)

func CalculateTTLFromNow(ttl string) (int64, error) {
	defaultTTL, err := time.ParseDuration(ttl)
	if err != nil {
		return 0, err
	}

	ttlInUnix := time.
		Now().
		Add(defaultTTL).
		Unix()

	return ttlInUnix, nil
}
