package utils

import (
	"time"

	"github.com/spf13/viper"
)

func CalculateTTLFromNow() (int64, error) {
	defaultTTL, err := time.ParseDuration(viper.GetString("limits.record.ttl"))
	if err != nil {
		return 0, err
	}

	ttlInUnix := time.
		Now().
		Add(defaultTTL).
		Unix()

	return ttlInUnix, nil
}
