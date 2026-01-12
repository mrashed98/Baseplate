package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/baseplate/baseplate/config"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrUserExists         = errors.New("user with this email already exists")
	ErrTeamExists         = errors.New("team with this slug already exists")
	ErrNotFound           = errors.New("not found")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrForbidden          = errors.New("forbidden")
)

type Service struct {
	repo   *Repository
	config *config.JWTConfig
}

func NewService(repo *Repository, cfg *config.JWTConfig) *Service {
	return &Service{repo: repo, config: cfg}
}

type JWTClaims struct {
	UserID uuid.UUID `json:"user_id"`
	Email  string    `json:"email"`
	jwt.RegisteredClaims
}

// User authentication
func (s *Service) Register(ctx context.Context, req *RegisterRequest) (*AuthResponse, error) {
	existing, err := s.repo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrUserExists
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &User{
		ID:           uuid.New(),
		Email:        req.Email,
		PasswordHash: string(hash),
		Name:         req.Name,
		Status:       "active",
	}

	if err := s.repo.CreateUser(ctx, user); err != nil {
		return nil, err
	}

	token, err := s.generateToken(user)
	if err != nil {
		return nil, err
	}

	return &AuthResponse{Token: token, User: user}, nil
}

func (s *Service) Login(ctx context.Context, req *LoginRequest) (*AuthResponse, error) {
	user, err := s.repo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	token, err := s.generateToken(user)
	if err != nil {
		return nil, err
	}

	return &AuthResponse{Token: token, User: user}, nil
}

func (s *Service) GetUserByID(ctx context.Context, id uuid.UUID) (*User, error) {
	return s.repo.GetUserByID(ctx, id)
}

func (s *Service) generateToken(user *User) (string, error) {
	claims := JWTClaims{
		UserID: user.ID,
		Email:  user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.config.ExpirationDuration())),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.config.Secret))
}

func (s *Service) ValidateToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.config.Secret), nil
	})
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, ErrUnauthorized
}

// Team management
func (s *Service) CreateTeam(ctx context.Context, userID uuid.UUID, req *CreateTeamRequest) (*Team, error) {
	existing, err := s.repo.GetTeamBySlug(ctx, req.Slug)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrTeamExists
	}

	team := &Team{
		ID:   uuid.New(),
		Name: req.Name,
		Slug: req.Slug,
	}

	if err := s.repo.CreateTeam(ctx, team); err != nil {
		return nil, err
	}

	// Create default roles
	adminRole := &Role{
		ID:          uuid.New(),
		TeamID:      team.ID,
		Name:        "admin",
		Permissions: AdminPermissions,
	}
	if err := s.repo.CreateRole(ctx, adminRole); err != nil {
		return nil, err
	}

	editorRole := &Role{
		ID:          uuid.New(),
		TeamID:      team.ID,
		Name:        "editor",
		Permissions: EditorPermissions,
	}
	if err := s.repo.CreateRole(ctx, editorRole); err != nil {
		return nil, err
	}

	viewerRole := &Role{
		ID:          uuid.New(),
		TeamID:      team.ID,
		Name:        "viewer",
		Permissions: ViewerPermissions,
	}
	if err := s.repo.CreateRole(ctx, viewerRole); err != nil {
		return nil, err
	}

	// Add creator as admin
	membership := &TeamMembership{
		ID:     uuid.New(),
		TeamID: team.ID,
		UserID: userID,
		RoleID: adminRole.ID,
	}
	if err := s.repo.CreateMembership(ctx, membership); err != nil {
		return nil, err
	}

	return team, nil
}

func (s *Service) GetTeam(ctx context.Context, id uuid.UUID) (*Team, error) {
	team, err := s.repo.GetTeamByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if team == nil {
		return nil, ErrNotFound
	}
	return team, nil
}

func (s *Service) GetTeamsByUser(ctx context.Context, userID uuid.UUID) ([]*Team, error) {
	return s.repo.GetTeamsByUserID(ctx, userID)
}

func (s *Service) UpdateTeam(ctx context.Context, team *Team) error {
	return s.repo.UpdateTeam(ctx, team)
}

func (s *Service) DeleteTeam(ctx context.Context, id uuid.UUID) error {
	return s.repo.DeleteTeam(ctx, id)
}

// Role management
func (s *Service) GetRoles(ctx context.Context, teamID uuid.UUID) ([]*Role, error) {
	return s.repo.GetRolesByTeamID(ctx, teamID)
}

func (s *Service) GetRole(ctx context.Context, id uuid.UUID) (*Role, error) {
	role, err := s.repo.GetRoleByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if role == nil {
		return nil, ErrNotFound
	}
	return role, nil
}

