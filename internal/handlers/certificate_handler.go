// internal/handlers/certificate_handler.go
package handlers

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/MuhammadZaidan1/jays-portfolio-api/internal/config"
	"github.com/MuhammadZaidan1/jays-portfolio-api/internal/models"
	"github.com/MuhammadZaidan1/jays-portfolio-api/internal/storage"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// GET /api/certificates — public
func GetCertificates(c *gin.Context) {
	var certificates []models.Certificate
	if err := config.DB.Order("date desc").Find(&certificates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data sertifikat: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "data": certificates})
}

// POST /api/admin/certificates
// Body: { title, issuer, date "YYYY-MM-DD", link_url? }
// File di-upload terpisah via POST /admin/certificates/:id/file
func CreateCertificate(c *gin.Context) {
	var input struct {
		Title   string  `json:"title"   binding:"required"`
		Issuer  string  `json:"issuer"  binding:"required"`
		Date    string  `json:"date"    binding:"required"` // "YYYY-MM-DD"
		LinkURL *string `json:"link_url"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	date, err := time.Parse("2006-01-02", input.Date)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format date tidak valid, gunakan YYYY-MM-DD"})
		return
	}

	cert := models.Certificate{
		ID:      uuid.New().String(),
		Title:   input.Title,
		Issuer:  input.Issuer,
		Date:    date,
		LinkURL: input.LinkURL,
	}

	if err := config.DB.Create(&cert).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan sertifikat"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"status": "success", "data": cert})
}

// PUT /api/admin/certificates/:id
// Body: { title, issuer, date "YYYY-MM-DD", link_url? }
// File di-upload/replace terpisah via POST /admin/certificates/:id/file
func UpdateCertificate(c *gin.Context) {
	id := c.Param("id")

	var cert models.Certificate
	if err := config.DB.First(&cert, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Sertifikat tidak ditemukan"})
		return
	}

	var input struct {
		Title   string  `json:"title"   binding:"required"`
		Issuer  string  `json:"issuer"  binding:"required"`
		Date    string  `json:"date"    binding:"required"`
		LinkURL *string `json:"link_url"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	date, err := time.Parse("2006-01-02", input.Date)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format date tidak valid, gunakan YYYY-MM-DD"})
		return
	}

	config.DB.Model(&cert).Updates(map[string]interface{}{
		"title":    input.Title,
		"issuer":   input.Issuer,
		"date":     date,
		"link_url": input.LinkURL,
	})

	// Reload
	config.DB.First(&cert, "id = ?", id)
	c.JSON(http.StatusOK, gin.H{"status": "success", "data": cert})
}

// DELETE /api/admin/certificates/:id
// Hapus record DB + file dari Supabase Storage
func DeleteCertificate(c *gin.Context) {
	id := c.Param("id")

	var cert models.Certificate
	if err := config.DB.First(&cert, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Sertifikat tidak ditemukan"})
		return
	}

	// Hapus file dari storage kalau ada
	if cert.StoragePath != "" {
		if err := storage.DeleteCertificateFile(cert.StoragePath); err != nil {
			fmt.Printf("Warning: failed to delete cert file from storage: %v\n", err)
		}
	}

	if err := config.DB.Delete(&cert).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menghapus sertifikat"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Sertifikat berhasil dihapus"})
}

// POST /api/admin/certificates/:id/file
// Upload atau replace file sertifikat (PDF/image)
// Form: multipart/form-data { file: <pdf|image> }
func UploadCertificateFile(c *gin.Context) {
	id := c.Param("id")

	var cert models.Certificate
	if err := config.DB.First(&cert, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Sertifikat tidak ditemukan"})
		return
	}

	fh, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File tidak ditemukan di request"})
		return
	}

	ext := strings.ToLower(filepath.Ext(fh.Filename))
	allowed := map[string]bool{".pdf": true, ".jpg": true, ".jpeg": true, ".png": true, ".webp": true}
	if !allowed[ext] {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Format tidak didukung: %s. Gunakan pdf, jpg, png, atau webp", ext)})
		return
	}

	f, err := fh.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuka file"})
		return
	}
	defer f.Close()

	// Nama file fixed per cert: cert_file.{ext}
	// upsert=true di storage helper → otomatis replace file lama
	filename := "cert_file" + ext
	publicURL, storagePath, err := storage.UploadCertificate(f, id, filename)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal upload ke storage: " + err.Error()})
		return
	}

	// Update URL dan path di DB
	config.DB.Model(&cert).Updates(map[string]interface{}{
		"file_url":     publicURL,
		"storage_path": storagePath,
	})

	config.DB.First(&cert, "id = ?", id)
	c.JSON(http.StatusOK, gin.H{"status": "success", "data": cert})
}