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

// User CRUD

// GET /admin/users
func (h *AdminHandler) GetAllUsers(c echo.Context) error {
	var users []models.User
	if err := database.GetDB().Find(&users).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to fetch users"})
	}
	return c.JSON(http.StatusOK, users)
}

// GET /admin/users/:id
func (h *AdminHandler) GetUser(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	var user models.User
	if err := database.GetDB().First(&user, id).Error; err != nil {
		return c.JSON(http.StatusNotFound, echo.Map{"error": "User not found"})
	}
	return c.JSON(http.StatusOK, user)
}

// PUT /admin/users/:id
func (h *AdminHandler) UpdateUser(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	var user models.User
	if err := database.GetDB().First(&user, id).Error; err != nil {
		return c.JSON(http.StatusNotFound, echo.Map{"error": "User not found"})
	}

	type UpdateUserRequest struct {
		Role     string `json:"role"`
		IsActive *bool  `json:"is_active"`
	}
	req := new(UpdateUserRequest)
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid input"})
	}

	if req.Role != "" {
		user.Role = req.Role
	}
	if req.IsActive != nil {
		user.IsActive = *req.IsActive
	}

	database.GetDB().Save(&user)
	return c.JSON(http.StatusOK, user)
}

// DELETE /admin/users/:id
func (h *AdminHandler) DeleteUser(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := database.GetDB().Delete(&models.User{}, id).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to delete user"})
	}
	return c.JSON(http.StatusOK, echo.Map{"message": "User deleted"})
}

// Team CRUD

// GET /admin/teams
func (h *AdminHandler) GetAllTeams(c echo.Context) error {
	var teams []models.Team
	if err := database.GetDB().Preload("Players").Find(&teams).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to fetch teams"})
	}
	return c.JSON(http.StatusOK, teams)
}

// POST /admin/teams
func (h *AdminHandler) CreateTeam(c echo.Context) error {
	var team models.Team
	if err := c.Bind(&team); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid input"})
	}
	if err := database.GetDB().Create(&team).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to create team"})
	}
	return c.JSON(http.StatusCreated, team)
}

// PUT /admin/teams/:id
func (h *AdminHandler) UpdateTeam(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	var team models.Team
	if err := database.GetDB().First(&team, id).Error; err != nil {
		return c.JSON(http.StatusNotFound, echo.Map{"error": "Team not found"})
	}

	if err := c.Bind(&team); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid input"})
	}

	database.GetDB().Save(&team)
	return c.JSON(http.StatusOK, team)
}

// DELETE /admin/teams/:id
func (h *AdminHandler) DeleteTeam(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := database.GetDB().Delete(&models.Team{}, id).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to delete team"})
	}
	return c.JSON(http.StatusOK, echo.Map{"message": "Team deleted"})
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

// Tournament CRUD Extensions

// PUT /admin/tournaments/:id
func (h *AdminHandler) UpdateTournament(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	var tournament models.Tournament
	if err := database.GetDB().First(&tournament, id).Error; err != nil {
		return c.JSON(http.StatusNotFound, echo.Map{"error": "Tournament not found"})
	}

	if err := c.Bind(&tournament); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid input"})
	}

	database.GetDB().Save(&tournament)
	return c.JSON(http.StatusOK, tournament)
}

// DELETE /admin/tournaments/:id
func (h *AdminHandler) DeleteTournament(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := database.GetDB().Delete(&models.Tournament{}, id).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to delete tournament"})
	}
	return c.JSON(http.StatusOK, echo.Map{"message": "Tournament deleted"})
}

// DELETE /admin/tournaments/:id/teams/:team_id
func (h *AdminHandler) RemoveTeamFromTournament(c echo.Context) error {
	tournamentID, _ := strconv.Atoi(c.Param("id"))
	teamID, _ := strconv.Atoi(c.Param("team_id"))

	if err := database.GetDB().Where("tournament_id = ? AND team_id = ?", tournamentID, teamID).Delete(&models.Standing{}).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to remove team from tournament"})
	}
	return c.JSON(http.StatusOK, echo.Map{"message": "Team removed from tournament"})
}

