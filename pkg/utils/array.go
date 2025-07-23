package utils

func ArrayUnique[A comparable](input []A) []A {
	size := len(input)
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
