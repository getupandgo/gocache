package pagecache_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/getupandgo/gocache/mocks"
	"github.com/getupandgo/gocache/server/controllers"
	"github.com/getupandgo/gocache/utils/structs"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPageUpsert(t *testing.T) {
	content, err := ioutil.ReadFile("../../../mocks/Example Domain.html")
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	fmt.Println([]byte(content))

	body, _ := json.Marshal(structs.Page{"/example", content})

	request, _ := http.NewRequest("PUT", "/cache", bytes.NewBuffer(body))
	response := httptest.NewRecorder()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cacheMock := cache_mock.NewMockCacheClient(ctrl)
	cacheMock.EXPECT().UpsertPage(gomock.Any()).Return(nil)

	controllers.InitRouter(cacheMock).ServeHTTP(response, request)

	assert.Equal(t, 200, response.Code, "Ok is expected")
}

func TestTopPages(t *testing.T) {
	request, _ := http.NewRequest("GET", "/top", nil)
	response := httptest.NewRecorder()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cacheMock := cache_mock.NewMockCacheClient(ctrl)
	cacheMock.EXPECT().GetTopPages()

	controllers.InitRouter(cacheMock).ServeHTTP(response, request)

	assert.Equal(t, 200, response.Code, "Ok is expected")
}

func TestPageDeletion(t *testing.T) {
	body, _ := json.Marshal(structs.RemovePageBody{"/example"})

	request, _ := http.NewRequest("DELETE", "/cache", bytes.NewBuffer(body))
	response := httptest.NewRecorder()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cacheMock := cache_mock.NewMockCacheClient(ctrl)
	cacheMock.EXPECT().RemovePage(gomock.Any()).Return(nil)

	controllers.InitRouter(cacheMock).ServeHTTP(response, request)

	assert.Equal(t, 200, response.Code, "Ok is expected")
}
