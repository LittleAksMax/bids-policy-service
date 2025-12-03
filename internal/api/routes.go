package api

import (
	"net/http"

	"github.com/LittleAksMax/bids-policy-service/internal/health"
	"github.com/go-chi/chi/v5"
)

// Health handler implementation that checks all registered services.
func Health(checkers map[string]health.HealthChecker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		statuses := make(map[string]interface{})
		allHealthy := true

		for name, checker := range checkers {
			if err := checker.HealthCheck(r.Context()); err != nil {
				statuses[name] = map[string]interface{}{
					"status": "unhealthy",
					"error":  err.Error(),
				}
				allHealthy = false
			} else {
				statuses[name] = map[string]interface{}{
					"status": "healthy",
				}
			}
		}

		// Determine HTTP status code based on health
		statusCode := http.StatusOK
		if !allHealthy {
			statusCode = http.StatusServiceUnavailable
		}

		writeJSON(w, statusCode, apiResponse{
			Success: true,
			Data:    statuses,
		})
	}
}

// RegisterRoutes registers all endpoint handlers using the controller methods.
func RegisterRoutes(r chi.Router, pc *PolicyController, healthCheckers map[string]health.HealthChecker) {
	// Health
	r.Get("/health", Health(healthCheckers))

	// Register policy routes
	r.Route("/policies", func(r chi.Router) {
		r.Get("/", pc.ListPoliciesHandler)
		r.With(ValidateRequest[CreatePolicyRequest]()).Post("/", pc.CreatePolicyHandler)
		r.Get("/{id}", pc.GetPolicyHandler)
		r.With(ValidateRequest[UpdatePolicyRequest]()).Put("/{id}", pc.UpdatePolicyHandler)
	})
}
