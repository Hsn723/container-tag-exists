package pkg

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	repoReplacementPattern = regexp.MustCompile(`[.:-]`)
)

// ExtractRegistryURL extracts a registry URL from an image name.
func ExtractRegistryURL(image string) (string, error) {
	frag := strings.Split(image, "/")
	if len(frag) < 2 {
		return "", fmt.Errorf("malformed image name %q", image)
	}
	return frag[0], nil
}

// ExtractImagePath extracts the image path from an image name.
func ExtractImagePath(image string) (string, error) {
	frag := strings.Split(image, "/")
	if len(frag) < 2 {
		return "", fmt.Errorf("malformed image name %q", image)
	}
	return strings.Join(frag[1:], "/"), nil
}

// NormalizeRegistryName converts the registry URL into a normalized name.
func NormalizeRegistryName(url string) string {
	return strings.ToUpper(repoReplacementPattern.ReplaceAllString(url, "_"))
}
