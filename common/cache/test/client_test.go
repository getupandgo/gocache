package test_test

import (
	"errors"
	"strconv"

	"github.com/getupandgo/gocache/mocks/test_data"

	"github.com/getupandgo/gocache/common/utils"

	"github.com/spf13/viper"

	"github.com/getupandgo/gocache/common/cache"
	"github.com/getupandgo/gocache/common/cache/impl"
	"github.com/go-redis/redis"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const serviceFieldsCount = 2 // "ttl" and "hits" records in sorted set

var _ = Describe("Client", func() {
	const (
		itemsToInsert = 100
		redisDevConn  = "0.0.0.0:32769"
		testPageURL   = "/test/1"
	)

	var (
		redisInstance *redis.Client
		client        cache.Page

		defaultTTL string
	)

	BeforeEach(func() {
		utils.ReadConfig("../../../config")
		defaultTTL = viper.GetString("limits.record.ttl")

		DBOptions := utils.ReadDbOptions()
		DBOptions.Connection = redisDevConn

		var err error
		client, err = impl.Init(DBOptions)

		redisInstance = redis.NewClient(&redis.Options{
			Addr: DBOptions.Connection,
			DB:   0,
		})

		Expect(client).NotTo(BeNil())
		Expect(err).To(BeNil())
	})

	AfterEach(func() {
		redisInstance.FlushAll()
	})

	It("must insert page to cache", func() {
		samplePage, err := test_data.PopulatePage(testPageURL, defaultTTL)
		Expect(err).NotTo(HaveOccurred())

		_, err = client.Upsert(&samplePage)
		Expect(err).NotTo(HaveOccurred())

		dbContent, err := redisInstance.HGet(testPageURL, "content").Result()
		Expect(err).NotTo(HaveOccurred())

		Expect(string(samplePage.Content)).To(Equal(dbContent))
	})

	It("must insert N pages", func() {
		for i := 0; i < itemsToInsert; i++ {
			strIdx := strconv.FormatInt(int64(i), 10)
			samplePage, err := test_data.PopulatePage("/test/ins/"+strIdx, defaultTTL)
			if err != nil {
				Fail(err.Error())
			}

			_, err = client.Upsert(&samplePage)
			if err != nil {
				Fail(err.Error())
			}
		}

		rdContent, err := redisInstance.Keys("*").Result()
		Expect(err).NotTo(HaveOccurred())

		Expect(len(rdContent) - serviceFieldsCount).To(Equal(itemsToInsert))
	})

	It("must return page content", func() {
		samplePage, err := test_data.PopulatePage(testPageURL, defaultTTL)
		Expect(err).NotTo(HaveOccurred())

		_, err = client.Upsert(&samplePage)
		Expect(err).NotTo(HaveOccurred())

		pContent, err := client.Get(testPageURL)
		Expect(err).NotTo(HaveOccurred())
		Expect(pContent).To(Equal(samplePage.Content))

		rdContent, err := redisInstance.HGet(testPageURL, "content").Result()
		Expect(err).NotTo(HaveOccurred())

		Expect(string(samplePage.Content)).To(Equal(rdContent))
	})

	It("must not return expired page content", func() {
		samplePage, err := test_data.PopulatePage(testPageURL, "-1000")
		Expect(err).NotTo(HaveOccurred())

		_, err = client.Upsert(&samplePage)
		Expect(err).NotTo(HaveOccurred())

		pContent, err := client.Get(testPageURL)
		Expect(err).NotTo(HaveOccurred())
		Expect(pContent).To(BeNil())

		rdContent, err := redisInstance.HGet(testPageURL, "content").Result()
		Expect(err).To(Equal(redis.Nil))
		Expect(rdContent).To(BeEmpty())
	})

	It("must expire items with zero ttl", func() {
		for i := 0; i < 100; i++ {
			strIdx := strconv.FormatInt(int64(i), 10)

			samplePage, err := test_data.PopulatePage("/test/ttl/"+strIdx, "0")
			if err != nil {
				Fail(err.Error())
			}
			_, err = client.Upsert(&samplePage)
			if err != nil {
				Fail(err.Error())
			}
		}

		_, err := client.Expire()
		Expect(err).NotTo(HaveOccurred())

		rdContent, err := redisInstance.Keys("*").Result()
		Expect(err).NotTo(HaveOccurred())

		Expect(len(rdContent)).To(BeZero())
	})

	It("must return top pages", func() {
		for i := 0; i < itemsToInsert; i++ {
			strIdx := strconv.FormatInt(int64(i), 10)
			samplePage, err := test_data.PopulatePage("/test/top/"+strIdx, defaultTTL)
			if err != nil {
				Fail(err.Error())
			}

			_, err = client.Upsert(&samplePage)
			if err != nil {
				Fail(err.Error())
			}
		}

		topItemsNum := viper.GetInt("cache.top_records_count")
		for i := 0; i < topItemsNum; i++ {
			strIdx := strconv.FormatInt(int64(i), 10)
			_, err := client.Get("/test/top/" + strIdx)
			if err != nil {
				Fail(err.Error())
			}
		}

		topPages, err := client.Top()
		Expect(err).NotTo(HaveOccurred())

		Expect(topItemsNum).To(Equal(len(topPages)))
		Expect(topPages[topItemsNum-1].Hits).To(BeNumerically("==", 2))
	})

	It("must delete page from cache", func() {
		samplePage, err := test_data.PopulatePage(testPageURL, defaultTTL)
		Expect(err).NotTo(HaveOccurred())

		_, err = client.Upsert(&samplePage)
		Expect(err).NotTo(HaveOccurred())

		bytesFreed, err := client.Remove(testPageURL)
		Expect(err).NotTo(HaveOccurred())
		Expect(bytesFreed).NotTo(BeZero())

		_, err = redisInstance.HGet(testPageURL, "content").Result()
		Expect(err).To(Equal(redis.Nil))
	})

	It("must evict records by len limit", func() {
		DBOptions := utils.ReadDbOptions()
		DBOptions.Connection = redisDevConn
		DBOptions.MaxCount = 20

		client, err := impl.Init(DBOptions)
		if err != nil {
			Fail(err.Error())
		}

		for i := 0; i < 30; i++ {
			strIdx := strconv.FormatInt(int64(i), 10)
			samplePage, err := test_data.PopulatePage("/test/len/"+strIdx, defaultTTL)
			if err != nil {
				Fail(err.Error())
			}

			_, err = client.Upsert(&samplePage)
			if err != nil {
				Fail(err.Error())
			}
		}

		rdContent, err := redisInstance.Keys("*").Result()
		Expect(err).NotTo(HaveOccurred())

		resultingLen := int64(len(rdContent) - serviceFieldsCount)

		Expect(DBOptions.MaxCount).To(Equal(resultingLen))
	})

	It("must evict records by size limit", func() {
		DBOptions := utils.ReadDbOptions()
		DBOptions.Connection = redisDevConn
		DBOptions.MaxSize = 2000000

		client, err := impl.Init(DBOptions)
		if err != nil {
			Fail(err.Error())
		}

		for i := 0; i <= itemsToInsert; i++ {
			strIdx := strconv.FormatInt(int64(i), 10)
			samplePage, err := test_data.PopulatePage("/test/size/"+strIdx, defaultTTL)
			if err != nil {
				Fail(err.Error())
			}

			_, err = client.Upsert(&samplePage)
			if err != nil {
				Fail(err.Error())
			}
		}

		memst, err := redisInstance.Do("MEMORY", "STATS").Result()
		Expect(err).NotTo(HaveOccurred())

		res, err := resToMap(memst)
		Expect(err).NotTo(HaveOccurred())

		totalAlloc := res["total.allocated"].(int64)
		maxAllowedAllocWithDelta := int64(DBOptions.MaxSize)

		Expect(maxAllowedAllocWithDelta > totalAlloc).To(BeTrue())
	})

})

func resToMap(res interface{}) (map[string]interface{}, error) {
	resSlice, ok := res.([]interface{})
	if !ok {
		return nil, errors.New("unsupportable value")
	}

	resMap := make(map[string]interface{})

	for i := 0; i < len(resSlice); i += 2 {
		k := resSlice[i].(string)

		resMap[k] = resSlice[i+1]
	}

	return resMap, nil
}
