// internal/handlers/contact_handler.go
package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/MuhammadZaidan1/jays-portfolio-api/internal/config"
	"github.com/MuhammadZaidan1/jays-portfolio-api/internal/models"
)

// POST /api/contact — public
func SendMessage(c *gin.Context) {
	var input struct {
		Name    string `json:"name"    binding:"required"`
		Email   string `json:"email"   binding:"required,email"`
		Message string `json:"message" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": err.Error()})
		return
	}

	msg := models.ContactMessage{
		ID:      uuid.New().String(),
		Name:    input.Name,
		Email:   input.Email,
		Message: input.Message,
		IsRead:  false,
		SentAt:  time.Now(),
	}
	config.DB.Create(&msg)
	c.JSON(http.StatusCreated, gin.H{"status": "success", "data": msg})
}

// GET /api/admin/messages — admin
func GetMessages(c *gin.Context) {
	var msgs []models.ContactMessage
	config.DB.Order("sent_at DESC").Find(&msgs)
	c.JSON(http.StatusOK, gin.H{"status": "success", "data": msgs})
}

// PATCH /api/admin/messages/:id/read — admin
func MarkMessageRead(c *gin.Context) {
	config.DB.Model(&models.ContactMessage{}).
		Where("id = ?", c.Param("id")).
		Update("is_read", true)
	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

// DELETE /api/admin/messages/:id — admin
func DeleteMessage(c *gin.Context) {
	config.DB.Delete(&models.ContactMessage{}, "id = ?", c.Param("id"))
	c.JSON(http.StatusOK, gin.H{"status": "success"})
}