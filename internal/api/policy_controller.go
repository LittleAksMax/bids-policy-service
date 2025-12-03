package api

import (
	"encoding/json"
	"net/http"

	"github.com/LittleAksMax/bids-policy-service/internal/service"
	"github.com/go-chi/chi/v5"
)

type PolicyController struct {
	service *service.PolicyService
}

func NewPolicyController(service *service.PolicyService) *PolicyController {
	return &PolicyController{service: service}
}

func (c *PolicyController) GetPolicyHandler(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	policy, err := c.service.GetPolicy(r.Context(), id)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if err := json.NewEncoder(w).Encode(policy); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (c *PolicyController) CreatePolicyHandler(w http.ResponseWriter, r *http.Request) {
	req := r.Context().Value(requestBodyKey).(*CreatePolicyRequest)
	policy := req.ToPolicy()
	if err := c.service.CreatePolicy(r.Context(), policy); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (c *PolicyController) ListPoliciesHandler(w http.ResponseWriter, r *http.Request) {
	policies, err := c.service.ListPolicies(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if err := json.NewEncoder(w).Encode(policies); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (c *PolicyController) UpdatePolicyHandler(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	req := r.Context().Value(requestBodyKey).(*UpdatePolicyRequest)
	// You may want to get userID from context/session if needed
	policy := req.ToPolicy(id, "")
	if err := c.service.UpdatePolicy(r.Context(), policy); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
