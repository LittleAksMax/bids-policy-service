package api

import (
	"net/http"
)

// RequireAPIKey is a middleware that validates the X-API-Key header.
func RequireAPIKey(apiKey string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			providedKey := r.Header.Get("X-API-Key")
			if providedKey == "" || providedKey != apiKey {
				writeJSON(w, http.StatusUnauthorized, apiResponse{Success: false, Error: "invalid or missing API key"})
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// All validation-related code and shared types/functions have been moved to middleware.go.
// This file now only contains the API package's routing and middleware registration logic.
