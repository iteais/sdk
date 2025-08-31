package utils

import (
	"testing"
)

func TestArrayUnique(t *testing.T) {
	tests := []struct {
		name     string
		input    []int
		expected []int
	}{
		{
			name:     "empty slice",
			input:    []int{},
			expected: []int{},
		},
		{
			name:     "no duplicates",
			input:    []int{1, 2, 3, 4},
			expected: []int{1, 2, 3, 4},
		},
		{
			name:     "with duplicates",
			input:    []int{1, 2, 2, 3, 4, 4, 4},
			expected: []int{1, 2, 3, 4},
		},
		{
			name:     "all duplicates",
			input:    []int{1, 1, 1, 1},
			expected: []int{1},
		},
		{
			name:     "unsorted with duplicates",
			input:    []int{4, 2, 3, 2, 1, 4},
			expected: []int{4, 2, 3, 1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ArrayUnique(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("ArrayUnique() = %v, want %v", result, tt.expected)
				return
			}
			for _, v := range result {
				if !contains(tt.expected, v) {
					t.Errorf("ArrayUnique() = %v, want %v", result, tt.expected)
				}
			}
		})
	}
}

func contains[A comparable](slice []A, value A) bool {
	for _, v := range slice {
		if v == value {
			return true
		}
	}
	return false
}
