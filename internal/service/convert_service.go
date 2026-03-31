package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/LittleAksMax/bids-policy-service/internal/convert"
	"github.com/LittleAksMax/bidscript"
	"github.com/LittleAksMax/bidscript/ast"
	"github.com/LittleAksMax/bidscript/lexer"
	"github.com/LittleAksMax/bidscript/parser"
)

type ConvertService struct {
}

type ConvertServiceInterface interface {
	TreeToScript(root *convert.Node) (string, error)
	ScriptToTree(source string) *convert.Node
}

func NewConvertService() *ConvertService {
	return &ConvertService{}
}

func (service *ConvertService) ScriptToTree(source string) *convert.Node {
	l := lexer.NewLexer(source)
	p := parser.NewParser(l)
	programAST := p.ParseProgram()

	// We can't convert if there are errors
	if len(p.Errors()) != 0 {
		return nil
	}

	data, err := ast.MarshalJSON(programAST)
	if err != nil {
		return nil
	}

	root, err := decodeJSONObject(data)
	if err != nil {
		return nil
	}
	if err := expectNodeType(root, "Program"); err != nil {
		return nil
	}

	statement, err := rawField(root, "statement")
	if err != nil {
		return nil
	}

	return nodeFromJSON(statement)
}

type rawJSONObject map[string]json.RawMessage

func nodeFromJSON(data json.RawMessage) *convert.Node {
	fields, err := decodeJSONObject(data)
	if err != nil {
		return nil
	}

	nodeType, err := stringField(fields, "node_type")
	if err != nil {
		return nil
	}

	switch nodeType {
	case "TerminalStatement":
		terminal := terminalNodeFromJSON(fields)
		if terminal == nil {
			return nil
		}
		return &convert.Node{Terminal: terminal}

	case "ConditionStatement":
		condition := conditionNodeFromJSON(fields)
		if condition == nil {
			return nil
		}
		return &convert.Node{Condition: condition}
	default:
		return nil
	}
}

func terminalNodeFromJSON(fields rawJSONObject) *convert.TerminalNode {
	if err := expectNodeType(fields, "TerminalStatement"); err != nil {
		return nil
	}

	operatorLiteral, err := stringField(fields, "operator")
	if err != nil {
		return nil
	}

	value, err := rawField(fields, "value")
	if err != nil {
		return nil
	}
	amount, err := numericLiteralFromJSON(value)
	if err != nil {
		return nil
	}

	percentage, err := boolField(fields, "percentage")
	if err != nil {
		return nil
	}

	operator, err := operatorFromLiteral(operatorLiteral)
	if err != nil {
		return nil
	}

	return &convert.TerminalNode{
		Operator:   convert.Operator(operator),
		Amount:     amount,
		Percentage: percentage,
	}
}

func conditionNodeFromJSON(fields rawJSONObject) *convert.ConditionNode {
	if err := expectNodeType(fields, "ConditionStatement"); err != nil {
		return nil
	}

	metricData, err := rawField(fields, "metric")
	if err != nil {
		return nil
	}
	metric, err := metricFromJSON(metricData)
	if err != nil {
		return nil
	}

	branchItems, err := arrayField(fields, "branches")
	if err != nil {
		return nil
	}
	branches := make([]convert.BranchNode, 0, len(branchItems))
	for _, branchData := range branchItems {
		branch := branchNodeFromJSON(branchData)
		if branch == nil {
			return nil
		}
		branches = append(branches, *branch)
	}

	var defaultNode *convert.Node
	if data, ok := fields["default"]; ok && len(data) > 0 && string(data) != "null" {
		node := nodeFromJSON(data)
		if node == nil {
			return nil
		}
		defaultNode = node
	}

	return &convert.ConditionNode{
		Metric:     metric,
		MetricType: metricType(metric),
		Branches:   branches,
		Default:    defaultNode,
	}
}

func branchNodeFromJSON(data json.RawMessage) *convert.BranchNode {
	fields, err := decodeJSONObject(data)
	if err != nil {
		return nil
	}
	if err := expectNodeType(fields, "Branch"); err != nil {
		return nil
	}

	intervalData, err := rawField(fields, "interval")
	if err != nil {
		return nil
	}
	lower, upper, err := intervalBoundsFromJSON(intervalData)
	if err != nil {
		return nil
	}

	nodeData, err := rawField(fields, "node")
	if err != nil {
		return nil
	}
	node := nodeFromJSON(nodeData)
	if node == nil {
		return nil
	}

	return &convert.BranchNode{
		Lower: lower,
		Upper: upper,
		Node:  *node,
	}
}

