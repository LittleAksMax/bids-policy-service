package service

import (
	"encoding/json"
	"fmt"

	"github.com/LittleAksMax/bids-policy-service/internal/convert"
	"github.com/LittleAksMax/bidscript"
)

func intervalBoundsFromJSON(data json.RawMessage) (*float64, *float64, error) {
	fields, err := decodeJSONObject(data)
	if err != nil {
		return nil, nil, err
	}
	if err := expectNodeType(fields, "Interval"); err != nil {
		return nil, nil, err
	}

	lowerData, err := rawField(fields, "lower")
	if err != nil {
		return nil, nil, err
	}
	lower, err := boundFromJSON(lowerData)
	if err != nil {
		return nil, nil, err
	}

	upperData, err := rawField(fields, "upper")
	if err != nil {
		return nil, nil, err
	}
	upper, err := boundFromJSON(upperData)
	if err != nil {
		return nil, nil, err
	}

	return lower, upper, nil
}

func metricFromJSON(data json.RawMessage) (convert.Metric, error) {
	fields, err := decodeJSONObject(data)
	if err != nil {
		return "", err
	}
	if err := expectNodeType(fields, "MetricRef"); err != nil {
		return "", err
	}

	name, err := stringField(fields, "name")
	if err != nil {
		return "", err
	}

	metric := convert.Metric(name)
	if !metric.IsValid() {
		return "", fmt.Errorf("unsupported metric %q", name)
	}

	return metric, nil
}

func boundFromJSON(data json.RawMessage) (*float64, error) {
	fields, err := decodeJSONObject(data)
	if err != nil {
		return nil, err
	}

	nodeType, err := stringField(fields, "node_type")
	if err != nil {
		return nil, err
	}

	switch nodeType {
	case "UnboundedLiteral":
		return nil, nil
	case "IntegerLiteral", "DecimalLiteral":
		value, err := numberField(fields, "value")
		if err != nil {
			return nil, err
		}
		return &value, nil
	default:
		return nil, fmt.Errorf("unsupported interval bound node_type %q", nodeType)
	}
}

func numericLiteralFromJSON(data json.RawMessage) (float64, error) {
	fields, err := decodeJSONObject(data)
	if err != nil {
		return 0, err
	}

	nodeType, err := stringField(fields, "node_type")
	if err != nil {
		return 0, err
	}
	if nodeType != "IntegerLiteral" && nodeType != "DecimalLiteral" {
		return 0, fmt.Errorf("unsupported numeric literal node_type %q", nodeType)
	}

	return numberField(fields, "value")
}

func operatorFromLiteral(literal string) (bidscript.Operator, error) {
	switch literal {
	case "+":
		return bidscript.OperatorAdd, nil
	case "-":
		return bidscript.OperatorSub, nil
	case "=":
		return bidscript.OperatorEq, nil
	default:
		return 0, fmt.Errorf("unsupported operator %q", literal)
	}
}

func decodeJSONObject(data []byte) (rawJSONObject, error) {
	var fields rawJSONObject
	if err := json.Unmarshal(data, &fields); err != nil {
		return nil, fmt.Errorf("decode json object: %w", err)
	}
	return fields, nil
}

func expectNodeType(fields rawJSONObject, want string) error {
	got, err := stringField(fields, "node_type")
	if err != nil {
		return err
	}
	if got != want {
		return fmt.Errorf("unexpected node_type %q, want %q", got, want)
	}
	return nil
}

func rawField(fields rawJSONObject, name string) (json.RawMessage, error) {
	value, ok := fields[name]
	if !ok || len(value) == 0 {
		return nil, fmt.Errorf("missing %q field", name)
	}
	return value, nil
}

func stringField(fields rawJSONObject, name string) (string, error) {
	value, err := rawField(fields, name)
	if err != nil {
		return "", err
	}

	var decoded string
	if err := json.Unmarshal(value, &decoded); err != nil {
		return "", fmt.Errorf("decode %q as string: %w", name, err)
	}
	return decoded, nil
}

func boolField(fields rawJSONObject, name string) (bool, error) {
	value, err := rawField(fields, name)
	if err != nil {
		return false, err
	}

	var decoded bool
	if err := json.Unmarshal(value, &decoded); err != nil {
		return false, fmt.Errorf("decode %q as bool: %w", name, err)
	}
	return decoded, nil
}

func numberField(fields rawJSONObject, name string) (float64, error) {
	value, err := rawField(fields, name)
	if err != nil {
		return 0, err
	}

	var decoded float64
	if err := json.Unmarshal(value, &decoded); err != nil {
		return 0, fmt.Errorf("decode %q as number: %w", name, err)
	}
	return decoded, nil
}

func arrayField(fields rawJSONObject, name string) ([]json.RawMessage, error) {
	value, err := rawField(fields, name)
	if err != nil {
		return nil, err
	}

	var decoded []json.RawMessage
	if err := json.Unmarshal(value, &decoded); err != nil {
		return nil, fmt.Errorf("decode %q as array: %w", name, err)
	}
	return decoded, nil
}
