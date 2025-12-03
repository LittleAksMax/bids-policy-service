package main

import (
	"context"
	"log"
	"os"

	"github.com/LittleAksMax/bids-policy-service/internal/cache"
	"github.com/LittleAksMax/bids-policy-service/internal/config"
	"github.com/LittleAksMax/bids-policy-service/internal/db"
	"github.com/joho/godotenv"
)

const (
	ModeDevelopment = "development"
	ModeProduction  = "production"
)

func main() {
	// Load development override file BEFORE config parsing if MODE indicates development.
	mode := os.Getenv("MODE")
	if mode != ModeDevelopment && mode != ModeProduction {
	}
	if mode == ModeDevelopment {
		if err := godotenv.Load(".env.Dev"); err != nil {
			log.Fatalf("Failed to load .env.Dev: %v", err)
		}
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	ctx := context.Background()

	dbCfg, err := db.Connect(ctx, cfg.PolicyDB)
	if err != nil {
		log.Fatalf("db connect error: %v", err)
	}
	defer func() {
		if err := dbCfg.Client.Disconnect(ctx); err != nil {
			log.Printf("db close error: %v", err)
		}
	}()

	cacheCfg, err := cache.NewRedisRefreshStore(ctx, cfg.PolicyCache)
	if err != nil {
		log.Fatalf("cache connect error: %v", err)
	}
	defer func() {
		if err := cacheCfg.Client.Close(); err != nil {
			log.Printf("redis close error: %v", err)
		}
	}()
}
