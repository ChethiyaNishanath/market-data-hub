package orderbook_test

import (
	"context"
	"net"
	"testing"

	pb "github.com/ChethiyaNishanath/market-data-hub/api/orderbook"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

const bufSize = 1024 * 1024

type mockOrderBookStore struct {
	snapshots map[string]*pb.GetSnapshotReply
}

func (m *mockOrderBookStore) GetSnapshot(symbol string) (*pb.GetSnapshotReply, error) {
	if symbol == "" {
		return nil, status.Error(codes.InvalidArgument, "symbol is empty")
	}
	snap, ok := m.snapshots[symbol]
	if !ok {
		return nil, status.Error(codes.NotFound, "symbol not found")
	}
	return snap, nil
}

type orderBookServer struct {
	pb.UnimplementedOrderBookServer
	store *mockOrderBookStore
}

func (s *orderBookServer) GetSnapshot(ctx context.Context, req *pb.OrderBookSnapshotRequest) (*pb.GetSnapshotReply, error) {
	return s.store.GetSnapshot(req.Symbol)
}

func setupServer(store *mockOrderBookStore) (*bufconn.Listener, *grpc.ClientConn, pb.OrderBookClient, func()) {
	lis := bufconn.Listen(bufSize)
	s := grpc.NewServer()
	srv := &orderBookServer{store: store}
	pb.RegisterOrderBookServer(s, srv)

	go func() {
		_ = s.Serve(lis)
	}()

	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(
		func(context.Context, string) (net.Conn, error) {
			return lis.Dial()
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	if err != nil {
		panic(err)
	}

	client := pb.NewOrderBookClient(conn)

	cleanup := func() {
		conn.Close()
		s.Stop()
	}

	return lis, conn, client, cleanup
}

func TestGetSnapshot_ValidRequest(t *testing.T) {
	store := &mockOrderBookStore{
		snapshots: map[string]*pb.GetSnapshotReply{
			"BNBBTC": {
				Symbol:       "BNBBTC",
				LastUpdateId: "123",
				Bids: []*pb.Order{
					{Price: "0.009", Amount: "100"},
				},
				Asks: []*pb.Order{
					{Price: "0.010", Amount: "150"},
				},
			},
		},
	}

	_, _, client, cleanup := setupServer(store)
	defer cleanup()

	resp, err := client.GetSnapshot(context.Background(),
		&pb.OrderBookSnapshotRequest{Symbol: "BNBBTC"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if resp.Symbol != "BNBBTC" {
		t.Errorf("expected symbol BNBBTC, got %s", resp.Symbol)
	}

	if len(resp.Bids) != 1 || len(resp.Asks) != 1 {
		t.Errorf("expected bids and asks length 1")
	}
}

func TestGetSnapshot_EmptySymbol(t *testing.T) {
	store := &mockOrderBookStore{snapshots: map[string]*pb.GetSnapshotReply{}}

	_, _, client, cleanup := setupServer(store)
	defer cleanup()

	_, err := client.GetSnapshot(context.Background(),
		&pb.OrderBookSnapshotRequest{Symbol: ""})

	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("expected InvalidArgument, got %v", err)
	}
}

func TestGetSnapshot_SymbolNotFound(t *testing.T) {
	store := &mockOrderBookStore{snapshots: map[string]*pb.GetSnapshotReply{}}

	_, _, client, cleanup := setupServer(store)
	defer cleanup()

	_, err := client.GetSnapshot(context.Background(),
		&pb.OrderBookSnapshotRequest{Symbol: "NOT_EXIST"})

	if status.Code(err) != codes.NotFound {
		t.Fatalf("expected NotFound error, got %v", err)
	}
}

func TestGetSnapshot_LargeOrderBook(t *testing.T) {
	largeBids := make([]*pb.Order, 10000)
	largeAsks := make([]*pb.Order, 10000)

	for i := range 10000 {
		largeBids[i] = &pb.Order{Price: "1.0", Amount: "1"}
		largeAsks[i] = &pb.Order{Price: "2.0", Amount: "1"}
	}

	store := &mockOrderBookStore{
		snapshots: map[string]*pb.GetSnapshotReply{
			"ETHBTC": {
				Symbol:       "ETHBTC",
				LastUpdateId: "9999",
				Bids:         largeBids,
				Asks:         largeAsks,
			},
		},
	}

	_, _, client, cleanup := setupServer(store)
	defer cleanup()

	resp, err := client.GetSnapshot(context.Background(),
		&pb.OrderBookSnapshotRequest{Symbol: "ETHBTC"})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(resp.Bids) != 10000 || len(resp.Asks) != 10000 {
		t.Fatalf("expected 10000 bids/asks, got %d/%d",
			len(resp.Bids), len(resp.Asks))
	}
}
