package utils

import (
	"golang.org/x/crypto/bcrypt"
)

// HashPassword mengubah string password menjadi hash yang aman
func HashPassword(password string) (string, error) {
	// Cost 10 adalah standar keseimbangan antara keamanan dan kecepatan
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	return string(bytes), err
}

// CheckPasswordHash membandingkan password input dengan hash di database
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}