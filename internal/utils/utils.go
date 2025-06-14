package utils

import (
	"regexp"
	"strings"
)

// SanitizeFilename removes or replaces characters that are problematic in filenames
func SanitizeFilename(filename string) string {
	// Replace problematic characters with underscores
	reg := regexp.MustCompile(`[<>:"/\\|?*]`)
	sanitized := reg.ReplaceAllString(filename, "_")
	
	// Replace multiple consecutive underscores with single underscore
	reg2 := regexp.MustCompile(`_+`)
	sanitized = reg2.ReplaceAllString(sanitized, "_")
	
	// Trim underscores from beginning and end
	sanitized = strings.Trim(sanitized, "_")
	
	// Limit length
	if len(sanitized) > 100 {
		sanitized = sanitized[:100]
	}
	
	// Ensure it's not empty
	if sanitized == "" {
		sanitized = "untitled"
	}
	
	return sanitized
}

// TruncateString truncates a string to a maximum length
func TruncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// ExtractCodeBlocks finds code blocks in text
func ExtractCodeBlocks(text string) []string {
	// Find code blocks marked with ```
	reg := regexp.MustCompile("```[\\s\\S]*?```")
	matches := reg.FindAllString(text, -1)
	
	// Also find inline code marked with `
	reg2 := regexp.MustCompile("`[^`]+`")
	inlineMatches := reg2.FindAllString(text, -1)
	
	return append(matches, inlineMatches...)
}

// Contains checks if a slice contains a string
func Contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}