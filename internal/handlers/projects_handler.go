// internal/handlers/projects_handler.go
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
	"gorm.io/gorm"
)

// parseProjectDate parse string "YYYY-MM" jadi time.Time (day default 01)
func parseProjectDate(s string) (time.Time, error) {
	return time.Parse("2006-01", s)
}

// ── PUBLIC ────────────────────────────────────────────────────────────────────

// GET /api/projects — urut project_date terbaru
func GetProjects(c *gin.Context) {
	var projects []models.Project
	err := config.DB.
		Preload("Images", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort_order asc")
		}).
		Preload("Links").
		Preload("TechStack").
		Order("project_date desc").
		Find(&projects).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data proyek: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "data": projects})
}

// GET /api/projects/:id
func GetProjectByID(c *gin.Context) {
	id := c.Param("id")
	var project models.Project

	err := config.DB.
		Preload("Images", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort_order asc")
		}).
		Preload("Links").
		Preload("TechStack").
		First(&project, "id = ?", id).Error

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project tidak ditemukan"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "data": project})
}

// ── ADMIN — PROJECT CRUD ──────────────────────────────────────────────────────

// POST /api/admin/projects
// Body: { title, short_story, description, is_featured, project_date "YYYY-MM", links[], tech_stack_ids[] }
func CreateProject(c *gin.Context) {
	var input struct {
		Title        string `json:"title"        binding:"required"`
		ShortStory   string `json:"short_story"`
		Description  string `json:"description"`
		IsFeatured   bool   `json:"is_featured"`
		ProjectDate  string `json:"project_date" binding:"required"` // "YYYY-MM"
		Links        []struct {
			Label string `json:"label"`
			URL   string `json:"url"`
		} `json:"links"`
		TechStackIDs []string `json:"tech_stack_ids"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	projectDate, err := parseProjectDate(input.ProjectDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format project_date tidak valid, gunakan YYYY-MM (contoh: 2024-03)"})
		return
	}

	projectID := uuid.New().String()

	links := make([]models.ProjectLink, 0, len(input.Links))
	for _, l := range input.Links {
		if l.Label == "" || l.URL == "" {
			continue
		}
		links = append(links, models.ProjectLink{
			ID:        uuid.New().String(),
			Label:     l.Label,
			URL:       l.URL,
			ProjectID: projectID,
		})
	}

	var techStacks []models.ProjectTechStack
	if len(input.TechStackIDs) > 0 {
		config.DB.Where("id IN ?", input.TechStackIDs).Find(&techStacks)
	}

	project := models.Project{
		ID:          projectID,
		Title:       input.Title,
		ShortStory:  input.ShortStory,
		Description: input.Description,
		IsFeatured:  input.IsFeatured,
		ProjectDate: projectDate,
		Links:       links,
		TechStack:   techStacks,
	}

	err = config.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Omit("TechStack.*").Create(&project).Error; err != nil {
			return err
		}
		if len(techStacks) > 0 {
			if err := tx.Model(&project).Association("TechStack").Replace(techStacks); err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan proyek: " + err.Error()})
		return
	}

	config.DB.Preload("Images").Preload("Links").Preload("TechStack").First(&project, "id = ?", projectID)
	c.JSON(http.StatusCreated, gin.H{"status": "success", "data": project})
}

// PUT /api/admin/projects/:id
func UpdateProject(c *gin.Context) {
	id := c.Param("id")

	var project models.Project
	if err := config.DB.First(&project, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Proyek tidak ditemukan"})
		return
	}

	var input struct {
		Title        string `json:"title"        binding:"required"`
		ShortStory   string `json:"short_story"`
		Description  string `json:"description"`
		IsFeatured   bool   `json:"is_featured"`
		ProjectDate  string `json:"project_date" binding:"required"` // "YYYY-MM"
		Links        []struct {
			Label string `json:"label"`
			URL   string `json:"url"`
		} `json:"links"`
		TechStackIDs []string `json:"tech_stack_ids"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	projectDate, err := parseProjectDate(input.ProjectDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format project_date tidak valid, gunakan YYYY-MM (contoh: 2024-03)"})
		return
	}

	var techStacks []models.ProjectTechStack
	if len(input.TechStackIDs) > 0 {
		config.DB.Where("id IN ?", input.TechStackIDs).Find(&techStacks)
	}

	err = config.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&project).Updates(map[string]interface{}{
			"title":        input.Title,
			"short_story":  input.ShortStory,
			"description":  input.Description,
			"is_featured":  input.IsFeatured,
			"project_date": projectDate,
		}).Error; err != nil {
			return err
		}

		if err := tx.Delete(&models.ProjectLink{}, "project_id = ?", id).Error; err != nil {
			return err
		}
		for _, l := range input.Links {
			if l.Label == "" || l.URL == "" {
				continue
			}
			if err := tx.Create(&models.ProjectLink{
				ID:        uuid.New().String(),
				Label:     l.Label,
				URL:       l.URL,
				ProjectID: id,
			}).Error; err != nil {
				return err
			}
		}

		if err := tx.Model(&project).Association("TechStack").Replace(techStacks); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengupdate proyek: " + err.Error()})
		return
	}

	config.DB.Preload("Images", func(db *gorm.DB) *gorm.DB {
		return db.Order("sort_order asc")
	}).Preload("Links").Preload("TechStack").First(&project, "id = ?", id)

	c.JSON(http.StatusOK, gin.H{"status": "success", "data": project})
}

