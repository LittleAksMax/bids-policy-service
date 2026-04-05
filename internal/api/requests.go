package api

import "github.com/LittleAksMax/bids-policy-service/internal/convert"

// CreatePolicyRequest is the request DTO for creating a policy
// UserID is not included in the JSON body; Marketplace is added
type CreatePolicyRequest struct {
	Marketplace string `json:"marketplace" validate:"required,marketplace"`
	Name        string `json:"name" validate:"required"`
	Script      string `json:"script" validate:"required,script"`
}

// UpdatePolicyRequest is the request DTO for updating a policy
// Only Name and Script can be updated; UserID, Marketplace are immutable
type UpdatePolicyRequest struct {
	Name   string `json:"name" validate:"required"`
	Script string `json:"script" validate:"required,script"`
}

type ConvertTreeToScriptRequest struct {
	Program convert.Node `json:"program" validate:"required,tree"`
}

type ConvertTreeToScriptResponse struct {
	Script string `json:"script"`
}

type ConvertScriptToTreeRequest struct {
	Script string `json:"script" validate:"required,script"`
}

type ConvertScriptToTreeResponse struct {
	Program convert.Node `json:"program" validate:"required,tree"`
}
