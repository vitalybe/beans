package tui

import (
	"context"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/hmans/beans/internal/bean"
	"github.com/hmans/beans/internal/beancore"
	"github.com/hmans/beans/internal/config"
	"github.com/hmans/beans/internal/graph"
	"github.com/hmans/beans/internal/ui"
)

// parentSelectedMsg is sent when a parent is selected from the picker
type parentSelectedMsg struct {
	beanIDs  []string // the beans being modified
	parentID string   // the new parent ID (empty string to clear parent)
}

// closeParentPickerMsg is sent when the parent picker is cancelled
type closeParentPickerMsg struct{}

// parentItem wraps a bean to implement list.Item for the parent picker
type parentItem struct {
	bean *bean.Bean
	cfg  *config.Config
}

func (i parentItem) Title() string       { return i.bean.Title }
func (i parentItem) Description() string { return i.bean.ID }
func (i parentItem) FilterValue() string { return i.bean.Title + " " + i.bean.ID }

// clearParentItem is a special item to clear the parent
type clearParentItem struct{}

func (i clearParentItem) Title() string       { return "(No Parent)" }
func (i clearParentItem) Description() string { return "Clear the parent assignment" }
func (i clearParentItem) FilterValue() string { return "no parent clear none" }

// parentItemDelegate handles rendering of parent picker items
type parentItemDelegate struct {
	cfg *config.Config
}

