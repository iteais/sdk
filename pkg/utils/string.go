package utils

import "math/rand"

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
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}
