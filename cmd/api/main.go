// cmd/api is the API gateway.  It exposes a Gin HTTP server and forwards every
// request to the User micro-service via NATS request-reply.
package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/chengtb/chengtb/config"
	"github.com/chengtb/chengtb/internal/infrastructure/messaging"
	"github.com/chengtb/chengtb/internal/infrastructure/transport"
	"github.com/chengtb/chengtb/internal/interfaces/api"
	"github.com/chengtb/chengtb/pkg/logger"
	"go.uber.org/zap"
)

func main() {
	cfg := config.Load()
	logger.Init(cfg.LogLevel)
	defer logger.Sync()

	log := logger.L()

	// Connect to NATS.
	nc, err := messaging.Connect(cfg.NATSAddr)
	if err != nil {
		log.Fatal("failed to connect to nats", zap.Error(err))
	}
	defer nc.Drain() //nolint:errcheck

	natsClient := messaging.NewClient(nc)

	// Build Gin engine and register routes.
	engine := transport.NewEngine(cfg.LogLevel == "debug")
	api.RegisterRoutes(engine, natsClient)

	srv := &http.Server{
		Addr:    cfg.HTTPAddr,
		Handler: engine,
	}

	// Start HTTP server in a goroutine.
	go func() {
		log.Info("api server starting", zap.String("addr", cfg.HTTPAddr))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal("server error", zap.Error(err))
		}
	}()

	// Wait for interrupt signal.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("shutting down api server")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Error("server forced to shutdown", zap.Error(err))
	}
}
