package api

import "github.com/LittleAksMax/policy-service/internal/repository"

// CreatePolicyRequest is the request DTO for creating a policy
type CreatePolicyRequest struct {
	UserID string   `json:"user_id" validate:"required,uuid"`
	Name   string   `json:"name" validate:"required"`
	Rules  []string `json:"rules" validate:"required"`
}

func (r *CreatePolicyRequest) ToPolicy() *repository.Policy {
	return &repository.Policy{
		UserID: r.UserID,
		Name:   r.Name,
		Rules:  r.Rules,
	}
}

// UpdatePolicyRequest is the request DTO for updating a policy
// ID is not in the body, but set from the URL in the controller
// (for validation, you can add it if needed)
type UpdatePolicyRequest struct {
	Name  string   `json:"name" validate:"required"`
	Rules []string `json:"rules" validate:"required"`
}

func (r *UpdatePolicyRequest) ToPolicy(id string, userID string) *repository.Policy {
	return &repository.Policy{
		ID:     id,
		UserID: userID,
		Name:   r.Name,
		Rules:  r.Rules,
	}
}
