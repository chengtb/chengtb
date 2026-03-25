package logger

import (
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	instance *zap.Logger
	once     sync.Once
)

// Init initialises the global logger with the given level.
func Init(level string) {
	once.Do(func() {
		lvl := zap.InfoLevel
		if err := lvl.UnmarshalText([]byte(level)); err != nil {
			lvl = zap.InfoLevel
		}
		cfg := zap.NewProductionConfig()
		cfg.Level = zap.NewAtomicLevelAt(lvl)
		cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		var err error
		instance, err = cfg.Build()
		if err != nil {
			panic(err)
		}
	})
}

// L returns the global logger, initialising it with "info" level if needed.
func L() *zap.Logger {
	if instance == nil {
		Init("info")
	}
	return instance
}

// Sync flushes any buffered log entries.
func Sync() {
	if instance != nil {
		_ = instance.Sync()
	}
}
