package handlers

import (
	"net/http"

	"github.com/MuhammadZaidan1/jays-portfolio-api/internal/config"
	"github.com/MuhammadZaidan1/jays-portfolio-api/internal/models"
	"github.com/gin-gonic/gin"
)

// --- PUBLIC ENDPOINTS ---

// GetSkills mengambil semua SkillCategory beserta Skills di dalamnya
func GetSkills(c *gin.Context) {
	var categories []models.SkillCategory

	err := config.DB.Preload("Skills").Find(&categories).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Gagal mengambil data skill: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   categories,
	})
}

// --- ADMIN ENDPOINTS (PROTECTED) ---

// CreateSkillCategory membuat kategori skill baru (Skills bisa disertakan sekaligus)
func CreateSkillCategory(c *gin.Context) {
	var category models.SkillCategory

	if err := c.ShouldBindJSON(&category); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := config.DB.Create(&category).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan kategori skill"})
		return
	}

	// Preload skills sebelum dikirim balik
	config.DB.Preload("Skills").First(&category, "id = ?", category.ID)
	c.JSON(http.StatusCreated, gin.H{"status": "success", "data": category})
}

// UpdateSkillCategory mengupdate nama kategori
func UpdateSkillCategory(c *gin.Context) {
	id := c.Param("id")
	var category models.SkillCategory

	if err := config.DB.First(&category, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Kategori tidak ditemukan"})
		return
	}

	if err := c.ShouldBindJSON(&category); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	config.DB.Save(&category)
	config.DB.Preload("Skills").First(&category, "id = ?", category.ID)
	c.JSON(http.StatusOK, gin.H{"status": "success", "data": category})
}

// DeleteSkillCategory menghapus kategori — Skills di dalamnya juga terhapus via FK cascade
func DeleteSkillCategory(c *gin.Context) {
	id := c.Param("id")

	if err := config.DB.Delete(&models.SkillCategory{}, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menghapus kategori"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Kategori berhasil dihapus"})
}

// CreateSkill menambahkan skill baru ke kategori yang sudah ada
func CreateSkill(c *gin.Context) {
	var skill models.Skill

	if err := c.ShouldBindJSON(&skill); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validasi category_id ada
	var category models.SkillCategory
	if err := config.DB.First(&category, "id = ?", skill.CategoryID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Kategori tidak ditemukan"})
		return
	}

	if err := config.DB.Create(&skill).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan skill"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"status": "success", "data": skill})
}

// UpdateSkill memperbarui data skill berdasarkan ID
func UpdateSkill(c *gin.Context) {
	id := c.Param("id")
	var skill models.Skill

	if err := config.DB.First(&skill, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Skill tidak ditemukan"})
		return
	}

	if err := c.ShouldBindJSON(&skill); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	config.DB.Save(&skill)
	c.JSON(http.StatusOK, gin.H{"status": "success", "data": skill})
}

// DeleteSkill menghapus skill berdasarkan ID
func DeleteSkill(c *gin.Context) {
	id := c.Param("id")

	if err := config.DB.Delete(&models.Skill{}, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menghapus skill"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Skill berhasil dihapus"})
}