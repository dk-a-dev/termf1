package cmd

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/dk-a-dev/termf1/v2/internal/config"
	"github.com/dk-a-dev/termf1/v2/internal/ui"
	"github.com/spf13/cobra"
)

// tuiCmd represents the tui command
var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Launch the full-featured Terminal UI (default)",
	Long: `Starts the full interactive terminal user interface, including Dashboard,
Standings, Schedule, Analysis, and AI Chat.`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Load()
		cfg.Version = version

		p := tea.NewProgram(
			ui.NewApp(cfg),
			tea.WithAltScreen(),
			tea.WithMouseCellMotion(),
		)

		if _, err := p.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Error running termf1: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(tuiCmd)
	// Make TUI the default command if none is specified
	rootCmd.Run = tuiCmd.Run
}
