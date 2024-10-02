//go:build generate
// +build generate

package main

import (
	"bufio"
	"os"
	"regexp"
	"strings"

	"github.com/charmbracelet/log"
)

func main() {
	extractCSSSelectors()
}

func extractCSSSelectors() {
	// Read the CSS file
	content, err := os.ReadFile("frontend/public/assets/css/kiosk.css")
	if err != nil {
		log.Error("reading file", "err", err)
		return
	}

	scanner := bufio.NewScanner(strings.NewReader(string(content)))
	var result strings.Builder

	selectorOrCommentRegex := regexp.MustCompile(`^([^{]+{$|/\*|\*/|[^:]+:[^;]+;)`)
	skipBlockRegex := regexp.MustCompile(`^(@media|@font-face|@keyframes)`)

	inSkipBlock := 0

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if skipBlockRegex.MatchString(line) {
			inSkipBlock++
		} else if inSkipBlock > 0 {
			if strings.Contains(line, "{") {
				inSkipBlock++
			} else if strings.Contains(line, "}") {
				inSkipBlock--
			}
		} else if selectorOrCommentRegex.MatchString(line) {
			if strings.HasPrefix(line, "/*") || strings.HasPrefix(line, "*/") {
				result.WriteString(line + "\n")
			} else if strings.HasSuffix(line, "{") {
				selector := strings.TrimSuffix(line, "{")
				result.WriteString(strings.TrimSpace(selector) + "{}\n")
			}
		}
	}

	if err := scanner.Err(); err != nil {
		log.Error("scanning file", "err", err)
		return
	}

	// Save the extracted content to custom.example.css
	err = os.WriteFile("custom.example.css", []byte(result.String()), 0644)
	if err != nil {
		log.Error("writing to custom.example.css", "err", err)
		return
	}
}
