package node

import "strings"

func countPrefix(text string, prefix rune) int {
	count := 0
	for char := range text {
		if char == '#' {
			count += 1
		} else {
			return count
		}
	}
	return count
}

func hasPrefix(block string, prefixes []string) bool {
	if prefixes == nil || len(prefixes) == 0 {
		return true
	}

	for _, prefix := range prefixes {
		if strings.HasPrefix(block, prefix) {
			return true
		}
	}
	return false
}

func trimPrefix(block string, prefixes []string) string {
	block = strings.TrimSpace(block)
	for _, prefix := range prefixes {
		block = strings.TrimPrefix(block, prefix)
	}
	return block
}
