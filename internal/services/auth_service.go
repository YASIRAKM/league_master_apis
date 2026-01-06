package services

import (
	"errors"
	"log"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/yourname/leaguemaster/internal/models"
	"github.com/yourname/leaguemaster/pkg/database"
	"github.com/yourname/leaguemaster/pkg/utils"
	"gorm.io/gorm"
)

type JWTClaims struct {
	UserID uint   `json:"user_id"`
	Role   string `json:"role"`
	TeamID *uint  `json:"team_id,omitempty"`
	jwt.RegisteredClaims
}

type AuthService struct {
	db *gorm.DB
}

func NewAuthService() *AuthService {
	return &AuthService{
		db: database.GetDB(),
	}
}

func (s *AuthService) RegisterCaptain(username, password, teamName string) (*models.User, error) {
	// Transaction to create User and Team
	tx := s.db.Begin()

	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	user := models.User{
		Username: username,
		Password: hashedPassword,
		Role:     "captain",
		IsActive: true,
	}

	if err := tx.Create(&user).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	team := models.Team{
		Name:      teamName,
		CaptainID: user.ID,
	}

	if err := tx.Create(&team).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// Update user with TeamID
	user.TeamID = &team.ID
	if err := tx.Save(&user).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	tx.Commit()
	return &user, nil
}

func (s *AuthService) Login(username, password string) (string, *models.User, error) {
	var user models.User
	if err := s.db.Where("username = ?", username).First(&user).Error; err != nil {
		log.Printf("Login failed: User '%s' not found", username)
		return "", nil, errors.New("invalid credentials")
	}

	if !utils.CheckPasswordHash(password, user.Password) {
		log.Printf("Login failed: Password mismatch for user '%s'", username)
		return "", nil, errors.New("invalid credentials")
	}

	// Generate Token
	claims := JWTClaims{
		UserID: user.ID,
		Role:   user.Role,
		TeamID: user.TeamID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(72 * time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		return "", nil, err
	}

	return tokenString, &user, nil
}

func (s *AuthService) CreateAdmin(username, password string) error {
	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		return err
	}

	var user models.User
	result := s.db.Where("role = ?", "admin").First(&user)

	if result.Error == gorm.ErrRecordNotFound {
		// Create
		admin := models.User{
			Username: username,
			Password: hashedPassword,
			Role:     "admin",
			IsActive: true,
		}
		if err := s.db.Create(&admin).Error; err != nil {
			return err
		}
		log.Printf("Admin user '%s' created successfully", username)
	} else if result.Error == nil {
		// Update password to ensure known state
		user.Password = hashedPassword
		user.Username = username // Ensure username matches (though role is unique admin? No, role is just enum)
		if err := s.db.Save(&user).Error; err != nil {
			return err
		}
		log.Printf("Admin user '%s' updated with default password", username)
	} else {
		return result.Error
	}

	return nil
}
