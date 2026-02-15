package bean

import (
	"strings"
	"testing"
)

func TestParseFilename(t *testing.T) {
	tests := []struct {
		name         string
		filename     string
		expectedID   string
		expectedSlug string
	}{
		// ID only
		{"id only with md", "abc.md", "abc", ""},
		{"id only with prefix", "beans-z5r9.md", "beans-z5r9", ""},
		{"id only no extension", "abc", "abc", ""},

		// Double-dash format (backward compat)
		{"double-dash basic", "abc--my-slug.md", "abc", "my-slug"},
		{"double-dash with prefix", "beans-z5r9--add-unit-tests.md", "beans-z5r9", "add-unit-tests"},
		{"double-dash long slug", "xyz--this-is-a-longer-slug.md", "xyz", "this-is-a-longer-slug"},

		// Dot format (backward compat)
		{"dot format basic", "abc.my-slug.md", "abc", "my-slug"},
		{"dot format with prefix", "beans-z5r9.add-unit-tests.md", "beans-z5r9", "add-unit-tests"},

		// Edge cases
		{"empty string", "", "", ""},
		{"just md extension", ".md", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotID, gotSlug := ParseFilename(tt.filename)
			if gotID != tt.expectedID || gotSlug != tt.expectedSlug {
				t.Errorf("ParseFilename(%q) = (%q, %q), want (%q, %q)",
					tt.filename, gotID, gotSlug, tt.expectedID, tt.expectedSlug)
			}
		})
	}
}

func TestBuildFilename(t *testing.T) {
	tests := []struct {
		name     string
		id       string
		expected string
	}{
		{"simple id", "abc", "abc.md"},
		{"with prefix", "beans-z5r9", "beans-z5r9.md"},
		{"another id", "xyz", "xyz.md"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BuildFilename(tt.id)
			if got != tt.expected {
				t.Errorf("BuildFilename(%q) = %q, want %q",
					tt.id, got, tt.expected)
			}
		})
	}
}

func TestNewID(t *testing.T) {
	t.Run("length without prefix", func(t *testing.T) {
		id := NewID("", 4)
		if len(id) != 4 {
			t.Errorf("NewID(\"\", 4) length = %d, want 4", len(id))
		}
	})

	t.Run("length with prefix", func(t *testing.T) {
		id := NewID("beans-", 4)
		if len(id) != 10 { // "beans-" (6) + 4
			t.Errorf("NewID(\"beans-\", 4) length = %d, want 10", len(id))
		}
	})

	t.Run("prefix preserved", func(t *testing.T) {
		prefix := "myapp-"
		id := NewID(prefix, 4)
		if !strings.HasPrefix(id, prefix) {
			t.Errorf("NewID(%q, 4) = %q, should start with prefix", prefix, id)
		}
	})

	t.Run("uses valid alphabet", func(t *testing.T) {
		id := NewID("", 100) // generate long ID to test alphabet
		for _, r := range id {
			if !strings.ContainsRune(idAlphabet, r) {
				t.Errorf("NewID contains invalid character %q, should only use %q", r, idAlphabet)
			}
		}
	})

	t.Run("generates unique IDs", func(t *testing.T) {
		seen := make(map[string]bool)
		for i := 0; i < 100; i++ {
			id := NewID("", 8)
			if seen[id] {
				t.Errorf("NewID generated duplicate: %q", id)
			}
			seen[id] = true
		}
	})
}

func TestParseFilenameAndBuildFilenameRoundtrip(t *testing.T) {
	tests := []struct {
		name string
		id   string
	}{
		{"simple id", "abc"},
		{"with prefix", "beans-z5r9"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filename := BuildFilename(tt.id)
			gotID, _ := ParseFilename(filename)
			if gotID != tt.id {
				t.Errorf("Roundtrip failed: BuildFilename(%q) = %q, ParseFilename ID = %q",
					tt.id, filename, gotID)
			}
		})
	}
}
