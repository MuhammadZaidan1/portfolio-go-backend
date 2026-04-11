// internal/handlers/project_techstack_handler.go
package handlers

import (
	"net/http"

	"github.com/MuhammadZaidan1/jays-portfolio-api/internal/config"
	"github.com/MuhammadZaidan1/jays-portfolio-api/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// GET /admin/project-tech-stacks
// Return semua techstack pool beserta info skill asalnya kalau ada
func GetProjectTechStacks(c *gin.Context) {
	var techStacks []models.ProjectTechStack
	if err := config.DB.Preload("Skill").Order("name asc").Find(&techStacks).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data techstack"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "data": techStacks})
}

// POST /admin/project-tech-stacks
// Tambah techstack manual (bukan dari skill)
// Body: { name, icon_url? }
func CreateProjectTechStack(c *gin.Context) {
	var input struct {
		Name    string  `json:"name"     binding:"required"`
		IconURL *string `json:"icon_url"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Cek duplikat nama
	var existing models.ProjectTechStack
	if err := config.DB.Where("name = ?", input.Name).First(&existing).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Techstack dengan nama ini sudah ada"})
		return
	}

	ts := models.ProjectTechStack{
		ID:      uuid.New().String(),
		Name:    input.Name,
		IconURL: input.IconURL,
		SkillID: nil, // manual add — tidak dari skill
	}

	if err := config.DB.Create(&ts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan techstack"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"status": "success", "data": ts})
}

// PUT /admin/project-tech-stacks/:id
// Update nama atau icon techstack
// Body: { name?, icon_url? }
func UpdateProjectTechStack(c *gin.Context) {
	id := c.Param("id")

	var ts models.ProjectTechStack
	if err := config.DB.First(&ts, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Techstack tidak ditemukan"})
		return
	}

	var input struct {
		Name    *string `json:"name"`
		IconURL *string `json:"icon_url"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates := map[string]interface{}{}
	if input.Name != nil {
		updates["name"] = *input.Name
	}
	if input.IconURL != nil {
		updates["icon_url"] = input.IconURL
	}

	if len(updates) > 0 {
		config.DB.Model(&ts).Updates(updates)
	}

	// Reload
	config.DB.Preload("Skill").First(&ts, "id = ?", id)
	c.JSON(http.StatusOK, gin.H{"status": "success", "data": ts})
}

// DELETE /admin/project-tech-stacks/:id
// Hapus dari pool — relasi ke project ikut terhapus via many2many
func DeleteProjectTechStack(c *gin.Context) {
	id := c.Param("id")

	var ts models.ProjectTechStack
	if err := config.DB.First(&ts, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Techstack tidak ditemukan"})
		return
	}

	// Hapus asosiasi many2many dulu baru delete record
	if err := config.DB.Model(&ts).Association("Projects").Clear(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menghapus relasi project"})
		return
	}

	if err := config.DB.Delete(&ts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menghapus techstack"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Techstack berhasil dihapus"})
}

// POST /admin/project-tech-stacks/import
// Import satu atau beberapa skill ke pool techstack
// Body: { skill_ids: ["uuid1", "uuid2", ...] }
// Skill yang sudah ada di pool (by skill_id) akan di-skip, tidak duplikat
func ImportSkillsToTechStack(c *gin.Context) {
	var input struct {
		SkillIDs []string `json:"skill_ids" binding:"required,min=1"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Ambil skills yang diminta
	var skills []models.Skill
	if err := config.DB.Where("id IN ?", input.SkillIDs).Find(&skills).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data skill"})
		return
	}

	// Cek mana yang sudah ada di pool (by skill_id)
	var existing []models.ProjectTechStack
	config.DB.Where("skill_id IN ?", input.SkillIDs).Find(&existing)

	existingSkillIDs := map[string]bool{}
	for _, e := range existing {
		if e.SkillID != nil {
			existingSkillIDs[*e.SkillID] = true
		}
	}

	// Insert yang belum ada
	imported := []models.ProjectTechStack{}
	skipped  := []string{}

	err := config.DB.Transaction(func(tx *gorm.DB) error {
		for _, skill := range skills {
			if existingSkillIDs[skill.ID] {
				skipped = append(skipped, skill.Name)
				continue
			}

			skillID := skill.ID
			ts := models.ProjectTechStack{
				ID:      uuid.New().String(),
				Name:    skill.Name,
				IconURL: skill.IconURL,
				SkillID: &skillID,
			}

			if err := tx.Create(&ts).Error; err != nil {
				return err
			}
			imported = append(imported, ts)
		}
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengimport skills: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":   "success",
		"imported": len(imported),
		"skipped":  skipped,
		"data":     imported,
	})
}

// GET /skills — sudah ada di skill_handler.go
// Endpoint ini untuk kebutuhan modal "Import from Skills" di frontend
// Cukup hit GET /api/skills yang sudah ada, tidak perlu endpoint baru