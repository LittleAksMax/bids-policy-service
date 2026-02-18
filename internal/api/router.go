package api

import (
	"net/http"

	"github.com/LittleAksMax/bids-policy-service/internal/cache"
	"github.com/LittleAksMax/bids-policy-service/internal/config"
	"github.com/LittleAksMax/bids-policy-service/internal/db"
	"github.com/LittleAksMax/bids-policy-service/internal/health"
	"github.com/LittleAksMax/bids-policy-service/internal/repository"
	"github.com/LittleAksMax/bids-policy-service/internal/service"

	"github.com/LittleAksMax/bids-util/requests"
	"github.com/go-chi/chi/v5"
)

// NewRouter constructs the main API router by wiring middleware and routes defined elsewhere.
func NewRouter(cfg *config.Config, dbCfg *db.Config, cacheCfg cache.RequestCache) http.Handler {
	r := chi.NewRouter()

	RegisterMiddleware(r)

	requests.ApplyCORS(
		r,
		cfg.AllowedOrigins,
		[]string{"GET", "POST", "PUT", "DELETE"},
		[]string{"Accept", "Authorization", "Content-Type", cfg.Auth.ClaimsHeader, cfg.Auth.TimestampHeader, cfg.Auth.SignatureHeader},
		[]string{"Set-Cookie"},
		true,
		300,
	)

	// Initialise layers for policies
	policyRepo := repository.NewMongoPolicyRepository(dbCfg.Database)
	policyService := service.NewPolicyService(policyRepo, cacheCfg)
	policyController := NewPolicyController(policyService, cfg.Auth.ClaimsHeader)

	// Create health checkers map
	healthCheckers := map[string]health.HealthChecker{
		"policy_db": dbCfg,
		"cache":     cacheCfg,
	}

	RegisterRoutes(r, policyController, healthCheckers, cfg.Auth)

	return r
}
