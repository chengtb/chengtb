package main

import (
	"log"

	"restaurant-system/config"
	"restaurant-system/database"
	"restaurant-system/router"
	"restaurant-system/services"
)

func main() {
	cfg := config.Load()

	if err := database.Init(cfg); err != nil {
		log.Fatalf("database init failed: %v", err)
	}

	if err := database.Migrate(); err != nil {
		log.Fatalf("database migration failed: %v", err)
	}

	// Start auto-dispatch background goroutine
	go services.RunAutoDispatch(cfg.DispatchIntervalSec, cfg.DispatchWindowMin)

	r := router.Setup()
	log.Printf("Server starting on :%s", cfg.ServerPort)
	if err := r.Run(":" + cfg.ServerPort); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
