package security

import "strings"

// MaskAPIKey masks an API key for display purposes, showing only the first 4 and last 4 characters
func MaskAPIKey(key string) string {
	if len(key) <= 8 {
		return "****"
	}
	return key[:4] + strings.Repeat("*", len(key)-8) + key[len(key)-4:]
}