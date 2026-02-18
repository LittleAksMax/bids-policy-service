package api

import (
	"net/http"

	"github.com/LittleAksMax/bids-policy-service/internal/repository"
	"github.com/LittleAksMax/bids-policy-service/internal/service"
	"github.com/LittleAksMax/bids-util/requests"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
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
	userID := r.Context().Value(uuidSubjectKey).(uuid.UUID)

	policy, err := pc.service.GetPolicy(r.Context(), userID, id)
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
	userID := r.Context().Value(uuidSubjectKey).(uuid.UUID)
	// Get marketplace from query parameter
	marketplace := r.URL.Query().Get("marketplace")

	var policies []*repository.Policy
	var err error = nil
	if marketplace != "" {
		// Filter by marketplace
		policies, err = pc.service.ListPoliciesByMarketplace(r.Context(), userID, marketplace)
	} else {
		// List all policies
		policies, err = pc.service.ListPolicies(r.Context(), userID)
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
	userID := r.Context().Value(uuidSubjectKey).(uuid.UUID)

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
	policy, err := createReq.ToPolicy(userID)
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
	userID := r.Context().Value(uuidSubjectKey).(uuid.UUID)

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
	policy, err := updateReq.ToPolicy(id, userID)
	if err != nil {
		requests.WriteJSON(w, http.StatusBadRequest, requests.APIResponse{
			Success: false,
			Error:   "invalid policy data: " + err.Error(),
		})
		return
	}

	// Fetch existing policy to preserve immutable fields
	existing, err := pc.service.GetPolicy(r.Context(), userID, id)
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
	err = pc.service.UpdatePolicy(r.Context(), userID, policy)
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
	userID := r.Context().Value(uuidSubjectKey).(uuid.UUID)

	// Verify policy exists
	existing, err := pc.service.GetPolicy(r.Context(), userID, id)
	if err != nil {
		requests.WriteJSON(w, http.StatusInternalServerError, requests.APIResponse{
			Success: false,
			Error:   "failed to delete policy",
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
	err = pc.service.DeletePolicy(r.Context(), userID, id)
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
