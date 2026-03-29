package service

import (
	"encoding/json"
	"testing"

	"github.com/LittleAksMax/bids-policy-service/internal/convert"
	"github.com/LittleAksMax/bidscript"
)

func TestConvertServiceTreeToScript(t *testing.T) {
	service := NewConvertService()

	tests := []struct {
		name string

		expected string
		root     *convert.Node
	}{
		{
			name:     "set terminal",
			expected: `=10.50`,
			root:     terminalNode(convert.OperatorSet, 10.50, false),
		},
		{
			name:     "add percentage terminal",
			expected: `+3.00%`,
			root:     terminalNode(convert.OperatorAdd, 3.00, true),
		},
		{
			name:     "subtract absolute terminal",
			expected: `-1.25`,
			root:     terminalNode(convert.OperatorSub, 1.25, false),
		},
		{
			name: "integer metric single branch",
			expected: `clicks
[0, 2](=1.25)`,
			root: conditionNode(convert.MetricClicks,
				[]convert.BranchNode{
					branchNode(float64Ptr(0), float64Ptr(2), terminalNode(convert.OperatorSet, 1.25, false)),
				},
				nil,
			),
		},
		{
			name: "integer metric with unbounded lower and default",
			expected: `orders
[_, 5](+1.00)
default (-2.00%)`,
			root: conditionNode(convert.MetricOrders,
				[]convert.BranchNode{
					branchNode(nil, float64Ptr(5), terminalNode(convert.OperatorAdd, 1.00, false)),
				},
				terminalNode(convert.OperatorSub, 2.00, true),
			),
		},
		{
			name: "decimal metric single branch",
			expected: `ctr
[0.00, 0.50](=1.25)`,
			root: conditionNode(convert.MetricCTR,
				[]convert.BranchNode{
					branchNode(float64Ptr(0.00), float64Ptr(0.50), terminalNode(convert.OperatorSet, 1.25, false)),
				},
				nil,
			),
		},
		{
			name: "multiple branches with terminal default",
			expected: `impressions
[0, 100](+1.00%)
[101, _](-0.50%)
default (=1.00)`,
			root: conditionNode(convert.MetricImpressions,
				[]convert.BranchNode{
					branchNode(float64Ptr(0), float64Ptr(100), terminalNode(convert.OperatorAdd, 1.00, true)),
					branchNode(float64Ptr(101), nil, terminalNode(convert.OperatorSub, 0.50, true)),
				},
				terminalNode(convert.OperatorSet, 1.00, false),
			),
		},
		{
			name: "nested condition inside branch",
			expected: `ctr
[0.00, 0.50](
  acos
  [1.00, 2.00](+3.00%)
  default (=4.00)
)
default (-1.00%)`,
			root: conditionNode(convert.MetricCTR,
				[]convert.BranchNode{
					branchNode(float64Ptr(0.00), float64Ptr(0.50), conditionNode(convert.MetricACoS,
						[]convert.BranchNode{
							branchNode(float64Ptr(1.00), float64Ptr(2.00), terminalNode(convert.OperatorAdd, 3.00, true)),
						},
						terminalNode(convert.OperatorSet, 4.00, false),
					)),
				},
				terminalNode(convert.OperatorSub, 1.00, true),
			),
		},
		{
			name: "nested condition inside default",
			expected: `sales
[0.00, 10.00](=1.00)
default (
  spend
  [_, 5.00](+0.50%)
)`,
			root: conditionNode(convert.MetricSales,
				[]convert.BranchNode{
					branchNode(float64Ptr(0.00), float64Ptr(10.00), terminalNode(convert.OperatorSet, 1.00, false)),
				},
				conditionNode(convert.MetricSpend,
					[]convert.BranchNode{
						branchNode(nil, float64Ptr(5.00), terminalNode(convert.OperatorAdd, 0.50, true)),
					},
					nil,
				),
			),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, treeErr := service.TreeToScript(test.root)
			if treeErr == nil && got == test.expected {
				return
			}

			treeJSON, marshalErr := json.MarshalIndent(test.root, "", "  ")
			if marshalErr != nil {
				t.Fatalf("marshal tree for failure output: %v", marshalErr)
			}

			if treeErr != nil {
				t.Fatalf("TreeToScript returned error for input tree:\n%s\n\nerror:\n%v", treeJSON, treeErr)
			}

			t.Fatalf("expected script:\n%s\n\ninput tree:\n%s\n\ngot script:\n%s", test.expected, treeJSON, got)
		})
	}
}

func TestConvertServiceTreeToScriptIgnoresMetricTypeField(t *testing.T) {
	service := NewConvertService()

	root := conditionNodeWithType(convert.MetricClicks, bidscript.MetricValueKindDecimal,
		[]convert.BranchNode{
			branchNode(float64Ptr(0), float64Ptr(2), terminalNode(convert.OperatorSet, 1.25, false)),
		},
		nil,
	)

	got, err := service.TreeToScript(root)
	if err != nil {
		t.Fatalf("TreeToScript returned error: %v", err)
	}

	expected := "clicks\n[0, 2](=1.25)"
	if got != expected {
		t.Fatalf("expected script %q, got %q", expected, got)
	}
}

