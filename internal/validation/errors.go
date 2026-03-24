package validation

import (
	"fmt"
	"strings"

	"github.com/LittleAksMax/bids-policy-service/internal/repository"
	utilsvalidation "github.com/LittleAksMax/bids-util/validation"
)

type RuleTypeValidationError struct {
	utilsvalidation.ValidationError
}

func (e *RuleTypeValidationError) Error() string {
	return strings.Join(e.Fields, ", ") + " must be one of: " + repository.NestedRuleType
}

type MarketplaceValidationError struct {
	utilsvalidation.ValidationError
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
	repository.MpBR,
	repository.MpAE,
	repository.MpBE,
	repository.MpEG,
	repository.MpIE,
	repository.MpIN,
	repository.MpNL,
	repository.MpPL,
	repository.MpSA,
	repository.MpSE,
	repository.MpTR,
	repository.MpZA,
	repository.MpAU,
	repository.MpJP,
	repository.MpSG,
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
