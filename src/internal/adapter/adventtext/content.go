package adventtext

import (
	"embed"
	"fmt"
	"strings"
)

// Each file in content is one door's surprise: 01.md to 24.md. Which door shows which file
// differs per player. Edit a file and rebuild to change it.
//
//go:embed content/*.md
var contentFiles embed.FS

// Content returns the surprise numbered item+1.
func Content(item int) string {
	data, err := contentFiles.ReadFile(fmt.Sprintf("content/%02d.md", item+1))
	if err != nil {
		return "This Advent surprise could not be found."
	}
	return strings.TrimSpace(string(data))
}
