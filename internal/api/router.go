package api

import (
	"net/http"

	"github.com/LittleAksMax/bids-policy-service/internal/cache"
	"github.com/LittleAksMax/bids-policy-service/internal/config"
	"github.com/LittleAksMax/bids-policy-service/internal/db"
	"github.com/LittleAksMax/bids-policy-service/internal/health"
	"github.com/LittleAksMax/bids-policy-service/internal/repository"
	"github.com/LittleAksMax/bids-policy-service/internal/service"
	"github.com/go-chi/chi/v5"
)

// NewRouter constructs the main API router by wiring middleware and routes defined elsewhere.
func NewRouter(cfg *config.Config, dbCfg *db.Config, cacheCfg cache.RequestCache) http.Handler {
	r := chi.NewRouter()

	RegisterMiddleware(r)

	// Initialise layers for policies
	policyRepo := repository.NewMongoPolicyRepository(dbCfg.Database)
	policyService := service.NewPolicyService(policyRepo, cacheCfg)
	policyController := NewPolicyController(policyService)

	// Create health checkers map
	healthCheckers := map[string]health.HealthChecker{
		"policy_db": dbCfg,
		"cache":     cacheCfg,
	}

	RegisterRoutes(r, policyController, healthCheckers)

	return r
}
