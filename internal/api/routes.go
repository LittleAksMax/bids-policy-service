package api

import (
	"net/http"

	"github.com/LittleAksMax/bids-policy-service/internal/config"
	"github.com/LittleAksMax/bids-policy-service/internal/health"
	"github.com/LittleAksMax/bids-policy-service/internal/validation"
	"github.com/LittleAksMax/bids-util/requests"
	utilsvalidation "github.com/LittleAksMax/bids-util/validation"
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

		requests.WriteJSON(w, statusCode, requests.APIResponse{
			Success: true,
			Data:    statuses,
		})
	}
}

const uuidSubjectKey = "uuidSubject"

// RegisterRoutes registers all endpoint handlers using the controller methods.
func RegisterRoutes(r chi.Router, pc *PolicyController, healthCheckers map[string]health.HealthChecker, authCfg *config.AuthConfig) {
	// Health
	r.Get("/health", Health(healthCheckers))

	// Register policy routes with AuthMiddleware
	r.Route("/policies", func(r chi.Router) {
		r.Use(
			requests.ValidateAccessToken(
				authCfg.SharedSecret,
				authCfg.AccessTokenSecret,
				authCfg.MaxSkew,
				authCfg.ClaimsHeader,
				authCfg.TimestampHeader,
				authCfg.SignatureHeader,
			),
			requests.EnsureValidSubject(
				authCfg.ClaimsHeader,
				uuidSubjectKey,
			),
		)

		validationFuncs := []func(T any) error{
			utilsvalidation.ValidateRequiredFields,
			validation.ValidateMarketplace,
			validation.ValidateType,
			utilsvalidation.ValidateEmails,
			utilsvalidation.ValidateUUIDs,
			utilsvalidation.ValidatePasswords,
			validation.ValidateRules,
		}
		r.Get("/", pc.ListPoliciesHandler)
		r.With(requests.ValidateRequest[CreatePolicyRequest](validationFuncs)).Post("/", pc.CreatePolicyHandler)
		r.Get("/{id}", pc.GetPolicyHandler)
		r.With(requests.ValidateRequest[UpdatePolicyRequest](validationFuncs)).Put("/{id}", pc.UpdatePolicyHandler)
		r.Delete("/{id}", pc.DeletePolicyHandler)
	})
}
