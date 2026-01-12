package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/baseplate/baseplate/config"
	"github.com/baseplate/baseplate/internal/core/auth"
	"github.com/baseplate/baseplate/internal/storage/postgres"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	// Read environment variables
	superAdminEmail := os.Getenv("SUPER_ADMIN_EMAIL")
	superAdminPassword := os.Getenv("SUPER_ADMIN_PASSWORD")

	if superAdminEmail == "" || superAdminPassword == "" {
		log.Fatal("SUPER_ADMIN_EMAIL and SUPER_ADMIN_PASSWORD environment variables are required")
	}

	// Load database configuration
	cfg := config.Load()

	// Connect to database
	db, err := postgres.NewClient(&cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	// Initialize repository
	authRepo := auth.NewRepository(db)

	// Check if super admin already exists
	existing, err := authRepo.GetUserByEmail(ctx, superAdminEmail)
	if err != nil {
		log.Fatalf("Failed to check for existing user: %v", err)
	}

	if existing != nil {
		if existing.IsSuperAdmin {
			fmt.Printf("Super admin user '%s' already exists\n", superAdminEmail)
			os.Exit(0)
		}
		// Promote existing user to super admin
		if err := promoteSuperAdmin(ctx, authRepo, existing.ID); err != nil {
			log.Fatalf("Failed to promote existing user to super admin: %v", err)
		}
		fmt.Printf("Promoted existing user '%s' to super admin\n", superAdminEmail)
		return
	}

	// Create password hash
	hash, err := bcrypt.GenerateFromPassword([]byte(superAdminPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("Failed to hash password: %v", err)
	}

	// Create new super admin user
	user := &auth.User{
		ID:           uuid.New(),
		Email:        superAdminEmail,
		PasswordHash: string(hash),
		Name:         "Super Admin",
		Status:       "active",
		IsSuperAdmin: true,
	}

	if err := authRepo.CreateUser(ctx, user); err != nil {
		log.Fatalf("Failed to create super admin user: %v", err)
	}

	fmt.Printf("Successfully created super admin user: %s\n", superAdminEmail)
}

// promoteSuperAdmin promotes an existing user to super admin
func promoteSuperAdmin(ctx context.Context, repo *auth.Repository, userID uuid.UUID) error {
	// Fetch the user first
	user, err := repo.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}
	if user == nil {
		return fmt.Errorf("user not found")
	}

	// Update user to be super admin
	now := time.Now()
	user.IsSuperAdmin = true
	user.SuperAdminPromotedAt = &now
	// SuperAdminPromotedBy is nil for initial setup

	return repo.UpdateUser(ctx, user)
}
