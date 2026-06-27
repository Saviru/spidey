package parser

import (
	"crypto/md5"
	"fmt"
	"strings"
)

type ParsedPage struct {
	GoLogic string
	HTML    string
	Scripts string
	Styles  string
}

// prepends [data-spidey="hash"] to all selectors
func ScopeCSS(css, scopeAttr string) string {
	var result strings.Builder
	var sel strings.Builder
	depth := 0
	keyframesDepth := -1

	for i := 0; i < len(css); i++ {
		c := css[i]
		if c == '{' {
			s := strings.TrimSpace(sel.String())
			sel.Reset()

			if strings.HasPrefix(s, "@") {
				if strings.HasPrefix(s, "@keyframes") {
					keyframesDepth = depth
				}
				result.WriteString(s + " {\n")
			} else {
				if keyframesDepth != -1 && depth > keyframesDepth {
					// Not scoping inside keyframes
					result.WriteString(s + " {\n")
				} else {
					parts := strings.Split(s, ",")
					for j, p := range parts {
						parts[j] = "[" + scopeAttr + "] " + strings.TrimSpace(p)
					}
					result.WriteString(strings.Join(parts, ", ") + " {\n")
				}
			}
			depth++
		} else if c == '}' {
			depth--
			if keyframesDepth != -1 && depth <= keyframesDepth {
				keyframesDepth = -1
			}
			s := sel.String()
			sel.Reset()
			result.WriteString(s + "}\n")
		} else {
			sel.WriteByte(c)
		}
	}
	result.WriteString(sel.String())
	return result.String()
}

// Extracts .spidey file and scopes CSS/HTML
func Parse(componentName string, content string) (*ParsedPage, error) {
	page := &ParsedPage{}

	if strings.HasPrefix(content, "---go") {
		parts := strings.SplitN(content, "---", 3)
		if len(parts) >= 3 {
			page.GoLogic = strings.TrimSpace(parts[1][2:]) // Remove 'go'
			content = parts[2]                             // Remaining content
		}
	}

	hash := fmt.Sprintf("data-spidey-%x", md5.Sum([]byte(componentName)))[:20] // e.g. data-spidey-1a2b3c4d...

	// Extract Styles
	if styleStart := strings.Index(content, "<style>"); styleStart != -1 {
		if styleEnd := strings.Index(content, "</style>"); styleEnd != -1 {
			rawStyles := content[styleStart+7 : styleEnd]
			page.Styles = ScopeCSS(rawStyles, hash)
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

	htmlContent := strings.TrimSpace(content)
	if page.Styles != "" {
		page.HTML = fmt.Sprintf("<div %s style=\"display: contents;\">\n%s\n</div>", hash, htmlContent)
	} else {
		page.HTML = htmlContent
	}

	return page, nil
}
