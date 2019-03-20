package cache

import (
	"github.com/getupandgo/gocache/common/cache/impl"
	"github.com/getupandgo/gocache/common/structs"
	"github.com/spf13/viper"
)

type CacheClient interface {
	GetPage(url string) ([]byte, error)
	UpsertPage(pg *structs.Page) error
	RemovePage(url string) (int64, error)
	GetTopPages() (map[int64]string, error)
	RemoveExpiredRecords() (int64, error)
}

func Init(conf *viper.Viper) (CacheClient, error) {
	return impl.Init(conf)
}
