package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/MuhammadZaidan1/jays-portfolio-api/internal/config"
	"github.com/MuhammadZaidan1/jays-portfolio-api/internal/models"
	"github.com/MuhammadZaidan1/jays-portfolio-api/internal/utils"
)

// LoginRequest adalah struktur input dari frontend
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func Login(c *gin.Context) {
	var input LoginRequest
	var admin models.Admin

	// 1. Bind JSON dari body request
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username dan password wajib diisi"})
		return
	}

	// 2. Cari Admin berdasarkan username
	err := config.DB.Where("username = ?", input.Username).First(&admin).Error
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Username atau password salah"})
		return
	}

	// 3. Verifikasi Password menggunakan Bcrypt
	if !utils.CheckPasswordHash(input.Password, admin.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Username atau password salah"})
		return
	}

	// 4. Generate Token JWT v5
	token, err := utils.GenerateToken(admin.ID, admin.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuat token akses"})
		return
	}

	// 5. Kirim Token ke Frontend
	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"token":  token,
	})
}