package api

import (
	"net/http"

	"github.com/LittleAksMax/bids-policy-service/internal/service"
	"github.com/LittleAksMax/bids-util/requests"
	"github.com/go-chi/chi/v5"
)

type PolicyController struct {
	service          service.PolicyServiceInterface
	claimsContextKey string
}

func NewPolicyController(service service.PolicyServiceInterface, claimsContextKey string) *PolicyController {
	return &PolicyController{
		service:          service,
		claimsContextKey: claimsContextKey,
	}
}

// GetPolicyHandler retrieves a single policy by ID (REST GET /policies/{id})
func (pc *PolicyController) GetPolicyHandler(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	policy, err := pc.service.GetPolicy(r.Context(), id)
	if err != nil {
		requests.WriteJSON(w, http.StatusInternalServerError, requests.APIResponse{
			Success: false,
			Error:   "failed to retrieve policy",
		})
		return
	}

	if policy == nil {
		requests.WriteJSON(w, http.StatusNotFound, requests.APIResponse{
			Success: false,
			Error:   "policy not found",
		})
		return
	}

	requests.WriteJSON(w, http.StatusOK, requests.APIResponse{
		Success: true,
		Data:    policy,
	})
}

// ListPoliciesHandler lists all policies for the user's marketplace (REST GET /policies)
func (pc *PolicyController) ListPoliciesHandler(w http.ResponseWriter, r *http.Request) {
	// Get marketplace from query parameter
	marketplace := r.URL.Query().Get("marketplace")

	var policies interface{}
	var err error

	if marketplace != "" {
		// Filter by marketplace
		policies, err = pc.service.ListPoliciesByMarketplace(r.Context(), marketplace)
	} else {
		// List all policies
		policies, err = pc.service.ListPolicies(r.Context())
	}

	if err != nil {
		requests.WriteJSON(w, http.StatusInternalServerError, requests.APIResponse{
			Success: false,
			Error:   "failed to retrieve policies",
		})
		return
	}

	requests.WriteJSON(w, http.StatusOK, requests.APIResponse{
		Success: true,
		Data:    policies,
	})
}

// CreatePolicyHandler creates a new policy (REST POST /policies)
func (pc *PolicyController) CreatePolicyHandler(w http.ResponseWriter, r *http.Request) {
	// Extract claims from context (guaranteed by middleware)
	claims := r.Context().Value(pc.claimsContextKey).(*requests.Claims)

	// Get the validated request from context
	createReq := GetRequestBody[CreatePolicyRequest](r)
	if createReq == nil {
		requests.WriteJSON(w, http.StatusBadRequest, requests.APIResponse{
			Success: false,
			Error:   "invalid request body",
		})
		return
	}

	// Convert request DTO to Policy model with user ID from claims
	policy, err := createReq.ToPolicy(claims.Subject)
	if err != nil {
		requests.WriteJSON(w, http.StatusBadRequest, requests.APIResponse{
			Success: false,
			Error:   "invalid policy data: " + err.Error(),
		})
		return
	}

	// Create policy in service
	err = pc.service.CreatePolicy(r.Context(), policy)
	if err != nil {
		requests.WriteJSON(w, http.StatusInternalServerError, requests.APIResponse{
			Success: false,
			Error:   "failed to create policy",
		})
		return
	}

	requests.WriteJSON(w, http.StatusCreated, requests.APIResponse{
		Success: true,
		Data:    policy,
	})
}

// UpdatePolicyHandler updates an existing policy (REST PUT /policies/{id})
func (pc *PolicyController) UpdatePolicyHandler(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	// Extract claims from context (guaranteed by middleware)
	claims := r.Context().Value(pc.claimsContextKey).(*requests.Claims)

	// Get the validated request from context
	updateReq := GetRequestBody[UpdatePolicyRequest](r)
	if updateReq == nil {
		requests.WriteJSON(w, http.StatusBadRequest, requests.APIResponse{
			Success: false,
			Error:   "invalid request body",
		})
		return
	}

	// Convert request DTO to Policy model
	policy, err := updateReq.ToPolicy(id, claims.Subject)
	if err != nil {
		requests.WriteJSON(w, http.StatusBadRequest, requests.APIResponse{
			Success: false,
			Error:   "invalid policy data: " + err.Error(),
		})
		return
	}

	// Fetch existing policy to preserve immutable fields
	existing, err := pc.service.GetPolicy(r.Context(), id)
	if err != nil {
		requests.WriteJSON(w, http.StatusInternalServerError, requests.APIResponse{
			Success: false,
			Error:   "failed to retrieve existing policy",
		})
		return
	}

	if existing == nil {
		requests.WriteJSON(w, http.StatusNotFound, requests.APIResponse{
			Success: false,
			Error:   "policy not found",
		})
		return
	}

	// Ensure immutable fields are preserved
	policy.UserID = existing.UserID
	policy.Marketplace = existing.Marketplace
	policy.Type = existing.Type

	// Update policy in service
	err = pc.service.UpdatePolicy(r.Context(), policy)
	if err != nil {
		requests.WriteJSON(w, http.StatusInternalServerError, requests.APIResponse{
			Success: false,
			Error:   "failed to update policy",
		})
		return
	}

	requests.WriteJSON(w, http.StatusOK, requests.APIResponse{
		Success: true,
		Data:    policy,
	})
}

// DeletePolicyHandler deletes a policy (REST DELETE /policies/{id})
func (pc *PolicyController) DeletePolicyHandler(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	// Verify policy exists
	existing, err := pc.service.GetPolicy(r.Context(), id)
	if err != nil {
		requests.WriteJSON(w, http.StatusInternalServerError, requests.APIResponse{
			Success: false,
			Error:   "failed to retrieve policy",
		})
		return
	}

	if existing == nil {
		requests.WriteJSON(w, http.StatusNotFound, requests.APIResponse{
			Success: false,
			Error:   "policy not found",
		})
		return
	}

	// Delete policy
	err = pc.service.DeletePolicy(r.Context(), id)
	if err != nil {
		requests.WriteJSON(w, http.StatusInternalServerError, requests.APIResponse{
			Success: false,
			Error:   "failed to delete policy",
		})
		return
	}

	requests.WriteJSON(w, http.StatusOK, requests.APIResponse{
		Success: true,
		Data:    map[string]string{"id": id},
	})
}
