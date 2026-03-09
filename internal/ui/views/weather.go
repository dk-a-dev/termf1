package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/devkeshwani/termf1/internal/api/openf1"
	"github.com/devkeshwani/termf1/internal/ui/styles"
)

// ── Messages ──────────────────────────────────────────────────────────────────

type weatherDataMsg struct {
	session  openf1.Session
	weathers []openf1.Weather
}
type weatherErrMsg struct{ err error }

// ── Model ─────────────────────────────────────────────────────────────────────

type WeatherView struct {
	client   *openf1.Client
	width    int
	height   int
	loading  bool
	err      error
	session  openf1.Session
	weathers []openf1.Weather
	spin     spinner.Model
}

func NewWeatherView(client *openf1.Client) *WeatherView {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(styles.ColorF1Red)
	return &WeatherView{client: client, loading: true, spin: s}
}

func (w *WeatherView) SetSize(width, h int) { w.width = width; w.height = h }

func (w *WeatherView) Init() tea.Cmd {
	return tea.Batch(w.spin.Tick, fetchWeatherCmd(w.client))
}

func (w *WeatherView) UpdateWeather(msg tea.Msg) (*WeatherView, tea.Cmd) {
	switch msg := msg.(type) {
	case spinner.TickMsg:
		if w.loading {
			var cmd tea.Cmd
			w.spin, cmd = w.spin.Update(msg)
			return w, cmd
		}
	case weatherDataMsg:
		w.loading = false
		w.err = nil
		w.session = msg.session
		w.weathers = msg.weathers
	case weatherErrMsg:
		w.loading = false
		w.err = msg.err
	}
	return w, nil
}

func (w *WeatherView) View() string {
	if w.loading && len(w.weathers) == 0 {
		return centred(w.width, w.height, w.spin.View()+" Loading weather…")
	}
	if w.err != nil && len(w.weathers) == 0 {
		return centred(w.width, w.height, styles.ErrorStyle.Render("⚠  "+w.err.Error()))
	}
	if len(w.weathers) == 0 {
		return centred(w.width, w.height, styles.DimStyle.Render("No weather data available."))
	}

	latest := w.weathers[len(w.weathers)-1]

	title := styles.Title.Render(fmt.Sprintf(" ⛅ Weather  —  %s %s",
		w.session.CountryName, w.session.SessionName))
	sep := styles.Divider.Render(strings.Repeat("─", w.width))

	// Main metrics grid
	cardW := 22
	cards := []string{
		w.metricCard("🌡  Air Temperature", fmt.Sprintf("%.1f °C", latest.AirTemperature),
			styles.ColorOrange, cardW),
		w.metricCard("🔥 Track Temperature", fmt.Sprintf("%.1f °C", latest.TrackTemperature),
			styles.ColorF1Red, cardW),
		w.metricCard("💧 Humidity", fmt.Sprintf("%.0f %%", latest.Humidity),
			styles.ColorBlue, cardW),
		w.metricCard("🌬  Wind Speed", fmt.Sprintf("%.1f m/s  %s",
			latest.WindSpeed, styles.WindDirection(latest.WindDirection)),
			styles.ColorTeal, cardW),
		w.metricCard("🧭 Pressure", fmt.Sprintf("%.1f hPa", latest.Pressure),
			styles.ColorSubtle, cardW),
		w.metricCard("🌧  Rainfall", rainfallLabel(latest.Rainfall),
			styles.ColorBlue, cardW),
	}

	row1 := lipgloss.JoinHorizontal(lipgloss.Top, cards[0], "  ", cards[1], "  ", cards[2])
	row2 := lipgloss.JoinHorizontal(lipgloss.Top, cards[3], "  ", cards[4], "  ", cards[5])
	metricsBlock := lipgloss.JoinVertical(lipgloss.Left, row1, "", row2)

	// Temperature trend chart
	trendTitle := lipgloss.NewStyle().Bold(true).Foreground(styles.ColorSubtle).Render("  Temperature Trend (last 20 samples)")
	airTrend := w.trendLine(w.weathers, func(wt openf1.Weather) float64 { return wt.AirTemperature },
		styles.ColorOrange, "Air", w.width-6)
	trackTrend := w.trendLine(w.weathers, func(wt openf1.Weather) float64 { return wt.TrackTemperature },
		styles.ColorF1Red, "Trk", w.width-6)

	return lipgloss.JoinVertical(lipgloss.Left,
		title, sep, "",
		"  "+metricsBlock, "",
		trendTitle, "",
		airTrend,
		trackTrend,
	)
}

func (w *WeatherView) metricCard(label, value string, color lipgloss.Color, width int) string {
	l := styles.DimStyle.Render(label)
	v := lipgloss.NewStyle().Bold(true).Foreground(color).Render(value)
	return styles.Card.Width(width).Render(
		lipgloss.JoinVertical(lipgloss.Left, l, v),
	)
}

// trendLine renders a compact sparkline for a weather metric.
func (w *WeatherView) trendLine(samples []openf1.Weather, extract func(openf1.Weather) float64,
	color lipgloss.Color, label string, width int) string {

	// Take last N samples
	n := 20
	if len(samples) < n {
		n = len(samples)
	}
	data := make([]float64, n)
	for i, s := range samples[len(samples)-n:] {
		data[i] = extract(s)
	}

	if n == 0 {
		return ""
	}

	minVal, maxVal := data[0], data[0]
	for _, v := range data {
		if v < minVal {
			minVal = v
		}
		if v > maxVal {
			maxVal = v
		}
	}

	// Braille-style sparkline using block chars
	levels := []string{"▁", "▂", "▃", "▄", "▅", "▆", "▇", "█"}
	bar := ""
	diff := maxVal - minVal
	for _, v := range data {
		idx := 0
		if diff > 0 {
			idx = int((v-minVal)/diff*float64(len(levels)-1) + 0.5)
		}
		if idx >= len(levels) {
			idx = len(levels) - 1
		}
		bar += levels[idx]
	}

	spark := lipgloss.NewStyle().Foreground(color).Render(bar)
	labelStr := lipgloss.NewStyle().Width(5).Foreground(styles.ColorSubtle).Render(label)
	minStr := lipgloss.NewStyle().Foreground(styles.ColorTextDim).Render(fmt.Sprintf(" %.1f", minVal))
	maxStr := lipgloss.NewStyle().Foreground(color).Bold(true).Render(fmt.Sprintf("→%.1f°C", maxVal))

	return "  " + labelStr + " " + minStr + " " + spark + " " + maxStr
}

// ── helpers ───────────────────────────────────────────────────────────────────

func rainfallLabel(r int) string {
	if r == 0 {
		return "Dry"
	}
	return lipgloss.NewStyle().Foreground(styles.ColorBlue).Render(fmt.Sprintf("%d mm", r))
}

func fetchWeatherCmd(client *openf1.Client) tea.Cmd {
	return func() tea.Msg {
		ctx := contextBG()
		session, err := client.GetLatestSession(ctx)
		if err != nil {
			return weatherErrMsg{err}
		}
		weathers, err := client.GetWeather(ctx, session.SessionKey)
		if err != nil {
			return weatherErrMsg{err}
		}
		return weatherDataMsg{session, weathers}
	}
}
