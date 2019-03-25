package test_test

import (
	"errors"
	"strconv"

	"github.com/getupandgo/gocache/common/utils"

	"github.com/spf13/viper"

	"github.com/getupandgo/gocache/common/cache"
	"github.com/getupandgo/gocache/common/cache/impl"
	"github.com/getupandgo/gocache/common/structs"
	"github.com/go-redis/redis"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rs/zerolog/log"
)

const SERVICE_FIELDS_COUNT = 2 // "ttl" and "hits" records in sorted set

var redisInstance *redis.Client

var N = 100

var examplePage = `<div>
    <h1>Example Domain</h1>
    <p>This domain is established to be used for illustrative examples in documents. You may use this
    domain in examples without prior coordination or asking for permission.</p>
    <p><a href="http://www.iana.org/domains/example">More information...</a></p>
</div>`

var samplePageContent = []byte(examplePage)

var testPageURL = "/test/1"

func ReadConfig() {
	viper.AutomaticEnv()
	viper.AddConfigPath("../../../../config")
	viper.SetConfigName("default")
	viper.SetConfigType("yaml")

	if err := viper.ReadInConfig(); err != nil {
		log.Fatal().
			Err(err).
			Msg("Failed to get config file")
	}
}

func Init() (*impl.RedisClient, error) {
	host := viper.GetString("redis.host")
	port := viper.GetString("redis.port")

	rc := &impl.RedisClient{}
	rc.Client = redis.NewClient(&redis.Options{
		Addr: host + ":" + port,
		DB:   0,
	})

	redisInstance = rc.Client

	_, err := rc.Ping().Result()
	if err != nil {
		return nil, err
	}

	return rc, nil

}

func populateSamplePage(url string) (structs.Page, error) {
	defaultTLL, err := utils.CalculateTTLFromNow()
	page := structs.Page{
		URL:       url,
		Content:   samplePageContent,
		TTL:       defaultTLL,
		TotalSize: len(samplePageContent),
	}

	return page, err
}

var _ = Describe("Client", func() {
	var client cache.Page

	BeforeEach(func() {
		ReadConfig()

		var err error
		client, err = Init()
		Expect(client).NotTo(BeNil())
		Expect(err).To(BeNil())
	})

	AfterEach(func() {
		redisInstance.FlushAll()

		ReadConfig()
	})

	It("must insert page to cache", func() {
		samplePage, err := populateSamplePage(testPageURL)
		Expect(err).NotTo(HaveOccurred())

		_, err = client.Upsert(&samplePage)
		Expect(err).NotTo(HaveOccurred())

		rdContent, err := redisInstance.HGet(testPageURL, "content").Result()
		Expect(err).NotTo(HaveOccurred())

		Expect(string(samplePageContent)).To(Equal(rdContent))
	})

	It("must insert N pages", func() {
		for i := 0; i < N; i++ {
			strIdx := strconv.FormatInt(int64(i), 10)
			samplePage, err := populateSamplePage("/test/ins/" + strIdx)
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

		Expect(len(rdContent) - SERVICE_FIELDS_COUNT).To(Equal(N))
	})

	It("must return page content", func() {
		samplePage, err := populateSamplePage(testPageURL)
		Expect(err).NotTo(HaveOccurred())

		_, err = client.Upsert(&samplePage)
		Expect(err).NotTo(HaveOccurred())

		pContent, err := client.Get(testPageURL)
		Expect(err).NotTo(HaveOccurred())
		Expect(pContent).To(Equal(samplePageContent))

		rdContent, err := redisInstance.HGet(testPageURL, "content").Result()
		Expect(err).NotTo(HaveOccurred())

		Expect(string(samplePageContent)).To(Equal(rdContent))
	})

	It("must expire items with zero ttl", func() {
		for i := 0; i < 100; i++ {
			strIdx := strconv.FormatInt(int64(i), 10)
			_, err := client.Upsert(&structs.Page{
				URL:       "/test/ttl/" + strIdx,
				Content:   samplePageContent,
				TTL:       0,
				TotalSize: len(samplePageContent)})

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
		for i := 0; i < N; i++ {
			strIdx := strconv.FormatInt(int64(i), 10)
			samplePage, err := populateSamplePage("/test/top/" + strIdx)
			if err != nil {
				Fail(err.Error())
			}

			_, err = client.Upsert(&samplePage)
			if err != nil {
				Fail(err.Error())
			}
		}

		topItemsNum := viper.GetInt("cache.top_records_number")
		for i := 0; i < topItemsNum; i++ {
			strIdx := strconv.FormatInt(int64(i), 10)
			_, _ = client.Get("/test/top/" + strIdx)
		}

		topPages, err := client.Top()
		Expect(err).NotTo(HaveOccurred())

		Expect(topItemsNum).To(Equal(len(topPages)))
		Expect(topPages[topItemsNum-1].Hits).To(BeNumerically("==", 2))
	})

	It("must delete page from cache", func() {
		samplePage, err := populateSamplePage(testPageURL)
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
		maxRecordLen := 20
		viper.Set("limits.record.max_number", maxRecordLen)

		for i := 0; i < 30; i++ {
			strIdx := strconv.FormatInt(int64(i), 10)
			samplePage, err := populateSamplePage("/test/len/" + strIdx)
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
		Expect(maxRecordLen).To(Equal(len(rdContent) - SERVICE_FIELDS_COUNT))
	})

	It("must evict records by size limit", func() {
		sampleSize := 2000000
		viper.Set("limits.max_size", sampleSize)

		for i := 0; i <= N; i++ {
			strIdx := strconv.FormatInt(int64(i), 10)
			samplePage, err := populateSamplePage("/test/size/" + strIdx)
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
		maxAllowedAllocWithDelta := int64(sampleSize + 1000)

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