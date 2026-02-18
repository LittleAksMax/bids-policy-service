package validation

import (
	"net/mail"
	"reflect"
	"strings"

	"github.com/LittleAksMax/bids-policy-service/internal/repository"
	"github.com/google/uuid"
)

// fieldName returns the JSON tag name if present; otherwise the struct field name.
func fieldName(field reflect.StructField) string {
	jsonTag := field.Tag.Get("json")
	name := strings.Split(jsonTag, ",")[0]
	if name == "" {
		name = field.Name
	}
	return name
}

// validateByTag scans struct fields and calls check(field, value) for each field whose validate tag contains tagContains.
// If check returns true, the field is appended to invalid list.
func validateByTag(v interface{}, tagContains string, check func(field reflect.StructField, value reflect.Value) bool) []string {
	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	typ := val.Type()
	var invalid []string
	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		fv := val.Field(i)
		validateTag := field.Tag.Get("validate")
		if !strings.Contains(validateTag, tagContains) {
			continue
		}
		if check(field, fv) {
			invalid = append(invalid, fieldName(field))
		}
	}
	return invalid
}

// ValidateRequiredFields checks if fields marked with validate:"required" are non-empty.
func ValidateRequiredFields(v interface{}) error {
	invalid := validateByTag(v, "required", func(field reflect.StructField, fv reflect.Value) bool {
		return fv.Kind() == reflect.String && fv.String() == ""
	})
	if len(invalid) > 0 {
		return &RequiredValidationError{validationError{Fields: invalid}}
	}
	return nil
}

// ValidateEmails checks if fields marked with validate:"email" have valid email format.
func ValidateEmails(v interface{}) error {
	invalid := validateByTag(v, "email", func(field reflect.StructField, fv reflect.Value) bool {
		if fv.Kind() != reflect.String {
			return false
		}
		_, err := mail.ParseAddress(fv.String())
		return err != nil
	})
	if len(invalid) > 0 {
		return &EmailValidationError{validationError{Fields: invalid}}
	}
	return nil
}

// ValidateUUIDs checks if fields marked with validate:"uuid" have valid UUID format.
func ValidateUUIDs(v interface{}) error {
	invalid := validateByTag(v, "uuid", func(field reflect.StructField, fv reflect.Value) bool {
		if fv.Kind() != reflect.String {
			return false
		}
		_, err := uuid.Parse(fv.String())
		return err != nil
	})
	if len(invalid) > 0 {
		return &UUIDValidationError{validationError{Fields: invalid}}
	}
	return nil
}

// ValidatePasswords checks if fields marked with validate:"password" meet minimum strength requirements.
func ValidatePasswords(v interface{}) error {
	invalid := validateByTag(v, "password", func(field reflect.StructField, fv reflect.Value) bool {
		if fv.Kind() != reflect.String {
			return false
		}
		return len(fv.String()) < 8
	})
	if len(invalid) > 0 {
		return &PasswordValidationError{validationError{Fields: invalid}}
	}
	return nil
}

// ValidateMarketplace checks fields with validate:"marketplace" using repository.IsValidMarketplace.
func ValidateMarketplace(v interface{}) error {
	invalid := validateByTag(v, "marketplace", func(field reflect.StructField, fv reflect.Value) bool {
		if fv.Kind() != reflect.String {
			return false
		}
		return !repository.IsValidMarketplace(fv.String())
	})
	if len(invalid) > 0 {
		return &MarketplaceValidationError{validationError{Fields: invalid}}
	}
	return nil
}

// ValidateType checks fields with validate:"type" via repository.IsValidRuleType.
func ValidateType(v interface{}) error {
	invalid := validateByTag(v, "type", func(field reflect.StructField, fv reflect.Value) bool {
		if fv.Kind() != reflect.String {
			return false
		}
		return !repository.IsValidRuleType(fv.String())
	})
	if len(invalid) > 0 {
		return &RuleTypeValidationError{validationError{Fields: invalid}}
	}
	return nil
}
