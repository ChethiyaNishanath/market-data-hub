package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/ChethiyaNishanath/market-data-hub/internal/app"
	"github.com/ChethiyaNishanath/market-data-hub/internal/config"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"
)

func main() {

	config.Init()
	cfg := config.GetConfig()

	slog.SetLogLoggerLevel(slog.LevelDebug)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	newApp := app.NewApp(&ctx)
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Use(middleware.Timeout(60 * time.Second))

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET"},
		AllowedHeaders:   []string{"Accept", "Content-Type"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	r.Use(httprate.LimitByIP(100, 1*time.Minute))

	newApp.RegisterRoutes(r)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%s", "8081"),
		Handler: r,
	}

	go func() {
		slog.Info(fmt.Sprintf("Server starting on port %s", "8081"))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Server failed", "error", err)
		}
	}()

	<-ctx.Done()
	slog.Info("Shutting down gracefully...")

	stop()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Duration(func() int {
		v, _ := strconv.Atoi(cfg.ShutdownTimeout)
		return v
	}())*time.Second)

	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		slog.Error("Forced server shutdown", "error", err)
	}

	slog.Info("Shutdown complete")

}
