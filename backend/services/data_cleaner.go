package services

import (
	"regexp"
	"strings"
)

type DataCleaner struct {
	multiSpaceRegex *regexp.Regexp
}

func NewDataCleaner() *DataCleaner {
	return &DataCleaner{
		multiSpaceRegex: regexp.MustCompile(`\s+`),
	}
}

// CleanText normalizes text by removing extra spaces, special characters, and standardizing casing
func (c *DataCleaner) CleanText(text string) string {
	// Trim leading and trailing spaces
	text = strings.TrimSpace(text)

	// Remove special characters (arrows, bullets, etc.) but keep letters, numbers, spaces, and common punctuation
	var builder strings.Builder
	for _, ch := range text {
		// Keep alphanumeric, spaces, hyphens, apostrophes, and periods
		if (ch >= 'a' && ch <= 'z') || 
		   (ch >= 'A' && ch <= 'Z') || 
		   (ch >= '0' && ch <= '9') || 
		   ch == ' ' || ch == '-' || ch == '\'' || ch == '.' || ch == '&' {
			builder.WriteRune(ch)
		}
	}
	text = builder.String()

	// Replace multiple spaces with single space
	text = c.multiSpaceRegex.ReplaceAllString(text, " ")

	// Trim again after cleaning
	text = strings.TrimSpace(text)

	// Convert to title case for consistency
	text = toTitleCase(text)

	return text
}

func toTitleCase(s string) string {
    words := strings.Fields(s)
    for i, word := range words {
        if len(word) > 0 {
            words[i] = strings.ToUpper(string(word[0])) + strings.ToLower(word[1:])
        }
    }
	return strings.Join(words, " ")
}