package opcua

import (
	"encoding/json"
	"math"
	"testing"

	"github.com/gopcua/opcua/ua"
)

func TestEncodeArgumentPreservesUnsignedInteger(t *testing.T) {
	got, err := encodeArgument(Argument{Type: "UInt64", Value: json.RawMessage("18446744073709551615")})
	if err != nil {
		t.Fatal(err)
	}
	if got.Type() != ua.TypeIDUint64 || got.Value() != uint64(math.MaxUint64) {
		t.Fatalf("got type %v and value %v", got.Type(), got.Value())
	}
}

func TestEncodeArgumentRejectsInvalidUInt32(t *testing.T) {
	for _, input := range []string{"-1", "4294967296", "1.5", `"1"`, "null"} {
		t.Run(input, func(t *testing.T) {
			if _, err := encodeArgument(Argument{Type: "UInt32", Value: json.RawMessage(input)}); err == nil {
				t.Fatalf("accepted %s as UInt32", input)
			}
		})
	}
}
