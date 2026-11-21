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
	v1 "github.com/simulator-bot/industrial-automation-connectors/internal/contracts/v1"
)

var ErrInvalidArguments = errors.New("invalid OPC UA method arguments")

type Argument = v1.OPCUAArgument

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
	decode, ok := argumentTypes[arg.Type]
	if !ok {
		return nil, fmt.Errorf("unsupported scalar type %q", arg.Type)
	}
	return decode(arg.Value)
}

var argumentTypes = map[string]func(json.RawMessage) (*ua.Variant, error){
	"Boolean": scalar[bool], "SByte": scalar[int8], "Byte": scalar[uint8],
	"Int16": scalar[int16], "UInt16": scalar[uint16], "Int32": scalar[int32],
	"UInt32": scalar[uint32], "Int64": scalar[int64], "UInt64": scalar[uint64],
	"Float": scalar[float32], "Double": scalar[float64], "String": scalar[string],
	"DateTime": scalar[time.Time], "ByteString": scalar[[]byte],
}

func SupportsArgumentType(kind string) bool {
	_, ok := argumentTypes[kind]
	return ok
}

func scalar[T any](raw json.RawMessage) (*ua.Variant, error) {
	var value T
	if err := json.Unmarshal(raw, &value); err != nil {
		return nil, err
	}
	return ua.NewVariant(value)
}