func (d parentItemDelegate) Height() int                             { return 1 }
func (d parentItemDelegate) Spacing() int                            { return 0 }
func (d parentItemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

func (d parentItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	var cursor string
	if index == m.Index() {
		cursor = lipgloss.NewStyle().Foreground(ui.ColorPrimary).Bold(true).Render("▌") + " "
	} else {
		cursor = "  "
	}

	switch item := listItem.(type) {
	case clearParentItem:
		text := ui.Muted.Render(item.Title())
		fmt.Fprint(w, cursor+text)

	case parentItem:
		// Get colors from config
		colors := d.cfg.GetBeanColors(item.bean.Status, item.bean.Type, item.bean.Priority)

		// Format: [type] title (id)
		typeBadge := ui.RenderTypeText(item.bean.Type, colors.TypeColor)
		title := item.bean.Title
		if colors.IsArchive {
			title = ui.Muted.Render(title)
		}
		id := ui.Muted.Render(" (" + item.bean.ID + ")")

		fmt.Fprint(w, cursor+typeBadge+" "+title+id)
	}
}

// parentPickerModel is the model for the parent picker view
type parentPickerModel struct {
	list          list.Model
	filterInput   textinput.Model
	allItems      []list.Item // all items (unfiltered)
	beanIDs       []string    // the beans we're setting the parent for
	beanTitle     string      // display title (single title or "N selected beans")
	beanTypes     []string    // types of the beans (to filter eligible parents)
	currentParent string      // current parent ID (to highlight, only for single bean)
	width         int
	height        int
}

func newParentPickerModel(beanIDs []string, beanTitle string, beanTypes []string, currentParent string, resolver *graph.Resolver, cfg *config.Config, width, height int) parentPickerModel {
	// Get valid parent types - for multi-select, find types valid for ALL beans
	var validParentTypes []string
	for i, beanType := range beanTypes {
		typeParents := beancore.ValidParentTypes(beanType)
		if i == 0 {
			validParentTypes = typeParents
		} else {
			// Intersect with existing valid types
			validParentTypes = intersectStrings(validParentTypes, typeParents)
		}
	}

	// Fetch all beans and filter to eligible parents
	allBeans, _ := resolver.Query().Beans(context.Background(), nil)

	// Collect all descendants of all selected beans (to prevent cycles)
	allDescendants := make(map[string]bool)
	for _, beanID := range beanIDs {
		for descID := range collectDescendants(beanID, allBeans) {
			allDescendants[descID] = true
		}
	}

	// Create set of selected bean IDs for quick lookup
	selectedSet := make(map[string]bool)
	for _, id := range beanIDs {
		selectedSet[id] = true
	}

	// Filter to eligible parents:
	// 1. Must be of a valid parent type for ALL selected beans
	// 2. Must not be any of the selected beans
	// 3. Must not be a descendant of any selected bean (to prevent cycles)
	var eligibleBeans []*bean.Bean
	for _, b := range allBeans {
		// Skip selected beans
		if selectedSet[b.ID] {
			continue
		}
		// Skip descendants (would create cycle)
		if allDescendants[b.ID] {
			continue
		}
		// Check if type is valid
		isValidType := false
		for _, validType := range validParentTypes {
			if b.Type == validType {
				isValidType = true
				break
			}
		}
		if !isValidType {
			continue
		}
		eligibleBeans = append(eligibleBeans, b)
	}

	// Sort by type order (milestone > epic > feature), then by title
	typeNames := cfg.TypeNames()
	typeOrder := make(map[string]int)
	for i, t := range typeNames {
		typeOrder[t] = i
	}
	sort.Slice(eligibleBeans, func(i, j int) bool {
		// Primary: type order
		ti, tj := typeOrder[eligibleBeans[i].Type], typeOrder[eligibleBeans[j].Type]
		if ti != tj {
			return ti < tj
		}
		// Secondary: title (case-insensitive)
		return strings.ToLower(eligibleBeans[i].Title) < strings.ToLower(eligibleBeans[j].Title)
	})

	delegate := parentItemDelegate{cfg: cfg}

	// Build items list - start with "clear parent" option
	items := make([]list.Item, 0, len(eligibleBeans)+1)
	items = append(items, clearParentItem{})

	selectedIndex := 0 // default to "No Parent"
	for i, b := range eligibleBeans {
		items = append(items, parentItem{bean: b, cfg: cfg})
		// If this is the current parent, remember its index (+1 for the clear option)
		if b.ID == currentParent {
			selectedIndex = i + 1
		}
	}

	// Calculate modal dimensions (matching View() function)
	modalWidth := max(40, min(80, width*60/100))
	modalHeight := max(10, min(20, height*60/100))
	// List dimensions within modal (account for border, padding, subtitle, help)
	listWidth := modalWidth - 6  // border (2) + padding (4)
	listHeight := modalHeight - 7 // border (2) + subtitle (1) + help (1) + padding (3)

	// Account for filter input box (border 2 lines) in list height
	listHeight -= 2

	l := list.New(items, delegate, listWidth, listHeight)
	l.Title = "Select Parent"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(false)
	l.SetShowPagination(false)
	l.Styles.Title = listTitleStyle
	l.Styles.TitleBar = lipgloss.NewStyle().Padding(0, 0, 0, 0)

	// Select the current parent if set
	if selectedIndex > 0 && selectedIndex < len(items) {
		l.Select(selectedIndex)
	}

	// Set up filter text input
	ti := textinput.New()
	ti.Placeholder = "Type to filter..."
	ti.CharLimit = 100
	ti.Width = listWidth - 2
	ti.Focus()
	ti.PromptStyle = lipgloss.NewStyle().Foreground(ui.ColorPrimary)
	ti.TextStyle = lipgloss.NewStyle()
	ti.PlaceholderStyle = lipgloss.NewStyle().Foreground(ui.ColorMuted)
	ti.Prompt = ""

	return parentPickerModel{
		list:          l,
		filterInput:   ti,
		allItems:      items,
		beanIDs:       beanIDs,
		beanTitle:     beanTitle,
		beanTypes:     beanTypes,
		currentParent: currentParent,
		width:         width,
		height:        height,
	}
}

// intersectStrings returns the intersection of two string slices
func intersectStrings(a, b []string) []string {
	set := make(map[string]bool)
	for _, s := range a {
		set[s] = true
	}
	var result []string
	for _, s := range b {
		if set[s] {
			result = append(result, s)
		}
	}
	return result
}

// collectDescendants returns a set of all bean IDs that are descendants of the given bean
func collectDescendants(beanID string, allBeans []*bean.Bean) map[string]bool {
	descendants := make(map[string]bool)

	// Build parent->children map
	children := make(map[string][]string)
	for _, b := range allBeans {
		if b.Parent != "" {
			children[b.Parent] = append(children[b.Parent], b.ID)
		}
	}

	// BFS to collect all descendants
	queue := children[beanID]
	for len(queue) > 0 {
		childID := queue[0]
		queue = queue[1:]
		if !descendants[childID] {
			descendants[childID] = true
			queue = append(queue, children[childID]...)
		}
	}

	return descendants
}

func (m parentPickerModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m parentPickerModel) Update(msg tea.Msg) (parentPickerModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		// Recalculate modal dimensions
		modalWidth := max(40, min(80, msg.Width*60/100))
		modalHeight := max(10, min(20, msg.Height*60/100))
		listWidth := modalWidth - 6
		listHeight := modalHeight - 7 - 2 // -2 for filter input border
		m.list.SetSize(listWidth, listHeight)
		m.filterInput.Width = listWidth - 2

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			switch item := m.list.SelectedItem().(type) {
			case clearParentItem:
				return m, func() tea.Msg {
					return parentSelectedMsg{beanIDs: m.beanIDs, parentID: ""}
				}
			case parentItem:
				return m, func() tea.Msg {
					return parentSelectedMsg{beanIDs: m.beanIDs, parentID: item.bean.ID}
				}
			}
		case tea.KeyEscape:
			return m, func() tea.Msg {
				return closeParentPickerMsg{}
			}
		case tea.KeyUp, tea.KeyDown:
			var cmd tea.Cmd
			m.list, cmd = m.list.Update(msg)
			return m, cmd
		default:
			// Send all other keys to the text input for filtering
			var cmd tea.Cmd
			m.filterInput, cmd = m.filterInput.Update(msg)
			m.applyFilter()
			return m, cmd
		}
	}

	var cmds []tea.Cmd
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	cmds = append(cmds, cmd)
	m.filterInput, cmd = m.filterInput.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