// DELETE /api/admin/projects/:id
func DeleteProject(c *gin.Context) {
	id := c.Param("id")

	var project models.Project
	if err := config.DB.Preload("Images").First(&project, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Proyek tidak ditemukan"})
		return
	}

	for _, img := range project.Images {
		if img.StoragePath != "" {
			_ = storage.DeleteProjectImage(img.StoragePath)
		}
	}

	config.DB.Model(&project).Association("TechStack").Clear()

	if err := config.DB.Delete(&project).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menghapus proyek"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Proyek berhasil dihapus"})
}

// ── ADMIN — PROJECT IMAGES ────────────────────────────────────────────────────

// POST /api/admin/projects/:id/images
func UploadProjectImage(c *gin.Context) {
	projectID := c.Param("id")

	var project models.Project
	if err := config.DB.First(&project, "id = ?", projectID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Proyek tidak ditemukan"})
		return
	}

	fh, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File tidak ditemukan di request"})
		return
	}

	ext := strings.ToLower(filepath.Ext(fh.Filename))
	allowed := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".webp": true, ".gif": true}
	if !allowed[ext] {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Format tidak didukung: %s", ext)})
		return
	}

	f, err := fh.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuka file"})
		return
	}
	defer f.Close()

	filename := uuid.New().String() + ext
	publicURL, storagePath, err := storage.UploadProjectImage(f, projectID, filename)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal upload ke storage: " + err.Error()})
		return
	}

	var count int64
	config.DB.Model(&models.ProjectImage{}).Where("project_id = ?", projectID).Count(&count)

	img := models.ProjectImage{
		ID:          uuid.New().String(),
		ImageURL:    publicURL,
		StoragePath: storagePath,
		SortOrder:   int(count),
		ProjectID:   projectID,
	}

	if err := config.DB.Create(&img).Error; err != nil {
		_ = storage.DeleteProjectImage(storagePath)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan data gambar"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"status": "success", "data": img})
}

// DELETE /api/admin/projects/images/:image_id
func DeleteProjectImage(c *gin.Context) {
	imageID := c.Param("image_id")

	var img models.ProjectImage
	if err := config.DB.First(&img, "id = ?", imageID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Gambar tidak ditemukan"})
		return
	}

	if img.StoragePath != "" {
		if err := storage.DeleteProjectImage(img.StoragePath); err != nil {
			fmt.Printf("Warning: failed to delete from storage: %v\n", err)
		}
	}

	if err := config.DB.Delete(&img).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menghapus gambar"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Gambar berhasil dihapus"})
}

// PATCH /api/admin/projects/images/:image_id/sort
func UpdateProjectImageSort(c *gin.Context) {
	imageID := c.Param("image_id")

	var input struct {
		SortOrder int `json:"sort_order"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := config.DB.Model(&models.ProjectImage{}).
		Where("id = ?", imageID).
		Update("sort_order", input.SortOrder).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal update urutan gambar"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success"})
}