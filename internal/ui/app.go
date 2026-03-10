package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/devkeshwani/termf1/internal/api/groq"
	"github.com/devkeshwani/termf1/internal/api/jolpica"
	"github.com/devkeshwani/termf1/internal/api/liveserver"
	"github.com/devkeshwani/termf1/internal/api/multiviewer"
	"github.com/devkeshwani/termf1/internal/api/openf1"
	"github.com/devkeshwani/termf1/internal/config"
	"github.com/devkeshwani/termf1/internal/ui/styles"
	"github.com/devkeshwani/termf1/internal/ui/views/analysis"
	"github.com/devkeshwani/termf1/internal/ui/views/chat"
	"github.com/devkeshwani/termf1/internal/ui/views/dashboard"
	"github.com/devkeshwani/termf1/internal/ui/views/driverstats"
	"github.com/devkeshwani/termf1/internal/ui/views/schedule"
	"github.com/devkeshwani/termf1/internal/ui/views/standings"
	"github.com/devkeshwani/termf1/internal/ui/views/trackmap"
	"github.com/devkeshwani/termf1/internal/ui/views/weather"
)

// viewID identifies which tab is active.
type viewID int

const (
	viewDashboard viewID = iota
	viewTrackMap
	viewDriverStats
	viewWeather
	viewSchedule
	viewStandings
	viewChat
	viewAnalysis
	numViews
)

var tabLabels = []string{
	" 1 Dashboard  ",
	" 2 Track Map  ",
	" 3 Driver Stats ",
	" 4 Weather    ",
	" 5 Schedule   ",
	" 6 Standings  ",
	" 7 Ask F1 AI  ",
	" 8 Analysis   ",
}

var tabKeys = []string{"1", "2", "3", "4", "5", "6", "7", "8"}

// App is the root bubbletea model. It owns navigation and delegates rendering
// to the current view sub-model.
type App struct {
	cfg     *config.Config
	width   int
	height  int
	current viewID

	dashboard   *dashboard.Dashboard2
	standings   *standings.Standings
	schedule    *schedule.Schedule
	weather     *weather.WeatherView
	chat        *chat.Chat
	trackMap    *trackmap.TrackMap
	driverStats *driverstats.DriverStats
	analysis    *analysis.Analysis
}

func NewApp(cfg *config.Config) *App {
	of1 := openf1.NewClient()
	joli := jolpica.NewClient()
	mv := multiviewer.NewClient()
	ls := liveserver.New(cfg.LiveServerAddr)
	groqClient := groq.NewClient(cfg.GroqAPIKey, cfg.GroqModel)

	return &App{
		cfg:         cfg,
		current:     viewDashboard,
		dashboard:   dashboard.NewDashboard2(ls, of1, joli, mv, cfg.RefreshRate),
		standings:   standings.NewStandings(joli),
		schedule:    schedule.NewSchedule(joli),
		weather:     weather.NewWeatherView(of1),
		chat:        chat.NewChat(groqClient),
		trackMap:    trackmap.NewTrackMap(of1),
		driverStats: driverstats.NewDriverStats(of1, joli),
		analysis:    analysis.NewAnalysis(of1),
	}
}

func (a *App) Init() tea.Cmd {
	return tea.Batch(
		a.standings.Init(),
		a.schedule.Init(),
	)
}

func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		contentH := a.height - headerHeight() - footerHeight()
		if contentH < 1 {
			contentH = 1
		}
		a.propagateSize(a.width, contentH)
		return a, nil

	case tea.KeyMsg:
		// When the AI chat input box is focused, let ALL keystrokes through to
		// the chat view — only Ctrl+C and Ctrl+Tab/Shift+Tab switch tabs.
		chatInputActive := a.current == viewChat && a.chat.InputFocused()

		switch msg.String() {
		case "ctrl+c":
			return a, tea.Quit

		case "q":
			if !chatInputActive {
				return a, tea.Quit
			}

		case "tab":
			if !chatInputActive {
				a.current = (a.current + 1) % numViews
				return a, a.initCurrentView()
			}

		case "shift+tab":
			if !chatInputActive {
				a.current = (a.current - 1 + numViews) % numViews
				return a, a.initCurrentView()
			}

		case "1":
			if !chatInputActive {
				a.current = viewDashboard
				return a, a.initCurrentView()
			}
		case "2":
			if !chatInputActive {
				a.current = viewTrackMap
				return a, a.initCurrentView()
			}
		case "3":
			if !chatInputActive {
				a.current = viewDriverStats
				return a, a.initCurrentView()
			}
		case "4":
			if !chatInputActive {
				a.current = viewWeather
				return a, a.initCurrentView()
			}
		case "5":
			if !chatInputActive {
				a.current = viewSchedule
				return a, a.initCurrentView()
			}
		case "6":
			if !chatInputActive {
				a.current = viewStandings
				return a, a.initCurrentView()
			}
		case "7":
			if !chatInputActive {
				a.current = viewChat
				return a, a.initCurrentView()
			}
		case "8":
			if !chatInputActive {
				a.current = viewAnalysis
				return a, a.initCurrentView()
			}
		case "r":
			if !chatInputActive {
				return a, a.initCurrentView()
			}
		}
	}

	// Delegate message to the current sub-model.
	var cmd tea.Cmd
	switch a.current {
	case viewDashboard:
		a.dashboard, cmd = a.dashboard.UpdateDash(msg)
	case viewTrackMap:
		a.trackMap, cmd = a.trackMap.UpdateTrackMap(msg)
	case viewDriverStats:
		a.driverStats, cmd = a.driverStats.UpdateDriverStats(msg)
	case viewWeather:
		a.weather, cmd = a.weather.UpdateWeather(msg)
	case viewSchedule:
		a.schedule, cmd = a.schedule.UpdateSchedule(msg)
	case viewStandings:
		a.standings, cmd = a.standings.UpdateStandings(msg)
	case viewChat:
		a.chat, cmd = a.chat.UpdateChat(msg)
	case viewAnalysis:
		a.analysis, cmd = a.analysis.UpdateAnalysis(msg)
	}
	return a, cmd
}

