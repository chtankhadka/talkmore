package main

import (
	"context"
	"log"
	"my-work/config"
	"my-work/middleware"
	"my-work/routes"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	// Set Gin mode based on environment
	if os.Getenv("GIN_MODE") == "release" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	// Initialize AppConfig (includes validator)
	app, err := config.Init()
	if err != nil {
		log.Fatalf("Failed to initialize AppConfig: %v", err)
	}
	defer func() {
		if err := app.Client.Disconnect(context.Background()); err != nil {
			log.Printf("Error disconnecting MongoDB client: %v", err)
		}
	}()

	// Get port from environment or default to 8000
	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	// Initialize Gin router
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	// Public routes (no authentication)
	routes.PublicRoutes(router, app)

	// Authorized routes (with authentication middleware)
	authorized := router.Group("/api")
	authorized.Use(middleware.Authentication(app))
	routes.UserRoutes(authorized, app)

	// Start server with graceful shutdown
	srv := &http.Server{
		Addr:    "0.0.0.0:" + port,
		Handler: router,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Server failed to start: %v", err)
		}
	}()

	// Handle shutdown signals
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Server shutdown failed: %v", err)
	} else {
		log.Println("Server shut down gracefully")
	}
}
