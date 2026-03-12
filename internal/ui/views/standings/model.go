package standings

import (
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbletea"
	"github.com/dk-a-dev/termf1/v2/internal/ui/views/common"
	"github.com/charmbracelet/lipgloss"
	"github.com/dk-a-dev/termf1/v2/internal/api/jolpica"
	"github.com/dk-a-dev/termf1/v2/internal/ui/styles"
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

// View returns the rendered standard output
func (s *Standings) View() string {
	if s.loading && len(s.drivers) == 0 {
		return common.Centred(s.width, s.height, s.spin.View()+" Loading standings…")
	}
	if s.err != nil && len(s.drivers) == 0 {
		return common.Centred(s.width, s.height, styles.ErrorStyle.Render("⚠  "+s.err.Error()))
	}

	return RenderStandings(s.drivers, s.constructors, s.width)
}



// ── command ───────────────────────────────────────────────────────────────────

func fetchStandingsCmd(client *jolpica.Client) tea.Cmd {
	return func() tea.Msg {
		ctx := common.ContextBG()
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