// applyFilter updates the list items based on the current filter input value
func (m *parentPickerModel) applyFilter() {
	term := m.filterInput.Value()

	// clearParentItem is always visible; filter only the parentItems
	var parentItems []list.Item
	for _, item := range m.allItems {
		if _, ok := item.(clearParentItem); !ok {
			parentItems = append(parentItems, item)
		}
	}

	filtered := fuzzyFilterItems(term, parentItems)

	// Prepend the clearParentItem
	result := make([]list.Item, 0, len(filtered)+1)
	result = append(result, clearParentItem{})
	result = append(result, filtered...)

	m.list.SetItems(result)
}

func (m parentPickerModel) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	// For multi-select, don't show individual bean ID
	var beanID string
	if len(m.beanIDs) == 1 {
		beanID = m.beanIDs[0]
	}

	// Render filter input with a border
	modalWidth := max(40, min(80, m.width*60/100))
	filterBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ui.ColorMuted).
		Padding(0, 1).
		Width(modalWidth - 6). // border(2) + padding(4)
		Render(m.filterInput.View())

	return renderPickerModal(pickerModalConfig{
		Title:       "Select Parent",
		BeanTitle:   m.beanTitle,
		BeanID:      beanID,
		FilterInput: filterBox,
		ListContent: m.list.View(),
		Width:       m.width,
		WidthPct:    60,
		MaxWidth:    80,
	})
}

// ModalView returns the picker rendered as a centered modal overlay on top of the background
func (m parentPickerModel) ModalView(bgView string, fullWidth, fullHeight int) string {
	modal := m.View()
	return overlayModal(bgView, modal, fullWidth, fullHeight)
}
