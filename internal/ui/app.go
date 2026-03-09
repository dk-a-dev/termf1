package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/devkeshwani/termf1/internal/api/groq"
	"github.com/devkeshwani/termf1/internal/api/jolpica"
	"github.com/devkeshwani/termf1/internal/api/openf1"
	"github.com/devkeshwani/termf1/internal/config"
	"github.com/devkeshwani/termf1/internal/ui/styles"
	"github.com/devkeshwani/termf1/internal/ui/views"
)

// viewID identifies which tab is active.
type viewID int

const (
	viewDashboard viewID = iota
	viewStandings
	viewSchedule
	viewWeather
	viewChat
	viewTrackMap
	viewDriverStats
	numViews
)

var tabLabels = []string{
	" 1 Dashboard [WIP] ",
	" 2 Standings ",
	" 3 Schedule  ",
	" 4 Weather   ",
	" 5 Ask AI    ",
	" 6 Track Map ",
	" 7 Driver Stats ",
}

var tabKeys = []string{"1", "2", "3", "4", "5", "6", "7"}

// App is the root bubbletea model. It owns navigation and delegates rendering
// to the current view sub-model.
type App struct {
	cfg     *config.Config
	width   int
	height  int
	current viewID

	dashboard   *views.Dashboard
	standings   *views.Standings
	schedule    *views.Schedule
	weather     *views.WeatherView
	chat        *views.Chat
	trackMap    *views.TrackMap
	driverStats *views.DriverStats
}

func NewApp(cfg *config.Config) *App {
	of1 := openf1.NewClient()
	joli := jolpica.NewClient()
	groqClient := groq.NewClient(cfg.GroqAPIKey, cfg.GroqModel)

	return &App{
		cfg:         cfg,
		current:     viewStandings,
		dashboard:   views.NewDashboard(of1, cfg.RefreshRate),
		standings:   views.NewStandings(joli),
		schedule:    views.NewSchedule(joli),
		weather:     views.NewWeatherView(of1),
		chat:        views.NewChat(groqClient),
		trackMap:    views.NewTrackMap(of1),
		driverStats: views.NewDriverStats(of1, joli),
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
				a.current = viewStandings
				return a, a.initCurrentView()
			}
		case "3":
			if !chatInputActive {
				a.current = viewSchedule
				return a, a.initCurrentView()
			}
		case "4":
			if !chatInputActive {
				a.current = viewWeather
				return a, a.initCurrentView()
			}
		case "5":
			if !chatInputActive {
				a.current = viewChat
				return a, a.initCurrentView()
			}
		case "6":
			if !chatInputActive {
				a.current = viewTrackMap
				return a, a.initCurrentView()
			}
		case "7":
			if !chatInputActive {
				a.current = viewDriverStats
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
	case viewStandings:
		a.standings, cmd = a.standings.UpdateStandings(msg)
	case viewSchedule:
		a.schedule, cmd = a.schedule.UpdateSchedule(msg)
	case viewWeather:
		a.weather, cmd = a.weather.UpdateWeather(msg)
	case viewChat:
		a.chat, cmd = a.chat.UpdateChat(msg)
	case viewTrackMap:
		a.trackMap, cmd = a.trackMap.UpdateTrackMap(msg)
	case viewDriverStats:
		a.driverStats, cmd = a.driverStats.UpdateDriverStats(msg)
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
		return lipgloss.NewStyle().Padding(4, 6).Render(
			styles.Title.Render(" 🚧 Dashboard — Work In Progress ") + "\n\n" +
			styles.DimStyle.Render("  Live timing will be powered by a custom backend (F1 live-timing protocol).\n") +
			styles.DimStyle.Render("  Use the other tabs in the meantime."),
		)
	case viewStandings:
		return a.standings.View()
	case viewSchedule:
		return a.schedule.View()
	case viewWeather:
		return a.weather.View()
	case viewChat:
		return a.chat.View()
	case viewTrackMap:
		return a.trackMap.View()
	case viewDriverStats:
		return a.driverStats.View()
	}
	return ""
}

func (a *App) initCurrentView() tea.Cmd {
	switch a.current {
	case viewDashboard:
		return a.dashboard.Init()
	case viewStandings:
		return a.standings.Init()
	case viewSchedule:
		return a.schedule.Init()
	case viewWeather:
		return a.weather.Init()
	case viewChat:
		return a.chat.Init()
	case viewTrackMap:
		return a.trackMap.Init()
	case viewDriverStats:
		return a.driverStats.Init()
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

	// Top row: brand + tabs
	topRow := lipgloss.JoinHorizontal(lipgloss.Top, brand, "  ", tabRow)

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
