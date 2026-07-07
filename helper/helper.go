package helper

import (
	"encoding/base64"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"time"
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

// Common time formats for parsing
var timeFormats = []string{
	time.RFC3339,               // "2006-01-02T15:04:05Z07:00"
	time.RFC3339Nano,           // "2006-01-02T15:04:05.999999999Z07:00"
	"2006-01-02T15:04:05",      // ISO format without timezone
	"2006-01-02 15:04:05",      // Common database format
	"2006-01-02T15:04:05.000Z", // JavaScript JSON format
	"2006-01-02",               // Date only
	"15:04:05",                 // Time only
	"15:04",                    // Time without seconds
}

// ParseTime attempts to parse a time string using common formats
func ParseTime(timeStr string) (time.Time, error) {
	if timeStr == "" {
		return time.Time{}, nil
	}

	for _, format := range timeFormats {
		if t, err := time.Parse(format, timeStr); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse time: %s", timeStr)
}

// FormatTime formats a time.Time to RFC3339 format
func FormatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(time.RFC3339)
}

// IsValidTimeString checks if a string can be parsed as a time
func IsValidTimeString(timeStr string) bool {
	_, err := ParseTime(timeStr)
	return err == nil
}

// TimeToStartOfDay returns the time at the start of the day (00:00:00)
func TimeToStartOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

// TimeToEndOfDay returns the time at the end of the day (23:59:59.999999999)
func TimeToEndOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 999999999, t.Location())
}

// stringToTimeHookFunc is a mapstructure decode hook that converts strings to time.Time
func stringToTimeHookFunc() mapstructure.DecodeHookFunc {
	return func(
		f reflect.Type,
		t reflect.Type,
		data interface{}) (interface{}, error) {

		if f.Kind() != reflect.String {
			return data, nil
		}

		if t != reflect.TypeOf(time.Time{}) {
			return data, nil
		}

		// Convert the string to time.Time
		return ParseTime(data.(string))
	}
}

// ApplyChanges function to decode map into struct with support for time.Time fields
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
		DecodeHook:  stringToTimeHookFunc(),
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

// CopyRambleUpload copies an uploaded ramble file from one ramble directory to another.
// Returns the new filename on success.
func CopyRambleUpload(basePath string, srcRambleID, dstRambleID uint, filename string) (string, error) {
	srcPath := filepath.Join(basePath, "uploads", "ramble", fmt.Sprintf("%d", srcRambleID), filename)
	data, err := os.ReadFile(srcPath)
	if err != nil {
		return "", fmt.Errorf("failed to read source file: %v", err)
	}

	ext := filepath.Ext(filename)
	dstRambleIDStr := fmt.Sprintf("%d", dstRambleID)
	var dstFilename string
	if strings.HasPrefix(filename, "document_") {
		dstFilename = "document_" + dstRambleIDStr + ext
	} else {
		dstFilename = dstRambleIDStr + ext
	}

	dstDir := filepath.Join(basePath, "uploads", "ramble", dstRambleIDStr)
	if err := os.MkdirAll(dstDir, os.ModePerm); err != nil {
		return "", fmt.Errorf("failed to create upload directory: %v", err)
	}

	dstPath := filepath.Join(dstDir, dstFilename)
	if err := os.WriteFile(dstPath, data, 0644); err != nil {
		return "", fmt.Errorf("failed to write destination file: %v", err)
	}

	return dstFilename, nil
}
