package operationcomment

import "strings"

const Marker = "oasgo:operation"

// Block is a raw operation comment block extracted from Go doc comments.
type Block struct {
	Lines []string
}

// ExtractBlock returns the comment lines after the oasgo operation marker.
func ExtractBlock(lines []string) (Block, bool) {
	for i, line := range lines {
		if strings.TrimSpace(line) == Marker {
			return Block{Lines: lines[i+1:]}, true
		}
	}

	return Block{}, false
}
