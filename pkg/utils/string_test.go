package utils

import (
	"testing"
)

func TestSliceString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{name: "short string", input: "abc", expected: "abc"},
		{name: "long string", input: "abcdefghijklmnopqrstuvwxyz", expected: "abcdefghij"},
		{name: "empty string", input: "", expected: ""},
		{name: "exactly 10 chars", input: "1234567890", expected: "1234567890"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SliceString(tt.input)
			if result != tt.expected {
				t.Errorf("SliceString(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestGenerateRandomString(t *testing.T) {
	tests := []struct {
		name     string
		length   int
		expected int
	}{
		{name: "zero length", length: 0, expected: 0},
		{name: "positive length", length: 5, expected: 5},
		{name: "large length", length: 100, expected: 100},
		{name: "negative length", length: -1, expected: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateRandomString(tt.length)
			if len(result) != tt.expected {
				t.Errorf("GenerateRandomString(%d) length = %d, want %d", tt.length, len(result), tt.expected)
			}
		})
	}
}

func TestToUpperCamelCase(t *testing.T) {
	tests := []struct {
		name   string
		string string
		want   string
	}{
		{name: "one_words", string: "hello", want: "Hello"},
		{name: "two_words", string: "hello_world", want: "HelloWorld"},
		{name: "three_words", string: "hello_world_test", want: "HelloWorldTest"},
		{name: "whi_prefix", string: "preHello_world", want: "PreHelloWorld"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ToUpperCamelCase(tt.string); got != tt.want {
				t.Errorf("ToUpperCamelCase() = %v, want %v", got, tt.want)
			}
		})
	}
}
