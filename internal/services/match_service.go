package services

import (
	"errors"

	"github.com/yourname/leaguemaster/internal/models"
	"github.com/yourname/leaguemaster/pkg/database"
	"gorm.io/gorm"
)

type MatchService struct {
	db *gorm.DB
}

func NewMatchService() *MatchService {
	return &MatchService{
		db: database.GetDB(),
	}
}

// AddMatchEvent handles the transactional logic for adding goals/cards
func (s *MatchService) AddMatchEvent(matchID, playerID uint, eventType string, minute int) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 1. Verify Match
		var match models.Match
		if err := tx.First(&match, matchID).Error; err != nil {
			return err
		}

		if match.Status == "completed" {
			return errors.New("cannot add events to completed match")
		}

		// 2. Verify Player belongs to one of the teams
		var player models.Player
		if err := tx.First(&player, playerID).Error; err != nil {
			return err
		}

		if player.TeamID != *match.TeamAID && player.TeamID != *match.TeamBID {
			return errors.New("player not playing in this match")
		}

		// 3. Create Event
		event := models.MatchEvent{
			MatchID:   matchID,
			PlayerID:  playerID,
			EventType: eventType,
			Minute:    minute,
		}
		if err := tx.Create(&event).Error; err != nil {
			return err
		}

		// 4. Update Score if Goal
		if eventType == "goal" {
			if player.TeamID == *match.TeamAID {
				match.ScoreA++
			} else {
				match.ScoreB++
			}
			if err := tx.Save(&match).Error; err != nil {
				return err
			}

			// Update Player Goals
			player.GoalsScored++
			if err := tx.Save(&player).Error; err != nil {
				return err
			}
		} else if eventType == "card_red" {
			player.RedCards++
			if err := tx.Save(&player).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

// ResolveMatch allows admins to force score and advance winner
func (s *MatchService) ResolveMatch(matchID uint, scoreA, scoreB int) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var match models.Match
		if err := tx.First(&match, matchID).Error; err != nil {
			return err
		}

		match.ScoreA = scoreA
		match.ScoreB = scoreB
		match.Status = "completed"

		if err := tx.Save(&match).Error; err != nil {
			return err
		}

		// Logic to advance winner to next match if applicable
		if match.NextMatchID != nil {
			var nextMatch models.Match
			if err := tx.First(&nextMatch, *match.NextMatchID).Error; err != nil {
				return err
			}

			// Determine winner
			var winnerID *uint
			if scoreA > scoreB {
				winnerID = match.TeamAID
			} else if scoreB > scoreA {
				winnerID = match.TeamBID
			}
			// If draw, handle via penalties or assumed logic?
			// Spec doesn't specify penalties, assume simple win/loss or draw allowed until bracket logic kicks in.
			// Bracket usually implies knockout, so a winner is needed.
			// For now, if draw, we don't advance anyone automatically or specific rules.

			if winnerID != nil {
				// logic to decide if it goes to TeamA or TeamB slot in next match
				// This implies the next match structure needs to know which slot this match feeds into.
				// Simple approach: NextMatch has TeamAID and TeamBID empty initially.
				// We need to know if this match is the 'top' or 'bottom' feeder.
				// Simpler: Just check which slot is empty or use MatchNumber.
				// For this simplified logic, we'll try to fill TeamA first, then TeamB.

				if nextMatch.TeamAID == nil {
					nextMatch.TeamAID = winnerID
				} else if nextMatch.TeamBID == nil {
					nextMatch.TeamBID = winnerID
				}
				if err := tx.Save(&nextMatch).Error; err != nil {
					return err
				}
			}
		}

		return nil
	})
}