func TestConvertServiceTreeToScriptReturnsErrorForFractionalIntegerBounds(t *testing.T) {
	service := NewConvertService()

	root := conditionNode(convert.MetricClicks,
		[]convert.BranchNode{
			branchNode(float64Ptr(0.5), float64Ptr(2), terminalNode(convert.OperatorSet, 1.25, false)),
		},
		nil,
	)

	got, err := service.TreeToScript(root)
	if err == nil {
		t.Fatalf("expected TreeToScript to fail for fractional integer bound, got script %q", got)
	}
	if got != "" {
		t.Fatalf("expected empty script on error, got %q", got)
	}
}

func TestConvertServiceScriptToTreeTerminalPrograms(t *testing.T) {
	service := NewConvertService()

	tests := []struct {
		name       string
		source     string
		operator   convert.Operator
		amount     float64
		percentage bool
	}{
		{
			name:       "set absolute terminal",
			source:     `=1.25`,
			operator:   convert.OperatorSet,
			amount:     1.25,
			percentage: false,
		},
		{
			name:       "add percentage terminal",
			source:     `+2.50%`,
			operator:   convert.OperatorAdd,
			amount:     2.50,
			percentage: true,
		},
		{
			name:       "subtract absolute terminal",
			source:     `-0.75`,
			operator:   convert.OperatorSub,
			amount:     0.75,
			percentage: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root := service.ScriptToTree(test.source)
			if root == nil {
				t.Fatalf("ScriptToTree returned nil for %q", test.source)
			}

			if root.Terminal == nil {
				t.Fatalf("expected terminal node for %q", test.source)
			}
			if root.Condition != nil {
				t.Fatalf("did not expect condition node for %q", test.source)
			}

			if root.Terminal.Operator != test.operator {
				t.Fatalf("operator mismatch. expected=%q, got=%q", string(test.operator), string(root.Terminal.Operator))
			}
			if root.Terminal.Amount != test.amount {
				t.Fatalf("amount mismatch. expected=%v, got=%v", test.amount, root.Terminal.Amount)
			}
			if root.Terminal.Percentage != test.percentage {
				t.Fatalf("percentage mismatch. expected=%t, got=%t", test.percentage, root.Terminal.Percentage)
			}
		})
	}
}

