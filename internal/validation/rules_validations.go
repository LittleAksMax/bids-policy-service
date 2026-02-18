package validation

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/LittleAksMax/bids-policy-service/internal/repository"
)

type nodeHeader struct {
	Type string `json:"type"`
}

type conditionDTO struct {
	Type     string          `json:"type"`
	Variable string          `json:"variable"`
	Min      float64         `json:"min"`
	Max      float64         `json:"max"`
	If       json.RawMessage `json:"if"`
	Else     json.RawMessage `json:"else"`
}

func UnmarshalRuleNodeJSON(data []byte) (repository.RuleNode, error) {
	var h nodeHeader
	if err := json.Unmarshal(data, &h); err != nil {
		return nil, err
	}

	switch h.Type {
	case "terminal":
		var t repository.Terminal
		if err := json.Unmarshal(data, &t); err != nil {
			return nil, err
		}

		if err := validateTerminal(&t); err != nil {
			return nil, err
		}

		return &t, nil
	case "condition":
		var dto conditionDTO
		if err := json.Unmarshal(data, &dto); err != nil {
			return nil, err
		}

		ifNode, err := UnmarshalRuleNodeJSON(dto.If)
		if err != nil {
			return nil, fmt.Errorf("if->%w", err)
		}
		if ifNode == nil {
			return nil, fmt.Errorf("if->node is nil")
		}
		elseNode, err := UnmarshalRuleNodeJSON(dto.Else)
		if err != nil {
			return nil, fmt.Errorf("else->%w", err)
		}
		if elseNode == nil {
			return nil, fmt.Errorf("else->node is nil")
		}

		c := repository.Condition{
			Type:     dto.Type,
			Variable: dto.Variable,
			Min:      dto.Min,
			Max:      dto.Max,
			If:       ifNode,
			Else:     elseNode,
		}

		if err := validateCondition(&c); err != nil {
			return nil, err
		}

		return &c, nil

	default:
		return nil, &RuleTypeValidationError{validationError{Fields: []string{h.Type}}}
	}
}

// ValidateRules checks fields with validate:"rules" contain valid JSON for a rule node.
// Special-case: if the field is missing or empty, return a RequiredValidationError for "rules".
// Otherwise, propagate specific errors from UnmarshalRuleNodeJSON (type, shape, recursion, etc.).
func ValidateRules(v interface{}) error {
	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	typ := val.Type()

	found := false
	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		fv := val.Field(i)
		validateTag := field.Tag.Get("validate")
		if !strings.Contains(validateTag, "rules") {
			continue
		}
		found = true
		// Expect json.RawMessage ([]byte) for rules
		if fv.Kind() == reflect.Slice && fv.Type().Elem().Kind() == reflect.Uint8 {
			b := fv.Bytes()
			if len(b) == 0 {
				return &RequiredValidationError{validationError{Fields: []string{"rules"}}}
			}
			// Propagate specific validation errors from unmarshal/recursive validation
			if _, err := UnmarshalRuleNodeJSON(b); err != nil {
				return err
			}
		} else {
			// Wrong type: treat as required violation to keep API surface simple
			return &RequiredValidationError{validationError{Fields: []string{"rules"}}}
		}
	}
	if !found {
		return &RequiredValidationError{validationError{Fields: []string{"rules"}}}
	}
	return nil
}

func validateTerminal(t *repository.Terminal) error {
	if t.Type != "terminal" {
		return &RuleNodeTypeValidationError{Type: t.Type}
	}

	if !repository.IsValidRuleOpType(t.Op.Type) {
		return &RuleNodeOpTypeValidationError{Type: t.Op.Type}
	}

	// NOTE: we don't validate t.Op.Amount, we just assume the user enters reasonable values
	return nil
}

func validateCondition(c *repository.Condition) error {
	if c.Type != "condition" {
		return &RuleNodeTypeValidationError{Type: c.Type}
	}

	if c.Min >= c.Max {
		return errors.New("min must be less than max")
	}

	if !repository.IsValidVariable(c.Variable) {
		return &RuleNodeVariableValidationError{Variable: c.Variable}
	}

	return nil
}
