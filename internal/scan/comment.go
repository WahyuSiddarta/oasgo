package scan

import (
	"go/ast"
	"strings"
)

// CommentLines returns normalized doc comment lines without Go comment markers.
func CommentLines(group *ast.CommentGroup) []string {
	if group == nil {
		return nil
	}

	lines := make([]string, 0, len(group.List))
	for _, comment := range group.List {
		text := strings.TrimPrefix(comment.Text, "//")
		text = strings.TrimPrefix(text, " ")
		lines = append(lines, text)
	}

	return lines
}
