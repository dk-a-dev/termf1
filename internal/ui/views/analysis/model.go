// Package analysis implements the Analysis tab: terminal charts from OpenF1 data.
package analysis

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dk-a-dev/termf1/v2/internal/api/openf1"
	"github.com/dk-a-dev/termf1/v2/internal/ui/styles"
	"github.com/dk-a-dev/termf1/v2/internal/ui/views/common"
)

// ── Chart IDs ─────────────────────────────────────────────────────────────────

const (
	ChartStrategy = iota
	// ChartLapLine   // 2-D lap-time line chart (driver comparison)
	ChartLapSpark  // per-driver sparklines
	ChartPace
	ChartSectors
	ChartSpeed
	ChartPositions // position changes per lap
	ChartTeamPace  // team pace comparison box plot
	ChartPitStops  // pit stop duration bar chart
	NumCharts
)

var chartNames = []string{
	"① Strategy",
	// "② Lap Lines",
	"③ Sparklines",
	"④ Pace",
	"⑤ Sectors",
	"⑥ Speed",
	"⑦ Positions",
	"⑧ Team Pace",
	"⑨ Pit Stops",
}

// ── Messages ──────────────────────────────────────────────────────────────────

// DataMsg carries fetched session data back to the Update loop.
type DataMsg struct {
	Session   openf1.Session
	Drivers   []openf1.Driver
	Laps      []openf1.Lap
	Stints    []openf1.Stint
	Positions []openf1.Position
	Pits      []openf1.Pit
}

// ErrMsg carries a fetch error.
type ErrMsg struct{ Err error }

// ── Model ─────────────────────────────────────────────────────────────────────

// Analysis is the bubbletea sub-model for the Analysis tab.
type Analysis struct {
	of1     *openf1.Client
	width   int
	height  int
	loading bool
	fetched bool
	err     error

	chart     int
	session   openf1.Session
	drivers   []openf1.Driver
	laps      []openf1.Lap
	stints    []openf1.Stint
	positions []openf1.Position
	pits      []openf1.Pit

	viewport viewport.Model
	spin     spinner.Model
}

// NewAnalysis constructs a ready-to-use Analysis model.
func NewAnalysis(of1 *openf1.Client) *Analysis {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(styles.ColorF1Red)
	return &Analysis{
		of1:      of1,
		loading:  true,
		viewport: viewport.New(80, 20),
		spin:     s,
	}
}

// SetSize propagates terminal dimensions to the model.
func (a *Analysis) SetSize(w, h int) {
	a.width = w
	a.height = h
	if !a.loading {
		a.reloadViewport()
	}
}

// Init starts the spinner and triggers the first data fetch.
func (a *Analysis) Init() tea.Cmd {
	if a.fetched {
		return nil
	}
	return tea.Batch(a.spin.Tick, FetchCmd(a.of1))
}

// UpdateAnalysis handles all messages destined for the Analysis tab.
func (a *Analysis) UpdateAnalysis(msg tea.Msg) (*Analysis, tea.Cmd) {
	switch msg := msg.(type) {
	case spinner.TickMsg:
		if a.loading {
			var cmd tea.Cmd
			a.spin, cmd = a.spin.Update(msg)
			return a, cmd
		}

	case DataMsg:
		a.loading = false
		a.fetched = true
		a.session = msg.Session
		a.drivers = msg.Drivers
		a.laps = msg.Laps
		a.stints = msg.Stints
		a.positions = msg.Positions
		a.pits = msg.Pits
		a.reloadViewport()
		return a, nil

	case ErrMsg:
		a.loading = false
		a.err = msg.Err
		return a, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "left", "h":
			a.chart = (a.chart - 1 + NumCharts) % NumCharts
			a.reloadViewport()
			return a, nil
		case "right", "l":
			a.chart = (a.chart + 1) % NumCharts
			a.reloadViewport()
			return a, nil
		case "r":
			a.loading = true
			a.fetched = false
			a.laps = nil
			a.stints = nil
			return a, tea.Batch(a.spin.Tick, FetchCmd(a.of1))
		}
	}
	var cmd tea.Cmd
	a.viewport, cmd = a.viewport.Update(msg)
	return a, cmd
}

// View renders the analysis tab.
func (a *Analysis) View() string {
	if a.loading {
		return common.Centred(a.width, a.height, a.spin.View()+" Loading analysis data…")
	}
	if a.err != nil {
		return common.Centred(a.width, a.height, styles.ErrorStyle.Render("⚠  "+a.err.Error()))
	}
	hdr := a.renderHeader()
	sep := styles.Divider.Render(strings.Repeat("─", a.width))
	hints := lipgloss.NewStyle().Width(a.width).Background(styles.ColorBgHeader).
		Render(styles.FooterStyle.Render("  ←/h  →/l  switch chart   ↑↓  scroll   r  refresh"))
	return lipgloss.JoinVertical(lipgloss.Left, hdr, sep, a.viewport.View(), hints)
}

// ── Internal helpers ──────────────────────────────────────────────────────────

func (a *Analysis) reloadViewport() {
	w := a.width - 2
	h := a.height - 5
	if w < 20 {
		w = 20
	}
	if h < 5 {
		h = 5
	}
	a.viewport.Width = w
	a.viewport.Height = h
	a.viewport.SetContent(a.renderChart(w, h))
}

func (a *Analysis) renderHeader() string {
	sessInfo := lipgloss.NewStyle().Foreground(styles.ColorYellow).Bold(true).
		Render(a.session.CountryName) +
		"  " + styles.DimStyle.Render(a.session.SessionName) +
		"  " + styles.DimStyle.Render(fmt.Sprintf("%d", a.session.Year))

	var pills []string
	for i, name := range chartNames {
		if i == a.chart {
			pills = append(pills, lipgloss.NewStyle().
				Background(styles.ColorF1Red).Foreground(lipgloss.Color("#ffffff")).
				Bold(true).Padding(0, 2).Render(name))
		} else {
			pills = append(pills, lipgloss.NewStyle().
				Background(styles.ColorBgCard).Foreground(styles.ColorTextDim).
				Padding(0, 2).Render(name))
		}
	}
	return lipgloss.NewStyle().Width(a.width).Background(styles.ColorBgHeader).Padding(0, 2).
		Render(lipgloss.JoinHorizontal(lipgloss.Left, sessInfo, "      ", strings.Join(pills, " ")))
}
