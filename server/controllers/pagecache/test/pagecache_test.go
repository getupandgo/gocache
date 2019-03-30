package test_test

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strconv"
	"strings"

	"github.com/getupandgo/gocache/mocks/test_data"

	"github.com/getupandgo/gocache/mocks/db"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"

	"github.com/getupandgo/gocache/common/structs"
	"github.com/getupandgo/gocache/common/utils"

	"github.com/getupandgo/gocache/server/controllers"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const testPageUrl = "/example/test/controllers"

var _ = Describe("Pagecache", func() {
	var (
		ctrl       *gomock.Controller
		cacheMock  *cache_mock.MockPage
		samplePage structs.Page
		testRouter *gin.Engine
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		cacheMock = cache_mock.NewMockPage(ctrl)

		utils.ReadConfig("../../../../config")
		defaultTTL := viper.GetString("limits.record.ttl")
		maxReqSize := viper.GetInt64("limits.record.max_size")

		testRouter = controllers.InitRouter(cacheMock, maxReqSize, defaultTTL)

		var err error
		samplePage, err = test_data.PopulatePage(testPageUrl, defaultTTL)
		Expect(err).To(BeNil())
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("should upsert new page", func() {
		request, err := test_data.PopulateUpsertReq("/cache/upsert", testPageUrl)
		Expect(err).To(BeNil())

		response := httptest.NewRecorder()

		cacheMock.EXPECT().Upsert(gomock.Any()).DoAndReturn(
			func(pg *structs.Page) (bool, error) {
				pageValue := *pg

				if reflect.DeepEqual(pageValue, samplePage) {
					return true, nil
				} else {
					Fail("Struct formed by controller is not equal to sample one")
				}

				return false, nil
			})

		testRouter.ServeHTTP(response, request)

		Expect(response.Code).To(Equal(200))
	})

	It("should retrieve page", func() {
		values := url.Values{}
		values.Add("url", testPageUrl)

		request, err := http.NewRequest("POST", "/cache/get", strings.NewReader(values.Encode()))
		request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		Expect(err).To(BeNil())

		response := httptest.NewRecorder()

		cacheMock.EXPECT().Get(gomock.Any()).DoAndReturn(
			func(url string) ([]byte, error) {
				if url == testPageUrl {
					return samplePage.Content, nil
				} else {
					Fail("Struct formed by controller is not equal to sample one")
				}

				return nil, nil
			})

		testRouter.ServeHTTP(response, request)

		responseBody, err := ioutil.ReadAll(response.Body)
		Expect(err).To(BeNil())

		responseString, err := strconv.Unquote(string(responseBody))
		Expect(err).To(BeNil())

		testContentString := string(samplePage.Content)

		Expect(response.Code).To(Equal(200))
		Expect(responseString).To(Equal(testContentString))
	})

	It("should return top pages", func() {
		request, _ := http.NewRequest("GET", "/cache/top", nil)
		response := httptest.NewRecorder()

		cacheMock.EXPECT().Top()

		testRouter.ServeHTTP(response, request)

		Expect(response.Code).To(Equal(200))
	})

	It("should delete page", func() {
		values := url.Values{}
		values.Add("url", testPageUrl)

		request, _ := http.NewRequest("POST", "/cache/delete", strings.NewReader(values.Encode()))
		request.Header.Add("Content-Type", "application/x-www-form-urlencoded")

		response := httptest.NewRecorder()

		cacheMock.EXPECT().Remove(gomock.Any()).Return(0, nil)

		testRouter.ServeHTTP(response, request)

		Expect(response.Code).To(Equal(200))
	})
})
