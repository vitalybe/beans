---
# beans-fqsb
title: use [[ ]] for relationships between beans to allow interopability with obsidian
status: completed
type: task
priority: normal
created_at: 2026-02-13T12:46:09Z
updated_at: 2026-02-13T14:25:38Z
---

## Tasks\n\n- [x] Add wrapWikilink/unwrapWikilink helpers\n- [x] Modify Parse() to unwrap wikilinks\n- [x] Modify Render() to wrap wikilinks\n- [x] Update tests\n- [x] Run tests

## Summary of Changes

Added Obsidian-compatible wikilink syntax (`[[bean-id]]`) for relationship fields in YAML frontmatter.

### Changes
- Added `wrapWikilink`/`unwrapWikilink` helper functions in `internal/bean/bean.go`
- Modified `Parse()` to strip `[[` `]]` from parent, blocking, and blocked_by fields (backward compatible)
- Modified `Render()` to wrap relationship IDs in `[[` `]]`
- Added tests for wikilink parsing, helpers, and backward compatibility
- Updated existing render tests to expect wikilink-wrapped output

### Migration
- Read: supports both old (plain ID) and new (`[[ID]]`) formats
- Write: always outputs new `[[ID]]` format
- Files migrate gradually as they are updated
