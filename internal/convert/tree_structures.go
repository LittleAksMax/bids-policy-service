package convert

import "github.com/LittleAksMax/bidscript"

type Node struct {
	Terminal  *TerminalNode  `json:"terminal,omitempty"`
	Condition *ConditionNode `json:"condition,omitempty"`
}

type MetricType = bidscript.MetricValueKind

type Operator byte

const (
	OperatorAdd Operator = '+'
	OperatorSub Operator = '-'
	OperatorSet Operator = '='
)

type Metric string

const (
	MetricImpressions Metric = "impressions"
	MetricClicks      Metric = "clicks"
	MetricOrders      Metric = "orders"
	MetricRoaS        Metric = "roas"
	MetricACoS        Metric = "acos"
	MetricCPC         Metric = "cpc"
	MetricCTR         Metric = "ctr"
	MetricSales       Metric = "sales"
	MetricSpend       Metric = "spend"
)

func (m Metric) UsesDecimal() bool {
	kind, ok := m.ValueKind()
	return ok && kind == bidscript.MetricValueKindDecimal
}

func (m Metric) UsesInteger() bool {
	kind, ok := m.ValueKind()
	return ok && kind == bidscript.MetricValueKindInteger
}

func (m Metric) ValueKind() (MetricType, bool) {
	return bidscript.MetricValueKindFor(string(m))
}

func (m Metric) IsValid() bool {
	switch m {
	case MetricImpressions, MetricClicks, MetricOrders, MetricRoaS, MetricACoS, MetricCPC, MetricCTR, MetricSales, MetricSpend:
		return true
	}
	return false
}

func IsValidMetricType(metricType MetricType) bool {
	switch metricType {
	case bidscript.MetricValueKindInteger, bidscript.MetricValueKindDecimal:
		return true
	default:
		return false
	}
}

type TerminalNode struct {
	Operator   Operator `json:"operator" validate:"required"`
	Amount     float64  `json:"amount" validate:"nonnegative"`
	Percentage bool     `json:"percentage"`
}

type BranchNode struct {
	Lower *float64 `json:"lower" validate:"nonnegative"`
	Upper *float64 `json:"upper" validate:"nonnegative"`
	Node  Node     `json:"node"`
}

type ConditionNode struct {
	Metric     Metric       `json:"metric" validate:"required"`
	MetricType MetricType   `json:"type"`
	Branches   []BranchNode `json:"branches"`
	Default    *Node        `json:"default,omitempty"`
}
