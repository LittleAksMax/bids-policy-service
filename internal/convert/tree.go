package convert

import (
	"errors"
	"fmt"
	"math"
)

func GetTreeErrors(programRoot *Node) []error {
	if programRoot == nil {
		return []error{errors.New("program node cannot be nil")}
	}

	if programRoot.Terminal == nil && programRoot.Condition == nil {
		return []error{errors.New("program node must define 'terminal' or 'condition' configuration")}
	}

	if programRoot.Terminal != nil && programRoot.Condition != nil {
		return []error{errors.New("program node must not define both 'terminal' and 'condition' types")}
	}

	if programRoot.Terminal != nil {
		return getTerminalNodeErrors(programRoot.Terminal)
	} else {
		// programRoot.Condition != nil
		return getConditionNodeErrors(programRoot.Condition)
	}
}

func getTerminalNodeErrors(terminal *TerminalNode) []error {
	if terminal == nil {
		return []error{errors.New("node cannot be nil")}
	}

	errs := make([]error, 0)

	if terminal.Operator != OperatorAdd && terminal.Operator != OperatorSub && terminal.Operator != OperatorSet {
		errs = append(errs, fmt.Errorf("unknown operator '%s'", terminal.Operator))
	}

	if terminal.Amount < 0.0 {
		errs = append(errs, fmt.Errorf("bid change amount must be non-negative"))
	}

	if terminal.Operator == OperatorSet && terminal.Percentage {
		errs = append(errs, fmt.Errorf("'=' operation cannot have a percentage"))
	}

	return errs
}

func getConditionNodeErrors(condition *ConditionNode) []error {
	if condition == nil {
		return []error{errors.New("node cannot be nil")}
	}

	errs := make([]error, 0)

	if !condition.Metric.IsValid() {
		errs = append(errs, fmt.Errorf("invalid metric '%s'", condition.Metric))
	}

	// This != nil guard is because I am unsure of how empty arrays are unmarshalled
	if condition.Branches != nil {
		for _, branch := range condition.Branches {
			errs = append(errs, getBranchNodeErrors(branch, condition.Metric)...)
		}
	}

	if condition.Default != nil {
		errs = append(errs, GetTreeErrors(condition.Default)...)
	}

	return errs
}

func getBranchNodeErrors(branch BranchNode, metric Metric) []error {
	errs := make([]error, 0)
	if branch.Lower != nil && branch.Upper != nil {
		if *branch.Lower > *branch.Upper {
			errs = append(errs, fmt.Errorf("left (%f) must be no greater than right (%f) of interval", *branch.Lower, *branch.Upper))
		}
	}

	if metric.UsesInteger() {
		if branch.Lower != nil && !isWholeNumber(*branch.Lower) {
			errs = append(errs, fmt.Errorf("metric '%s' requires lower bound to be a whole number, got %f", metric, *branch.Lower))
		}
		if branch.Upper != nil && !isWholeNumber(*branch.Upper) {
			errs = append(errs, fmt.Errorf("metric '%s' requires upper bound to be a whole number, got %f", metric, *branch.Upper))
		}
	}

	return append(errs, GetTreeErrors(&branch.Node)...)
}

func isWholeNumber(value float64) bool {
	return math.Trunc(value) == value
}
