package utils

import "strings"

func IsUrl(inputPath string) bool {
	return strings.HasPrefix(inputPath, "http://") || strings.HasPrefix(inputPath, "https://")
}
