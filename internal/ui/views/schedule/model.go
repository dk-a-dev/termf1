package schedule

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/bubbletea"
	"github.com/devkeshwani/termf1/internal/ui/views/common"
	"github.com/charmbracelet/lipgloss"
	"github.com/devkeshwani/termf1/internal/api/jolpica"
	"github.com/devkeshwani/termf1/internal/ui/styles"
)

// ── Messages ──────────────────────────────────────────────────────────────────

type scheduleDataMsg struct{ races []jolpica.Race }
type scheduleErrMsg struct{ err error }

// ── Model ─────────────────────────────────────────────────────────────────────

type Schedule struct {
	client   *jolpica.Client
	width    int
	height   int
	loading  bool
	err      error
	races    []jolpica.Race
	viewport viewport.Model
	spin     spinner.Model
	cursor   int // currently selected race index
}

func NewSchedule(client *jolpica.Client) *Schedule {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(styles.ColorF1Red)
	vp := viewport.New(80, 30)
	return &Schedule{client: client, loading: true, spin: s, viewport: vp}
}

func (s *Schedule) SetSize(w, h int) {
	s.width = w
	s.height = h
	s.viewport.Width = w
	s.viewport.Height = h - 4
}

func (s *Schedule) Init() tea.Cmd {
	return tea.Batch(s.spin.Tick, fetchScheduleCmd(s.client))
}

func (s *Schedule) UpdateSchedule(msg tea.Msg) (*Schedule, tea.Cmd) {
	switch msg := msg.(type) {
	case spinner.TickMsg:
		if s.loading {
			var cmd tea.Cmd
			s.spin, cmd = s.spin.Update(msg)
			return s, cmd
		}
	case scheduleDataMsg:
		s.loading = false
		s.err = nil
		s.races = msg.races
		s.cursor = s.nextRaceIdx()
		content := s.buildContent()
		s.viewport.SetContent(content)
		s.scrollToNextRace()
	case scheduleErrMsg:
		s.loading = false
		s.err = msg.err

	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if s.cursor > 0 {
				s.cursor--
				s.viewport.SetContent(s.buildContent())
			}
			return s, nil
		case "down", "j":
			if s.cursor < len(s.races)-1 {
				s.cursor++
				s.viewport.SetContent(s.buildContent())
			}
			return s, nil
		case "enter", " ":
			s.viewport.SetContent(s.buildContent())
			return s, nil
		}
	}

	var cmd tea.Cmd
	s.viewport, cmd = s.viewport.Update(msg)
	return s, cmd
}

func (s *Schedule) View() string {
	if s.loading && len(s.races) == 0 {
		return common.Centred(s.width, s.height, s.spin.View()+" Loading schedule…")
	}
	if s.err != nil && len(s.races) == 0 {
		return common.Centred(s.width, s.height, styles.ErrorStyle.Render("⚠  "+s.err.Error()))
	}

	year := raceYear(s.races)
	title := styles.Title.Render(" 🏁 " + year + " F1 Season Calendar")
	sep := styles.Divider.Render(strings.Repeat("─", s.width))
	hint := styles.FooterStyle.Render("  ↑/↓ navigate  │  ↑↓/PgUp/PgDn scroll  │  Enter expand sessions")

	if len(s.races) > 0 {
		s.viewport.SetContent(s.buildContent())
	}

	return lipgloss.JoinVertical(lipgloss.Left,
		title, sep,
		s.viewport.View(),
		hint,
	)
}

// buildContent renders all race cards grouped by month into a scrollable string.
func (s *Schedule) buildContent() string {
	now := time.Now()
	cardW := s.width - 4
	if cardW < 40 {
		cardW = 40
	}

	type monthGroup struct {
		month string
		cards []string
	}

	groups := map[string]*monthGroup{}
	order := []string{}

	for i, race := range s.races {
		raceTime := parseDate(race.Date, race.Time)
		monthKey := "Unknown"
		if !raceTime.IsZero() {
			monthKey = raceTime.Format("January 2006")
		}

		if _, exists := groups[monthKey]; !exists {
			groups[monthKey] = &monthGroup{month: monthKey}
			order = append(order, monthKey)
		}
		groups[monthKey].cards = append(groups[monthKey].cards, s.raceCard(i, race, now, cardW))
	}

	var sb strings.Builder
	for _, month := range order {
		g := groups[month]
		monthHdr := lipgloss.NewStyle().
			Bold(true).
			Foreground(styles.ColorYellow).
			Background(lipgloss.Color("#1a1a1a")).
			Padding(0, 2).
			Width(s.width).
			Render("  " + g.month)
		sb.WriteString(monthHdr + "\n")
		for _, card := range g.cards {
			sb.WriteString(card + "\n")
		}
	}
	return sb.String()
}

