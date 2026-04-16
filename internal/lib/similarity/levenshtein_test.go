package similarity

import "testing"

func TestLevenshteinDistance(t *testing.T) {
	tests := []struct {
		s1, s2   string
		expected int
	}{
		{"", "", 0},
		{"a", "", 1},
		{"", "a", 1},
		{"abc", "abc", 0},
		{"abc", "def", 3},
		{"kitten", "sitting", 3},
	}

	for _, tt := range tests {
		result := LevenshteinDistance(tt.s1, tt.s2)
		if result != tt.expected {
			t.Errorf("LevenshteinDistance(%q, %q) = %d, want %d",
				tt.s1, tt.s2, result, tt.expected)
		}
	}
}

func TestCalculateSimilarity(t *testing.T) {
	tests := []struct {
		s1, s2   string
		expected float64
	}{
		{"abc", "abc", 1.0},
		{"abc", "xyz", 0.0},
		{"uber trip", "uber viagem", 0.4545454545454546},
	}

	for _, tt := range tests {
		result := CalculateSimilarity(tt.s1, tt.s2)
		if result < tt.expected-0.01 || result > tt.expected+0.01 {
			t.Errorf("CalculateSimilarity(%q, %q) = %f, want %f",
				tt.s1, tt.s2, result, tt.expected)
		}
	}
}
