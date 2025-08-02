package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"portal64api/internal/api"
	"portal64api/internal/cache"
	"portal64api/internal/config"
	"portal64api/internal/database"

	"github.com/gin-gonic/gin"
)

// @title Portal64 API
// @version 1.0
// @description REST API for DWZ (Deutsche Wertungszahl) chess rating system
// @description Provides access to player ratings, club information, and tournament data
// @description from the SVW (Schachverband Württemberg) chess federation databases.

// @contact.name API Support
// @contact.email support@svw.info

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /

// @schemes http https

// @tag.name players
// @tag.description Player and rating operations

// @tag.name clubs
// @tag.description Club and organization operations

// @tag.name tournaments
// @tag.description Tournament operations

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Set Gin mode based on environment
	if cfg.Server.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Connect to databases
	dbs, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to databases: %v", err)
	}
	defer dbs.Close()

	// Initialize cache service
	cacheService, err := cache.NewCacheService(cfg.Cache)
	if err != nil {
		log.Fatalf("Failed to initialize cache service: %v", err)
	}
	defer func() {
		if closeErr := cacheService.Close(); closeErr != nil {
			log.Printf("Error closing cache service: %v", closeErr)
		}
	}()

	// Test cache connectivity if enabled
	if cfg.Cache.Enabled {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if pingErr := cacheService.Ping(ctx); pingErr != nil {
			log.Printf("Warning: Cache connectivity test failed: %v", pingErr)
		} else {
			log.Println("Cache service connected successfully")
		}
	} else {
		log.Println("Cache service disabled")
	}

	// Setup routes
	router := api.SetupRoutes(dbs, cacheService)

	// Create HTTP server
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Starting server on %s", addr)
		
		if cfg.Server.EnableHTTPS {
			if cfg.Server.CertFile == "" || cfg.Server.KeyFile == "" {
				log.Fatal("HTTPS enabled but cert_file or key_file not configured")
			}
			log.Printf("Starting HTTPS server on %s", addr)
			if err := srv.ListenAndServeTLS(cfg.Server.CertFile, cfg.Server.KeyFile); err != nil && err != http.ErrServerClosed {
				log.Fatalf("Failed to start HTTPS server: %v", err)
			}
		} else {
			log.Printf("Starting HTTP server on %s", addr)
			if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Fatalf("Failed to start HTTP server: %v", err)
			}
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Give outstanding requests 30 seconds to complete
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}
