---
# beans-2p05
title: Add --sort and --sortr flags to beans list and beans tui
status: completed
type: feature
priority: normal
created_at: 2026-02-15T11:45:22Z
updated_at: 2026-02-15T11:49:27Z
---

Add reverse sort (--sortr) flag to beans list, and add both --sort and --sortr flags to beans tui. Extract sortBeans to shared bean.SortBeans function.

## Summary of Changes\n\n- Extracted sorting logic from `cmd/list.go` into a shared `bean.SortBeans()` function in `internal/bean/sort.go`\n- Added `SortConfig` interface so sorting doesn't depend on the config package directly\n- Added `--sortr` flag to `beans list` for reverse sorting\n- Added `--sort` and `--sortr` flags to `beans tui`\n- Made `--sort` and `--sortr` mutually exclusive in both commands\n- Updated all tests and added reverse sort test cases
