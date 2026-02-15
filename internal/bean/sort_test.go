package bean

import (
	"testing"
	"time"
)

// mockSortConfig implements SortConfig for testing
type mockSortConfig struct {
	statuses   []string
	priorities []string
	types      []string
}

func (m mockSortConfig) StatusNames() []string   { return m.statuses }
func (m mockSortConfig) PriorityNames() []string  { return m.priorities }
func (m mockSortConfig) TypeNames() []string      { return m.types }

func TestSortBeans(t *testing.T) {
	cfg := mockSortConfig{
		statuses:   []string{"in-progress", "todo", "draft", "completed", "scrapped"},
		priorities: []string{"critical", "high", "normal", "low", "deferred"},
		types:      []string{"milestone", "epic", "bug", "feature", "task"},
	}

	t.Run("sort by id", func(t *testing.T) {
		beans := []*Bean{
			{ID: "c3"},
			{ID: "a1"},
			{ID: "b2"},
		}
		SortBeans(beans, "id", false, cfg)
		if beans[0].ID != "a1" || beans[1].ID != "b2" || beans[2].ID != "c3" {
			t.Errorf("got [%s, %s, %s], want [a1, b2, c3]", beans[0].ID, beans[1].ID, beans[2].ID)
		}
	})

	t.Run("sort by id reversed", func(t *testing.T) {
		beans := []*Bean{
			{ID: "c3"},
			{ID: "a1"},
			{ID: "b2"},
		}
		SortBeans(beans, "id", true, cfg)
		if beans[0].ID != "c3" || beans[1].ID != "b2" || beans[2].ID != "a1" {
			t.Errorf("got [%s, %s, %s], want [c3, b2, a1]", beans[0].ID, beans[1].ID, beans[2].ID)
		}
	})

	t.Run("sort by created", func(t *testing.T) {
		now := time.Now()
		earlier := now.Add(-1 * time.Hour)
		evenEarlier := now.Add(-2 * time.Hour)
		beans := []*Bean{
			{ID: "old", CreatedAt: &evenEarlier},
			{ID: "new", CreatedAt: &now},
			{ID: "mid", CreatedAt: &earlier},
		}
		SortBeans(beans, "created", false, cfg)
		if beans[0].ID != "new" || beans[1].ID != "mid" || beans[2].ID != "old" {
			t.Errorf("got [%s, %s, %s], want [new, mid, old]", beans[0].ID, beans[1].ID, beans[2].ID)
		}
	})

	t.Run("sort by created reversed", func(t *testing.T) {
		now := time.Now()
		earlier := now.Add(-1 * time.Hour)
		evenEarlier := now.Add(-2 * time.Hour)
		beans := []*Bean{
			{ID: "old", CreatedAt: &evenEarlier},
			{ID: "new", CreatedAt: &now},
			{ID: "mid", CreatedAt: &earlier},
		}
		SortBeans(beans, "created", true, cfg)
		if beans[0].ID != "old" || beans[1].ID != "mid" || beans[2].ID != "new" {
			t.Errorf("got [%s, %s, %s], want [old, mid, new]", beans[0].ID, beans[1].ID, beans[2].ID)
		}
	})

	t.Run("sort by status", func(t *testing.T) {
		beans := []*Bean{
			{ID: "c1", Status: "completed"},
			{ID: "t1", Status: "todo"},
			{ID: "i1", Status: "in-progress"},
		}
		SortBeans(beans, "status", false, cfg)
		expected := []string{"i1", "t1", "c1"}
		for i, want := range expected {
			if beans[i].ID != want {
				t.Errorf("beans[%d].ID = %q, want %q", i, beans[i].ID, want)
			}
		}
	})

	t.Run("sort by status reversed", func(t *testing.T) {
		beans := []*Bean{
			{ID: "c1", Status: "completed"},
			{ID: "t1", Status: "todo"},
			{ID: "i1", Status: "in-progress"},
		}
		SortBeans(beans, "status", true, cfg)
		expected := []string{"c1", "t1", "i1"}
		for i, want := range expected {
			if beans[i].ID != want {
				t.Errorf("beans[%d].ID = %q, want %q", i, beans[i].ID, want)
			}
		}
	})

	t.Run("sort by priority", func(t *testing.T) {
		beans := []*Bean{
			{ID: "low", Status: "todo", Priority: "low"},
			{ID: "crit", Status: "todo", Priority: "critical"},
			{ID: "high", Status: "todo", Priority: "high"},
		}
		SortBeans(beans, "priority", false, cfg)
		expected := []string{"crit", "high", "low"}
		for i, want := range expected {
			if beans[i].ID != want {
				t.Errorf("beans[%d].ID = %q, want %q", i, beans[i].ID, want)
			}
		}
	})

	t.Run("sort by priority reversed", func(t *testing.T) {
		beans := []*Bean{
			{ID: "low", Status: "todo", Priority: "low"},
			{ID: "crit", Status: "todo", Priority: "critical"},
			{ID: "high", Status: "todo", Priority: "high"},
		}
		SortBeans(beans, "priority", true, cfg)
		expected := []string{"low", "high", "crit"}
		for i, want := range expected {
			if beans[i].ID != want {
				t.Errorf("beans[%d].ID = %q, want %q", i, beans[i].ID, want)
			}
		}
	})

	t.Run("default sort (empty sortBy)", func(t *testing.T) {
		beans := []*Bean{
			{ID: "1", Status: "completed", Type: "bug"},
			{ID: "2", Status: "todo", Type: "feature"},
			{ID: "3", Status: "todo", Type: "bug"},
		}
		SortBeans(beans, "", false, cfg)
		// todo beans first (by type order: bug before feature), then completed
		expected := []string{"3", "2", "1"}
		for i, want := range expected {
			if beans[i].ID != want {
				t.Errorf("beans[%d].ID = %q, want %q", i, beans[i].ID, want)
			}
		}
	})

	t.Run("default sort reversed", func(t *testing.T) {
		beans := []*Bean{
			{ID: "1", Status: "completed", Type: "bug"},
			{ID: "2", Status: "todo", Type: "feature"},
			{ID: "3", Status: "todo", Type: "bug"},
		}
		SortBeans(beans, "", true, cfg)
		// Reversed default: completed first, then todo (feature before bug)
		expected := []string{"1", "2", "3"}
		for i, want := range expected {
			if beans[i].ID != want {
				t.Errorf("beans[%d].ID = %q, want %q", i, beans[i].ID, want)
			}
		}
	})

	t.Run("empty slice", func(t *testing.T) {
		beans := []*Bean{}
		SortBeans(beans, "id", false, cfg)
		if len(beans) != 0 {
			t.Errorf("expected empty slice, got %d elements", len(beans))
		}
	})

	t.Run("single element reversed", func(t *testing.T) {
		beans := []*Bean{{ID: "only"}}
		SortBeans(beans, "id", true, cfg)
		if beans[0].ID != "only" {
			t.Errorf("got %q, want \"only\"", beans[0].ID)
		}
	})
}

