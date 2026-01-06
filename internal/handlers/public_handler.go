package handlers

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/yourname/leaguemaster/internal/models"
	"github.com/yourname/leaguemaster/pkg/database"
)

type PublicHandler struct{}

func NewPublicHandler() *PublicHandler {
	return &PublicHandler{}
}

// GET /tournaments
func (h *PublicHandler) GetTournaments(c echo.Context) error {
	var tournaments []models.Tournament
	if err := database.GetDB().Find(&tournaments).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to fetch tournaments"})
	}
	return c.JSON(http.StatusOK, tournaments)
}

// GET /tournaments/:id/matches
func (h *PublicHandler) GetTournamentMatches(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	var matches []models.Match
	// Preload teams to show names
	if err := database.GetDB().Preload("TeamA").Preload("TeamB").Where("tournament_id = ?", id).Find(&matches).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to fetch matches"})
	}
	return c.JSON(http.StatusOK, matches)
}

// GET /tournaments/:id/standings
func (h *PublicHandler) GetStandings(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	var standings []models.Standing
	if err := database.GetDB().Preload("Team").Where("tournament_id = ?", id).Order("points desc, goals_for desc").Find(&standings).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to fetch standings"})
	}
	return c.JSON(http.StatusOK, standings)
}

// GET /tournaments/:id/teams
func (h *PublicHandler) GetTournamentTeams(c echo.Context) error {
	// This might require a join if teams are not directly linked to tournament in a simple way in the model provided,
	// but usually we can find teams that have matches in this tournament or if there's a registration table.
	// Based on the models, there isn't a direct "TournamentTeams" link explicit in the structs other than Matches.
	// HOWEVER, usually there's a registration or we assume all teams in the system are available,
	// OR we query matches.
	// Let's assume for now we list all teams, or if the user implied a relation.
	// The prompt says "List all teams participating".
	// The current models don't have a specific join table for Tournament<->Team active registration,
	// but Standings are a good proxy for participation.
	id, _ := strconv.Atoi(c.Param("id"))
	var standings []models.Standing
	if err := database.GetDB().Where("tournament_id = ?", id).Preload("Team").Find(&standings).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to fetch participating teams"})
	}

	teams := make([]models.Team, len(standings))
	for i, s := range standings {
		teams[i] = s.Team
	}

	return c.JSON(http.StatusOK, teams)
}

// GET /teams/:id
func (h *PublicHandler) GetTeam(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	var team models.Team
	if err := database.GetDB().Preload("Players").First(&team, id).Error; err != nil {
		return c.JSON(http.StatusNotFound, echo.Map{"error": "Team not found"})
	}
	return c.JSON(http.StatusOK, team)
}

// GET /players/:id
func (h *PublicHandler) GetPlayer(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	var player models.Player
	if err := database.GetDB().First(&player, id).Error; err != nil {
		return c.JSON(http.StatusNotFound, echo.Map{"error": "Player not found"})
	}
	return c.JSON(http.StatusOK, player)
}
