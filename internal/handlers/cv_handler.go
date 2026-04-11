// internal/handlers/cv_handler.go
package handlers

import (
	"net/http"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/MuhammadZaidan1/jays-portfolio-api/internal/config"
	"github.com/MuhammadZaidan1/jays-portfolio-api/internal/models"
	"github.com/MuhammadZaidan1/jays-portfolio-api/internal/storage"
)

// GET /api/cv — public
func GetCV(c *gin.Context) {
	var cv models.CV
	if err := config.DB.Order("updated_at DESC").First(&cv).Error; err != nil {
		// Belum ada CV sama sekali, return null bukan error
		c.JSON(http.StatusOK, gin.H{"status": "success", "data": nil})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "data": cv})
}

// POST /api/admin/cv — admin, multipart/form-data { file: PDF }
func UploadCV(c *gin.Context) {
	fh, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "No file provided"})
		return
	}

	if filepath.Ext(fh.Filename) != ".pdf" {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Only PDF files are accepted"})
		return
	}

	f, err := fh.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to read file"})
		return
	}
	defer f.Close()

	// Pakai nama fixed supaya file lama otomatis ke-replace di Supabase Storage
	publicURL, err := storage.UploadCV(f, "cv_latest.pdf")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Upload to storage failed"})
		return
	}

	// Upsert: ambil record lama kalau ada, kalau tidak buat baru
	var cv models.CV
	result := config.DB.Order("updated_at DESC").First(&cv)

	if result.Error != nil {
		// Belum ada record sama sekali → buat baru
		cv = models.CV{
			ID:            uuid.New().String(),
			FileURL:       publicURL,
			DownloadCount: 0,
			UpdatedAt:     time.Now(),
		}
		config.DB.Create(&cv)
	} else {
		// Sudah ada → update URL dan timestamp saja, download_count tetap
		config.DB.Model(&cv).Updates(map[string]interface{}{
			"file_url":   publicURL,
			"updated_at": time.Now(),
		})
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "data": cv})
}

// PATCH /api/cv/download — public, increment download counter
func IncrementCVDownload(c *gin.Context) {
	config.DB.Model(&models.CV{}).
		Where("1 = 1").
		UpdateColumn("download_count", gorm.Expr("download_count + 1"))
	c.JSON(http.StatusOK, gin.H{"status": "success"})
}