up:
	docker-compose up --build
redis-test:
	docker build -t redis-test ./mocks/redis_docker_test
	docker run -d -p 32769:6379 redis-test
test:
	mockgen -destination ./mocks/db/cache.go -package cache_mock github.com/getupandgo/gocache/common/cache Page
	export GIN_MODE=test; ginkgo -r
