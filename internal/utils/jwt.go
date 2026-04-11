package utils

import (
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// AdminClaims mendefinisikan isi data di dalam token
type AdminClaims struct {
	AdminID  string `json:"admin_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// GenerateToken membuat token baru untuk Admin yang berhasil login
func GenerateToken(adminID string, username string) (string, error) {
	secretKey := []byte(os.Getenv("JWT_SECRET"))

	// REVISI: Token sekarang HANYA berlaku selama 30 menit
	claims := AdminClaims{
		adminID,
		username,
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(30 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secretKey)
}

// ValidateToken membedah dan memvalidasi apakah token tersebut asli/belum expired
func ValidateToken(tokenString string) (*AdminClaims, error) {
	secretKey := []byte(os.Getenv("JWT_SECRET"))

	token, err := jwt.ParseWithClaims(tokenString, &AdminClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Pastikan algoritma signing-nya sesuai (HS256)
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return secretKey, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*AdminClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}