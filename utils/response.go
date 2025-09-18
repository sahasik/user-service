package utils

import (
	"fmt"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

// ValidateImageFile validates uploaded image file
func ValidateImageFile(file *multipart.FileHeader, maxSize int64) error {
	// Check file size
	if file.Size > maxSize {
		return fmt.Errorf("file too large: max size is %d bytes", maxSize)
	}

	// Check file extension
	ext := strings.ToLower(filepath.Ext(file.Filename))
	allowedExts := []string{".jpg", ".jpeg", ".png"}

	isAllowed := false
	for _, allowedExt := range allowedExts {
		if ext == allowedExt {
			isAllowed = true
			break
		}
	}

	if !isAllowed {
		return fmt.Errorf("invalid file type: only jpg, jpeg, png are allowed")
	}

	return nil
}

// ValidateDocumentFile validates uploaded document file
func ValidateDocumentFile(file *multipart.FileHeader, maxSize int64) error {
	// Check file size
	if file.Size > maxSize {
		return fmt.Errorf("file too large: max size is %d bytes", maxSize)
	}

	// Check file extension
	ext := strings.ToLower(filepath.Ext(file.Filename))
	allowedExts := []string{".pdf", ".doc", ".docx", ".jpg", ".jpeg", ".png"}

	isAllowed := false
	for _, allowedExt := range allowedExts {
		if ext == allowedExt {
			isAllowed = true
			break
		}
	}

	if !isAllowed {
		return fmt.Errorf("invalid file type: only pdf, doc, docx, jpg, jpeg, png are allowed")
	}

	return nil
}

// SaveUploadedFile saves uploaded file to specified directory
func SaveUploadedFile(file *multipart.FileHeader, category string, userID int, basePath string) (string, error) {
	// Create directory structure: uploads/category/year/month/
	now := time.Now()
	dir := filepath.Join(basePath, category, fmt.Sprintf("%d", now.Year()), fmt.Sprintf("%02d", now.Month()))

	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory: %v", err)
	}

	// Generate unique filename
	ext := filepath.Ext(file.Filename)
	filename := fmt.Sprintf("%s_%d_%s%s",
		category, userID, uuid.New().String()[:8], ext)

	fullPath := filepath.Join(dir, filename)

	// Open uploaded file
	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	// Create destination file
	dst, err := os.Create(fullPath)
	if err != nil {
		return "", err
	}
	defer dst.Close()

	// Copy file content
	_, err = dst.ReadFrom(src)
	if err != nil {
		return "", err
	}

	// Return relative path
	relativePath := filepath.Join(category, fmt.Sprintf("%d", now.Year()), fmt.Sprintf("%02d", now.Month()), filename)
	return relativePath, nil
}

// DeleteFile deletes a file from filesystem
func DeleteFile(filePath, basePath string) error {
	if filePath == "" {
		return nil
	}

	fullPath := filepath.Join(basePath, filePath)

	// Check if file exists
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return nil // File doesn't exist, nothing to delete
	}

	return os.Remove(fullPath)
}

// GetFileURL generates file URL for serving
func GetFileURL(filePath, baseURL string) string {
	if filePath == "" {
		return ""
	}

	return fmt.Sprintf("%s/files/%s", baseURL, filePath)
}
