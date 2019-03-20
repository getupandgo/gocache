package cache

import (
	"github.com/getupandgo/gocache/common/cache/impl"
	"github.com/getupandgo/gocache/common/structs"
)

type CacheClient interface {
	GetPage(url string) ([]byte, error)
	UpsertPage(pg *structs.Page) error
	RemovePage(url string) (int64, error)
	GetTopPages() (map[int64]string, error)
	RemoveExpiredRecords() (int64, error)
}

func Init() (CacheClient, error) {
	return impl.Init()
}
