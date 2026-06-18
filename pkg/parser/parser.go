package parser

import (
	"strings"
)

type ParsedPage struct {
	GoLogic string
	HTML    string
	Scripts string
	Styles  string
}

// extracts --go block from .spidey file
func Parse(content string) (*ParsedPage, error) {
	page := &ParsedPage{}

	if strings.HasPrefix(content, "---go") {
		parts := strings.SplitN(content, "---", 3)
		if len(parts) >= 3 {
			page.GoLogic = strings.TrimSpace(parts[1][2:]) // Remove 'go'
			content = parts[2]                             // Remaining content
		}
	}

	// Extract Styles
	if styleStart := strings.Index(content, "<style>"); styleStart != -1 {
		if styleEnd := strings.Index(content, "</style>"); styleEnd != -1 {
			page.Styles = content[styleStart+7 : styleEnd]
			content = content[:styleStart] + content[styleEnd+8:]
		}
	}

	if scriptStart := strings.Index(content, `<script data-island`); scriptStart != -1 {
		if scriptEnd := strings.Index(content, "</script>"); scriptEnd != -1 {
			// Extract everything inside the script tag
			fullScriptTag := content[scriptStart : scriptEnd+9]
			page.Scripts = fullScriptTag
			content = strings.Replace(content, fullScriptTag, "", 1)
		}
	}

	page.HTML = strings.TrimSpace(content)

	return page, nil
}
