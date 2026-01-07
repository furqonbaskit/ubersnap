package main

import (
	"log"
	"ubersnap/internal/config"
	"ubersnap/internal/database"
	"ubersnap/internal/models"
	"ubersnap/internal/routes"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.LoadConfig()

	if err := database.InitDatabase(cfg); err != nil {
		log.Fatalf("Database initialization failed: %v", err)
	}
	defer func() {
		if err := database.CloseDatabase(); err != nil {
			log.Printf("Error closing database: %v", err)
		}
	}()

	if err := database.AutoMigrate(&models.Coupon{}, &models.Claim{}); err != nil {
		log.Fatalf("Database migration failed: %v", err)
	}

	r := gin.Default()

	// Get database instance
	db := database.GetDB()

	// Setup routes
	routes.SetupRoutes(r, db)

	// Start server
	if err := r.Run(":" + cfg.ServerPort); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

}
