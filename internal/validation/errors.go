package validation

import (
	"fmt"
	"strings"

	"github.com/LittleAksMax/bids-policy-service/internal/repository"
)

// validationError represents validation errors.
type validationError struct {
	Fields []string
}

type RequiredValidationError struct {
	validationError
}

func (e *RequiredValidationError) Error() string {
	return strings.Join(e.Fields, ", ") + " required"
}

// EmailValidationError represents email validation errors.
type EmailValidationError struct {
	validationError
}

func (e *EmailValidationError) Error() string {
	return strings.Join(e.Fields, ", ") + " must be valid email address(es)"
}

// UUIDValidationError represents UUID validation errors.
type UUIDValidationError struct {
	validationError
}

func (e *UUIDValidationError) Error() string {
	return strings.Join(e.Fields, ", ") + " must be valid UUID(s)"
}

// PasswordValidationError represents password validation errors.
type PasswordValidationError struct {
	validationError
}

func (e *PasswordValidationError) Error() string {
	return strings.Join(e.Fields, ", ") + " must be at least 8 characters"
}

type RuleTypeValidationError struct {
	validationError
}

func (e *RuleTypeValidationError) Error() string {
	return strings.Join(e.Fields, ", ") + " must be one of: " + repository.NestedRuleType
}

type MarketplaceValidationError struct {
	validationError
}

var allowedMarketplacesStr = strings.Join([]string{
	repository.MpUK,
	repository.MpDE,
	repository.MpFR,
	repository.MpIT,
	repository.MpES,
	repository.MpUS,
	repository.MpCA,
	repository.MpMX,
}, ", ")

func (e *MarketplaceValidationError) Error() string {
	return strings.Join(e.Fields, ", ") + " must be one of: " + allowedMarketplacesStr
}

type RuleNodeTypeValidationError struct {
	Type string
}

var allowedRuleNodeTypes = strings.Join([]string{repository.ConditionType, repository.TerminalType}, ", ")

func (e *RuleNodeTypeValidationError) Error() string {
	// Empty string denotes missing type field, rather than incorrect value
	if e.Type == "" {
		return "Missing rule node type"
	}

	return "Invalid rule node type: " + e.Type + ". Must be one of: " + allowedRuleNodeTypes
}

type RuleNodeOpTypeValidationError struct {
	Type string
}

var allowedRuleNodeOpTypes = strings.Join([]string{repository.OpMul, repository.OpAdd}, ", ")

func (e *RuleNodeOpTypeValidationError) Error() string {
	return "Invalid rule node operator type: " + e.Type + ". Must be one of: " + allowedRuleNodeOpTypes
}

type RuleNodeVariableValidationError struct {
	Variable string
}

func (e *RuleNodeVariableValidationError) Error() string {
	return "Invalid rule node variable: " + e.Variable
}

type RuleNodeRangeValidationError struct {
	Min float64
	Max float64
}

func (e *RuleNodeRangeValidationError) Error() string {
	return "Invalid rule node range: min(" + formatFloat(e.Min) + ") > max(" + formatFloat(e.Max) + ")"
}

type RuleNodeBranchMissingValidationError struct {
	Branch string
}

func (e *RuleNodeBranchMissingValidationError) Error() string {
	return "Missing rule node branch: " + e.Branch
}

// formatFloat provides a minimal float formatting for error messages.
func formatFloat(f float64) string {
	// Trim excessive decimals for readability; you can adjust as needed
	return strings.TrimRight(strings.TrimRight(fmt.Sprintf("%0.6f", f), "0"), ".")
}
