package helper

import (
	"encoding/base64"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"

	"github.com/mitchellh/mapstructure"
	"golang.org/x/text/runes"
	"golang.org/x/text/unicode/norm"
)

func toCamelCase(input string) string {
	isToUpper := false
	var result string
	for i, v := range input {
		if i == 0 {
			result += strings.ToLower(string(v))
		} else if v == '_' {
			isToUpper = true
		} else {
			if isToUpper {
				result += strings.ToUpper(string(v))
				isToUpper = false
			} else {
				result += string(v)
			}
		}
	}
	return result
}

// ApplyChanges function to decode map into struct
func ApplyChanges(changes map[string]interface{}, to interface{}) error {
	camelCaseKeys := make(map[string]interface{})
	for k, v := range changes {
		camelCaseKeys[toCamelCase(k)] = v
	}

	dec, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		ErrorUnused: true,
		TagName:     "json",
		Result:      to,
		ZeroFields:  true,
	})

	if err != nil {
		return err
	}

	return dec.Decode(camelCaseKeys)
}

func Contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}

// ToSlug transforme une chaîne en un slug
func ToSlug(input string) string {
	// Normaliser les caractères pour supprimer les accents
	t := norm.NFD.String(input)
	t = strings.ToLower(t)
	t = runes.Remove(runes.In(unicode.Mn)).String(t) // Supprimer les diacritiques

	// Remplacer les caractères non alphanumériques par des tirets
	reg, _ := regexp.Compile(`[^a-z0-9]+`)
	t = reg.ReplaceAllString(t, "-")

	// Supprimer les tirets en début et fin
	t = strings.Trim(t, "-")

	return t
}

func GenerateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[rand.Intn(len(charset))]
	}
	return string(result)
}

// SaveBase64Image saves a base64 encoded image to disk
// Returns the filename (without path) on success, or error on failure
func SaveBase64Image(base64Data, basePath, entityType, entityID string) (string, error) {
	// Parse the base64 data to determine file type
	var imageData []byte
	var fileExtension string
	var err error

	// Check if the base64 data has a data URL prefix (e.g., "data:image/png;base64,")
	if strings.HasPrefix(base64Data, "data:") {
		// Extract the MIME type and base64 data
		parts := strings.Split(base64Data, ",")
		if len(parts) != 2 {
			return "", fmt.Errorf("invalid base64 data format")
		}

		// Determine file extension from MIME type
		mimeType := parts[0]
		if strings.Contains(mimeType, "image/jpeg") || strings.Contains(mimeType, "image/jpg") {
			fileExtension = ".jpg"
		} else if strings.Contains(mimeType, "image/png") {
			fileExtension = ".png"
		} else if strings.Contains(mimeType, "image/gif") {
			fileExtension = ".gif"
		} else if strings.Contains(mimeType, "image/webp") {
			fileExtension = ".webp"
		} else {
			return "", fmt.Errorf("unsupported image type")
		}

		// Decode the base64 data
		imageData, err = base64.StdEncoding.DecodeString(parts[1])
		if err != nil {
			return "", fmt.Errorf("failed to decode base64 data: %v", err)
		}
	} else {
		// Assume it's plain base64 without prefix, default to .jpg
		fileExtension = ".jpg"
		imageData, err = base64.StdEncoding.DecodeString(base64Data)
		if err != nil {
			return "", fmt.Errorf("failed to decode base64 data: %v", err)
		}
	}

	// Create the directory structure
	uploadDir := filepath.Join(basePath, "uploads", entityType, entityID)
	err = os.MkdirAll(uploadDir, os.ModePerm)
	if err != nil {
		return "", fmt.Errorf("failed to create upload directory: %v", err)
	}

	// Generate the filename
	filename := entityID + fileExtension
	filePath := filepath.Join(uploadDir, filename)

	// Write the file
	err = os.WriteFile(filePath, imageData, 0644)
	if err != nil {
		return "", fmt.Errorf("failed to write image file: %v", err)
	}

	return filename, nil
}

// SaveBase64Document saves a base64 encoded document to disk
// Returns the filename (without path) on success, or error on failure
func SaveBase64Document(base64Data, basePath, entityType, entityID, prefix string) (string, error) {
	// Parse the base64 data to determine file type
	var documentData []byte
	var fileExtension string
	var err error

	// Check if the base64 data has a data URL prefix (e.g., "data:application/pdf;base64,")
	if strings.HasPrefix(base64Data, "data:") {
		// Extract the MIME type and base64 data
		parts := strings.Split(base64Data, ",")
		if len(parts) != 2 {
			return "", fmt.Errorf("invalid base64 data format")
		}

		// Determine file extension from MIME type
		mimeType := parts[0]
		if strings.Contains(mimeType, "application/pdf") {
			fileExtension = ".pdf"
		} else if strings.Contains(mimeType, "application/msword") {
			fileExtension = ".doc"
		} else if strings.Contains(mimeType, "application/vnd.openxmlformats-officedocument.wordprocessingml.document") {
			fileExtension = ".docx"
		} else if strings.Contains(mimeType, "text/plain") {
			fileExtension = ".txt"
		} else if strings.Contains(mimeType, "image/") {
			// Handle images as well
			if strings.Contains(mimeType, "image/jpeg") || strings.Contains(mimeType, "image/jpg") {
				fileExtension = ".jpg"
			} else if strings.Contains(mimeType, "image/png") {
				fileExtension = ".png"
			} else if strings.Contains(mimeType, "image/gif") {
				fileExtension = ".gif"
			} else if strings.Contains(mimeType, "image/webp") {
				fileExtension = ".webp"
			} else {
				fileExtension = ".jpg" // default for unknown image types
			}
		} else {
			fileExtension = ".bin" // default for unknown types
		}

		// Decode the base64 data
		documentData, err = base64.StdEncoding.DecodeString(parts[1])
		if err != nil {
			return "", fmt.Errorf("failed to decode base64 data: %v", err)
		}
	} else {
		// Assume it's plain base64 without prefix, default to .pdf
		fileExtension = ".pdf"
		documentData, err = base64.StdEncoding.DecodeString(base64Data)
		if err != nil {
			return "", fmt.Errorf("failed to decode base64 data: %v", err)
		}
	}

	// Create the directory structure
	uploadDir := filepath.Join(basePath, "uploads", entityType, entityID)
	err = os.MkdirAll(uploadDir, os.ModePerm)
	if err != nil {
		return "", fmt.Errorf("failed to create upload directory: %v", err)
	}

	// Generate the filename with prefix
	filename := prefix + entityID + fileExtension
	filePath := filepath.Join(uploadDir, filename)

	// Write the file
	err = os.WriteFile(filePath, documentData, 0644)
	if err != nil {
		return "", fmt.Errorf("failed to write document file: %v", err)
	}

	return filename, nil
}
