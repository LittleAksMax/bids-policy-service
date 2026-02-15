package cache

import (
	"context"
	"time"

	"github.com/LittleAksMax/bids-policy-service/internal/health"
)

// RequestCache is a key-value store for caching web requests for policies.
type RequestCache interface {
	health.HealthChecker
	Save(ctx context.Context, key string, value string, expiresAt time.Time) error
	Get(ctx context.Context, key string) (value string, expiresAt time.Time, err error)
	Delete(ctx context.Context, key string) error
}
