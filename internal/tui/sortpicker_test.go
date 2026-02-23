package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestSortPickerShortcutKeys(t *testing.T) {
	tests := []struct {
		key      string
		wantSort string
	}{
		{"d", ""},
		{"c", "created"},
		{"u", "updated"},
		{"s", "status"},
		{"p", "priority"},
		{"i", "id"},
	}

	for _, tt := range tests {
		t.Run("key_"+tt.key, func(t *testing.T) {
			m := newSortPickerModel("", 80, 24)
			_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(tt.key)})
			if cmd == nil {
				t.Fatal("expected command from shortcut key")
			}
			msg := cmd()
			sel, ok := msg.(sortSelectedMsg)
			if !ok {
				t.Fatalf("expected sortSelectedMsg, got %T", msg)
			}
			if sel.sortBy != tt.wantSort {
				t.Errorf("got sortBy %q, want %q", sel.sortBy, tt.wantSort)
			}
		})
	}
}

func TestSortPickerCursorNavigation(t *testing.T) {
	m := newSortPickerModel("", 80, 24)

	// Should start at 0 (default)
	if m.cursor != 0 {
		t.Fatalf("expected initial cursor 0, got %d", m.cursor)
	}

	// Move down
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	if m.cursor != 1 {
		t.Errorf("after down: got cursor %d, want 1", m.cursor)
	}

	// Move down with j
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	if m.cursor != 2 {
		t.Errorf("after j: got cursor %d, want 2", m.cursor)
	}

	// Move up
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	if m.cursor != 1 {
		t.Errorf("after up: got cursor %d, want 1", m.cursor)
	}

	// Move up with k
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
	if m.cursor != 0 {
		t.Errorf("after k: got cursor %d, want 0", m.cursor)
	}

	// Can't go above 0
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	if m.cursor != 0 {
		t.Errorf("after up at top: got cursor %d, want 0", m.cursor)
	}

	// Can't go below last item
	for i := 0; i < 10; i++ {
		m, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	}
	if m.cursor != len(sortOptions)-1 {
		t.Errorf("after many downs: got cursor %d, want %d", m.cursor, len(sortOptions)-1)
	}
}

func TestSortPickerInitialCursor(t *testing.T) {
	tests := []struct {
		currentSort string
		wantCursor  int
	}{
		{"", 0},         // default
		{"created", 1},  // second option
		{"updated", 2},  // third option
		{"status", 3},   // fourth option
		{"priority", 4}, // fifth option
		{"id", 5},       // sixth option
	}

	for _, tt := range tests {
		t.Run("sort_"+tt.currentSort, func(t *testing.T) {
			m := newSortPickerModel(tt.currentSort, 80, 24)
			if m.cursor != tt.wantCursor {
				t.Errorf("got cursor %d, want %d", m.cursor, tt.wantCursor)
			}
		})
	}
}

func TestSortPickerEnterConfirms(t *testing.T) {
	m := newSortPickerModel("", 80, 24)

	// Move to "updated" (index 2)
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected command from enter")
	}
	msg := cmd()
	sel, ok := msg.(sortSelectedMsg)
	if !ok {
		t.Fatalf("expected sortSelectedMsg, got %T", msg)
	}
	if sel.sortBy != "updated" {
		t.Errorf("got sortBy %q, want %q", sel.sortBy, "updated")
	}
}

func TestSortPickerEscCloses(t *testing.T) {
	m := newSortPickerModel("", 80, 24)

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if cmd == nil {
		t.Fatal("expected command from esc")
	}
	msg := cmd()
	if _, ok := msg.(closeSortPickerMsg); !ok {
		t.Fatalf("expected closeSortPickerMsg, got %T", msg)
	}
}

func TestSortPickerSCloses(t *testing.T) {
	m := newSortPickerModel("", 80, 24)

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("S")})
	if cmd == nil {
		t.Fatal("expected command from S")
	}
	msg := cmd()
	if _, ok := msg.(closeSortPickerMsg); !ok {
		t.Fatalf("expected closeSortPickerMsg, got %T", msg)
	}
}

func TestSortLabel(t *testing.T) {
	tests := []struct {
		sortBy string
		want   string
	}{
		{"", "default"},
		{"created", "created"},
		{"updated", "updated"},
		{"status", "status"},
		{"priority", "priority"},
		{"id", "id"},
		{"unknown", "default"},
	}

	for _, tt := range tests {
		t.Run(tt.sortBy, func(t *testing.T) {
			got := sortLabel(tt.sortBy)
			if got != tt.want {
				t.Errorf("sortLabel(%q) = %q, want %q", tt.sortBy, got, tt.want)
			}
		})
	}
}
