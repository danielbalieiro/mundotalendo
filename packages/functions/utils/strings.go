package utils

import (
	"strings"
)

// CleanEmojis remove todos os emojis de uma string
func CleanEmojis(s string) string {
	var result strings.Builder
	for _, r := range s {
		// Remove emojis (ranges Unicode)
		if !isEmoji(r) {
			result.WriteRune(r)
		}
	}
	return strings.TrimSpace(result.String())
}

// isEmoji verifica se o rune é um emoji
func isEmoji(r rune) bool {
	// Emoji ranges principais (Unicode 13.0+)
	return (r >= 0x1F300 && r <= 0x1F9FF) || // Emoticons, Symbols
		(r >= 0x1F1E6 && r <= 0x1F1FF) || // Regional Indicator Symbols (flags)
		(r >= 0x2600 && r <= 0x26FF) || // Misc Symbols
		(r >= 0x2700 && r <= 0x27BF) || // Dingbats
		(r == 0xFE0F) || (r == 0x200D) // Variation Selector, Zero Width Joiner
}
