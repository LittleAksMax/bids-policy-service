package repository

type RuleNode interface {
	isRuleNode()
}

const (
	terminalType  = "terminal"
	conditionType = "condition"
)

func IsValidRuleNodeType(t string) bool {
	return t == terminalType || t == conditionType
}

const (
	opAdd = "add"
	opMul = "mul"
)

func IsValidRuleOpType(t string) bool {
	return t == opAdd || t == opMul
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

const (
	NestedRuleType = "nested"
)

// NOTE: can use binary search to optimise eventually
func IsValidVariable(v string) bool {
	switch v {
	case Impressions, Clicks, CTR, Spend, CPC, Orders, Sales, ACoS, RoAS:
		return true
	default:
		return false
	}
}

type Condition struct {
	Type     string   `bson:"type"` // "condition"
	Variable string   `bson:"variable"`
	Min      float64  `bson:"min"`
	Max      float64  `bson:"max"`
	If       RuleNode `bson:"if"`
	Else     RuleNode `bson:"else"`
}

func (*Condition) isRuleNode() {}

type Terminal struct {
	Type string `bson:"type"` // "terminal"
	Op   RuleOp `bson:"op"`
}

func (*Terminal) isRuleNode() {}

type RuleOp struct {
	Type   string     `bson:"type"` // "add" or "mul"
	Amount RuleAmount `bson:"amount"`
}

type RuleAmount struct {
	Neg    bool    `bson:"neg"`
	Amount float64 `bson:"amount"`
	Perc   bool    `bson:"perc"`
}

type Policy struct {
	ID          string   `bson:"_id,omitempty"`
	UserID      string   `bson:"user_id"`
	Marketplace string   `bson:"marketplace"`
	Name        string   `bson:"name"`
	Type        string   `bson:"type"`
	Rules       RuleNode `bson:"rules"`
}
