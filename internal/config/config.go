package config

import (
	"errors"

	"github.com/LittleAksMax/policy-service/internal/cache"
	"github.com/LittleAksMax/policy-service/internal/db"
)

type Config struct {
	Port        int
	PolicyDB    *db.MongoConnectionConfig
	PolicyCache *cache.RedisConnectionConfig
}

func Load() (cfg *Config, err error) {
	defer func() {
		if recover() != nil {
			err = errors.New("failed to load configuration -- maybe an expected environment variable is missing")
		}
	}()
	return &Config{
		Port: readPort("PORT"),
		PolicyDB: &db.MongoConnectionConfig{
			Host:     getStrFromEnv("MONGO_HOST"),
			Port:     readPort("MONGO_PORT"),
			User:     getStrFromEnv("MONGO_USERNAME"),
			Passwd:   getStrFromEnv("MONGO_PASSWORD"),
			Database: getStrFromEnv("MONGO_DATABASE"),
		},
		PolicyCache: &cache.RedisConnectionConfig{
			RedisHost:     getStrFromEnv("REDIS_HOST"),
			RedisPort:     readPort("REDIS_PORT"),
			RedisPassword: getStrFromEnv("REDIS_PASSWORD"),
		},
	}, nil
}
