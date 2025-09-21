// morningstarl2504/balkanid_repo/BalkanID_repo-f1fc3ed1d312b9c79/backend/internal/utils/hash.go
package utils

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
)

func CalculateSHA256(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()
	return CalculateSHA256Reader(file)
}

func CalculateSHA256Reader(reader io.Reader) (string, error) {
	hash := sha256.New()
	if _, err := io.Copy(hash, reader); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}