redis-dev:
	docker build -t redis-dev ./mocks/redis_dev_img
	docker run -d -p 32769:6379 redis-dev
test:
	mockgen -destination ./mocks/mock_cache.go -package cache_mock github.com/getupandgo/gocache/common/cache Page
	go test ./...