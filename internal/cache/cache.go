package cache

import (
	"context"
	"fmt"
	"time"

	"golang-boilerplate/internal/config"
	"golang-boilerplate/internal/constants"
	"golang-boilerplate/internal/errors"
)

// Cache defines the interface for cache operations
type Cache interface {
	// Get retrieves a value from cache
	Get(ctx context.Context, key string) (string, error)

	// Set stores a value in cache with expiration
	Set(ctx context.Context, key string, value string, expiration time.Duration) error

	// Delete removes a value from cache
	Delete(ctx context.Context, key string) error

	// Exists checks if a key exists in cache
	Exists(ctx context.Context, key string) (bool, error)

	// Close closes the cache connection
	Close() error
}

func ProvideCache(cfg *config.Config) (Cache, error) {
	switch cfg.CacheProvider {
	case constants.CacheProviderRedis:
		redisCache, err := NewRedisCache(cfg)
		if err != nil {
			return nil, errors.CacheError("Failed to initialize Redis cache", err).
				WithOperation("initialize_cache").
				WithResource("cache")
		}
		return redisCache, nil
	default:
		return nil, errors.InternalError("Invalid cache provider", fmt.Errorf("invalid cache provider: %s", cfg.CacheProvider)).
			WithOperation("initialize_cache").
			WithResource("cache").
			WithContext("cache_provider", cfg.CacheProvider)
	}
}
