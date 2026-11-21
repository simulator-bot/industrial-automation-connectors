package opcua

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gopcua/opcua/ua"
)

var ErrInvalidArguments = errors.New("invalid OPC UA method arguments")

type Argument struct {
	Type  string          `json:"type"`
	Value json.RawMessage `json:"value"`
}

type MethodCall struct {
	ObjectNodeID string     `json:"objectNodeId"`
	Arguments    []Argument `json:"arguments"`
}

func (s *session) Invoke(ctx context.Context, target string, parameters json.RawMessage) (json.RawMessage, error) {
	var call MethodCall
	if err := json.Unmarshal(parameters, &call); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidArguments, err)
	}
	nodes, err := parseNodes([]string{call.ObjectNodeID, target})
	if err != nil {
		return nil, err
	}
	req := &ua.CallMethodRequest{ObjectID: nodes[0], MethodID: nodes[1]}
	for i, arg := range call.Arguments {
		input, err := encodeArgument(arg)
		if err != nil {
			return nil, fmt.Errorf("%w: argument %d (%s): %v", ErrInvalidArguments, i+1, arg.Type, err)
		}
		req.InputArguments = append(req.InputArguments, input)
	}
	if err := s.Health(ctx); err != nil {
		return nil, err
	}
	res, err := s.client.Call(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("OPC UA Call: %w", err)
	}
	if res == nil {
		return nil, ua.StatusBadUnknownResponse
	}
	if res.StatusCode != ua.StatusOK {
		return nil, fmt.Errorf("OPC UA Call: %w (input argument results: %v)", res.StatusCode, res.InputArgumentResults)
	}
	outputs := make([]Argument, len(res.OutputArguments))
	for i, output := range res.OutputArguments {
		if output == nil {
			return nil, ua.StatusBadUnknownResponse
		}
		payload, err := json.Marshal(output.Value())
		if err != nil {
			return nil, fmt.Errorf("encode OPC UA output argument %d: %w", i+1, err)
		}
		kind := strings.TrimPrefix(output.Type().String(), "TypeID")
		kind = strings.Replace(kind, "Uint", "UInt", 1)
		if output.Has(ua.VariantArrayValues) {
			kind += "[]"
		}
		outputs[i] = Argument{Type: kind, Value: payload}
	}
	return json.Marshal(struct {
		StatusCode      string     `json:"statusCode"`
		OutputArguments []Argument `json:"outputArguments"`
	}{formatStatus(res.StatusCode), outputs})
}

func encodeArgument(arg Argument) (*ua.Variant, error) {
	if len(arg.Value) == 0 || bytes.Equal(bytes.TrimSpace(arg.Value), []byte("null")) {
		return nil, fmt.Errorf("a non-null value is required")
	}
	switch arg.Type {
	case "Boolean":
		return scalar[bool](arg.Value)
	case "SByte":
		return scalar[int8](arg.Value)
	case "Byte":
		return scalar[uint8](arg.Value)
	case "Int16":
		return scalar[int16](arg.Value)
	case "UInt16":
		return scalar[uint16](arg.Value)
	case "Int32":
		return scalar[int32](arg.Value)
	case "UInt32":
		return scalar[uint32](arg.Value)
	case "Int64":
		return scalar[int64](arg.Value)
	case "UInt64":
		return scalar[uint64](arg.Value)
	case "Float":
		return scalar[float32](arg.Value)
	case "Double":
		return scalar[float64](arg.Value)
	case "String":
		return scalar[string](arg.Value)
	case "DateTime":
		return scalar[time.Time](arg.Value)
	case "ByteString":
		return scalar[[]byte](arg.Value)
	default:
		return nil, fmt.Errorf("unsupported scalar type %q", arg.Type)
	}
}

func scalar[T any](raw json.RawMessage) (*ua.Variant, error) {
	var value T
	if err := json.Unmarshal(raw, &value); err != nil {
		return nil, err
	}
	return ua.NewVariant(value)
}
