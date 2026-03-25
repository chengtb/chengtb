// cmd/service is the User micro-service.  It subscribes to NATS subjects and
// processes User commands / queries using the DDD application layer.
package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/chengtb/chengtb/config"
	appuser "github.com/chengtb/chengtb/internal/application/user"
	"github.com/chengtb/chengtb/internal/infrastructure/messaging"
	"github.com/chengtb/chengtb/internal/infrastructure/persistence"
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

	// Wire up DDD layers.
	repo := persistence.NewInMemoryUserRepository()
	userSvc := appuser.NewService(repo)
	subscriber := messaging.NewSubscriber(nc, userSvc)

	if err := subscriber.Subscribe(); err != nil {
		log.Fatal("failed to subscribe", zap.Error(err))
	}

	log.Info("user service ready")

	// Wait for interrupt.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("shutting down user service")
	subscriber.Drain()
}
