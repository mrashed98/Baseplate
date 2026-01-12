package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/baseplate/baseplate/config"
	"github.com/baseplate/baseplate/internal/api"
	"github.com/baseplate/baseplate/internal/api/handlers"
	"github.com/baseplate/baseplate/internal/core/auth"
	"github.com/baseplate/baseplate/internal/core/blueprint"
	"github.com/baseplate/baseplate/internal/core/entity"
	"github.com/baseplate/baseplate/internal/core/validation"
	"github.com/baseplate/baseplate/internal/storage/postgres"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Validate critical configuration
	if cfg.JWT.Secret == "" {
		log.Fatalf("JWT_SECRET environment variable is required")
	}

	// Connect to database
	db, err := postgres.NewClient(&cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	log.Println("Connected to database")

	// Initialize repositories
	authRepo := auth.NewRepository(db)
	blueprintRepo := blueprint.NewRepository(db)
	entityRepo := entity.NewRepository(db)

	// Initialize services
	authService := auth.NewService(authRepo, &cfg.JWT)
	blueprintService := blueprint.NewService(blueprintRepo)
	validator := validation.NewValidator()
	entityService := entity.NewService(entityRepo, blueprintService, validator)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authService)
	teamHandler := handlers.NewTeamHandler(authService)
	blueprintHandler := handlers.NewBlueprintHandler(blueprintService)
	entityHandler := handlers.NewEntityHandler(entityService)

	// Setup router
	router := api.NewRouter(
		authService,
		authHandler,
		teamHandler,
		blueprintHandler,
		entityHandler,
	)

	engine := router.Setup(cfg.Server.Mode)

	// Graceful shutdown
	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit
		log.Println("Shutting down server...")
		db.Close()
		os.Exit(0)
	}()

	// Start server
	log.Printf("Starting server on port %s", cfg.Server.Port)
	if err := engine.Run(":" + cfg.Server.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
