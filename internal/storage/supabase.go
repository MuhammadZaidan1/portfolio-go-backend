// internal/storage/supabase.go
package storage

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func supabaseURL() string    { return os.Getenv("SUPABASE_URL") }
func serviceRoleKey() string { return os.Getenv("SUPABASE_SERVICE_KEY") }

// upload adalah helper internal — kirim file ke Supabase Storage bucket tertentu.
// path = path di dalam bucket, misal "cv_latest.pdf" atau "abc123/uuid.jpg"
// contentType = "application/pdf", "image/jpeg", dll
// upsert = true → replace kalau sudah ada
func upload(bucket, path string, data []byte, contentType string, upsert bool) (string, error) {
	uploadURL := fmt.Sprintf("%s/storage/v1/object/%s/%s", supabaseURL(), bucket, path)

	req, err := http.NewRequest(http.MethodPost, uploadURL, bytes.NewReader(data))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+serviceRoleKey())
	req.Header.Set("Content-Type", contentType)
	if upsert {
		req.Header.Set("x-upsert", "true")
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("upload request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("supabase storage error %d: %s", resp.StatusCode, string(body))
	}

	publicURL := fmt.Sprintf("%s/storage/v1/object/public/%s/%s", supabaseURL(), bucket, path)
	return publicURL, nil
}

// deleteFile adalah helper internal — hapus file dari Supabase Storage.
// path = path di dalam bucket, misal "abc123/uuid.jpg"
func deleteFile(bucket, path string) error {
	deleteURL := fmt.Sprintf("%s/storage/v1/object/%s/%s", supabaseURL(), bucket, path)

	req, err := http.NewRequest(http.MethodDelete, deleteURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create delete request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+serviceRoleKey())

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("delete request failed: %w", err)
	}
	defer resp.Body.Close()

	// 200 atau 404 dianggap ok (file mungkin sudah tidak ada)
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusNotFound {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("supabase delete error %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// ── CV ────────────────────────────────────────────────────────────────────────

// UploadCV upload file PDF ke bucket "cv".
// Selalu pakai nama fixed "cv_latest.pdf" → auto-replace file lama.
func UploadCV(fileData io.Reader, filename string) (string, error) {
	data, err := io.ReadAll(fileData)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}
	return upload("cv", filename, data, "application/pdf", true)
}

// ── Project Images ────────────────────────────────────────────────────────────

// UploadProjectImage upload gambar ke bucket "project-images".
// Disimpan di path: {projectID}/{filename} supaya mudah di-manage per project.
// originalFilename dipakai untuk deteksi content-type dari extension.
// Return: public URL dan storage path (untuk keperluan delete nanti).
func UploadProjectImage(fileData io.Reader, projectID, originalFilename string) (publicURL, storagePath string, err error) {
	data, readErr := io.ReadAll(fileData)
	if readErr != nil {
		return "", "", fmt.Errorf("failed to read file: %w", readErr)
	}

	// Deteksi content type dari extension
	ext := strings.ToLower(filepath.Ext(originalFilename))
	contentType := extensionToContentType(ext)
	if contentType == "" {
		return "", "", fmt.Errorf("unsupported image format: %s", ext)
	}

	// Path: {projectID}/{originalFilename}
	// Caller bertanggung jawab kasih filename yang unik (pakai UUID)
	storagePath = fmt.Sprintf("%s/%s", projectID, originalFilename)

	publicURL, err = upload("project-images", storagePath, data, contentType, false)
	if err != nil {
		return "", "", err
	}

	return publicURL, storagePath, nil
}

// DeleteProjectImage hapus gambar dari bucket "project-images".
// storagePath = nilai yang disimpan di DB, format: {projectID}/{filename}
func DeleteProjectImage(storagePath string) error {
	return deleteFile("project-images", storagePath)
}

// ExtractProjectImagePath ekstrak storage path dari public URL.
// Berguna kalau di DB hanya simpan public URL, bukan path terpisah.
// Input:  "https://xxx.supabase.co/storage/v1/object/public/project-images/abc/img.jpg"
// Output: "abc/img.jpg"
func ExtractProjectImagePath(publicURL string) string {
	marker := "/project-images/"
	idx := strings.Index(publicURL, marker)
	if idx == -1 {
		return ""
	}
	return publicURL[idx+len(marker):]
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func extensionToContentType(ext string) string {
	switch ext {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".webp":
		return "image/webp"
	case ".gif":
		return "image/gif"
	case ".svg":
		return "image/svg+xml"
	default:
		return ""
	}
}

// ── Certificate Files ─────────────────────────────────────────────────────────

// UploadCertificate upload file sertifikat ke bucket "certificates".
// Simpan di path: {certID}/{filename} — support PDF dan image.
// Return: public URL dan storage path (untuk keperluan delete).
func UploadCertificate(fileData io.Reader, certID, originalFilename string) (publicURL, storagePath string, err error) {
	data, readErr := io.ReadAll(fileData)
	if readErr != nil {
		return "", "", fmt.Errorf("failed to read file: %w", readErr)
	}

	ext := strings.ToLower(filepath.Ext(originalFilename))
	contentType := certContentType(ext)
	if contentType == "" {
		return "", "", fmt.Errorf("unsupported format: %s (gunakan pdf, jpg, png, atau webp)", ext)
	}

	storagePath = fmt.Sprintf("%s/%s", certID, originalFilename)
	publicURL, err = upload("certificates", storagePath, data, contentType, true) // upsert=true → replace
	if err != nil {
		return "", "", err
	}

	return publicURL, storagePath, nil
}

// DeleteCertificateFile hapus file sertifikat dari bucket "certificates".
func DeleteCertificateFile(storagePath string) error {
	return deleteFile("certificates", storagePath)
}

// ExtractCertFilePath ekstrak storage path dari public URL.
func ExtractCertFilePath(publicURL string) string {
	marker := "/certificates/"
	idx := strings.Index(publicURL, marker)
	if idx == -1 {
		return ""
	}
	return publicURL[idx+len(marker):]
}

func certContentType(ext string) string {
	switch ext {
	case ".pdf":
		return "application/pdf"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".webp":
		return "image/webp"
	default:
		return ""
	}
}