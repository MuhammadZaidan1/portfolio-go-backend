package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/MuhammadZaidan1/jays-portfolio-api/internal/config"
	"github.com/MuhammadZaidan1/jays-portfolio-api/internal/models"
)

// GET /api/experiences — public
func GetExperiences(c *gin.Context) {
	var experiences []models.Experience
	err := config.DB.Preload("Points").Order("start_date desc").Find(&experiences).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menarik data pengalaman: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "data": experiences})
}

// POST /api/admin/experiences
func CreateExperience(c *gin.Context) {
	var input struct {
		Title     string     `json:"title"      binding:"required"`
		Location  string     `json:"location"   binding:"required"`
		StartDate string     `json:"start_date" binding:"required"`
		EndDate   *string    `json:"end_date"`   // nullable
		Points    []struct {
			Content string `json:"content"`
		} `json:"points"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Parse dates
	startDate, err := time.Parse("2006-01-02", input.StartDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start_date format, use YYYY-MM-DD"})
		return
	}

	var endDate *time.Time
	if input.EndDate != nil && *input.EndDate != "" {
		t, err := time.Parse("2006-01-02", *input.EndDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end_date format, use YYYY-MM-DD"})
			return
		}
		endDate = &t
	}

	// Build points
	points := make([]models.ExperiencePoint, 0, len(input.Points))
	for _, p := range input.Points {
		if p.Content == "" {
			continue
		}
		points = append(points, models.ExperiencePoint{
			ID:      uuid.New().String(),
			Content: p.Content,
		})
	}

	exp := models.Experience{
		ID:        uuid.New().String(),
		Title:     input.Title,
		Location:  input.Location,
		StartDate: startDate,
		EndDate:   endDate,
		Points:    points,
	}

	if err := config.DB.Create(&exp).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan pengalaman: " + err.Error()})
		return
	}

	// Reload dengan preload agar response lengkap
	config.DB.Preload("Points").First(&exp, "id = ?", exp.ID)
	c.JSON(http.StatusCreated, gin.H{"status": "success", "data": exp})
}

// PUT /api/admin/experiences/:id
func UpdateExperience(c *gin.Context) {
	id := c.Param("id")

	// Pastikan data ada
	var exp models.Experience
	if err := config.DB.First(&exp, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Data pengalaman tidak ditemukan"})
		return
	}

	var input struct {
		Title     string  `json:"title"      binding:"required"`
		Location  string  `json:"location"   binding:"required"`
		StartDate string  `json:"start_date" binding:"required"`
		EndDate   *string `json:"end_date"`  // nullable
		Points    []struct {
			Content string `json:"content"`
		} `json:"points"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Parse dates
	startDate, err := time.Parse("2006-01-02", input.StartDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start_date format, use YYYY-MM-DD"})
		return
	}

	var endDate *time.Time
	if input.EndDate != nil && *input.EndDate != "" {
		t, err := time.Parse("2006-01-02", *input.EndDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end_date format, use YYYY-MM-DD"})
			return
		}
		endDate = &t
	}

	// Jalankan dalam transaksi agar atomic
	err = config.DB.Transaction(func(tx *gorm.DB) error {
		// 1. Update field utama experience
		if err := tx.Model(&exp).Updates(map[string]interface{}{
			"title":      input.Title,
			"location":   input.Location,
			"start_date": startDate,
			"end_date":   endDate, // akan di-set NULL kalau nil
		}).Error; err != nil {
			return err
		}

		// 2. Hapus semua points lama milik experience ini
		if err := tx.Delete(&models.ExperiencePoint{}, "experience_id = ?", id).Error; err != nil {
			return err
		}

		// 3. Insert points baru
		for _, p := range input.Points {
			if p.Content == "" {
				continue
			}
			point := models.ExperiencePoint{
				ID:           uuid.New().String(),
				Content:      p.Content,
				ExperienceID: id,
			}
			if err := tx.Create(&point).Error; err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengupdate pengalaman: " + err.Error()})
		return
	}

	// Reload dengan preload untuk response yang lengkap
	config.DB.Preload("Points").First(&exp, "id = ?", id)
	c.JSON(http.StatusOK, gin.H{"status": "success", "data": exp})
}

// DELETE /api/admin/experiences/:id
func DeleteExperience(c *gin.Context) {
	id := c.Param("id")

	if err := config.DB.Delete(&models.Experience{}, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menghapus data"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Data pengalaman berhasil dihapus"})
}