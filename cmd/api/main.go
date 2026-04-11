package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/MuhammadZaidan1/jays-portfolio-api/internal/config"
	"github.com/MuhammadZaidan1/jays-portfolio-api/internal/handlers"
	"github.com/MuhammadZaidan1/jays-portfolio-api/internal/middleware"
)

func main() {
	// 1. Load Environment Variables
	config.LoadEnv()

	// 2. Hubungkan ke Database & Jalankan Auto-Migration
	config.ConnectDB()

	// 3. Setup Gin Engine
	r := gin.Default()
	r.SetTrustedProxies(nil)

	// 4. Setup CORS
	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:5173"
	}
	allowedOrigins := strings.Split(frontendURL, ",")

	r.Use(cors.New(cors.Config{
		AllowOrigins:     allowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "Accept", "X-Requested-With"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// 5. API Routes
	api := r.Group("/api")
	{
		api.GET("/ping", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"message": "Backend Jay is officially ALIVE!",
				"status":  "Connected to Supabase",
			})
		})

		// --- PUBLIC ROUTES ---
		api.GET("/skills",               handlers.GetSkills)
		api.GET("/experiences",          handlers.GetExperiences)
		api.GET("/projects",             handlers.GetProjects)
		api.GET("/projects/:id",         handlers.GetProjectByID)
		api.GET("/project-tech-stacks",  handlers.GetProjectTechStacks) // public untuk filter bar
		api.GET("/certificates",         handlers.GetCertificates)
		api.POST("/login",               handlers.Login)
		api.GET("/cv",                   handlers.GetCV)
		api.POST("/contact",             handlers.SendMessage)
		api.PATCH("/cv/download",        handlers.IncrementCVDownload)

		// --- PROTECTED ADMIN ROUTES ---
		admin := api.Group("/admin")
		admin.Use(middleware.AuthMiddleware())
		{
			// Projects
			admin.POST("/projects",       handlers.CreateProject)
			admin.PUT("/projects/:id",    handlers.UpdateProject)
			admin.DELETE("/projects/:id", handlers.DeleteProject)

			// Project Images
			admin.POST("/projects/:id/images",             handlers.UploadProjectImage)
			admin.DELETE("/projects/images/:image_id",     handlers.DeleteProjectImage)
			admin.PATCH("/projects/images/:image_id/sort", handlers.UpdateProjectImageSort)

			// Project TechStack Pool
			admin.GET("/project-tech-stacks",         handlers.GetProjectTechStacks)
			admin.POST("/project-tech-stacks",        handlers.CreateProjectTechStack)
			admin.POST("/project-tech-stacks/import", handlers.ImportSkillsToTechStack)
			admin.PUT("/project-tech-stacks/:id",     handlers.UpdateProjectTechStack)
			admin.DELETE("/project-tech-stacks/:id",  handlers.DeleteProjectTechStack)

			// Experiences
			admin.POST("/experiences",       handlers.CreateExperience)
			admin.PUT("/experiences/:id",    handlers.UpdateExperience)
			admin.DELETE("/experiences/:id", handlers.DeleteExperience)

			// Certificates
			admin.POST("/certificates",            handlers.CreateCertificate)
			admin.PUT("/certificates/:id",         handlers.UpdateCertificate)
			admin.DELETE("/certificates/:id",      handlers.DeleteCertificate)
			admin.POST("/certificates/:id/file",   handlers.UploadCertificateFile)

			// Skill Categories
			admin.POST("/skill-categories",       handlers.CreateSkillCategory)
			admin.PUT("/skill-categories/:id",    handlers.UpdateSkillCategory)
			admin.DELETE("/skill-categories/:id", handlers.DeleteSkillCategory)

			// Skills
			admin.POST("/skills",       handlers.CreateSkill)
			admin.PUT("/skills/:id",    handlers.UpdateSkill)
			admin.DELETE("/skills/:id", handlers.DeleteSkill)

			// CV & Messages
			admin.POST("/cv",                 handlers.UploadCV)
			admin.GET("/messages",            handlers.GetMessages)
			admin.PATCH("/messages/:id/read", handlers.MarkMessageRead)
			admin.DELETE("/messages/:id",     handlers.DeleteMessage)
		}
	}

	// 6. Jalankan Server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("🚀 Server running on http://localhost:%s\n", port)
	r.Run(":" + port)
}