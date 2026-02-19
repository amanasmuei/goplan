package claude

import "fmt"

// ExtractJSON extracts a JSON object from text that might contain other content.
// It handles responses where JSON may be wrapped in markdown code blocks or
// surrounded by explanatory text. Returns the first complete JSON object found.
func ExtractJSON(text string) (string, error) {
	// Find the first { and matching }
	start := -1
	depth := 0

	for i, c := range text {
		if c == '{' {
			if start == -1 {
				start = i
			}
			depth++
		} else if c == '}' {
			depth--
			if depth == 0 && start != -1 {
				return text[start : i+1], nil
			}
		}
	}

	return "", fmt.Errorf("no valid JSON object found in text")
}
