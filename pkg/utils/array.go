package utils

// ArrayUnique returns a new slice containing only unique elements from the input slice.
// The order of elements in the output slice is preserved based on their first occurrence in the input.
// The input slice must contain comparable elements (e.g., int, string, etc.).
// If the input is nil, it returns an empty slice.
func ArrayUnique[A comparable](input []A) []A {
	if input == nil {
		return []A{}
	}

	size := len(input)
	if size == 0 {
		return []A{}
	}

	seen := make(map[A]bool, size)
	var result []A

	for _, v := range input {
		if !seen[v] {
			seen[v] = true
			result = append(result, v)
		}
	}

	return result
}