// Player CRUD (Admin)

// GET /admin/players
func (h *AdminHandler) GetAllPlayers(c echo.Context) error {
	var players []models.Player
	if err := database.GetDB().Find(&players).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to fetch players"})
	}
	return c.JSON(http.StatusOK, players)
}

// POST /admin/players
func (h *AdminHandler) CreatePlayer(c echo.Context) error {
	var player models.Player
	if err := c.Bind(&player); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid input"})
	}
	if err := database.GetDB().Create(&player).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to create player"})
	}
	return c.JSON(http.StatusCreated, player)
}

// GET /admin/players/:id
func (h *AdminHandler) GetPlayer(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	var player models.Player
	if err := database.GetDB().First(&player, id).Error; err != nil {
		return c.JSON(http.StatusNotFound, echo.Map{"error": "Player not found"})
	}
	return c.JSON(http.StatusOK, player)
}

// PUT /admin/players/:id
func (h *AdminHandler) UpdatePlayer(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	var player models.Player
	if err := database.GetDB().First(&player, id).Error; err != nil {
		return c.JSON(http.StatusNotFound, echo.Map{"error": "Player not found"})
	}

	if err := c.Bind(&player); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid input"})
	}

	database.GetDB().Save(&player)
	return c.JSON(http.StatusOK, player)
}

// DELETE /admin/players/:id
func (h *AdminHandler) DeletePlayer(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := database.GetDB().Delete(&models.Player{}, id).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to delete player"})
	}
	return c.JSON(http.StatusOK, echo.Map{"message": "Player deleted"})
}

// Staff CRUD

// GET /admin/staff
func (h *AdminHandler) GetAllStaff(c echo.Context) error {
	var staff []models.Staff
	if err := database.GetDB().Find(&staff).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to fetch staff"})
	}
	return c.JSON(http.StatusOK, staff)
}

// POST /admin/staff
func (h *AdminHandler) CreateStaff(c echo.Context) error {
	var staff models.Staff
	if err := c.Bind(&staff); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid input"})
	}
	if err := database.GetDB().Create(&staff).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to create staff"})
	}
	return c.JSON(http.StatusCreated, staff)
}

// GET /admin/staff/:id
func (h *AdminHandler) GetStaff(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	var staff models.Staff
	if err := database.GetDB().First(&staff, id).Error; err != nil {
		return c.JSON(http.StatusNotFound, echo.Map{"error": "Staff not found"})
	}
	return c.JSON(http.StatusOK, staff)
}

// PUT /admin/staff/:id
func (h *AdminHandler) UpdateStaff(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	var staff models.Staff
	if err := database.GetDB().First(&staff, id).Error; err != nil {
		return c.JSON(http.StatusNotFound, echo.Map{"error": "Staff not found"})
	}
	if err := c.Bind(&staff); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid input"})
	}
	database.GetDB().Save(&staff)
	return c.JSON(http.StatusOK, staff)
}

// DELETE /admin/staff/:id
func (h *AdminHandler) DeleteStaff(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := database.GetDB().Delete(&models.Staff{}, id).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to delete staff"})
	}
	return c.JSON(http.StatusOK, echo.Map{"message": "Staff deleted"})
}

// Captains List
func (h *AdminHandler) GetAllCaptains(c echo.Context) error {
	var captains []models.User
	if err := database.GetDB().Where("role = ?", "captain").Find(&captains).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to fetch captains"})
	}
	return c.JSON(http.StatusOK, captains)
}

// Notifications

// POST /admin/notifications
func (h *AdminHandler) SendNotification(c echo.Context) error {
	type NotificationRequest struct {
		UserID  uint   `json:"user_id" form:"user_id"`
		Message string `json:"message" form:"message"`
	}
	req := new(NotificationRequest)
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid input"})
	}

	notification := models.Notification{
		UserID:  req.UserID,
		Message: req.Message,
		IsRead:  false,
	}

	if err := database.GetDB().Create(&notification).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to send notification"})
	}

	return c.JSON(http.StatusCreated, notification)
}