func (s *Service) CreateRole(ctx context.Context, teamID uuid.UUID, name string, permissions []string) (*Role, error) {
	role := &Role{
		ID:          uuid.New(),
		TeamID:      teamID,
		Name:        name,
		Permissions: permissions,
	}
	if err := s.repo.CreateRole(ctx, role); err != nil {
		return nil, err
	}
	return role, nil
}

func (s *Service) UpdateRole(ctx context.Context, role *Role) error {
	return s.repo.UpdateRole(ctx, role)
}

// Membership management
func (s *Service) GetMembership(ctx context.Context, teamID, userID uuid.UUID) (*TeamMembership, error) {
	return s.repo.GetMembership(ctx, teamID, userID)
}

func (s *Service) GetMemberships(ctx context.Context, teamID uuid.UUID) ([]*TeamMembership, error) {
	return s.repo.GetMembershipsByTeamID(ctx, teamID)
}

func (s *Service) AddMember(ctx context.Context, teamID uuid.UUID, userEmail string, roleID uuid.UUID) (*TeamMembership, error) {
	user, err := s.repo.GetUserByEmail(ctx, userEmail)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrNotFound
	}

	membership := &TeamMembership{
		ID:     uuid.New(),
		TeamID: teamID,
		UserID: user.ID,
		RoleID: roleID,
	}
	if err := s.repo.CreateMembership(ctx, membership); err != nil {
		return nil, err
	}
	return membership, nil
}

func (s *Service) RemoveMember(ctx context.Context, teamID, userID uuid.UUID) error {
	return s.repo.DeleteMembership(ctx, teamID, userID)
}

func (s *Service) GetUserPermissions(ctx context.Context, teamID, userID uuid.UUID) ([]string, error) {
	membership, err := s.repo.GetMembership(ctx, teamID, userID)
	if err != nil {
		return nil, err
	}
	if membership == nil {
		return nil, ErrForbidden
	}

	role, err := s.repo.GetRoleByID(ctx, membership.RoleID)
	if err != nil {
		return nil, err
	}
	if role == nil {
		return nil, ErrForbidden
	}

	return role.Permissions, nil
}

func (s *Service) HasPermission(ctx context.Context, teamID, userID uuid.UUID, permission string) (bool, error) {
	permissions, err := s.GetUserPermissions(ctx, teamID, userID)
	if err != nil {
		return false, err
	}

	for _, p := range permissions {
		if p == permission {
			return true, nil
		}
	}
	return false, nil
}

// API Key management
func (s *Service) CreateAPIKey(ctx context.Context, teamID uuid.UUID, userID *uuid.UUID, req *CreateAPIKeyRequest) (*CreateAPIKeyResponse, error) {
	rawKey := make([]byte, 32)
	if _, err := rand.Read(rawKey); err != nil {
		return nil, err
	}
	keyString := "bp_" + hex.EncodeToString(rawKey)

	hash := sha256.Sum256([]byte(keyString))
	keyHash := hex.EncodeToString(hash[:])

	var expiresAt *time.Time
	if req.ExpiresAt != nil {
		t, err := time.Parse(time.RFC3339, *req.ExpiresAt)
		if err != nil {
			return nil, fmt.Errorf("invlid expiration date format: %w", err)
		}
		expiresAt = &t
	}

	apiKey := &APIKey{
		ID:          uuid.New(),
		TeamID:      teamID,
		UserID:      userID,
		Name:        req.Name,
		KeyHash:     keyHash,
		Permissions: req.Permissions,
		ExpiresAt:   expiresAt,
	}

	if err := s.repo.CreateAPIKey(ctx, apiKey); err != nil {
		return nil, err
	}

	return &CreateAPIKeyResponse{
		APIKey: apiKey,
		Key:    keyString,
	}, nil
}

func (s *Service) ValidateAPIKey(ctx context.Context, keyString string) (*APIKey, error) {
	hash := sha256.Sum256([]byte(keyString))
	keyHash := hex.EncodeToString(hash[:])

	apiKey, err := s.repo.GetAPIKeyByHash(ctx, keyHash)
	if err != nil {
		return nil, err
	}
	if apiKey == nil {
		return nil, ErrUnauthorized
	}

	if apiKey.ExpiresAt != nil && apiKey.ExpiresAt.Before(time.Now()) {
		return nil, ErrUnauthorized
	}

	// Update last used
	go s.repo.UpdateAPIKeyLastUsed(context.Background(), apiKey.ID)

	return apiKey, nil
}

func (s *Service) GetAPIKeys(ctx context.Context, teamID uuid.UUID) ([]*APIKey, error) {
	return s.repo.GetAPIKeysByTeamID(ctx, teamID)
}

func (s *Service) DeleteAPIKey(ctx context.Context, id uuid.UUID) error {
	return s.repo.DeleteAPIKey(ctx, id)
}
