gen:
	@echo "  >  Generating dependency files..."
	mockgen -destination ./mocks/mock_cache.go -package cache_mock github.com/getupandgo/gocache/utils/cache CacheClient