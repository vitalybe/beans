package cmd

import (
	"fmt"

	"github.com/hmans/beans/internal/tui"
	"github.com/spf13/cobra"
)

var (
	tuiSort     string
	tuiSortr    string
	tuiShowDone bool
)

var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Open the interactive TUI",
	Long:  `Opens an interactive terminal user interface for browsing and managing beans.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if tuiSort != "" && tuiSortr != "" {
			return fmt.Errorf("--sort and --sortr are mutually exclusive")
		}

		sortOpts := tui.SortOptions{
			ShowDone: tuiShowDone,
		}
		if tuiSort != "" {
			sortOpts.SortBy = tuiSort
		}
		if tuiSortr != "" {
			sortOpts.SortBy = tuiSortr
			sortOpts.Reverse = true
		}

		return tui.Run(core, cfg, sortOpts)
	},
}

func init() {
	tuiCmd.Flags().StringVar(&tuiSort, "sort", "", "Sort by: created, updated, status, priority, id (default: status, priority, type, title)")
	tuiCmd.Flags().StringVar(&tuiSortr, "sortr", "", "Sort by (reversed): created, updated, status, priority, id")
	tuiCmd.Flags().BoolVar(&tuiShowDone, "show-done", false, "Include completed and scrapped beans")
	rootCmd.AddCommand(tuiCmd)
}
