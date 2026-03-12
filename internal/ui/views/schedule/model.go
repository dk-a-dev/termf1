package schedule

import (
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/bubbletea"
	"github.com/dk-a-dev/termf1/internal/ui/views/common"
	"github.com/charmbracelet/lipgloss"
	"github.com/dk-a-dev/termf1/internal/api/jolpica"
	"github.com/dk-a-dev/termf1/internal/ui/styles"
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
		s.cursor = NextRaceIdx(s.races, time.Now())
		content := RenderScheduleCards(s.races, time.Now(), s.width, s.cursor)
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
				s.viewport.SetContent(RenderScheduleCards(s.races, time.Now(), s.width, s.cursor))
			}
			return s, nil
		case "down", "j":
			if s.cursor < len(s.races)-1 {
				s.cursor++
				s.viewport.SetContent(RenderScheduleCards(s.races, time.Now(), s.width, s.cursor))
			}
			return s, nil
		case "enter", " ":
			s.viewport.SetContent(RenderScheduleCards(s.races, time.Now(), s.width, s.cursor))
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

	year := RaceYear(s.races)
	title := styles.Title.Render(" 🏁 " + year + " F1 Season Calendar")
	sep := styles.Divider.Render(strings.Repeat("─", s.width))
	hint := styles.FooterStyle.Render("  ↑/↓ navigate  │  ↑↓/PgUp/PgDn scroll  │  Enter expand sessions")

	if len(s.races) > 0 {
		s.viewport.SetContent(RenderScheduleCards(s.races, time.Now(), s.width, s.cursor))
	}

	return lipgloss.JoinVertical(lipgloss.Left,
		title, sep,
		s.viewport.View(),
		hint,
	)
}

// ── helpers ───────────────────────────────────────────────────────────────────

func (s *Schedule) scrollToNextRace() {
	idx := NextRaceIdx(s.races, time.Now())
	if idx > 2 {
		s.viewport.SetYOffset(idx * 6)
	}
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
