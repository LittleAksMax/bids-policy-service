package validation

import (
	"testing"

	"github.com/LittleAksMax/bids-policy-service/internal/convert"
	"github.com/LittleAksMax/bidscript"
)

func TestValidateTreeAcceptsValidMetricType(t *testing.T) {
	root := convert.Node{
		Condition: &convert.ConditionNode{
			Metric:     convert.MetricClicks,
			MetricType: bidscript.MetricValueKindInteger,
			Branches: []convert.BranchNode{
				{
					Lower: float64Ptr(0),
					Upper: float64Ptr(10),
					Node: convert.Node{
						Terminal: &convert.TerminalNode{
							Operator:   convert.OperatorSet,
							Amount:     1.25,
							Percentage: false,
						},
					},
				},
			},
		},
	}

	if err := validateTree(root); err != nil {
		t.Fatalf("validateTree returned error: %v", err)
	}
}

func TestValidateTreeRejectsInvalidMetricType(t *testing.T) {
	tests := []struct {
		name       string
		metricType convert.MetricType
	}{
		{
			name:       "missing type",
			metricType: "",
		},
		{
			name:       "unsupported type",
			metricType: "whole-number",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root := convert.Node{
				Condition: &convert.ConditionNode{
					Metric:     convert.MetricClicks,
					MetricType: test.metricType,
					Branches: []convert.BranchNode{
						{
							Lower: float64Ptr(0),
							Upper: float64Ptr(10),
							Node: convert.Node{
								Terminal: &convert.TerminalNode{
									Operator:   convert.OperatorSet,
									Amount:     1.25,
									Percentage: false,
								},
							},
						},
					},
				},
			}

			if err := validateTree(root); err == nil {
				t.Fatal("expected validateTree to reject invalid metric type")
			}
		})
	}
}

func float64Ptr(value float64) *float64 {
	return &value
}
