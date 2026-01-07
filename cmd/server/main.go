package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
	"github.com/yourname/leaguemaster/internal/handlers"
	"github.com/yourname/leaguemaster/internal/middleware"
	models "github.com/yourname/leaguemaster/internal/models"
	"github.com/yourname/leaguemaster/internal/services"
	"github.com/yourname/leaguemaster/pkg/database"
)

func main() {
	// Load .env
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// Connect to Database
	database.Connect()

	// Auto Migrate
	db := database.GetDB()
	err := db.AutoMigrate(
		&models.User{},
		&models.Team{},
		&models.Player{},
		&models.Tournament{},
		&models.Match{},
		&models.MatchEvent{},
		&models.Standing{},
		&models.Notification{},
	)
	if err != nil {
		log.Fatal("Failed to migrate database: ", err)
	}

	// Initialize Echo
	e := echo.New()

	// Middleware
	e.Use(echoMiddleware.Logger())
	e.Use(echoMiddleware.Recover())
	e.Use(echoMiddleware.CORS())

	// Initialize Handlers
	authHandler := handlers.NewAuthHandler()

	// Seed Admin
	services.NewAuthService().CreateAdmin("admin", "admin123")

	publicHandler := handlers.NewPublicHandler()
	captainHandler := handlers.NewCaptainHandler()
	adminHandler := handlers.NewAdminHandler()
	notificationHandler := handlers.NewNotificationHandler()

	// Routes
	v1 := e.Group("/api/v1")

	// Auth (Open)
	auth := v1.Group("/auth")
	auth.POST("/register", authHandler.Register)
	auth.POST("/login", authHandler.Login)

	// Public Routes (Open to All)
	public := v1.Group("") // or just attach to v1 if no prefix desired, but spec says "Group A", usually implied under API root
	// Spec: GET /tournaments
	public.GET("/tournaments", publicHandler.GetTournaments)
	public.GET("/tournaments/:id/matches", publicHandler.GetTournamentMatches)
	public.GET("/tournaments/:id/standings", publicHandler.GetStandings)
	public.GET("/tournaments/:id/teams", publicHandler.GetTournamentTeams)
	public.GET("/teams/:id", publicHandler.GetTeam)
	public.GET("/players/:id", publicHandler.GetPlayer)

	// Mobile Routes (Protected: Captain Role)
	mobile := v1.Group("/mobile")
	mobile.Use(middleware.AuthMiddleware)
	mobile.Use(middleware.CaptainOnly)

	// Note: Spec says "/my-team", so it's /api/v1/my-team
	mobile.GET("/my-team", captainHandler.GetMyTeam)
	mobile.PUT("/my-team", captainHandler.UpdateTeam)
	mobile.POST("/my-team/players", captainHandler.AddPlayer)
	mobile.DELETE("/my-team/players/:id", captainHandler.RemovePlayer)
	mobile.POST("/matches/:id/events", captainHandler.AddMatchEvent)

	// Mobile Notification Routes
	mobile.GET("/notifications", notificationHandler.GetMyNotifications)
	mobile.POST("/notifications/:id/read", notificationHandler.MarkNotificationRead)

	// Admin Routes (Protected: Admin Role)
	admin := v1.Group("/admin")
	admin.Use(middleware.AuthMiddleware)
	admin.Use(middleware.AdminOnly)

	// Admin User CRUD
	admin.GET("/users", adminHandler.GetAllUsers)
	admin.GET("/users/:id", adminHandler.GetUser)
	admin.PUT("/users/:id", adminHandler.UpdateUser)
	admin.DELETE("/users/:id", adminHandler.DeleteUser)

	// Admin Team CRUD
	admin.GET("/teams", adminHandler.GetAllTeams)
	admin.POST("/teams", adminHandler.CreateTeam)
	admin.PUT("/teams/:id", adminHandler.UpdateTeam)
	admin.DELETE("/teams/:id", adminHandler.DeleteTeam)

	// Admin Tournament Extensions
	admin.POST("/tournaments", adminHandler.CreateTournament)
	admin.PUT("/tournaments/:id", adminHandler.UpdateTournament)
	admin.DELETE("/tournaments/:id", adminHandler.DeleteTournament)
	admin.POST("/tournaments/:id/teams", adminHandler.AddTeamToTournament)
	admin.DELETE("/tournaments/:id/teams/:team_id", adminHandler.RemoveTeamFromTournament)

	admin.POST("/tournaments/:id/generate", adminHandler.GenerateBracket)
	admin.POST("/matches/:id/resolve", adminHandler.ResolveMatch)
	admin.GET("/dashboard/stats", adminHandler.GetDashboardStats)
	admin.POST("/users/:id/ban", adminHandler.BanUser)
	admin.POST("/players/:id/ban", adminHandler.BanPlayer)

	// Start Server
	// Start Server
	port := os.Getenv("PORT")
	if port == "" {
		port = ":8080"
	}
	if len(port) > 0 && port[0] != ':' {
		port = ":" + port
	}
	e.Logger.Fatal(e.Start(port))
}
