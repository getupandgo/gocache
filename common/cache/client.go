package cache

import (
	"github.com/getupandgo/gocache/common/cache/impl"
	"github.com/getupandgo/gocache/common/structs"
)

type CacheClient interface {
	GetPage(url string) ([]byte, error)
	UpsertPage(pg *structs.Page) error
	RemovePage(url string) (int, error)
	GetTopPages() ([]structs.ScoredPage, error)
	RemoveExpiredRecords() (int, error)
}

func Init() (CacheClient, error) {
	return impl.Init()
}
