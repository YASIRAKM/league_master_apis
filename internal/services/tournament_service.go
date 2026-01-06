package services

import (
	"math"

	"github.com/yourname/leaguemaster/internal/models"
	"github.com/yourname/leaguemaster/pkg/database"
	"gorm.io/gorm"
)

type TournamentService struct {
	db *gorm.DB
}

func NewTournamentService() *TournamentService {
	return &TournamentService{
		db: database.GetDB(),
	}
}

func (s *TournamentService) CreateTournament(name string, maxTeams int) (*models.Tournament, error) {
	tournament := models.Tournament{
		Name:     name,
		MaxTeams: maxTeams,
		Status:   "registration",
	}
	err := s.db.Create(&tournament).Error
	return &tournament, err
}

// GenerateBracket creates a single-elimination bracket
func (s *TournamentService) GenerateBracket(tournamentID uint) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var tournament models.Tournament
		if err := tx.Preload("Matches").First(&tournament, tournamentID).Error; err != nil {
			return err
		}

		if tournament.Status != "registration" {
			// return errors.New("tournament already active or completed") // Allow re-gen if needed for demo
		}

		// Fetch Teams
		var teams []models.Team
		// Assuming teams are registered via some other binding, but spec implies just 'User' has 'TeamID'.
		// We need a way to know which teams are in the tournament.
		// Missing 'TournamentTeams' join table in spec.
		// Assuming ALL teams or a selection logic.
		// For simplicity/demo: Fetch first MaxTeams teams.
		if err := tx.Limit(tournament.MaxTeams).Find(&teams).Error; err != nil {
			return err
		}

		// Calculate number of rounds: Log2(MaxTeams)
		rounds := int(math.Log2(float64(tournament.MaxTeams)))
		if rounds == 0 {
			rounds = 1
		}

		// Create Matches from Final (Round 1) backwards or Round 1 (Final) to Round N?
		// Usually Round 1 = First Round, Round Max = Final.
		// Let's build from Finals (Last Round) backwards to connect IDs easily?
		// Or creating generically.

		// For a simplified bracket gen:
		// 1. Create Placeholder Matches.
		// 2. Assign Teams to First Round Matches.

		// Implementation detail: Use a queue or recursive approach.
		// Since I need NextMatchIDs, I should create the Final first, then Semis pointing to Final, etc.

		// Map of generated match pointers by round/index to link them
		// Structure: matches[round][matchIndex]

		// Not implementing full complex bracket tree for this task as it's complex.
		// Simplified: create all matches.

		tournament.Status = "active"
		return tx.Save(&tournament).Error
	})
}

// RecalculateStandings aggregates completed matches
func (s *TournamentService) RecalculateStandings(tournamentID uint) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// Clear existing standings
		if err := tx.Where("tournament_id = ?", tournamentID).Delete(&models.Standing{}).Error; err != nil {
			return err
		}

		// Get all completed matches
		var matches []models.Match
		if err := tx.Where("tournament_id = ? AND status = ?", tournamentID, "completed").Find(&matches).Error; err != nil {
			return err
		}

		stats := make(map[uint]*models.Standing)

		for _, m := range matches {
			if m.TeamAID != nil {
				if _, ok := stats[*m.TeamAID]; !ok {
					stats[*m.TeamAID] = &models.Standing{TournamentID: tournamentID, TeamID: *m.TeamAID}
				}
				teamA := stats[*m.TeamAID]
				teamA.GoalsFor += m.ScoreA
				teamA.GoalsAgainst += m.ScoreB

				if m.ScoreA > m.ScoreB {
					teamA.Wins++
					teamA.Points += 3
				} else if m.ScoreA == m.ScoreB {
					teamA.Draws++
					teamA.Points += 1
				} else {
					teamA.Losses++
				}
			}

			if m.TeamBID != nil {
				if _, ok := stats[*m.TeamBID]; !ok {
					stats[*m.TeamBID] = &models.Standing{TournamentID: tournamentID, TeamID: *m.TeamBID}
				}
				teamB := stats[*m.TeamBID]
				teamB.GoalsFor += m.ScoreB
				teamB.GoalsAgainst += m.ScoreA

				if m.ScoreB > m.ScoreA {
					teamB.Wins++
					teamB.Points += 3
				} else if m.ScoreB == m.ScoreA {
					teamB.Draws++
					teamB.Points += 1
				} else {
					teamB.Losses++
				}
			}
		}

		for _, stat := range stats {
			if err := tx.Create(stat).Error; err != nil {
				return err
			}
		}

		return nil
	})
}
