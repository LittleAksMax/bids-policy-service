package api

import (
	"encoding/json"

	"github.com/LittleAksMax/bids-policy-service/internal/repository"
	"github.com/LittleAksMax/bids-policy-service/internal/validation"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// RuleNodeRequest is the interface for rule tree requests
// It is used to unmarshal either a ConditionRequest or TerminalRequest
// and convert to repository.RuleNode

// CreatePolicyRequest is the request DTO for creating a policy
// Now uses Rule json.RawMessage for polymorphic rule tree
// UserID is not included in the JSON body; Marketplace is added
type CreatePolicyRequest struct {
	Marketplace string          `json:"marketplace" validate:"required,marketplace"`
	Name        string          `json:"name" validate:"required"`
	Type        string          `json:"type" validate:"required,type"`
	Rules       json.RawMessage `json:"rules" validate:"required,rules"`
}

// ToPolicy now takes userID as a parameter
func (r *CreatePolicyRequest) ToPolicy(userID uuid.UUID) (*repository.Policy, error) {
	ruleNode, err := validation.UnmarshalRuleNodeRequest(r.Rules)
	if err != nil {
		return nil, err
	}
	return &repository.Policy{
		ID:          primitive.NewObjectID(), // new Object ID for new policy
		UserID:      userID.String(),
		Marketplace: r.Marketplace,
		Name:        r.Name,
		Type:        r.Type,
		Rules:       ruleNode,
	}, nil
}

// UpdatePolicyRequest is the request DTO for updating a policy
// Only Name and Rule can be updated; UserID, Marketplace, and Type are immutable
type UpdatePolicyRequest struct {
	Name  string          `json:"name" validate:"required"`
	Rules json.RawMessage `json:"rule" validate:"required,rules"`
}

func (r *UpdatePolicyRequest) ToPolicy(id string, userID uuid.UUID) (*repository.Policy, error) {
	hexID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	ruleNode, err := validation.UnmarshalRuleNodeRequest(r.Rules)
	if err != nil {
		return nil, err
	}
	return &repository.Policy{
		ID:     hexID,
		UserID: userID.String(),
		Name:   r.Name,
		Rules:  ruleNode,
	}, nil
}
