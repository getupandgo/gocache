package pagecache_test

import (
	"bytes"
	"encoding/json"
	"github.com/getupandgo/gocache/common/structs"
	"github.com/getupandgo/gocache/mocks"
	"github.com/getupandgo/gocache/server/controllers"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
)

func sampleMultipartReq(uri string) (*http.Request, error) {
	content, err := ioutil.ReadFile("../../../mocks/Example Domain.html")
	if err != nil {
		return nil, err
	}

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("content", "example_file_name")
	if err != nil {
		return nil, err
	}
	part.Write(content)

	_ = writer.WriteField("url", "/example/test")

	if err = writer.Close(); err != nil {
		return nil, err
	}

	newReq, err := http.NewRequest("PUT", uri, body)
	if err != nil {
		return nil, err
	}

	newReq.Header.Add("Content-Type", writer.FormDataContentType())

	return newReq, nil
}

func TestPageUpsert(t *testing.T) {
	request, err := sampleMultipartReq("/cache")
	if err != nil {
		assert.Fail(t, err.Error())
	}

	response := httptest.NewRecorder()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cacheMock := cache_mock.NewMockCacheClient(ctrl)
	cacheMock.EXPECT().UpsertPage(gomock.Any()).Return(nil)

	controllers.InitRouter(cacheMock).ServeHTTP(response, request)

	assert.Equal(t, 200, response.Code, "Ok is expected")
}

func TestTopPagesRetrieval(t *testing.T) {
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
	cacheMock.EXPECT().RemovePage(gomock.Any()).Return(int64(0), nil)

	controllers.InitRouter(cacheMock).ServeHTTP(response, request)

	assert.Equal(t, 200, response.Code, "Ok is expected")
}
