package main

import (
	"fmt"
	"log"

	"github.com/MuhammadZaidan1/jays-portfolio-api/internal/config"
	"github.com/MuhammadZaidan1/jays-portfolio-api/internal/models"
	"github.com/MuhammadZaidan1/jays-portfolio-api/internal/utils"
)

func main() {
	// 1. Load env dan koneksi database
	config.LoadEnv()
	config.ConnectDB()

	// 2. Data Admin Baru
	username := "jeyjouw123"
	plainPassword := "niggafkmeup456"

	// 3. Hash password menggunakan Bcrypt (Utility yang sudah kita buat)
	hashedPassword, err := utils.HashPassword(plainPassword)
	if err != nil {
		log.Fatal("Gagal melakukan hashing password:", err)
	}

	// 4. Masukkan ke Database
	admin := models.Admin{
		Username: username,
		Password: hashedPassword,
	}

	// Cek apakah username sudah ada untuk menghindari duplikat
	var existingAdmin models.Admin
	if err := config.DB.Where("username = ?", username).First(&existingAdmin).Error; err == nil {
		fmt.Println("⚠️ Admin dengan username ini sudah ada!")
		return
	}

	if err := config.DB.Create(&admin).Error; err != nil {
		log.Fatal("Gagal membuat akun admin:", err)
	}

	fmt.Printf("✅ Admin berhasil dibuat!\nUsername: %s\n", username)
}