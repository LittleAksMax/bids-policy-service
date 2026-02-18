package repository

import (
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
)

// conditionDTOBSON is a lightweight DTO for BSON decoding of condition nodes,
// carrying children as bson.Raw for recursive decoding.
type conditionDTOBSON struct {
	Type     string   `bson:"type"`
	Variable string   `bson:"variable"`
	Min      float64  `bson:"min"`
	Max      float64  `bson:"max"`
	If       bson.Raw `bson:"if"`
	Else     bson.Raw `bson:"else"`
}

// UnmarshalRuleNodeBSON decodes a polymorphic RuleNode from BSON, recursively handling condition branches.
func UnmarshalRuleNodeBSON(rules bson.Raw) (RuleNode, error) {
	if len(rules) == 0 {
		return nil, nil
	}

	// Peek at the discriminator first
	var header struct {
		Type string `bson:"type"`
	}
	if err := bson.Unmarshal(rules, &header); err != nil {
		return nil, fmt.Errorf("decode header: %w", err)
	}

	switch header.Type {
	case TerminalType:
		var t Terminal
		if err := bson.Unmarshal(rules, &t); err != nil {
			return nil, fmt.Errorf("decode terminal: %w", err)
		}
		return &t, nil
	case ConditionType:
		var dto conditionDTOBSON
		if err := bson.Unmarshal(rules, &dto); err != nil {
			return nil, fmt.Errorf("decode condition dto: %w", err)
		}
		ifNode, err := UnmarshalRuleNodeBSON(dto.If)
		if err != nil {
			return nil, fmt.Errorf("if->%w", err)
		}
		elseNode, err := UnmarshalRuleNodeBSON(dto.Else)
		if err != nil {
			return nil, fmt.Errorf("else->%w", err)
		}
		c := Condition{
			Type:     dto.Type,
			Variable: dto.Variable,
			Min:      dto.Min,
			Max:      dto.Max,
			If:       ifNode,
			Else:     elseNode,
		}
		return &c, nil
	default:
		return nil, fmt.Errorf("invalid rule node type: %s", header.Type)
	}
}
