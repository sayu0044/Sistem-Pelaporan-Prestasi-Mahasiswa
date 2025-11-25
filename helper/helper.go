package helper

import (
	"strings"
)

// TrimSpace menghapus whitespace dari string
func TrimSpace(s string) string {
	return strings.TrimSpace(s)
}

// IsEmpty memeriksa apakah string kosong
func IsEmpty(s string) bool {
	return strings.TrimSpace(s) == ""
}

// Contains memeriksa apakah slice mengandung value tertentu
func Contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

