package config

import (
	"time"

	"github.com/LittleAksMax/bids-policy-service/internal/cache"
	"github.com/LittleAksMax/bids-policy-service/internal/db"
	"github.com/LittleAksMax/bids-util/env"
)

type AuthConfig struct {
	AccessTokenSecret string
	SharedSecret      string
	MaxSkew           time.Duration
	ClaimsHeader      string
	TimestampHeader   string
	SignatureHeader   string
	APIKey            string
}

type Config struct {
	Port           int
	AllowedOrigins []string
	Auth           *AuthConfig
	PolicyDB       *db.MongoConnectionConfig
	PolicyCache    *cache.RedisConnectionConfig
}

func Load() (cfg *Config, err error) {
	port := env.ReadPort("PORT")
	allowedOrigins := env.GetStrListFromEnv("ALLOWED_ORIGINS")
	accessTokenSecret := env.GetStrFromEnv("ACCESS_TOKEN_SECRET")
	sharedSecret := env.GetStrFromEnv("X_AUTH_SIG_SECRET")
	maxSkew := env.ParseDurationEnv("MAX_SKEW")
	claimsHeader := env.GetStrFromEnv("CLAIMS_HEADER")
	timestampHeader := env.GetStrFromEnv("TIMESTAMP_HEADER")
	signatureHeader := env.GetStrFromEnv("SIGNATURE_HEADER")
	apiKey := env.GetStrFromEnv("API_KEY")
	mongoHost := env.GetStrFromEnv("MONGO_HOST")
	mongoPort := env.ReadPort("MONGO_PORT")
	mongoUsername := env.GetStrFromEnv("MONGO_USERNAME")
	mongoPassword := env.GetStrFromEnv("MONGO_PASSWORD")
	mongoDatabase := env.GetStrFromEnv("MONGO_DATABASE")
	redisHost := env.GetStrFromEnv("REDIS_HOST")
	redisPort := env.ReadPort("REDIS_PORT")
	redisPassword := env.GetStrFromEnv("REDIS_PASSWORD")

	return &Config{
		Port:           port,
		AllowedOrigins: allowedOrigins,
		Auth: &AuthConfig{
			AccessTokenSecret: accessTokenSecret,
			SharedSecret:      sharedSecret,
			MaxSkew:           maxSkew,
			ClaimsHeader:      claimsHeader,
			TimestampHeader:   timestampHeader,
			SignatureHeader:   signatureHeader,
			APIKey:            apiKey,
		},
		PolicyDB: &db.MongoConnectionConfig{
			Host:     mongoHost,
			Port:     mongoPort,
			User:     mongoUsername,
			Passwd:   mongoPassword,
			Database: mongoDatabase,
		},
		PolicyCache: &cache.RedisConnectionConfig{
			RedisHost:     redisHost,
			RedisPort:     redisPort,
			RedisPassword: redisPassword,
		},
	}, nil
}
