package api

import (
	"encoding/json"
	"errors"

	"github.com/LittleAksMax/bids-policy-service/internal/repository"
)

// RuleNodeRequest is the interface for rule tree requests
// It is used to unmarshal either a ConditionRequest or TerminalRequest
// and convert to repository.RuleNode

type RuleNodeRequest interface {
	ToRuleNode() (repository.RuleNode, error)
}

type ConditionRequest struct {
	Type     string           `json:"type" validate:"required"`
	Variable string           `json:"variable"`
	Min      float64          `json:"min"`
	Max      float64          `json:"max"`
	If       *json.RawMessage `json:"if"`
	Else     *json.RawMessage `json:"else"`
}

type TerminalRequest struct {
	Type string        `json:"type" validate:"required"`
	Op   RuleOpRequest `json:"op"`
}

type RuleOpRequest struct {
	Type   string            `json:"type"`
	Amount RuleAmountRequest `json:"amount"`
}

type RuleAmountRequest struct {
	Neg    bool    `json:"neg"`
	Amount float64 `json:"amount"`
	Perc   bool    `json:"perc"`
}

// UnmarshalJSON for RuleNodeRequest (handles polymorphic deserialization)
func UnmarshalRuleNodeRequest(data []byte) (RuleNodeRequest, error) {
	var typeField struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(data, &typeField); err != nil {
		return nil, err
	}
	if typeField.Type == "condition" {
		var cond ConditionRequest
		if err := json.Unmarshal(data, &cond); err != nil {
			return nil, err
		}
		return &cond, nil
	} else if typeField.Type == "terminal" {
		var term TerminalRequest
		if err := json.Unmarshal(data, &term); err != nil {
			return nil, err
		}
		return &term, nil
	}
	return nil, errors.New("invalid rule node type")
}

func (c *ConditionRequest) ToRuleNode() (repository.RuleNode, error) {
	var ifNode, elseNode repository.RuleNode
	if c.If != nil {
		ifReq, err := UnmarshalRuleNodeRequest(*c.If)
		if err != nil {
			return nil, err
		}
		ifNode, err = ifReq.ToRuleNode()
		if err != nil {
			return nil, err
		}
	}
	if c.Else != nil {
		elseReq, err := UnmarshalRuleNodeRequest(*c.Else)
		if err != nil {
			return nil, err
		}
		elseNode, err = elseReq.ToRuleNode()
		if err != nil {
			return nil, err
		}
	}
	return &repository.Condition{
		Type:     c.Type,
		Variable: c.Variable,
		Min:      c.Min,
		Max:      c.Max,
		If:       ifNode,
		Else:     elseNode,
	}, nil
}

func (t *TerminalRequest) ToRuleNode() (repository.RuleNode, error) {
	return &repository.Terminal{
		Type: t.Type,
		Op: repository.RuleOp{
			Type: t.Op.Type,
			Amount: repository.RuleAmount{
				Neg:    t.Op.Amount.Neg,
				Amount: t.Op.Amount.Amount,
				Perc:   t.Op.Amount.Perc,
			},
		},
	}, nil
}

// CreatePolicyRequest is the request DTO for creating a policy
// Now uses Rule json.RawMessage for polymorphic rule tree
// UserID is not included in the JSON body; Marketplace is added
type CreatePolicyRequest struct {
	Marketplace string          `json:"marketplace" validate:"required"`
	Name        string          `json:"name" validate:"required"`
	Rule        json.RawMessage `json:"rule" validate:"required"`
}

// ToPolicy now takes userID as a parameter
func (r *CreatePolicyRequest) ToPolicy(userID string) (*repository.Policy, error) {
	ruleReq, err := UnmarshalRuleNodeRequest(r.Rule)
	if err != nil {
		return nil, err
	}
	ruleNode, err := ruleReq.ToRuleNode()
	if err != nil {
		return nil, err
	}
	return &repository.Policy{
		UserID:      userID,
		Marketplace: r.Marketplace,
		Name:        r.Name,
		Rules:       ruleNode,
	}, nil
}

// UpdatePolicyRequest is the request DTO for updating a policy
// Only Name and Rule can be updated; UserID, Marketplace, and Type are immutable
type UpdatePolicyRequest struct {
	Name string          `json:"name" validate:"required"`
	Rule json.RawMessage `json:"rule" validate:"required"`
}

func (r *UpdatePolicyRequest) ToPolicy(id string, userID string) (*repository.Policy, error) {
	ruleReq, err := UnmarshalRuleNodeRequest(r.Rule)
	if err != nil {
		return nil, err
	}
	ruleNode, err := ruleReq.ToRuleNode()
	if err != nil {
		return nil, err
	}
	return &repository.Policy{
		ID:     id,
		UserID: userID,
		Name:   r.Name,
		Rules:  ruleNode,
	}, nil
}
