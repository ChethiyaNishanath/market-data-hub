package grpc

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net"
	"strconv"

	pb "github.com/ChethiyaNishanath/market-data-hub/api/orderbook"
	"github.com/ChethiyaNishanath/market-data-hub/internal/domain/orderbook"
	"github.com/ChethiyaNishanath/market-data-hub/internal/store/memory"
	"google.golang.org/grpc"
)

var (
	port = flag.Int("port", 50051, "The server port")
)

type server struct {
	pb.UnimplementedOrderBookServer
}

func (s *server) GetSnapshot(_ context.Context, in *pb.OrderBookSnapshotRequest) (*pb.GetSnapshotReply, error) {
	symbol := in.GetSymbol()
	store := memory.GetOrderBookStore()
	orderBook, ok := store.GetItem(symbol)

	if !ok {
		slog.Warn("orderbook not found", "symbol", symbol)
		return nil, fmt.Errorf("orderbook not found:: %s", symbol)
	}

	resp := MapOrderBookToSnapshot(symbol, orderBook)

	return resp, nil
}

func RunGrpcServer() {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		slog.Error("Failed to listen grpc: %v", "error", err)
	}
	s := grpc.NewServer()
	pb.RegisterOrderBookServer(s, &server{})
	slog.Info("GRPC server listening at " + lis.Addr().String())
	if err := s.Serve(lis); err != nil {
		slog.Error("Failed to serve grpc: %v", "error", err)
	}
}

func MapOrderBookToSnapshot(symbol string, ob *orderbook.OrderBook) *pb.GetSnapshotReply {
	return &pb.GetSnapshotReply{
		Symbol:       symbol,
		LastUpdateId: strconv.Itoa(ob.LastUpdateID),
		Bids:         mapLevels(ob.Bids),
		Asks:         mapLevels(ob.Asks),
	}
}

func mapLevels(levels [][]string) []*pb.Order {
	orders := make([]*pb.Order, 0, len(levels))

	for _, lvl := range levels {
		if len(lvl) < 2 {
			continue
		}

		orders = append(orders, &pb.Order{
			Price:  lvl[0],
			Amount: lvl[1],
		})
	}

	return orders
}
