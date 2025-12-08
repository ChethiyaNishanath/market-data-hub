package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	pb "github.com/ChethiyaNishanath/market-data-hub/api/orderbook"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var snapshotCmd = &cobra.Command{
	Use:   "snapshot",
	Short: "Retrieve a snapshot of the current order book",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runOrderBookGrpcClient()
	},
}

func init() {
	rootCmd.AddCommand(snapshotCmd)

	snapshotCmd.Flags().String("addr", "0.0.0.0:50051", "gRPC server address")
	snapshotCmd.Flags().String("symbol", "BTCUSDT", "Symbol to fetch the snapshot for")

	viper.BindPFlag("addr", snapshotCmd.Flags().Lookup("addr"))
	viper.BindPFlag("symbol", snapshotCmd.Flags().Lookup("symbol"))
}

func runOrderBookGrpcClient() error {
	addr := viper.GetString("addr")
	sym := viper.GetString("symbol")

	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		slog.Error("did not connect: %v", "error", err)
	}

	defer conn.Close()

	c := pb.NewOrderBookClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	res, err := c.GetSnapshot(ctx, &pb.OrderBookSnapshotRequest{Symbol: sym})

	if err != nil {
		slog.Error("snapshot not received %w", "error", err)
		return fmt.Errorf("snapshot not received: %w", err)
	}

	slog.Info("Snapshot received", "symbol", sym, "data", res)

	return nil
}
