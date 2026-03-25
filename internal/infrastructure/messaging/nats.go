// Package messaging provides NATS infrastructure utilities.
package messaging

import (
	"time"

	"github.com/chengtb/chengtb/pkg/logger"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

const (
	defaultReconnectWait  = 2 * time.Second
	defaultMaxReconnects  = 60
	defaultRequestTimeout = 5 * time.Second
)

// Connect establishes a NATS connection with automatic reconnection.
func Connect(addr string) (*nats.Conn, error) {
	opts := []nats.Option{
		nats.ReconnectWait(defaultReconnectWait),
		nats.MaxReconnects(defaultMaxReconnects),
		nats.DisconnectErrHandler(func(_ *nats.Conn, err error) {
			logger.L().Warn("nats disconnected", zap.Error(err))
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			logger.L().Info("nats reconnected", zap.String("url", nc.ConnectedUrl()))
		}),
		nats.ClosedHandler(func(_ *nats.Conn) {
			logger.L().Warn("nats connection closed")
		}),
	}

	nc, err := nats.Connect(addr, opts...)
	if err != nil {
		return nil, err
	}
	logger.L().Info("nats connected", zap.String("addr", addr))
	return nc, nil
}

// RequestTimeout is the default timeout for NATS request-reply calls.
var RequestTimeout = defaultRequestTimeout
