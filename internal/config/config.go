package config

import (
	"errors"
	"time"

	"github.com/LittleAksMax/bids-policy-service/internal/cache"
	"github.com/LittleAksMax/bids-policy-service/internal/db"
	"github.com/LittleAksMax/bids-util/env"
)

type AuthConfig struct {
	SharedSecret    string
	MaxSkew         time.Duration
	ClaimsHeader    string
	TimestampHeader string
	SignatureHeader string
}

type Config struct {
	Port           int
	AllowedOrigins []string
	Auth           *AuthConfig
	PolicyDB       *db.MongoConnectionConfig
	PolicyCache    *cache.RedisConnectionConfig
}

func Load() (cfg *Config, err error) {
	defer func() {
		if recover() != nil {
			err = errors.New("failed to load configuration -- maybe an expected environment variable is missing")
		}
	}()
	return &Config{
		Port:           env.ReadPort("PORT"),
		AllowedOrigins: env.GetStrListFromEnv("ALLOWED_ORIGINS"),
		Auth: &AuthConfig{
			SharedSecret:    env.GetStrFromEnv("X_AUTH_SIG_SECRET"),
			MaxSkew:         env.ParseDurationEnv("MAX_SKEW"),
			ClaimsHeader:    env.GetStrFromEnv("CLAIMS_HEADER"),
			TimestampHeader: env.GetStrFromEnv("CLAIMS_HEADER"),
			SignatureHeader: env.GetStrFromEnv("SIGNATURE_HEADER"),
		},
		PolicyDB: &db.MongoConnectionConfig{
			Host:     env.GetStrFromEnv("MONGO_HOST"),
			Port:     env.ReadPort("MONGO_PORT"),
			User:     env.GetStrFromEnv("MONGO_USERNAME"),
			Passwd:   env.GetStrFromEnv("MONGO_PASSWORD"),
			Database: env.GetStrFromEnv("MONGO_DATABASE"),
		},
		PolicyCache: &cache.RedisConnectionConfig{
			RedisHost:     env.GetStrFromEnv("REDIS_HOST"),
			RedisPort:     env.ReadPort("REDIS_PORT"),
			RedisPassword: env.GetStrFromEnv("REDIS_PASSWORD"),
		},
	}, nil
}
