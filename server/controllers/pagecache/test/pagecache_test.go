package test_test

import (
	"bytes"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"

	"github.com/getupandgo/gocache/mocks"
	"github.com/getupandgo/gocache/server/controllers"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Pagecache", func() {
	const testRequestSize int64 = 20480000

	var (
		ctrl      *gomock.Controller
		cacheMock *cache_mock.MockPage
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		cacheMock = cache_mock.NewMockPage(ctrl)
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("should retrieve page", func() {
		request, err := sampleMultipartReq("/cache/upsert")
		Expect(err).To(BeNil())

		response := httptest.NewRecorder()

		cacheMock.EXPECT().Upsert(gomock.Any()).Return(true, nil)

		controllers.InitRouter(cacheMock, testRequestSize).ServeHTTP(response, request)

		Expect(response.Code).To(Equal(200))
	})

	It("should upsert new page", func() {
		values := url.Values{}
		values.Add("url", "/example/test")

		request, _ := http.NewRequest("POST", "/cache/get", strings.NewReader(values.Encode()))
		request.Header.Add("Content-Type", "application/x-www-form-urlencoded")

		response := httptest.NewRecorder()

		cacheMock.EXPECT().Get(gomock.Any()).Return([]byte{}, nil)

		controllers.InitRouter(cacheMock, testRequestSize).ServeHTTP(response, request)

		Expect(response.Code).To(Equal(200))
	})

	It("should return top pages", func() {
		request, _ := http.NewRequest("GET", "/cache/top", nil)
		response := httptest.NewRecorder()

		cacheMock.EXPECT().Top()

		controllers.InitRouter(cacheMock, testRequestSize).ServeHTTP(response, request)

		Expect(response.Code).To(Equal(200))
	})

	It("should delete page", func() {
		values := url.Values{}
		values.Add("url", "/example/test")

		request, _ := http.NewRequest("POST", "/cache/delete", strings.NewReader(values.Encode()))
		request.Header.Add("Content-Type", "application/x-www-form-urlencoded")

		response := httptest.NewRecorder()

		cacheMock.EXPECT().Remove(gomock.Any()).Return(0, nil)

		controllers.InitRouter(cacheMock, testRequestSize).ServeHTTP(response, request)

		Expect(response.Code).To(Equal(200))
	})
})

func sampleMultipartReq(uri string) (*http.Request, error) {
	content, err := ioutil.ReadFile("../../../../mocks/Example Domain.html")
	if err != nil {
		return nil, err
	}

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)

	urlField, err := writer.CreateFormField("url")
	if err != nil {
		return nil, err
	}
	_, err = urlField.Write([]byte("/example/test"))
	if err != nil {
		return nil, err
	}

	part, err := writer.CreateFormFile("content", "example_file_name")
	if err != nil {
		return nil, err
	}
	_, err = part.Write(content)
	if err != nil {
		return nil, err
	}

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
