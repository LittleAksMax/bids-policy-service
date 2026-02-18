package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/LittleAksMax/bids-policy-service/internal/api"
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
		log.Fatalf("Unknown mode: '%s'", mode)
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

	r := api.NewRouter(cfg, dbCfg, cacheCfg)
	addr := fmt.Sprintf(":%d", cfg.Port)

	log.Printf("starting server on %s (mode=%s)", addr, mode)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Printf("server stopped: %v", err)
		os.Exit(1)
	}
}
