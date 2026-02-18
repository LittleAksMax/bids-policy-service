package api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/LittleAksMax/bids-policy-service/internal/validation"
	"github.com/LittleAksMax/bids-util/requests"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// contextKey type for context keys to avoid collisions.
type contextKey string

const requestBodyKey contextKey = "requestBody"

// ValidateRequest is a generic middleware that validates request body fields.
// It decodes the JSON body into the provided type T and validates all marked fields.
func ValidateRequest[T any]() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var reqValue T
			if err := json.NewDecoder(r.Body).Decode(&reqValue); err != nil {
				requests.WriteJSON(w, http.StatusBadRequest, requests.APIResponse{Success: false, Error: "invalid request body"})
				return
			}

			validationFunctions := []func(T any) error{
				validation.ValidateRequiredFields,
				validation.ValidateMarketplace,
				validation.ValidateType,
				validation.ValidateEmails,
				validation.ValidateUUIDs,
				validation.ValidatePasswords,
				validation.ValidateRules,
			}

			for _, validate := range validationFunctions {
				if err := validate(&reqValue); err != nil {
					requests.WriteJSON(w, http.StatusBadRequest, requests.APIResponse{Success: false, Error: err.Error()})
					return
				}
			}

			ctx := context.WithValue(r.Context(), requestBodyKey, &reqValue)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetRequestBody retrieves the validated request body from the context.
func GetRequestBody[T any](r *http.Request) *T {
	if body := r.Context().Value(requestBodyKey); body != nil {
		if typedBody, ok := body.(*T); ok {
			return typedBody
		}
	}
	return nil
}

func RegisterMiddleware(r chi.Router) {
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
}
