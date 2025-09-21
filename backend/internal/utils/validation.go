// morningstarl2504/balkanid_repo/BalkanID_repo-f1fc3ed153144eb6d79e3c90f73a0f3d312b9c79/backend/internal/utils/validation.go
package utils

import (
	"fmt"
	"mime/multipart"
	"net/http"
	"strings"
)

// ValidateFileType checks the actual content type of a file against its declared MIME type.
func ValidateFileType(fileHeader *multipart.FileHeader) error {
	declaredMimeType := fileHeader.Header.Get("Content-Type")

	// Open the file to read its content
	file, err := fileHeader.Open()
	if err != nil {
		return fmt.Errorf("could not open file for validation: %w", err)
	}
	defer file.Close()

	// Read the first 512 bytes to detect the content type
	buffer := make([]byte, 512)
	_, err = file.Read(buffer)
	if err != nil {
		return fmt.Errorf("could not read file for validation: %w", err)
	}

	// Detect the actual MIME type from the content
	detectedMimeType := http.DetectContentType(buffer)

	// Allow for some flexibility (e.g., 'text/plain' vs 'text/csv')
	// We compare the base type (e.g., 'text', 'image', 'application')
	declaredBase := strings.Split(declaredMimeType, "/")[0]
	detectedBase := strings.Split(detectedMimeType, "/")[0]

	if declaredBase != detectedBase {
		// A special case for common document types that can be detected as 'application/zip'
		if detectedMimeType == "application/zip" {
			switch declaredMimeType {
			case "application/vnd.openxmlformats-officedocument.wordprocessingml.document", // .docx
				"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",      // .xlsx
				"application/vnd.openxmlformats-officedocument.presentationml.presentation": // .pptx
				return nil // Allow these as they are zip-based formats
			}
		}
		return fmt.Errorf("file content mismatch: declared as %s, but detected as %s", declaredMimeType, detectedMimeType)
	}

	return nil
}