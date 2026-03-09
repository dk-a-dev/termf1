package views

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/devkeshwani/termf1/internal/api/jolpica"
	"github.com/devkeshwani/termf1/internal/ui/styles"
)

// ── Messages ──────────────────────────────────────────────────────────────────

type standingsDataMsg struct {
	drivers      []jolpica.DriverStanding
	constructors []jolpica.ConstructorStanding
}
type standingsErrMsg struct{ err error }

// ── Model ─────────────────────────────────────────────────────────────────────

type Standings struct {
	client       *jolpica.Client
	width        int
	height       int
	loading      bool
	err          error
	drivers      []jolpica.DriverStanding
	constructors []jolpica.ConstructorStanding
	spin         spinner.Model
}

func NewStandings(client *jolpica.Client) *Standings {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(styles.ColorF1Red)
	return &Standings{client: client, loading: true, spin: s}
}

func (s *Standings) SetSize(w, h int) { s.width = w; s.height = h }

func (s *Standings) Init() tea.Cmd {
	return tea.Batch(s.spin.Tick, fetchStandingsCmd(s.client))
}

func (s *Standings) UpdateStandings(msg tea.Msg) (*Standings, tea.Cmd) {
	switch msg := msg.(type) {
	case spinner.TickMsg:
		if s.loading {
			var cmd tea.Cmd
			s.spin, cmd = s.spin.Update(msg)
			return s, cmd
		}
	case standingsDataMsg:
		s.loading = false
		s.err = nil
		s.drivers = msg.drivers
		s.constructors = msg.constructors
	case standingsErrMsg:
		s.loading = false
		s.err = msg.err
	}
	return s, nil
}

func (s *Standings) View() string {
	if s.loading && len(s.drivers) == 0 {
		return centred(s.width, s.height, s.spin.View()+" Loading standings…")
	}
	if s.err != nil && len(s.drivers) == 0 {
		return centred(s.width, s.height, styles.ErrorStyle.Render("⚠  "+s.err.Error()))
	}

	half := s.width / 2
	if half < 30 {
		return lipgloss.JoinVertical(lipgloss.Left,
			s.driverStandingsTable(s.width),
			"",
			s.constructorStandingsTable(s.width),
		)
	}

	left := s.driverStandingsTable(half - 1)
	right := s.constructorStandingsTable(s.width - half - 1)
	return lipgloss.JoinHorizontal(lipgloss.Top, left, " ", right)
}

// ── Driver standings ──────────────────────────────────────────────────────────

func (s *Standings) driverStandingsTable(w int) string {
	title := styles.Title.Render("  Driver Championship")
	sep := styles.Divider.Render(strings.Repeat("─", w))

	// find max points for bar scaling
	maxPts := 1.0
	for _, d := range s.drivers {
		if p, _ := strconv.ParseFloat(d.Points, 64); p > maxPts {
			maxPts = p
		}
	}

	barW := w - 28
	if barW < 4 {
		barW = 4
	}

	lines := []string{title, sep}
	for _, d := range s.drivers {
		pts, _ := strconv.ParseFloat(d.Points, 64)
		pos, _ := strconv.Atoi(d.Position)

		posStr := fmt.Sprintf("%2d", pos)
		code := d.Driver.Code
		if code == "" {
			code = d.Driver.FamilyName
			if len(code) > 3 {
				code = code[:3]
			}
		}

		// Team colour for the bar
		teamCol := styles.ColorSubtle
		if len(d.Constructors) > 0 {
			teamCol = styles.TeamColor(d.Constructors[0].Name)
		}

		bar := styles.BarChart(pts, maxPts, barW, teamCol)
		ptsStr := fmt.Sprintf("%4.0f", pts)

		posC := lipgloss.NewStyle().Width(3).Foreground(posColor(pos)).Bold(pos <= 3).Render(posStr)
		codeC := lipgloss.NewStyle().Width(5).Bold(true).Foreground(styles.ColorText).Render(code)
		ptsC := lipgloss.NewStyle().Width(5).Align(lipgloss.Right).Foreground(styles.ColorYellow).Render(ptsStr)

		line := posC + codeC + bar + " " + ptsC
		lines = append(lines, line)
	}
	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

// ── Constructor standings ─────────────────────────────────────────────────────

func (s *Standings) constructorStandingsTable(w int) string {
	title := styles.Title.Render("  Constructor Championship")
	sep := styles.Divider.Render(strings.Repeat("─", w))

	maxPts := 1.0
	for _, c := range s.constructors {
		if p, _ := strconv.ParseFloat(c.Points, 64); p > maxPts {
			maxPts = p
		}
	}

	barW := w - 28
	if barW < 4 {
		barW = 4
	}

	lines := []string{title, sep}
	for _, c := range s.constructors {
		pts, _ := strconv.ParseFloat(c.Points, 64)
		pos, _ := strconv.Atoi(c.Position)

		name := c.Constructor.Name
		if len(name) > 12 {
			name = name[:12]
		}

		teamCol := styles.TeamColor(c.Constructor.Name)
		bar := styles.BarChart(pts, maxPts, barW, teamCol)
		ptsStr := fmt.Sprintf("%4.0f", pts)

		posC := lipgloss.NewStyle().Width(3).Foreground(posColor(pos)).Bold(pos <= 3).Render(fmt.Sprintf("%2d", pos))
		nameC := lipgloss.NewStyle().Width(14).Bold(true).Foreground(styles.ColorText).Render(name)
		ptsC := lipgloss.NewStyle().Width(5).Align(lipgloss.Right).Foreground(styles.ColorYellow).Render(ptsStr)

		line := posC + nameC + bar + " " + ptsC
		lines = append(lines, line)
	}
	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

// ── command ───────────────────────────────────────────────────────────────────

func fetchStandingsCmd(client *jolpica.Client) tea.Cmd {
	return func() tea.Msg {
		ctx := contextBG()
		drivers, err := client.GetDriverStandings(ctx)
		if err != nil {
			return standingsErrMsg{err}
		}
		constructors, err := client.GetConstructorStandings(ctx)
		if err != nil {
			return standingsErrMsg{err}
		}
		return standingsDataMsg{drivers, constructors}
	}
}