func (service *ConvertService) TreeToScript(root *convert.Node) (string, error) {
	if errs := convert.GetTreeErrors(root); len(errs) > 0 {
		return "", errors.Join(errs...)
	}

	builder := strings.Builder{}

	if err := treeToScript(root, 0, &builder); err != nil {
		return "", err
	}

	return builder.String(), nil
}

func treeToScript(root *convert.Node, indentation int, builder *strings.Builder) error {
	if root == nil {
		return errors.New("program node cannot be nil")
	}

	if root.Terminal != nil {
		writeTerminal(*root.Terminal, builder)
	} else {
		if root.Condition == nil {
			return errors.New("program node must define 'terminal' or 'condition' configuration")
		}
		if err := writeCondition(*root.Condition, indentation, builder); err != nil {
			return err
		}
	}

	return nil
}

// 2 spaces is the indent
const indent = "  "

func writeCondition(condition convert.ConditionNode, indentation int, builder *strings.Builder) error {
	kind, ok := condition.Metric.ValueKind()
	if !ok {
		return fmt.Errorf("unsupported metric %q", condition.Metric)
	}

	builder.WriteString(strings.Repeat(indent, indentation))
	builder.WriteString(string(condition.Metric))

	for _, branch := range condition.Branches {
		builder.WriteByte('\n')
		builder.WriteString(strings.Repeat(indent, indentation))
		builder.WriteByte('[')

		if err := writeIntervalValue(branch.Lower, kind, builder); err != nil {
			return err
		}

		builder.WriteString(", ")

		if err := writeIntervalValue(branch.Upper, kind, builder); err != nil {
			return err
		}

		builder.WriteString("] (")

		// We allow inlining if the next statement is terminal
		if branch.Node.Terminal != nil {
			writeTerminal(*branch.Node.Terminal, builder)
		} else {
			if branch.Node.Condition == nil {
				return errors.New("branch node must define 'terminal' or 'condition' configuration")
			}
			// Newline for next section
			builder.WriteByte('\n')
			if err := writeCondition(*branch.Node.Condition, indentation+1, builder); err != nil {
				return err
			}
			builder.WriteByte('\n')
			builder.WriteString(strings.Repeat(indent, indentation))
		}

		// Close statement regardless
		builder.WriteByte(')')
	}

	if condition.Default != nil {
		builder.WriteByte('\n')
		builder.WriteString(strings.Repeat(indent, indentation))
		builder.WriteString("default (")
		// We allow inlining if the next statement is terminal
		if condition.Default.Terminal != nil {
			writeTerminal(*condition.Default.Terminal, builder)
		} else {
			if condition.Default.Condition == nil {
				return errors.New("default node must define 'terminal' or 'condition' configuration")
			}
			// Newline for next section
			builder.WriteByte('\n')
			if err := writeCondition(*condition.Default.Condition, indentation+1, builder); err != nil {
				return err
			}
			builder.WriteByte('\n')
			builder.WriteString(strings.Repeat(indent, indentation))
		}

		// Close statement regardless
		builder.WriteByte(')')
	}

	return nil
}

func writeIntervalValue(bound *float64, kind bidscript.MetricValueKind, builder *strings.Builder) error {
	if bound != nil {
		if kind == bidscript.MetricValueKindInteger {
			if math.Trunc(*bound) != *bound {
				return fmt.Errorf("integer metric bounds must be whole numbers, got %f", *bound)
			}
			builder.WriteString(strconv.FormatInt(int64(*bound), 10))
		} else if kind == bidscript.MetricValueKindDecimal {
			builder.WriteString(strconv.FormatFloat(*bound, 'f', 2, 64))
		} else {
			return fmt.Errorf("unsupported metric kind %q", kind)
		}
	} else {
		builder.WriteByte('_')
	}

	return nil
}

func writeTerminal(terminal convert.TerminalNode, builder *strings.Builder) {
	// Terminals never need an indentation as per the opinionated nature of our formatter
	// Write operator
	builder.WriteByte(byte(terminal.Operator))

	// Write float as string, to 2d.p.
	builder.WriteString(strconv.FormatFloat(terminal.Amount, 'f', 2, 64))

	// Only write percentage if its set
	if terminal.Percentage {
		builder.WriteByte('%')
	}
}

func metricType(metric convert.Metric) convert.MetricType {
	kind, ok := metric.ValueKind()
	if !ok {
		return ""
	}
	return kind
}
