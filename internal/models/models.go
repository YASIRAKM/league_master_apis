package models

import (
	"time"
)

type User struct {
	ID        uint      `gorm:"primaryKey" json:"id" form:"id"`
	Username  string    `gorm:"unique;not null" json:"username" form:"username"`
	Password  string    `gorm:"not null" json:"-" form:"password"` // Hashed, but used for binding in login/register? No, specific structs used there.
	Role      string    `gorm:"type:enum('admin','captain');not null" json:"role" form:"role"`
	IsActive  bool      `gorm:"default:true" json:"is_active" form:"is_active"`
	IsBanned  bool      `gorm:"default:false" json:"is_banned" form:"is_banned"`
	TeamID    *uint     `json:"team_id,omitempty" form:"team_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Team struct {
	ID        uint      `gorm:"primaryKey" json:"id" form:"id"`
	Name      string    `gorm:"unique;not null" json:"name" form:"name"`
	LogoURL   string    `json:"logo_url" form:"logo_url"`
	CaptainID uint      `gorm:"unique" json:"captain_id" form:"captain_id"`
	Players   []Player  `gorm:"foreignKey:TeamID" json:"players,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Player struct {
	ID           uint      `gorm:"primaryKey" json:"id" form:"id"`
	TeamID       uint      `gorm:"not null;index" json:"team_id" form:"team_id"`
	Name         string    `gorm:"not null" json:"name" form:"name"`
	JerseyNumber int       `json:"jersey_number" form:"jersey_number"`
	GoalsScored  int       `gorm:"default:0" json:"goals_scored" form:"goals_scored"`
	RedCards     int       `gorm:"default:0" json:"red_cards" form:"red_cards"`
	IsBanned     bool      `gorm:"default:false" json:"is_banned" form:"is_banned"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type Tournament struct {
	ID        uint      `gorm:"primaryKey" json:"id" form:"id"`
	Name      string    `gorm:"not null" json:"name" form:"name"`
	Status    string    `gorm:"type:enum('registration','active','completed');default:'registration'" json:"status" form:"status"`
	MaxTeams  int       `gorm:"default:16" json:"max_teams" form:"max_teams"`
	Matches   []Match   `gorm:"foreignKey:TournamentID" json:"matches,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Match struct {
	ID           uint   `gorm:"primaryKey" json:"id" form:"id"`
	TournamentID uint   `gorm:"not null;index" json:"tournament_id" form:"tournament_id"`
	Round        int    `json:"round" form:"round"`
	MatchNumber  int    `json:"match_number" form:"match_number"`
	TeamAID      *uint  `json:"team_a_id" form:"team_a_id"`
	TeamBID      *uint  `json:"team_b_id" form:"team_b_id"`
	ScoreA       int    `gorm:"default:0" json:"score_a" form:"score_a"`
	ScoreB       int    `gorm:"default:0" json:"score_b" form:"score_b"`
	Status       string `gorm:"type:enum('scheduled','pending_verification','disputed','completed');default:'scheduled'" json:"status" form:"status"`
	NextMatchID  *uint  `json:"next_match_id" form:"next_match_id"`

	// Relationships
	TeamA       *Team        `gorm:"foreignKey:TeamAID" json:"team_a,omitempty"`
	TeamB       *Team        `gorm:"foreignKey:TeamBID" json:"team_b,omitempty"`
	MatchEvents []MatchEvent `gorm:"foreignKey:MatchID" json:"match_events,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type MatchEvent struct {
	ID        uint      `gorm:"primaryKey" json:"id" form:"id"`
	MatchID   uint      `gorm:"not null;index" json:"match_id" form:"match_id"`
	PlayerID  uint      `gorm:"not null" json:"player_id" form:"player_id"`
	EventType string    `gorm:"type:enum('goal','card_yellow','card_red');not null" json:"event_type" form:"event_type"`
	Minute    int       `json:"minute" form:"minute"`
	CreatedAt time.Time `json:"created_at"`
}

type Standing struct {
	TournamentID uint `gorm:"primaryKey" json:"tournament_id"`
	TeamID       uint `gorm:"primaryKey" json:"team_id"`
	Points       int  `gorm:"default:0" json:"points"`
	Wins         int  `gorm:"default:0" json:"wins"`
	Losses       int  `gorm:"default:0" json:"losses"`
	Draws        int  `gorm:"default:0" json:"draws"`
	GoalsFor     int  `gorm:"default:0" json:"goals_for"` // Goals Scored
	GoalsAgainst int  `gorm:"default:0" json:"goals_against"`

	Team Team `gorm:"foreignKey:TeamID" json:"team_name,omitempty"`
}

type Notification struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"not null;index" json:"user_id"`
	Message   string    `gorm:"not null" json:"message"`
	IsRead    bool      `gorm:"default:false" json:"is_read"`
	CreatedAt time.Time `json:"created_at"`
}