func (a *App) View() string {
	if a.width == 0 {
		return "Loading…"
	}
	if a.width < 80 || a.height < 20 {
		return styles.ErrorStyle.Render(
			fmt.Sprintf("Terminal too small (%dx%d). Resize to at least 80×20.", a.width, a.height),
		)
	}

	header := a.renderHeader()
	footer := a.renderFooter()
	content := a.currentView()

	full := lipgloss.JoinVertical(lipgloss.Left, header, content, footer)
	return full
}

// ── helpers ──────────────────────────────────────────────────────────────────

func (a *App) currentView() string {
	switch a.current {
	case viewDashboard:
		return a.dashboard.View()
	case viewTrackMap:
		return a.trackMap.View()
	case viewDriverStats:
		return a.driverStats.View()
	case viewWeather:
		return a.weather.View()
	case viewSchedule:
		return a.schedule.View()
	case viewStandings:
		return a.standings.View()
	case viewChat:
		return a.chat.View()
	case viewAnalysis:
		return a.analysis.View()
	}
	return ""
}

func (a *App) initCurrentView() tea.Cmd {
	switch a.current {
	case viewDashboard:
		return a.dashboard.Init()
	case viewTrackMap:
		return a.trackMap.Init()
	case viewDriverStats:
		return a.driverStats.Init()
	case viewWeather:
		return a.weather.Init()
	case viewSchedule:
		return a.schedule.Init()
	case viewStandings:
		return a.standings.Init()
	case viewChat:
		return a.chat.Init()
	case viewAnalysis:
		return a.analysis.Init()
	}
	return nil
}

func (a *App) propagateSize(w, h int) {
	a.dashboard.SetSize(w, h)
	a.standings.SetSize(w, h)
	a.schedule.SetSize(w, h)
	a.weather.SetSize(w, h)
	a.chat.SetSize(w, h)
	a.trackMap.SetSize(w, h)
	a.driverStats.SetSize(w, h)
	a.analysis.SetSize(w, h)
}

func headerHeight() int { return 3 }
func footerHeight() int { return 1 }

func (a *App) renderHeader() string {
	// Brand
	brand := lipgloss.NewStyle().
		Bold(true).
		Foreground(styles.ColorF1Red).
		Padding(0, 1).
		Render("🏎 termf1")

	// Version badge
	v := a.cfg.Version
	verBadge := styles.DimStyle.Render(" " + v + " ")

	// Tabs
	tabs := make([]string, len(tabLabels))
	for i, label := range tabLabels {
		if viewID(i) == a.current {
			tabs[i] = styles.ActiveTab.Render(label)
		} else {
			tabs[i] = styles.InactiveTab.Render(label)
		}
	}
	tabRow := lipgloss.JoinHorizontal(lipgloss.Top, tabs...)

	// Top row: brand + version + tabs
	topRow := lipgloss.JoinHorizontal(lipgloss.Top, brand, verBadge, "  ", tabRow)

	// Separator line
	sep := styles.Divider.Render(strings.Repeat("─", a.width))

	// PaddingTop(1) makes the header bar 3 lines (blank + content + sep),
	// matching headerHeight()=3 and giving the tab strip visual breathing room.
	return lipgloss.JoinVertical(lipgloss.Left,
		lipgloss.NewStyle().Background(styles.ColorBgHeader).Width(a.width).PaddingTop(1).Render(topRow),
		sep,
	)
}

func (a *App) renderFooter() string {
	hints := []string{
		"q quit",
		"tab/1-7 navigate",
		"r refresh",
		"↑↓/PgUp/PgDn scroll",
	}
	text := styles.FooterStyle.Render("  " + strings.Join(hints, "  │  "))
	return lipgloss.NewStyle().Width(a.width).Background(styles.ColorBgHeader).Render(text)
}
