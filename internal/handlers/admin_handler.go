package handlers

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/yourname/leaguemaster/internal/models"
	"github.com/yourname/leaguemaster/pkg/database"
)

type AdminHandler struct{}

func NewAdminHandler() *AdminHandler {
	return &AdminHandler{}
}

// POST /tournaments/:id/teams
func (h *AdminHandler) AddTeamToTournament(c echo.Context) error {
	tournamentID, _ := strconv.Atoi(c.Param("id"))

	type AddTeamRequest struct {
		TeamID uint `json:"team_id" form:"team_id"`
	}
	req := new(AddTeamRequest)
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid input"})
	}

	// Verify Tournament exists
	var tournament models.Tournament
	if err := database.GetDB().First(&tournament, tournamentID).Error; err != nil {
		return c.JSON(http.StatusNotFound, echo.Map{"error": "Tournament not found"})
	}

	// Verify Team exists
	var team models.Team
	if err := database.GetDB().First(&team, req.TeamID).Error; err != nil {
		return c.JSON(http.StatusNotFound, echo.Map{"error": "Team not found"})
	}

	// Check if already registered
	var count int64
	database.GetDB().Model(&models.Standing{}).Where("tournament_id = ? AND team_id = ?", tournamentID, req.TeamID).Count(&count)
	if count > 0 {
		return c.JSON(http.StatusConflict, echo.Map{"error": "Team already in tournament"})
	}

	// Create Standing entry to register team
	standing := models.Standing{
		TournamentID: uint(tournamentID),
		TeamID:       req.TeamID,
		Points:       0,
	}

	if err := database.GetDB().Create(&standing).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to add team to tournament"})
	}

	return c.JSON(http.StatusCreated, echo.Map{"message": "Team added to tournament"})
}

// POST /tournaments
func (h *AdminHandler) CreateTournament(c echo.Context) error {
	var tournament models.Tournament
	if err := c.Bind(&tournament); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid input"})
	}

	if err := database.GetDB().Create(&tournament).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to create tournament"})
	}

	return c.JSON(http.StatusCreated, tournament)
}

// POST /tournaments/:id/generate
// POST /tournaments/:id/generate
func (h *AdminHandler) GenerateBracket(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))

	// Check if tournament exists
	var tournament models.Tournament
	if err := database.GetDB().First(&tournament, id).Error; err != nil {
		return c.JSON(http.StatusNotFound, echo.Map{"error": "Tournament not found"})
	}

	// Fetch confirmed teams (via Standings, assuming registration creates a Standing entry)
	var standings []models.Standing
	if err := database.GetDB().Where("tournament_id = ?", id).Find(&standings).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to fetch teams"})
	}

	if len(standings) < 2 {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Not enough teams to generate bracket (need at least 2)"})
	}

	// Simple pairing logic (1 vs 2, 3 vs 4, etc.)
	// In a real app, you might shuffle or seed them.
	matchCount := 0
	for i := 0; i < len(standings)-1; i += 2 {
		teamA := standings[i].TeamID
		teamB := standings[i+1].TeamID

		match := models.Match{
			TournamentID: tournament.ID,
			TeamAID:      &teamA,
			TeamBID:      &teamB,
			MatchNumber:  matchCount + 1,
			Round:        1,
			Status:       "scheduled",
		}

		if err := database.GetDB().Create(&match).Error; err != nil {
			return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to create match"})
		}
		matchCount++
	}

	// Update tournament status
	tournament.Status = "active"
	database.GetDB().Save(&tournament)

	return c.JSON(http.StatusOK, echo.Map{
		"message":         "Bracket generated and tournament started",
		"matches_created": matchCount,
	})
}

// POST /matches/:id/resolve
func (h *AdminHandler) ResolveMatch(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))

	type ResolveRequest struct {
		Status string `json:"status" form:"status"` // e.g., "completed"
		ScoreA int    `json:"score_a" form:"score_a"`
		ScoreB int    `json:"score_b" form:"score_b"`
	}

	req := new(ResolveRequest)
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid request"})
	}

	var match models.Match
	if err := database.GetDB().First(&match, id).Error; err != nil {
		return c.JSON(http.StatusNotFound, echo.Map{"error": "Match not found"})
	}

	match.Status = req.Status
	match.ScoreA = req.ScoreA
	match.ScoreB = req.ScoreB

	if err := database.GetDB().Save(&match).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to update match"})
	}

	return c.JSON(http.StatusOK, match)
}

// GET /dashboard/stats
func (h *AdminHandler) GetDashboardStats(c echo.Context) error {
	var totalUsers int64
	var totalMatches int64
	var totalTournaments int64

	database.GetDB().Model(&models.User{}).Count(&totalUsers)
	database.GetDB().Model(&models.Match{}).Count(&totalMatches)
	database.GetDB().Model(&models.Tournament{}).Count(&totalTournaments)

	return c.JSON(http.StatusOK, echo.Map{
		"total_users":       totalUsers,
		"total_matches":     totalMatches,
		"total_tournaments": totalTournaments,
	})
}
func (h *AdminHandler) BanUser(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))

	// Optional: Parse body to support unban? Or toggle?
	// For simplicity, let's assume it sets Banned=true (or reads from body if complex)
	// Let's implement toggle or specific instructions? "ban player or captain"
	// I'll make it explicit: POST /ban sets to true. If they want unban, maybe specific route or body.

	// Parse optional body for state
	type BanRequest struct {
		IsBanned *bool `json:"is_banned" form:"is_banned"`
	}
	req := new(BanRequest)
	c.Bind(req) // If fails, we default to banning

	var user models.User
	if err := database.GetDB().First(&user, id).Error; err != nil {
		return c.JSON(http.StatusNotFound, echo.Map{"error": "User not found"})
	}

	if req.IsBanned != nil {
		user.IsBanned = *req.IsBanned
	} else {
		user.IsBanned = true // Default action
	}

	database.GetDB().Save(&user)
	return c.JSON(http.StatusOK, echo.Map{"message": "User ban status updated", "is_banned": user.IsBanned})
}

// POST /players/:id/ban
func (h *AdminHandler) BanPlayer(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))

	type BanRequest struct {
		IsBanned *bool `json:"is_banned" form:"is_banned"`
	}
	req := new(BanRequest)
	c.Bind(req)

	var player models.Player
	if err := database.GetDB().First(&player, id).Error; err != nil {
		return c.JSON(http.StatusNotFound, echo.Map{"error": "Player not found"})
	}

	if req.IsBanned != nil {
		player.IsBanned = *req.IsBanned
	} else {
		player.IsBanned = true
	}

	database.GetDB().Save(&player)
	return c.JSON(http.StatusOK, echo.Map{"message": "Player ban status updated", "is_banned": player.IsBanned})
}
