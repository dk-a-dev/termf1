package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbletea"
	"github.com/devkeshwani/termf1/internal/config"
	"github.com/devkeshwani/termf1/internal/ui"
)

func main() {
	cfg := config.Load()

	p := tea.NewProgram(
		ui.NewApp(cfg),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running termf1: %v\n", err)
		os.Exit(1)
	}
}