func TestConvertServiceScriptToTreeConditionProgramRecursively(t *testing.T) {
	service := NewConvertService()

	source := `ctr
[_, 0.50](
  clicks
  [0, 10](+2.00)
  default (=1.00)
)
default (-0.10%)`

	rootNode := service.ScriptToTree(source)
	if rootNode == nil {
		t.Fatal("ScriptToTree returned nil")
	}

	if rootNode.Condition == nil {
		t.Fatal("expected root condition node")
	}
	if rootNode.Terminal != nil {
		t.Fatal("did not expect root terminal node")
	}

	root := rootNode.Condition
	if root.Metric != convert.MetricCTR {
		t.Fatalf("metric mismatch. expected=%q, got=%q", convert.MetricCTR, root.Metric)
	}
	if root.MetricType != bidscript.MetricValueKindDecimal {
		t.Fatalf("metric type mismatch. expected=%q, got=%q", bidscript.MetricValueKindDecimal, root.MetricType)
	}
	if len(root.Branches) != 1 {
		t.Fatalf("branch count mismatch. expected=%d, got=%d", 1, len(root.Branches))
	}

	rootBranch := root.Branches[0]
	if rootBranch.Lower != nil {
		t.Fatalf("expected unbounded lower interval, got=%v", *rootBranch.Lower)
	}
	if rootBranch.Upper == nil || *rootBranch.Upper != 0.50 {
		t.Fatalf("upper interval mismatch. expected=%v, got=%v", 0.50, derefFloat(rootBranch.Upper))
	}
	if rootBranch.Node.Condition == nil {
		t.Fatal("expected nested condition node")
	}
	if rootBranch.Node.Terminal != nil {
		t.Fatal("did not expect nested root terminal node")
	}

	nested := rootBranch.Node.Condition
	if nested.Metric != convert.MetricClicks {
		t.Fatalf("nested metric mismatch. expected=%q, got=%q", convert.MetricClicks, nested.Metric)
	}
	if nested.MetricType != bidscript.MetricValueKindInteger {
		t.Fatalf("nested metric type mismatch. expected=%q, got=%q", bidscript.MetricValueKindInteger, nested.MetricType)
	}
	if len(nested.Branches) != 1 {
		t.Fatalf("nested branch count mismatch. expected=%d, got=%d", 1, len(nested.Branches))
	}

	nestedBranch := nested.Branches[0]
	if nestedBranch.Lower == nil || *nestedBranch.Lower != 0 {
		t.Fatalf("nested lower interval mismatch. expected=%v, got=%v", 0.0, derefFloat(nestedBranch.Lower))
	}
	if nestedBranch.Upper == nil || *nestedBranch.Upper != 10 {
		t.Fatalf("nested upper interval mismatch. expected=%v, got=%v", 10.0, derefFloat(nestedBranch.Upper))
	}
	if nestedBranch.Node.Terminal == nil {
		t.Fatal("expected nested terminal node")
	}
	if nestedBranch.Node.Condition != nil {
		t.Fatal("did not expect nested terminal to also contain a condition")
	}
	if nestedBranch.Node.Terminal.Operator != convert.OperatorAdd {
		t.Fatalf("nested operator mismatch. expected=%q, got=%q", string(convert.OperatorAdd), string(nestedBranch.Node.Terminal.Operator))
	}
	if nestedBranch.Node.Terminal.Amount != 2 {
		t.Fatalf("nested amount mismatch. expected=%v, got=%v", 2.0, nestedBranch.Node.Terminal.Amount)
	}
	if nestedBranch.Node.Terminal.Percentage {
		t.Fatal("did not expect nested percentage terminal")
	}

	if nested.Default == nil || nested.Default.Terminal == nil {
		t.Fatal("expected nested default terminal node")
	}
	if nested.Default.Condition != nil {
		t.Fatal("did not expect nested default condition node")
	}
	if nested.Default.Terminal.Operator != convert.OperatorSet {
		t.Fatalf("nested default operator mismatch. expected=%q, got=%q", string(convert.OperatorSet), string(nested.Default.Terminal.Operator))
	}
	if nested.Default.Terminal.Amount != 1 {
		t.Fatalf("nested default amount mismatch. expected=%v, got=%v", 1.0, nested.Default.Terminal.Amount)
	}
	if nested.Default.Terminal.Percentage {
		t.Fatal("did not expect nested default percentage terminal")
	}

	if root.Default == nil || root.Default.Terminal == nil {
		t.Fatal("expected root default terminal node")
	}
	if root.Default.Condition != nil {
		t.Fatal("did not expect root default condition node")
	}
	if root.Default.Terminal.Operator != convert.OperatorSub {
		t.Fatalf("root default operator mismatch. expected=%q, got=%q", string(convert.OperatorSub), string(root.Default.Terminal.Operator))
	}
	if root.Default.Terminal.Amount != 0.10 {
		t.Fatalf("root default amount mismatch. expected=%v, got=%v", 0.10, root.Default.Terminal.Amount)
	}
	if !root.Default.Terminal.Percentage {
		t.Fatal("expected root default percentage terminal")
	}
}

func TestConvertServiceScriptToTreeReturnsNilForInvalidScript(t *testing.T) {
	service := NewConvertService()

	root := service.ScriptToTree("ctr\n[0.00, 0.50](+1.00")
	if root != nil {
		treeJSON, err := json.MarshalIndent(root, "", "  ")
		if err != nil {
			t.Fatalf("marshal tree for failure output: %v", err)
		}

		t.Fatalf("expected nil tree for invalid script, got:\n%s", treeJSON)
	}
}

/* The functions below are factory functions for the different types of Node for ease of use */

func terminalNode(operator convert.Operator, amount float64, percentage bool) *convert.Node {
	return &convert.Node{
		Terminal: &convert.TerminalNode{
			Operator:   operator,
			Amount:     amount,
			Percentage: percentage,
		},
	}
}

func conditionNode(metric convert.Metric, branches []convert.BranchNode, defaultNode *convert.Node) *convert.Node {
	return conditionNodeWithType(metric, mustMetricType(metric), branches, defaultNode)
}

func conditionNodeWithType(metric convert.Metric, metricType convert.MetricType, branches []convert.BranchNode, defaultNode *convert.Node) *convert.Node {
	return &convert.Node{
		Condition: &convert.ConditionNode{
			Metric:     metric,
			MetricType: metricType,
			Branches:   branches,
			Default:    defaultNode,
		},
	}
}

func branchNode(lower, upper *float64, node *convert.Node) convert.BranchNode {
	return convert.BranchNode{
		Lower: lower,
		Upper: upper,
		Node:  *node,
	}
}

/* The functions below are simple utilities */

// float64Ptr is used to get a pointer from a constant float64 (essentially by copying)
func float64Ptr(value float64) *float64 {
	return &value
}

// derefFloat dereferences non-null floats, but leaves nil floats as is
func derefFloat(value *float64) interface{} {
	if value == nil {
		return nil
	}
	return *value
}

func mustMetricType(metric convert.Metric) convert.MetricType {
	kind, ok := metric.ValueKind()
	if !ok {
		panic("metric kind lookup failed in test helper")
	}
	return kind
}
