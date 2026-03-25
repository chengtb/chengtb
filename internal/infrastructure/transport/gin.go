// Package transport sets up the Gin HTTP engine with global middleware.
package transport

import (
	"net/http"
	"time"

	"github.com/chengtb/chengtb/pkg/logger"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// NewEngine creates a Gin engine with logger, recovery and CORS middleware.
func NewEngine(debug bool) *gin.Engine {
	if !debug {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(zapLogger())
	r.Use(gin.Recovery())
	r.Use(cors())

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "time": time.Now().UTC()})
	})

	return r
}

// zapLogger is a Gin middleware that logs each request using Zap.
func zapLogger() gin.HandlerFunc {
	log := logger.L()
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path

		c.Next()

		log.Info("http",
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.Int("status", c.Writer.Status()),
			zap.Duration("latency", time.Since(start)),
			zap.String("client_ip", c.ClientIP()),
		)
	}
}

// cors adds permissive CORS headers — tighten for production.
func cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
