package utils

import "testing"

func TestCleanEmojis(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Country with location emoji prefix",
			input:    "📍Marrocos",
			expected: "Marrocos",
		},
		{
			name:     "Category with rainbow emoji",
			input:    "🌈Novembro",
			expected: "Novembro",
		},
		{
			name:     "Category with Christmas tree emoji",
			input:    "🎄Dezembro",
			expected: "Dezembro",
		},
		{
			name:     "String without emojis",
			input:    "Brasil",
			expected: "Brasil",
		},
		{
			name:     "String with multiple emojis",
			input:    "📍🇧🇷Brasil🇧🇷📍",
			expected: "Brasil",
		},
		{
			name:     "Empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "Only emoji",
			input:    "📍🌈🎄",
			expected: "",
		},
		{
			name:     "Emoji in the middle",
			input:    "Mar📍rocos",
			expected: "Marrocos",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CleanEmojis(tt.input)
			if result != tt.expected {
				t.Errorf("CleanEmojis(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
