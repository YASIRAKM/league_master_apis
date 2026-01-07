package handlers

import (
	"net/http"
	"strconv"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"github.com/yourname/leaguemaster/internal/models"
	"github.com/yourname/leaguemaster/pkg/database"
)

type NotificationHandler struct{}

func NewNotificationHandler() *NotificationHandler {
	return &NotificationHandler{}
}

// GET /mobile/notifications
func (h *NotificationHandler) GetMyNotifications(c echo.Context) error {
	user := c.Get("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	userID := uint(claims["user_id"].(float64))

	var notifications []models.Notification
	if err := database.GetDB().Where("user_id = ?", userID).Order("created_at desc").Find(&notifications).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to fetch notifications"})
	}

	return c.JSON(http.StatusOK, notifications)
}

// POST /mobile/notifications/:id/read
func (h *NotificationHandler) MarkNotificationRead(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	user := c.Get("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	userID := uint(claims["user_id"].(float64))

	var notification models.Notification
	if err := database.GetDB().Where("id = ? AND user_id = ?", id, userID).First(&notification).Error; err != nil {
		return c.JSON(http.StatusNotFound, echo.Map{"error": "Notification not found"})
	}

	notification.IsRead = true
	database.GetDB().Save(&notification)

	return c.JSON(http.StatusOK, echo.Map{"message": "Notification marked as read"})
}
