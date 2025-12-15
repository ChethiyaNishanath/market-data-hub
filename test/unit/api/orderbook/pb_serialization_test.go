package orderbook_test

import (
	"testing"

	pb "github.com/ChethiyaNishanath/market-data-hub/api/orderbook"
	"google.golang.org/protobuf/proto"
)

func TestOrderBookSnapshotRequestSerialization(t *testing.T) {
	msg := &pb.OrderBookSnapshotRequest{
		Symbol: "BNBBTC",
	}

	data, err := proto.Marshal(msg)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	var out pb.OrderBookSnapshotRequest
	if err := proto.Unmarshal(data, &out); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if !proto.Equal(msg, &out) {
		t.Errorf("proto not equal: expected %v, got %v", msg, &out)
	}
}
