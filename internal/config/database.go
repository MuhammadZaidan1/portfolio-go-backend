package config

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/MuhammadZaidan1/jays-portfolio-api/internal/models"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// LoadEnv mengisi environment variables dari file .env
func LoadEnv() {
	err := godotenv.Load()
	if err != nil {
		// Ubah dari Fatal ke Println. 
		// Kenapa? Karena saat deploy ke production, file .env biasanya nggak ada 
		// (env diset langsung di server). Kalau pakai Fatal, app lu bakal crash pas di-deploy.
		log.Println("⚠️  Warning: .env file not found. Using system environment variables if available.")
	}
}

// ConnectDB membuka koneksi ke Supabase dan melakukan Auto-Migrate
func ConnectDB() {
	// 1. Ambil Connection String dari .env
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("❌ DATABASE_URL is not set in environment variables")
	}

	// 2. Buka Koneksi menggunakan GORM dan Driver Postgres
	// KUNCI UTAMA SUPABASE POOLER: PreferSimpleProtocol WAJIB true, dipadu dengan PrepareStmt: false
	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  dsn,
		PreferSimpleProtocol: true, // Mencegah driver nge-hang saat lewat Supavisor/PgBouncer
	}), &gorm.Config{
		PrepareStmt: false, // Mencegah GORM membuat cache prepared statement
		Logger:      logger.Default.LogMode(logger.Warn), // Opsional: Ubah ke Info kalau mau lihat log SQL
	})
	if err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}

	// 3. Konfigurasi Connection Pool Bawaan Go (Best Practice)
	// Biar koneksi nggak bocor dan membebani pooler Supabase
	sqlDB, err := db.DB()
	if err == nil {
		sqlDB.SetMaxIdleConns(10)           // Maksimal koneksi nganggur
		sqlDB.SetMaxOpenConns(50)           // Maksimal koneksi berjalan
		sqlDB.SetConnMaxLifetime(time.Hour) // Umur maksimal satu koneksi
	}

	fmt.Println("✅ Successfully connected to Supabase PostgreSQL!")

	// 4. Auto-Migration: Sinkronisasi Model Go ke Tabel Database
	fmt.Println("🚀 Running Auto-Migration...")
	err = db.AutoMigrate(
		&models.Admin{},
		&models.CV{},
		&models.SkillCategory{},
		&models.Skill{},
		&models.Project{},
		&models.ProjectImage{},
		&models.ProjectLink{},
		&models.Experience{},
		&models.ExperiencePoint{},
		&models.Certificate{},
		&models.ContactMessage{},
	)

	if err != nil {
		log.Fatalf("❌ Migration Failed: %v", err)
	}

	fmt.Println("✨ Database Migration Completed Successfully!")

	DB = db
}