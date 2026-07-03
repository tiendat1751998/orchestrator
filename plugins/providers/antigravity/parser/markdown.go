// Package parser implements output formats normalizers and structured text extractions.
package parser

import (
	"strings"
)

// ParseMarkdown cleans and extracts content from raw markdown output.
// It normalizes spacing and filters out sentinel command blocks.
func ParseMarkdown(input string) (string, error) {
	lines := strings.Split(input, "\n")
	var cleanedLines []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Filter out internal delimiter markers
		if strings.Contains(trimmed, "---END---") {
			continue
		}

		// Filter out backtick formatting fences if they are alone
		if strings.HasPrefix(trimmed, "```") && len(trimmed) <= 6 {
			continue
		}

		cleanedLines = append(cleanedLines, line)
	}

	result := strings.Join(cleanedLines, "\n")
	return strings.TrimSpace(result), nil
}

// ExtractCodeBlock searches and returns the contents of the first code block matching target language.
func ExtractCodeBlock(input, lang string) string {
	lines := strings.Split(input, "\n")
	var block []string
	inBlock := false

	lowerLang := strings.ToLower(lang)

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "```") {
			if inBlock {
				break // End of targeted block reached
			}
			blockLang := strings.ToLower(strings.TrimPrefix(trimmed, "```"))
			if blockLang == lowerLang || (lang == "" && blockLang == "") {
				inBlock = true
				continue
			}
		}

		if inBlock {
			block = append(block, line)
		}
	}

	if !inBlock {
		return ""
	}

	return strings.Join(block, "\n")
}
