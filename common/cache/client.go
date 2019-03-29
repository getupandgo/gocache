package cache

import (
	"github.com/getupandgo/gocache/common/cache/impl"
	"github.com/getupandgo/gocache/common/structs"
	"github.com/getupandgo/gocache/common/utils"
)

type Page interface {
	Get(url string) ([]byte, error)
	Upsert(pg *structs.Page) (bool, error)
	Remove(url string) (int, error)
	Top() ([]structs.ScoredPage, error)
	Expire() (int, error)
}

func Init(options utils.DBOptions) (Page, error) {
	return impl.Init(options)
}
