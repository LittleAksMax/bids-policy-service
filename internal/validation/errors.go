package validation

import (
	"errors"
	"strings"

	"github.com/LittleAksMax/bids-policy-service/internal/repository"
	utilsvalidation "github.com/LittleAksMax/bids-util/validation"
)

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

type policyValidationError struct {
	utilsvalidation.ValidationError
	Details []error
}

type ScriptValidationError policyValidationError

type TreeValidationError policyValidationError

func (e *ScriptValidationError) Error() string {
	if len(e.Details) > 0 {
		return errors.Join(e.Details...).Error()
	}

	return strings.Join(e.Fields, ", ") + " must be a valid script string"
}

func (e *TreeValidationError) Error() string {
	if len(e.Details) > 0 {
		return errors.Join(e.Details...).Error()
	}

	return strings.Join(e.Fields, ", ") + " must be a valid tree object"
}
