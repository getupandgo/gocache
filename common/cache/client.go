package cache

import (
	"github.com/getupandgo/gocache/common/cache/impl"
	"github.com/getupandgo/gocache/common/structs"
)

type Page interface {
	Get(url string) ([]byte, error)
	Upsert(pg *structs.Page) error
	Remove(url string) (int, error)
	Top() ([]structs.ScoredPage, error)
	Expire() (int, error)
}

func Init() (Page, error) {
	return impl.Init()
}
