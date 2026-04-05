package cache

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisConnectionConfig struct {
	RedisHost     string
	RedisPort     int
	RedisPassword string
}

func (cfg *RedisConnectionConfig) DSN() string {
	return fmt.Sprintf("%s:%d", cfg.RedisHost, cfg.RedisPort)
}

var (
	ErrCacheMiss    = errors.New("cache entry not found")
	ErrCacheExpired = errors.New("cache entry expired")
)

const redisRequestCacheNamespace = "policy-service:request-cache"

type RedisRequestCache struct {
	Client *redis.Client
	keyNS  string // namespace prefix e.g. "policy-service:request-cache"
}

// NewRedisRequestCache creates a Redis-backed request cache using values from Config.
func NewRedisRequestCache(ctx context.Context, cfg *RedisConnectionConfig) (*RedisRequestCache, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.DSN(),
		Password: cfg.RedisPassword,
	})
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		err := client.Close()
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("redis ping failed: %w", err)
	}

	log.Println("Pinged your deployment. You successfully connected to Redis!")

	return &RedisRequestCache{Client: client, keyNS: redisRequestCacheNamespace}, nil
}

func (s *RedisRequestCache) buildKey(key string) string {
	return s.keyNS + ":" + key
}

// Set stores a cache value with TTL.
func (s *RedisRequestCache) Set(ctx context.Context, key string, value string, expiresIn time.Duration) error {
	if key == "" || value == "" {
		return errors.New("key and value required")
	}
	cacheKey := s.buildKey(key)
	return s.Client.Set(ctx, cacheKey, value, expiresIn).Err()
}

// Get retrieves a cached value and calculates expiresAt using TTL.
func (s *RedisRequestCache) Get(ctx context.Context, key string) (string, time.Time, error) {
	cacheKey := s.buildKey(key)
	val, err := s.Client.Get(ctx, cacheKey).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", time.Time{}, ErrCacheMiss
		}
		return "", time.Time{}, err
	}
	ttl, err := s.Client.TTL(ctx, cacheKey).Result()
	if err != nil {
		return "", time.Time{}, err
	}
	if ttl <= 0 {
		return "", time.Time{}, ErrCacheExpired
	}
	expiresAt := time.Now().Add(ttl)
	return val, expiresAt, nil
}

// Delete removes a cached key.
func (s *RedisRequestCache) Delete(ctx context.Context, key string) error {
	cacheKey := s.buildKey(key)
	return s.Client.Del(ctx, cacheKey).Err()
}

// HealthCheck checks if the Redis connection is healthy.
func (s *RedisRequestCache) HealthCheck(ctx context.Context) error {
	return s.Client.Ping(ctx).Err()
}
