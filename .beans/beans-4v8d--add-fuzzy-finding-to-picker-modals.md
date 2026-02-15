---
# beans-4v8d
title: Add fuzzy finding to picker modals
status: completed
type: feature
priority: normal
created_at: 2026-02-15T11:34:33Z
updated_at: 2026-02-15T11:41:04Z
---

Replace built-in list filtering with always-on fuzzy search input in parent and blocking picker modals. Typing immediately filters items (fzf-style) instead of requiring '/' to activate filtering.

## Summary of Changes\n\nAdded always-on fuzzy search to parent and blocking picker modals:\n\n- Added `fuzzyFilterItems()` helper in `modal.go` using `list.DefaultFilter` (sahilm/fuzzy)\n- Added `FilterInput` and `HelpText` fields to `pickerModalConfig`\n- Parent picker: typing immediately filters, arrows navigate, Enter selects, Esc closes. "(No Parent)" always visible regardless of filter\n- Blocking picker: same fuzzy input, Space toggles items (not sent to textinput), toggled state preserved across filter changes\n- Disabled built-in `list.SetFilteringEnabled` in favor of custom keyboard routing\n- Removed "/ filter" from default help text
