package utils

import (
	"math/rand"
	"strings"
	"unicode"
)

func SliceString(s string) string {
	n := 10

	// Convert the string to a slice of runes
	runes := []rune(s)

	// Check if the string has at least n runes
	if len(runes) > n {
		// Slice the rune slice and convert back to a string
		return string(runes[:n])
	} else {
		// If the string is shorter than n, print the whole string
		return s
	}
}

func GenerateRandomString(length int) string {
	if length <= 0 {
		return ""
	}

	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

// ToUpperCamelCase converts a snake_case string to UpperCamelCase.
func ToUpperCamelCase(s string) string {
	parts := strings.Split(s, "_")
	if len(parts) == 0 {
		return ""
	}

	result := []rune{}
	for _, part := range parts {
		if len(part) > 0 {
			// Capitalize the first letter of all parts
			result = append(result, unicode.ToUpper(rune(part[0])))
			result = append(result, []rune(part[1:])...)
		}
	}
	return string(result)
}