func TestSortByStatusPriorityAndType(t *testing.T) {
	statusNames := []string{"draft", "todo", "in-progress", "completed"}
	priorityNames := []string{"critical", "high", "normal", "low", "deferred"}
	typeNames := []string{"bug", "feature", "task"}

	t.Run("sorts by status first", func(t *testing.T) {
		beans := []*Bean{
			{ID: "1", Title: "A", Status: "completed", Priority: "critical"},
			{ID: "2", Title: "B", Status: "todo", Priority: "low"},
			{ID: "3", Title: "C", Status: "draft", Priority: "high"},
		}

		SortByStatusPriorityAndType(beans, statusNames, priorityNames, typeNames)

		if beans[0].Status != "draft" {
			t.Errorf("First bean status = %q, want \"draft\"", beans[0].Status)
		}
		if beans[1].Status != "todo" {
			t.Errorf("Second bean status = %q, want \"todo\"", beans[1].Status)
		}
		if beans[2].Status != "completed" {
			t.Errorf("Third bean status = %q, want \"completed\"", beans[2].Status)
		}
	})

	t.Run("sorts by priority within same status", func(t *testing.T) {
		beans := []*Bean{
			{ID: "1", Title: "E Low", Status: "todo", Priority: "low"},
			{ID: "2", Title: "A Critical", Status: "todo", Priority: "critical"},
			{ID: "3", Title: "B High", Status: "todo", Priority: "high"},
			{ID: "4", Title: "C Normal", Status: "todo", Priority: "normal"},
			{ID: "5", Title: "D No Priority", Status: "todo", Priority: ""},
		}

		SortByStatusPriorityAndType(beans, statusNames, priorityNames, typeNames)

		// Order by priority: critical, high, normal (and empty), low, deferred
		// Within same priority, order by title alphabetically
		expectedOrder := []string{"A Critical", "B High", "C Normal", "D No Priority", "E Low"}
		for i, expected := range expectedOrder {
			if beans[i].Title != expected {
				t.Errorf("beans[%d].Title = %q, want %q", i, beans[i].Title, expected)
			}
		}
	})

	t.Run("empty priority treated as normal", func(t *testing.T) {
		beans := []*Bean{
			{ID: "1", Title: "Low", Status: "todo", Priority: "low"},
			{ID: "2", Title: "Empty", Status: "todo", Priority: ""},
			{ID: "3", Title: "Normal", Status: "todo", Priority: "normal"},
			{ID: "4", Title: "High", Status: "todo", Priority: "high"},
		}

		SortByStatusPriorityAndType(beans, statusNames, priorityNames, typeNames)

		// High should come first, then Normal and Empty (same priority level), then Low
		if beans[0].Title != "High" {
			t.Errorf("First bean = %q, want \"High\"", beans[0].Title)
		}
		if beans[3].Title != "Low" {
			t.Errorf("Last bean = %q, want \"Low\"", beans[3].Title)
		}
		// Empty and Normal should be adjacent (both at normal priority level)
		normalIdx, emptyIdx := -1, -1
		for i, b := range beans {
			if b.Title == "Normal" {
				normalIdx = i
			}
			if b.Title == "Empty" {
				emptyIdx = i
			}
		}
		if normalIdx != 1 && normalIdx != 2 {
			t.Errorf("Normal should be at index 1 or 2, got %d", normalIdx)
		}
		if emptyIdx != 1 && emptyIdx != 2 {
			t.Errorf("Empty should be at index 1 or 2, got %d", emptyIdx)
		}
	})

	t.Run("sorts by type after priority", func(t *testing.T) {
		beans := []*Bean{
			{ID: "1", Title: "Task", Status: "todo", Priority: "high", Type: "task"},
			{ID: "2", Title: "Bug", Status: "todo", Priority: "high", Type: "bug"},
			{ID: "3", Title: "Feature", Status: "todo", Priority: "high", Type: "feature"},
		}

		SortByStatusPriorityAndType(beans, statusNames, priorityNames, typeNames)

		if beans[0].Type != "bug" {
			t.Errorf("First bean type = %q, want \"bug\"", beans[0].Type)
		}
		if beans[1].Type != "feature" {
			t.Errorf("Second bean type = %q, want \"feature\"", beans[1].Type)
		}
		if beans[2].Type != "task" {
			t.Errorf("Third bean type = %q, want \"task\"", beans[2].Type)
		}
	})

	t.Run("sorts by title after type", func(t *testing.T) {
		beans := []*Bean{
			{ID: "1", Title: "Zebra", Status: "todo", Priority: "high", Type: "bug"},
			{ID: "2", Title: "Apple", Status: "todo", Priority: "high", Type: "bug"},
			{ID: "3", Title: "Mango", Status: "todo", Priority: "high", Type: "bug"},
		}

		SortByStatusPriorityAndType(beans, statusNames, priorityNames, typeNames)

		if beans[0].Title != "Apple" {
			t.Errorf("First bean title = %q, want \"Apple\"", beans[0].Title)
		}
		if beans[1].Title != "Mango" {
			t.Errorf("Second bean title = %q, want \"Mango\"", beans[1].Title)
		}
		if beans[2].Title != "Zebra" {
			t.Errorf("Third bean title = %q, want \"Zebra\"", beans[2].Title)
		}
	})

	t.Run("handles nil priority names gracefully", func(t *testing.T) {
		beans := []*Bean{
			{ID: "1", Title: "A", Status: "todo", Priority: "high"},
			{ID: "2", Title: "B", Status: "todo", Priority: ""},
		}

		// Should not panic with nil priorityNames
		SortByStatusPriorityAndType(beans, statusNames, nil, typeNames)

		// Both should be sorted by status, type, then title
		if beans[0].Title != "A" {
			t.Errorf("First bean title = %q, want \"A\"", beans[0].Title)
		}
	})
}

