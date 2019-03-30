package test_data

import (
	"io/ioutil"
	"os"
	"path"

	"github.com/getupandgo/gocache/common/structs"
	"github.com/getupandgo/gocache/common/utils"
)

func PopulatePage(url string, ttl string) (structs.Page, error) {
	unixTTL, err := utils.CalculateTTLFromNow(ttl)

	samplePagePath := path.Join(os.Getenv("GOPATH"), "src/github.com/getupandgo/gocache/mocks/Example Domain.html")

	content, err := ioutil.ReadFile(samplePagePath)
	if err != nil {
		return structs.Page{}, err
	}

	page := structs.Page{
		URL:       url,
		Content:   content,
		TTL:       unixTTL,
		TotalSize: len(url) + len(content),
	}

	return page, err
}
