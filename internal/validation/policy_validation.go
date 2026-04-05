package validation

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"

	"github.com/LittleAksMax/bids-policy-service/internal/convert"
	utilsvalidation "github.com/LittleAksMax/bids-util/validation"
)

func ValidateScript(v interface{}) error {
	var details []error
	invalid := utilsvalidation.ValidateByTag(v, "script", func(field reflect.StructField, fv reflect.Value) bool {
		fieldName := utilsvalidation.FieldName(field)
		if fv.Kind() != reflect.String {
			details = append(details, fmt.Errorf("%s must be a valid script string", fieldName))
			return true
		}

		if err := convert.GetScriptErrors(fv.String()); err != nil {
			details = append(details, fmt.Errorf("%s must be a valid script string: %w", fieldName, err))
			return true
		}

		return false
	})
	if len(invalid) > 0 {
		return &ScriptValidationError{
			ValidationError: utilsvalidation.ValidationError{Fields: invalid},
			Details:         details,
		}
	}

	return nil
}

func ValidateTree(v interface{}) error {
	var details []error
	invalid := utilsvalidation.ValidateByTag(v, "tree", func(field reflect.StructField, fv reflect.Value) bool {
		fieldName := utilsvalidation.FieldName(field)
		if err := validateTree(fv.Interface()); err != nil {
			details = append(details, fmt.Errorf("%s must be a valid tree object: %w", fieldName, err))
			return true
		}

		return false
	})
	if len(invalid) > 0 {
		return &TreeValidationError{
			ValidationError: utilsvalidation.ValidationError{Fields: invalid},
			Details:         details,
		}
	}

	return nil
}

func validateTree(tree any) error {
	// Convert into JSON
	payload, err := json.Marshal(tree)
	if err != nil {
		return err
	}

	// Decode into struct to find any structural errors
	var root convert.Node
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.DisallowUnknownFields() // error of non-specified fields are found
	if err := decoder.Decode(&root); err != nil {
		return err
	}

	errs := convert.GetTreeErrors(&root)
	errs = append(errs, getMetricTypeErrors(&root)...)
	if len(errs) == 0 {
		return nil
	}

	return errors.Join(errs...)
}

func getMetricTypeErrors(root *convert.Node) []error {
	if root == nil || root.Condition == nil {
		return nil
	}

	errs := make([]error, 0)
	condition := root.Condition
	expectedMetricType, metricValid := condition.Metric.ValueKind()
	if !convert.IsValidMetricType(condition.MetricType) {
		errs = append(errs, fmt.Errorf("condition metric '%s' must define type as %q or %q, got %q",
			condition.Metric,
			convert.MetricType("integer"),
			convert.MetricType("decimal"),
			condition.MetricType,
		))
	} else if metricValid && condition.MetricType != expectedMetricType {
		errs = append(errs, fmt.Errorf("condition metric '%s' must define type as %q, got %q",
			condition.Metric,
			expectedMetricType,
			condition.MetricType,
		))
	}

	for _, branch := range condition.Branches {
		errs = append(errs, getMetricTypeErrors(&branch.Node)...)
	}

	if condition.Default != nil {
		errs = append(errs, getMetricTypeErrors(condition.Default)...)
	}

	return errs
}
