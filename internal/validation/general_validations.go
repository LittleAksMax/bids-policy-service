package validation

import (
	"reflect"

	"github.com/LittleAksMax/bids-policy-service/internal/repository"
	utilsvalidation "github.com/LittleAksMax/bids-util/validation"
)

// ValidateMarketplace checks fields with validate:"marketplace" using repository.IsValidMarketplace.
func ValidateMarketplace(v interface{}) error {
	invalid := utilsvalidation.ValidateByTag(v, "marketplace", func(field reflect.StructField, fv reflect.Value) bool {
		if fv.Kind() != reflect.String {
			return false
		}
		return !repository.IsValidMarketplace(fv.String())
	})
	if len(invalid) > 0 {
		return &MarketplaceValidationError{utilsvalidation.ValidationError{Fields: invalid}}
	}
	return nil
}

// ValidateType checks fields with validate:"type" via repository.IsValidRuleType.
func ValidateType(v interface{}) error {
	invalid := utilsvalidation.ValidateByTag(v, "type", func(field reflect.StructField, fv reflect.Value) bool {
		if fv.Kind() != reflect.String {
			return false
		}
		return !repository.IsValidRuleType(fv.String())
	})
	if len(invalid) > 0 {
		return &RuleTypeValidationError{utilsvalidation.ValidationError{Fields: invalid}}
	}
	return nil
}
