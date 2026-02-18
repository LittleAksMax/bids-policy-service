package repository

import "go.mongodb.org/mongo-driver/bson/primitive"

type RuleNode interface {
	IsRuleNode()
}

const (
	TerminalType  = "terminal"
	ConditionType = "condition"
)

func IsValidRuleNodeType(t string) bool {
	return t == TerminalType || t == ConditionType
}

const (
	OpAdd = "add"
	OpMul = "mul"
)

func IsValidRuleOpType(t string) bool {
	return t == OpAdd || t == OpMul
}

const (
	Impressions = "impressions"
	Clicks      = "clicks"
	CTR         = "ctr"
	Spend       = "spend"
	CPC         = "cpc"
	Orders      = "orders"
	Sales       = "sales"
	ACoS        = "acos"
	RoAS        = "roas"
)

func IsValidVariable(v string) bool {
	switch v {
	case Impressions, Clicks, CTR, Spend, CPC, Orders, Sales, ACoS, RoAS:
		return true
	default:
		return false
	}
}

const (
	NestedRuleType = "nested"
)

func IsValidRuleType(t string) bool {
	switch t {
	case NestedRuleType:
		return true
	default:
		return false
	}
}

type Condition struct {
	Type     string   `bson:"type" json:"type"` // "condition"
	Variable string   `bson:"variable" json:"variable"`
	Min      float64  `bson:"min" json:"min"`
	Max      float64  `bson:"max" json:"max"`
	If       RuleNode `bson:"if" json:"if"`
	Else     RuleNode `bson:"else" json:"else"`
}

func (*Condition) IsRuleNode() {}

type Terminal struct {
	Type string `bson:"type" json:"type"` // "terminal"
	Op   RuleOp `bson:"op" json:"op"`
}

func (*Terminal) IsRuleNode() {}

type RuleOp struct {
	Type   string     `bson:"type" json:"type"` // "add" or "mul"
	Amount RuleAmount `bson:"amount" json:"amount"`
}

type RuleAmount struct {
	Neg    bool    `bson:"neg" json:"neg"`
	Amount float64 `bson:"amount" json:"amount"`
	Perc   bool    `bson:"perc" json:"perc"`
}

type Policy struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	UserID      string             `bson:"user_id" json:"user_id"`
	Marketplace string             `bson:"marketplace" json:"marketplace"`
	Name        string             `bson:"name" json:"name"`
	Type        string             `bson:"type" json:"type"`
	Rules       RuleNode           `bson:"rules" json:"rules"`
}

const (
	MpUK = "UK"
	MpDE = "DE"
	MpFR = "FR"
	MpIT = "IT"
	MpES = "ES"
	MpUS = "US"
	MpCA = "CA"
	MpMX = "MX"
)

func IsValidMarketplace(marketplace string) bool {
	switch marketplace {
	case MpUK, MpDE, MpFR, MpIT, MpES, MpUS, MpCA, MpMX:
		return true
	default:
		return false
	}
}
