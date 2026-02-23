package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/hmans/beans/internal/ui"
)

// openSortPickerMsg requests opening the sort picker
type openSortPickerMsg struct {
	currentSortBy string
}

// closeSortPickerMsg is sent when the sort picker is cancelled
type closeSortPickerMsg struct{}

// sortSelectedMsg is sent when a sort option is selected
type sortSelectedMsg struct {
	sortBy string
}

// sortOption represents a single sort option in the picker
type sortOption struct {
	key         string // shortcut key (d, c, u, s, p, i)
	sortBy      string // value for SortOptions.SortBy ("" for default)
	label       string // display label
	description string // short description
}

var sortOptions = []sortOption{
	{key: "d", sortBy: "", label: "Default", description: "status > priority > type > title"},
	{key: "c", sortBy: "created", label: "Created", description: "newest first"},
	{key: "u", sortBy: "updated", label: "Updated", description: "recently updated first"},
	{key: "s", sortBy: "status", label: "Status", description: "by status order"},
	{key: "p", sortBy: "priority", label: "Priority", description: "by priority order"},
	{key: "i", sortBy: "id", label: "ID", description: "alphabetical by ID"},
}

// sortPickerModel is the model for the sort picker overlay
type sortPickerModel struct {
	cursor        int
	currentSortBy string
	width         int
	height        int
}

func newSortPickerModel(currentSortBy string, width, height int) sortPickerModel {
	// Set cursor to current sort option
	cursor := 0
	for i, opt := range sortOptions {
		if opt.sortBy == currentSortBy {
			cursor = i
			break
		}
	}
	return sortPickerModel{
		cursor:        cursor,
		currentSortBy: currentSortBy,
		width:         width,
		height:        height,
	}
}

func (m sortPickerModel) Init() tea.Cmd {
	return nil
}

func (m sortPickerModel) Update(msg tea.Msg) (sortPickerModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "S":
			return m, func() tea.Msg { return closeSortPickerMsg{} }
		case "enter":
			opt := sortOptions[m.cursor]
			return m, func() tea.Msg { return sortSelectedMsg{sortBy: opt.sortBy} }
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(sortOptions)-1 {
				m.cursor++
			}
		default:
			// Check shortcut keys
			key := msg.String()
			for _, opt := range sortOptions {
				if key == opt.key {
					return m, func() tea.Msg { return sortSelectedMsg{sortBy: opt.sortBy} }
				}
			}
		}
	}

	return m, nil
}

func (m sortPickerModel) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	modalWidth := max(40, min(50, m.width*40/100))

	// Title
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(ui.ColorPrimary).
		Render("Sort By")

	// Render options
	var lines []string
	for i, opt := range sortOptions {
		// Cursor indicator
		var cursor string
		if i == m.cursor {
			cursor = lipgloss.NewStyle().Foreground(ui.ColorPrimary).Bold(true).Render("▌") + " "
		} else {
			cursor = "  "
		}

		// Shortcut key
		keyStr := helpKeyStyle.Render(opt.key)

		// Label
		labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#fff"))
		labelStr := labelStyle.Render(opt.label)

		// Description
		descStr := ui.Muted.Render(opt.description)

		// Current indicator
		var currentStr string
		if opt.sortBy == m.currentSortBy {
			currentStr = lipgloss.NewStyle().Foreground(ui.ColorSuccess).Render(" (current)")
		}

		lines = append(lines, fmt.Sprintf("%s%s  %s  %s%s", cursor, keyStr, labelStr, descStr, currentStr))
	}

	content := title + "\n\n" + strings.Join(lines, "\n") + "\n\n"

	// Footer
	footer := helpKeyStyle.Render("d/c/u/s/p/i") + " " + helpStyle.Render("select") + "  " +
		helpKeyStyle.Render("enter") + " " + helpStyle.Render("confirm") + "  " +
		helpKeyStyle.Render("esc") + " " + helpStyle.Render("cancel")

	content += footer

	border := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ui.ColorPrimary).
		Padding(1, 2).
		Width(modalWidth)

	return border.Render(content)
}

// ModalView returns the sort picker as a centered modal on top of the background
func (m sortPickerModel) ModalView(bgView string, fullWidth, fullHeight int) string {
	modal := m.View()
	return overlayModal(bgView, modal, fullWidth, fullHeight)
}

// sortLabel returns a display label for a sort field value
func sortLabel(sortBy string) string {
	for _, opt := range sortOptions {
		if opt.sortBy == sortBy {
			return strings.ToLower(opt.label)
		}
	}
	return "default"
}
