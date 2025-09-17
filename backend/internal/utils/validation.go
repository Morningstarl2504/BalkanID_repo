package utils

import (
	"fmt"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func ValidateFileType(filePath, declaredMimeType string) error {
	// Check file content
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Read first 512 bytes to detect content type
	buffer := make([]byte, 512)
	_, err = file.Read(buffer)
	if err != nil {
		return err
	}

	detectedMimeType := http.DetectContentType(buffer)
	
	// Also check by file extension
	ext := filepath.Ext(filePath)
	expectedMimeType := mime.TypeByExtension(ext)

	// Normalize mime types for comparison
	detectedBase := strings.Split(detectedMimeType, "/")[0]
	declaredBase := strings.Split(declaredMimeType, "/")[0]
	expectedBase := strings.Split(expectedMimeType, "/")[0]

	// Allow if any match or if it's a text file (flexible)
	if detectedBase == declaredBase || expectedBase == declaredBase || 
	   detectedMimeType == declaredMimeType || expectedMimeType == declaredMimeType {
		return nil
	}

	return fmt.Errorf("file content type mismatch: declared=%s, detected=%s", 
		declaredMimeType, detectedMimeType)
}
