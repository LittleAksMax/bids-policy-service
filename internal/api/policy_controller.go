package api

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/LittleAksMax/bids-policy-service/internal/cache"
	"github.com/LittleAksMax/bids-policy-service/internal/repository"
	"github.com/LittleAksMax/bids-policy-service/internal/service"
	"github.com/LittleAksMax/bids-util/requests"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

const policyCacheTTL = 5 * time.Minute

type PolicyController struct {
	service service.PolicyServiceInterface
	cache   cache.RequestCache
}

func NewPolicyController(service service.PolicyServiceInterface, cache cache.RequestCache) *PolicyController {
	return &PolicyController{
		service: service,
		cache:   cache,
	}
}

// GetPolicyHandler retrieves a single policy by ID (REST GET /policies/{id})
func (pc *PolicyController) GetPolicyHandler(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if !isValidPolicyID(id) {
		requests.WriteJSON(w, http.StatusBadRequest, requests.APIResponse{
			Success: false,
			Error:   "invalid policy id",
		})
		return
	}
	userID := r.Context().Value(uuidSubjectKey).(uuid.UUID)

	if policy, ok := pc.getCachedPolicy(r.Context(), userID, id); ok {
		requests.WriteJSON(w, http.StatusOK, requests.APIResponse{
			Success: true,
			Data:    policy,
		})
		return
	}

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

	pc.cachePolicy(r.Context(), userID, id, policy)

	requests.WriteJSON(w, http.StatusOK, requests.APIResponse{
		Success: true,
		Data:    policy,
	})
}

// ListPoliciesHandler lists all policies for the user's marketplace
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
	createReq := requests.GetRequestBody[CreatePolicyRequest](r)
	if createReq == nil {
		requests.WriteJSON(w, http.StatusBadRequest, requests.APIResponse{
			Success: false,
			Error:   "invalid request body",
		})
		return
	}

	// Call service directly with extracted fields
	policy, err := pc.service.CreatePolicy(r.Context(), userID, createReq.Marketplace, createReq.Name, strings.ToLower(createReq.Script))
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
	if !isValidPolicyID(id) {
		requests.WriteJSON(w, http.StatusBadRequest, requests.APIResponse{
			Success: false,
			Error:   "invalid policy id",
		})
		return
	}
	userID := r.Context().Value(uuidSubjectKey).(uuid.UUID)

	// Get the validated request from context
	updateReq := requests.GetRequestBody[UpdatePolicyRequest](r)
	if updateReq == nil {
		requests.WriteJSON(w, http.StatusBadRequest, requests.APIResponse{
			Success: false,
			Error:   "invalid request body",
		})
		return
	}

	// Call service directly with extracted fields
	policy, err := pc.service.UpdatePolicy(r.Context(), userID, id, updateReq.Name, strings.ToLower(updateReq.Script))
	if err != nil {
		requests.WriteJSON(w, http.StatusInternalServerError, requests.APIResponse{
			Success: false,
			Error:   "failed to update policy",
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

	pc.invalidatePolicyCache(r.Context(), userID, id)

	requests.WriteJSON(w, http.StatusOK, requests.APIResponse{
		Success: true,
		Data:    policy,
	})
}

// DeletePolicyHandler deletes a policy (REST DELETE /policies/{id})
func (pc *PolicyController) DeletePolicyHandler(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if !isValidPolicyID(id) {
		requests.WriteJSON(w, http.StatusBadRequest, requests.APIResponse{
			Success: false,
			Error:   "invalid policy id",
		})
		return
	}
	userID := r.Context().Value(uuidSubjectKey).(uuid.UUID)

	// Delete policy
	ok, err := pc.service.DeletePolicy(r.Context(), userID, id)
	if err != nil {
		requests.WriteJSON(w, http.StatusInternalServerError, requests.APIResponse{
			Success: false,
			Error:   "failed to delete policy",
		})
		return
	}
	if !ok {
		requests.WriteJSON(w, http.StatusNotFound, requests.APIResponse{
			Success: false,
			Error:   "policy not found",
		})
		return
	}

	pc.invalidatePolicyCache(r.Context(), userID, id)

	requests.WriteJSON(w, http.StatusOK, requests.APIResponse{
		Success: true,
		Data:    map[string]string{"id": id},
	})
}

func (pc *PolicyController) getCachedPolicy(ctx context.Context, userID uuid.UUID, id string) (*repository.Policy, bool) {
	cached, expiresAt, err := pc.cache.Get(ctx, policyCacheKey(userID, id))
	if err != nil || cached == "" || !expiresAt.After(time.Now()) {
		return nil, false
	}

	var policy repository.Policy
	if err := json.Unmarshal([]byte(cached), &policy); err != nil {
		return nil, false
	}

	return &policy, true
}

func (pc *PolicyController) cachePolicy(ctx context.Context, userID uuid.UUID, id string, policy *repository.Policy) {
	if policy == nil {
		return
	}

	payload, err := json.Marshal(policy)
	if err != nil {
		return
	}

	_ = pc.cache.Set(ctx, policyCacheKey(userID, id), string(payload), policyCacheTTL)
}

func (pc *PolicyController) invalidatePolicyCache(ctx context.Context, userID uuid.UUID, id string) {
	_ = pc.cache.Delete(ctx, policyCacheKey(userID, id))
}

func policyCacheKey(userID uuid.UUID, id string) string {
	return userID.String() + ":policy:" + strings.ToLower(id)
}
