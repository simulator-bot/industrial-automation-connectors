package opcua

import (
	"testing"
	"time"

	"github.com/gopcua/opcua/ua"
	"github.com/simulator-bot/industrial-automation-connectors/internal/protocol"
)

func TestDecodeValuePreservesStatusSeverity(t *testing.T) {
	for _, test := range []struct {
		status  ua.StatusCode
		quality string
	}{
		{ua.StatusCode(0x00000480), "GOOD"},
		{ua.StatusUncertain, "UNCERTAIN"},
		{ua.StatusBadNodeIDUnknown, "BAD"},
	} {
		got, err := decodeValue("i=2258", &ua.DataValue{Status: test.status}, time.Now().UTC())
		if err != nil {
			t.Fatal(err)
		}
		if got.Quality != test.quality || got.ProtocolStatus != formatStatus(test.status) ||
			string(got.Value) != "null" || got.SourceTimestamp != nil {
			t.Fatalf("unexpected value for %s: %+v", test.status, got)
		}
	}
}

func TestQueueUpdateStopsOnBackpressure(t *testing.T) {
	updates := make(chan protocol.Update, 2)
	if !queueUpdate(updates, protocol.Update{Value: protocol.Value{TargetID: "i=2258"}}) {
		t.Fatal("initial update was rejected")
	}
	if queueUpdate(updates, protocol.Update{}) {
		t.Fatal("overflowing update was accepted")
	}
	if got := <-updates; got.Value.TargetID != "i=2258" {
		t.Fatalf("initial update was lost: %+v", got)
	}
	if got := <-updates; got.Err == nil {
		t.Fatal("the stream did not report backpressure")
	}
}
