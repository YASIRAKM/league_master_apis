package handlers

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/yourname/leaguemaster/internal/models"
	"github.com/yourname/leaguemaster/internal/services"
	"github.com/yourname/leaguemaster/pkg/database"
)

type CaptainHandler struct{}

func NewCaptainHandler() *CaptainHandler {
	return &CaptainHandler{}
}

// Helper to get User ID from context
func getUserID(c echo.Context) uint {
	user := c.Get("user").(*services.JWTClaims)
	return user.UserID
}

// GET /my-team
func (h *CaptainHandler) GetMyTeam(c echo.Context) error {
	userID := getUserID(c)
	var user models.User
	if err := database.GetDB().First(&user, userID).Error; err != nil {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "User not found"})
	}

	if user.TeamID == nil {
		return c.JSON(http.StatusNotFound, echo.Map{"error": "No team assigned to this captain"})
	}

	var team models.Team
	if err := database.GetDB().Preload("Players").First(&team, *user.TeamID).Error; err != nil {
		return c.JSON(http.StatusNotFound, echo.Map{"error": "Team not found"})
	}

	return c.JSON(http.StatusOK, team)
}

// PUT /my-team
type UpdateTeamRequest struct {
	Name    string `json:"name" form:"name"`
	LogoURL string `json:"logo_url" form:"logo_url"`
}

func (h *CaptainHandler) UpdateTeam(c echo.Context) error {
	userID := getUserID(c)
	var user models.User
	database.GetDB().First(&user, userID)
	if user.TeamID == nil {
		return c.JSON(http.StatusNotFound, echo.Map{"error": "No team assigned"})
	}

	req := new(UpdateTeamRequest)
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid request"})
	}

	var team models.Team
	if err := database.GetDB().First(&team, *user.TeamID).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Database error"})
	}

	if req.Name != "" {
		team.Name = req.Name
	}
	if req.LogoURL != "" {
		team.LogoURL = req.LogoURL
	}

	database.GetDB().Save(&team)
	return c.JSON(http.StatusOK, team)
}

// POST /my-team/players
func (h *CaptainHandler) AddPlayer(c echo.Context) error {
	userID := getUserID(c)
	var user models.User
	database.GetDB().First(&user, userID)
	if user.TeamID == nil {
		return c.JSON(http.StatusNotFound, echo.Map{"error": "No team assigned"})
	}

	var player models.Player
	if err := c.Bind(&player); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid player data"})
	}

	player.TeamID = *user.TeamID // Force assignment to captain's team

	if err := database.GetDB().Create(&player).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to add player"})
	}

	return c.JSON(http.StatusCreated, player)
}

// DELETE /my-team/players/:id
func (h *CaptainHandler) RemovePlayer(c echo.Context) error {
	userID := getUserID(c)
	var user models.User
	database.GetDB().First(&user, userID)
	if user.TeamID == nil {
		return c.JSON(http.StatusNotFound, echo.Map{"error": "No team assigned"})
	}

	playerID, _ := strconv.Atoi(c.Param("id"))

	// Ensure player belongs to captain's team
	var player models.Player
	if err := database.GetDB().Where("id = ? AND team_id = ?", playerID, *user.TeamID).First(&player).Error; err != nil {
		return c.JSON(http.StatusNotFound, echo.Map{"error": "Player not found in your team"})
	}

	database.GetDB().Delete(&player)
	return c.JSON(http.StatusOK, echo.Map{"message": "Player removed"})
}

// POST /matches/:id/events
// POST /matches/:id/events
func (h *CaptainHandler) AddMatchEvent(c echo.Context) error {
	matchID, _ := strconv.Atoi(c.Param("id"))
	userID := getUserID(c)

	// Get Captain's Team
	var user models.User
	if err := database.GetDB().First(&user, userID).Error; err != nil {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "User not found"})
	}
	if user.TeamID == nil {
		return c.JSON(http.StatusForbidden, echo.Map{"error": "No team assigned"})
	}
	captainTeamID := *user.TeamID

	// Verify Match exists and Captain's Team is playing
	var match models.Match
	if err := database.GetDB().First(&match, matchID).Error; err != nil {
		return c.JSON(http.StatusNotFound, echo.Map{"error": "Match not found"})
	}

	if (match.TeamAID == nil || *match.TeamAID != captainTeamID) && (match.TeamBID == nil || *match.TeamBID != captainTeamID) {
		return c.JSON(http.StatusForbidden, echo.Map{"error": "Your team is not participating in this match"})
	}

	var event models.MatchEvent
	if err := c.Bind(&event); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid event data"})
	}
	event.MatchID = uint(matchID)

	// Validate Player belongs to one of the teams in the match
	var player models.Player
	if err := database.GetDB().First(&player, event.PlayerID).Error; err != nil {
		return c.JSON(http.StatusNotFound, echo.Map{"error": "Player not found"})
	}

	// Check if player's team is in the match
	isTeamA := match.TeamAID != nil && player.TeamID == *match.TeamAID
	isTeamB := match.TeamBID != nil && player.TeamID == *match.TeamBID

	if !isTeamA && !isTeamB {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Player does not belong to any team in this match"})
	}

	if err := database.GetDB().Create(&event).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to create event"})
	}

	return c.JSON(http.StatusCreated, event)
}
