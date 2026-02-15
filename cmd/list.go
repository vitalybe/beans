package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/hmans/beans/internal/bean"
	"github.com/hmans/beans/internal/graph"
	"github.com/hmans/beans/internal/graph/model"
	"github.com/hmans/beans/internal/output"
	"github.com/hmans/beans/internal/ui"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var (
	listJSON       bool
	listSearch     string
	listStatus     []string
	listNoStatus   []string
	listType       []string
	listNoType     []string
	listPriority   []string
	listNoPriority []string
	listTag        []string
	listNoTag      []string
	listHasParent   bool
	listNoParent    bool
	listParentID    string
	listHasBlocking bool
	listNoBlocking  bool
	listIsBlocked   bool
	listReady      bool
	listQuiet      bool
	listSort       string
	listSortr      string
	listFull       bool
	listShowDone   bool
)

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List all beans",
	Long: `Lists all beans in the .beans directory.

Search Syntax (--search/-S):
  The search flag supports Bleve query string syntax:

  login          Exact term match
  login~         Fuzzy match (1 edit distance, finds "loggin", "logins")
  login~2        Fuzzy match (2 edit distance)
  log*           Wildcard prefix match
  "user login"   Exact phrase match
  user AND login Both terms required
  user OR login  Either term matches
  slug:auth      Search only in slug field
  title:login    Search only in title field
  body:auth      Search only in body field`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// --sort and --sortr are mutually exclusive
		if listSort != "" && listSortr != "" {
			return fmt.Errorf("--sort and --sortr are mutually exclusive")
		}

		// Build GraphQL filter from CLI flags
		filter := &model.BeanFilter{
			Status:          listStatus,
			ExcludeStatus:   listNoStatus,
			Type:            listType,
			ExcludeType:     listNoType,
			Priority:        listPriority,
			ExcludePriority: listNoPriority,
			Tags:            listTag,
			ExcludeTags:     listNoTag,
		}

		// Add search filter if provided
		if listSearch != "" {
			filter.Search = &listSearch
		}

		// Add parent/blocks filters
		if listHasParent {
			filter.HasParent = &listHasParent
		}
		if listNoParent {
			filter.NoParent = &listNoParent
		}
		if listParentID != "" {
			filter.ParentID = &listParentID
		}
		if listHasBlocking {
			filter.HasBlocking = &listHasBlocking
		}
		if listNoBlocking {
			filter.NoBlocking = &listNoBlocking
		}
		// --ready and --is-blocked are mutually exclusive
		if listReady && listIsBlocked {
			return fmt.Errorf("--ready and --is-blocked are mutually exclusive")
		}

		if listIsBlocked {
			filter.IsBlocked = &listIsBlocked
		}

		// Default: hide completed/scrapped beans unless --show-done or explicit --status is set
		if !listShowDone && len(listStatus) == 0 && !listReady {
			filter.ExcludeStatus = append(filter.ExcludeStatus, "completed", "scrapped")
		}

		// --ready: beans available to start (not blocked, excludes in-progress/completed/scrapped/draft)
		if listReady {
			isBlocked := false
			filter.IsBlocked = &isBlocked
			filter.ExcludeStatus = append(filter.ExcludeStatus, "in-progress", "completed", "scrapped", "draft")
		}

		// Execute query via GraphQL resolver
		resolver := &graph.Resolver{Core: core}
		beans, err := resolver.Query().Beans(context.Background(), filter)
		if err != nil {
			return fmt.Errorf("querying beans: %w", err)
		}

		// Sort beans
		sortBy := listSort
		reverse := false
		if listSortr != "" {
			sortBy = listSortr
			reverse = true
		}
		bean.SortBeans(beans, sortBy, reverse, cfg)

		// JSON output (flat list)
		if listJSON {
			if !listFull {
				for _, b := range beans {
					b.Body = ""
				}
			}
			return output.SuccessMultiple(beans)
		}

		// Quiet mode: just IDs (flat)
		if listQuiet {
			for _, b := range beans {
				fmt.Println(b.ID)
			}
			return nil
		}

		// Default: tree view
		// We need all beans to find ancestors for context
		allBeans, err := resolver.Query().Beans(context.Background(), nil)
		if err != nil {
			return fmt.Errorf("querying all beans for tree: %w", err)
		}

		// Create sort function for tree building
		sortFn := func(b []*bean.Bean) {
			bean.SortBeans(b, sortBy, reverse, cfg)
		}

		// Build tree
		tree := ui.BuildTree(beans, allBeans, sortFn)

		if len(tree) == 0 {
			fmt.Println(ui.Muted.Render("No beans found. Create one with: beans new <title>"))
			return nil
		}

		// Calculate max ID width from all beans in tree
		maxIDWidth := 2
		for _, b := range allBeans {
			if len(b.ID) > maxIDWidth {
				maxIDWidth = len(b.ID)
			}
		}
		maxIDWidth += 2

		// Check if any beans have tags
		hasTags := false
		for _, b := range beans {
			if len(b.Tags) > 0 {
				hasTags = true
				break
			}
		}

		// Detect terminal width (default to 80 if not a terminal)
		termWidth := 80
		if w, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil && w > 0 {
			termWidth = w
		}

		fmt.Print(ui.RenderTree(tree, cfg, maxIDWidth, hasTags, termWidth))
		return nil
	},
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func init() {
	listCmd.Flags().BoolVar(&listJSON, "json", false, "Output as JSON")
	listCmd.Flags().StringVarP(&listSearch, "search", "S", "", "Full-text search in title and body")
	listCmd.Flags().StringArrayVarP(&listStatus, "status", "s", nil, "Filter by status (can be repeated)")
	listCmd.Flags().StringArrayVar(&listNoStatus, "no-status", nil, "Exclude by status (can be repeated)")
	listCmd.Flags().StringArrayVarP(&listType, "type", "t", nil, "Filter by type (can be repeated)")
	listCmd.Flags().StringArrayVar(&listNoType, "no-type", nil, "Exclude by type (can be repeated)")
	listCmd.Flags().StringArrayVarP(&listPriority, "priority", "p", nil, "Filter by priority (can be repeated)")
	listCmd.Flags().StringArrayVar(&listNoPriority, "no-priority", nil, "Exclude by priority (can be repeated)")
	listCmd.Flags().StringArrayVar(&listTag, "tag", nil, "Filter by tag (can be repeated, OR logic)")
	listCmd.Flags().StringArrayVar(&listNoTag, "no-tag", nil, "Exclude beans with tag (can be repeated)")
	listCmd.Flags().BoolVar(&listHasParent, "has-parent", false, "Filter beans with a parent")
	listCmd.Flags().BoolVar(&listNoParent, "no-parent", false, "Filter beans without a parent")
	listCmd.Flags().StringVar(&listParentID, "parent", "", "Filter by parent ID")
	listCmd.Flags().BoolVar(&listHasBlocking, "has-blocking", false, "Filter beans that are blocking others")
	listCmd.Flags().BoolVar(&listNoBlocking, "no-blocking", false, "Filter beans that aren't blocking others")
	listCmd.Flags().BoolVar(&listIsBlocked, "is-blocked", false, "Filter beans that are blocked by others")
	listCmd.Flags().BoolVar(&listReady, "ready", false, "Filter beans available to start (not blocked, excludes in-progress/completed/scrapped/draft)")
	listCmd.Flags().BoolVarP(&listQuiet, "quiet", "q", false, "Only output IDs (one per line)")
	listCmd.Flags().StringVar(&listSort, "sort", "", "Sort by: created, updated, status, priority, id (default: status, priority, type, title)")
	listCmd.Flags().StringVar(&listSortr, "sortr", "", "Sort by (reversed): created, updated, status, priority, id")
	listCmd.Flags().BoolVar(&listFull, "full", false, "Include bean body in JSON output")
	listCmd.Flags().BoolVar(&listShowDone, "show-done", false, "Include completed and scrapped beans")
	rootCmd.AddCommand(listCmd)
}