func (s *Schedule) raceCard(idx int, race jolpica.Race, now time.Time, w int) string {
	raceTime := parseDate(race.Date, race.Time)
	isPast := !raceTime.IsZero() && raceTime.Before(now)
	isNext := !isPast && s.isNextRaceIdx(idx)
	isSelected := idx == s.cursor

	var statusStr string
	switch {
	case isPast:
		statusStr = lipgloss.NewStyle().Foreground(styles.ColorMuted).Render("  ✓ PAST")
	case isNext:
		statusStr = lipgloss.NewStyle().Foreground(styles.ColorF1Red).Bold(true).Render("  ► NEXT RACE")
	default:
		statusStr = lipgloss.NewStyle().Foreground(styles.ColorSubtle).Render("  upcoming")
	}

	roundStr := lipgloss.NewStyle().Foreground(styles.ColorSubtle).Render(fmt.Sprintf("R%s", race.Round))
	nameStr := lipgloss.NewStyle().Bold(true).Foreground(styles.ColorF1White).Render(race.RaceName)
	headerLine := roundStr + "  " + nameStr + statusStr

	locStr := styles.DimStyle.Render(fmt.Sprintf("  📍 %s  ·  %s, %s",
		race.Circuit.CircuitName, race.Circuit.Location.Locality, race.Circuit.Location.Country))

	dateStr := ""
	if !raceTime.IsZero() {
		switch {
		case isNext:
			dateStr = lipgloss.NewStyle().Foreground(styles.ColorYellow).
				Render(fmt.Sprintf("  📅 %s  (%s)",
					raceTime.Format("Mon 2 Jan 2006 15:04 UTC"), countdownStr(raceTime, now)))
		case isPast:
			dateStr = styles.DimStyle.Render("  📅 " + raceTime.Format("Mon 2 Jan 2006"))
		default:
			dateStr = lipgloss.NewStyle().Foreground(styles.ColorTextDim).
				Render("  📅 " + raceTime.Format("Mon 2 Jan 2006 15:04 UTC"))
		}
	}

	lines := []string{headerLine, locStr, dateStr}

	// Show session detail when this card is selected or is the next race
	if isSelected || isNext {
		sessLines := sessionDetailLines(race, now)
		if sessLines != "" {
			lines = append(lines, "")
			lines = append(lines, strings.Split(sessLines, "\n")...)
		}
	}

	body := strings.Join(lines, "\n")
	if isPast {
		body = lipgloss.NewStyle().Foreground(styles.ColorSubtle).Render(body)
	}

	borderCol := styles.ColorBorder
	switch {
	case isNext:
		borderCol = styles.ColorF1Red
	case isSelected && !isPast:
		borderCol = styles.ColorBlue
	case isPast:
		borderCol = styles.ColorMuted
	}

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderCol).
		Padding(0, 1).
		Width(w).
		Render(body)
}

// ── helpers ───────────────────────────────────────────────────────────────────

func sessionDetailLines(race jolpica.Race, now time.Time) string {
	type sessEntry struct {
		label string
		s     *jolpica.SessionTime
	}
	entries := []sessEntry{
		{"FP1      ", race.FirstPractice},
		{"FP2      ", race.SecondPractice},
		{"FP3      ", race.ThirdPractice},
		{"Sprint Q ", race.SprintQualifying},
		{"Sprint   ", race.Sprint},
		{"Qualifying", race.Qualifying},
	}
	var parts []string
	for _, e := range entries {
		if e.s == nil {
			continue
		}
		t := parseDate(e.s.Date, e.s.Time)
		if t.IsZero() {
			continue
		}
		dot := "○"
		col := styles.ColorSubtle
		if t.Before(now) {
			dot = "●"
			col = styles.ColorMuted
		}
		parts = append(parts, lipgloss.NewStyle().Foreground(col).Render(
			fmt.Sprintf("    %s %-10s  %s", dot, e.label, t.Format("Mon 2 Jan  15:04 UTC")),
		))
	}
	return strings.Join(parts, "\n")
}

func parseDate(date, timeStr string) time.Time {
	if date == "" {
		return time.Time{}
	}
	combined := date
	if timeStr != "" && !strings.Contains(date, "T") {
		combined = date + "T" + timeStr
	}
	formats := []string{
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
		"2006-01-02",
	}
	for _, f := range formats {
		if t, err := time.Parse(f, combined); err == nil {
			return t
		}
	}
	if t, err := time.Parse("2006-01-02", date); err == nil {
		return t
	}
	return time.Time{}
}

func (s *Schedule) isNextRaceIdx(idx int) bool {
	now := time.Now()
	for i, r := range s.races {
		t := parseDate(r.Date, r.Time)
		if !t.IsZero() && t.After(now) {
			return i == idx
		}
	}
	return false
}

func (s *Schedule) nextRaceIdx() int {
	now := time.Now()
	for i, r := range s.races {
		t := parseDate(r.Date, r.Time)
		if !t.IsZero() && t.After(now) {
			return i
		}
	}
	return 0
}

func (s *Schedule) scrollToNextRace() {
	idx := s.nextRaceIdx()
	if idx > 2 {
		s.viewport.SetYOffset(idx * 6)
	}
}

func countdownStr(future, now time.Time) string {
	d := future.Sub(now)
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	if days > 0 {
		return fmt.Sprintf("in %dd %dh", days, hours)
	}
	return fmt.Sprintf("in %dh %dm", int(d.Hours()), int(d.Minutes())%60)
}

func raceYear(races []jolpica.Race) string {
	if len(races) == 0 {
		return ""
	}
	return races[0].Season
}

func fetchScheduleCmd(client *jolpica.Client) tea.Cmd {
	return func() tea.Msg {
		races, err := client.GetSchedule(common.ContextBG())
		if err != nil {
			return scheduleErrMsg{err}
		}
		return scheduleDataMsg{races}
	}
}
