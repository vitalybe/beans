package bean

import (
	"strings"

	gonanoid "github.com/matoous/go-nanoid/v2"
)

const idAlphabet = "0123456789abcdefghijklmnopqrstuvwxyz"

// NewID generates a new NanoID for a bean with an optional prefix and configurable length.
func NewID(prefix string, length int) string {
	id, err := gonanoid.Generate(idAlphabet, length)
	if err != nil {
		panic(err) // should never happen with valid alphabet
	}
	return prefix + id
}

// ParseFilename extracts the ID and optional slug from a bean filename.
// Supports multiple formats for backward compatibility:
//   - ID only: "beans-z5r9.md" -> ("beans-z5r9", "")
//   - Double-dash format: "beans-z5r9--user-registration.md" -> ("beans-z5r9", "user-registration")
//   - Dot format: "f7g.user-registration.md" -> ("f7g", "user-registration")
func ParseFilename(name string) (id, slug string) {
	// Remove .md extension
	name = strings.TrimSuffix(name, ".md")

	// Try double-dash format: id--slug
	if idx := strings.Index(name, "--"); idx > 0 {
		return name[:idx], name[idx+2:]
	}

	// Try dot format: id.slug (but not plain "id" which has no dot after .md removal)
	if idx := strings.Index(name, "."); idx > 0 {
		return name[:idx], name[idx+1:]
	}

	// ID only (no slug)
	return name, ""
}

// BuildFilename constructs a filename from an ID.
func BuildFilename(id string) string {
	return id + ".md"
}
