package api

import (
	"github.com/gin-gonic/gin"

	"github.com/baseplate/baseplate/internal/api/handlers"
	"github.com/baseplate/baseplate/internal/api/middleware"
	"github.com/baseplate/baseplate/internal/core/auth"
)

type Router struct {
	engine           *gin.Engine
	authMiddleware   *middleware.AuthMiddleware
	authHandler      *handlers.AuthHandler
	teamHandler      *handlers.TeamHandler
	blueprintHandler *handlers.BlueprintHandler
	entityHandler    *handlers.EntityHandler
	adminHandler     *handlers.AdminHandler
}

func NewRouter(
	authService *auth.Service,
	authHandler *handlers.AuthHandler,
	teamHandler *handlers.TeamHandler,
	blueprintHandler *handlers.BlueprintHandler,
	entityHandler *handlers.EntityHandler,
	adminHandler *handlers.AdminHandler,
) *Router {
	return &Router{
		authMiddleware:   middleware.NewAuthMiddleware(authService),
		authHandler:      authHandler,
		teamHandler:      teamHandler,
		blueprintHandler: blueprintHandler,
		entityHandler:    entityHandler,
		adminHandler:     adminHandler,
	}
}

func (r *Router) Setup(mode string) *gin.Engine {
	gin.SetMode(mode)
	r.engine = gin.New()
	r.engine.Use(gin.Recovery())
	r.engine.Use(gin.Logger())
	r.engine.Use(middleware.ErrorHandler())
	r.engine.Use(middleware.AuditMiddleware())

	r.setupRoutes()
	return r.engine
}

func (r *Router) setupRoutes() {
	api := r.engine.Group("/api")

	// Health check
	api.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Auth routes (public)
	authRoutes := api.Group("/auth")
	{
		authRoutes.POST("/register", r.authHandler.Register)
		authRoutes.POST("/login", r.authHandler.Login)
	}

	// Protected routes
	protected := api.Group("")
	protected.Use(r.authMiddleware.Authenticate())
	{
		// Current user
		protected.GET("/auth/me", r.authHandler.Me)

		// Teams (requires auth, no specific team)
		teams := protected.Group("/teams")
		{
			teams.POST("", r.teamHandler.Create)
			teams.GET("", r.teamHandler.List)
		}

		// Team-specific routes
		team := protected.Group("/teams/:teamId")
		team.Use(r.authMiddleware.RequireTeam())
		{
			team.GET("", r.teamHandler.Get)
			team.PUT("", r.authMiddleware.RequirePermission(auth.PermTeamManage), r.teamHandler.Update)
			team.DELETE("", r.authMiddleware.RequirePermission(auth.PermTeamManage), r.teamHandler.Delete)

			// Roles
			team.GET("/roles", r.teamHandler.ListRoles)
			team.POST("/roles", r.authMiddleware.RequirePermission(auth.PermTeamManage), r.teamHandler.CreateRole)

			// Members
			team.GET("/members", r.teamHandler.ListMembers)
			team.POST("/members", r.authMiddleware.RequirePermission(auth.PermTeamManage), r.teamHandler.AddMember)
			team.DELETE("/members/:userId", r.authMiddleware.RequirePermission(auth.PermTeamManage), r.teamHandler.RemoveMember)

			// API Keys
			team.GET("/api-keys", r.teamHandler.ListAPIKeys)
			team.POST("/api-keys", r.authMiddleware.RequirePermission(auth.PermTeamManage), r.teamHandler.CreateAPIKey)
		}

		// API key deletion (not team-scoped in URL)
		protected.DELETE("/api-keys/:keyId", r.authMiddleware.RequireTeam(), r.authMiddleware.RequirePermission(auth.PermTeamManage), r.teamHandler.DeleteAPIKey)

		// Blueprints (team required via header or param)
		blueprints := protected.Group("/blueprints")
		blueprints.Use(r.authMiddleware.RequireTeam())
		{
			blueprints.POST("", r.authMiddleware.RequirePermission(auth.PermBlueprintWrite), r.blueprintHandler.Create)
			blueprints.GET("", r.authMiddleware.RequirePermission(auth.PermBlueprintRead), r.blueprintHandler.List)
			blueprints.GET("/:id", r.authMiddleware.RequirePermission(auth.PermBlueprintRead), r.blueprintHandler.Get)
			blueprints.PUT("/:id", r.authMiddleware.RequirePermission(auth.PermBlueprintWrite), r.blueprintHandler.Update)
			blueprints.DELETE("/:id", r.authMiddleware.RequirePermission(auth.PermBlueprintDelete), r.blueprintHandler.Delete)

			// Entities under blueprint
			blueprints.POST("/:blueprintId/entities", r.authMiddleware.RequirePermission(auth.PermEntityWrite), r.entityHandler.Create)
			blueprints.GET("/:blueprintId/entities", r.authMiddleware.RequirePermission(auth.PermEntityRead), r.entityHandler.List)
			blueprints.POST("/:blueprintId/entities/search", r.authMiddleware.RequirePermission(auth.PermEntityRead), r.entityHandler.Search)
			blueprints.GET("/:blueprintId/entities/by-identifier/:identifier", r.authMiddleware.RequirePermission(auth.PermEntityRead), r.entityHandler.GetByIdentifier)
		}

		// Entity direct access (by ID)
		entities := protected.Group("/entities")
		entities.Use(r.authMiddleware.RequireTeam())
		{
			entities.GET("/:id", r.authMiddleware.RequirePermission(auth.PermEntityRead), r.entityHandler.Get)
			entities.PUT("/:id", r.authMiddleware.RequirePermission(auth.PermEntityWrite), r.entityHandler.Update)
			entities.DELETE("/:id", r.authMiddleware.RequirePermission(auth.PermEntityDelete), r.entityHandler.Delete)
		}

		// Admin routes (super admin only)
		admin := protected.Group("/admin")
		admin.Use(r.authMiddleware.RequireSuperAdmin())
		{
			// Team management
			admin.GET("/teams", r.adminHandler.ListTeams)
			admin.GET("/teams/:teamId", r.adminHandler.GetTeamDetail)

			// User management
			admin.GET("/users", r.adminHandler.ListUsers)
			admin.GET("/users/:userId", r.adminHandler.GetUserDetail)
			admin.PUT("/users/:userId", r.adminHandler.UpdateUser)
			admin.POST("/users/:userId/promote", r.adminHandler.PromoteUser)
			admin.POST("/users/:userId/demote", r.adminHandler.DemoteUser)

			// Audit logs
			admin.GET("/audit-logs", r.adminHandler.QueryAuditLogs)
		}
	}
}
