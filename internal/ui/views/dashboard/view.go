package dashboard

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/dk-a-dev/termf1/v2/internal/ui/views/common"
	"github.com/dk-a-dev/termf1/v2/internal/api/liveserver"
	"github.com/dk-a-dev/termf1/v2/internal/ui/styles"
)

// View renders the full dashboard.
func (d *Dashboard2) View() string {
	if d.loading && d.liveState == nil && d.fallback == nil {
		return common.Centred(d.width, d.height,
			d.spin.View()+" Connecting to live timing…")
	}
	if d.err != nil && d.liveState == nil && d.fallback == nil {
		return common.Centred(d.width, d.height,
			styles.ErrorStyle.Render("⚠  "+d.err.Error()))
	}

	lines := []string{
		d.renderTopBar(),
		d.renderMainArea(),
		d.renderTicker(),
	}
	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

// ── Top bar ───────────────────────────────────────────────────────────────────

func (d *Dashboard2) renderTopBar() string {
	var session, circuit, lapInfo, statusPill, weatherPill, serverPill string

	hasLive := d.serverAlive && len(d.liveRows) > 0

	switch {
	case hasLive:
		// Active race/session.
		st := d.liveState
		session = styles.BoldWhite.Render(st.SessionInfo.Meeting.Name)
		if st.SessionInfo.Name != "" {
			session += "  " + lipgloss.NewStyle().Foreground(styles.ColorSubtle).Render(st.SessionInfo.Name)
		}
		circuit = lipgloss.NewStyle().Foreground(styles.ColorTeal).Render("  ⬡ " + st.SessionInfo.Meeting.Circuit.ShortName)
		if st.LapCount.TotalLaps > 0 {
			lapInfo = lipgloss.NewStyle().Foreground(styles.ColorYellow).
				Render(fmt.Sprintf("  LAP %d / %d", st.LapCount.CurrentLap, st.LapCount.TotalLaps))
		} else if st.LapCount.CurrentLap > 0 {
			lapInfo = lipgloss.NewStyle().Foreground(styles.ColorYellow).
				Render(fmt.Sprintf("  LAP %d", st.LapCount.CurrentLap))
		}
		statusPill = d.liveStatusPillStr(st.SessionStatus, st.RaceControlMessages)
		weatherPill = d.liveWeatherPill(st.Weather)
		serverPill = lipgloss.NewStyle().
			Foreground(styles.ColorGreen).Background(lipgloss.Color("#052e16")).
			Padding(0, 1).Render("● LIVE")

	case d.serverAlive && d.fallback != nil:
		// Server connected, no active session — showing last race.
		fb := d.fallback
		session = styles.BoldWhite.Render(fb.Session.CountryName)
		if fb.Session.SessionName != "" {
			session += "  " + lipgloss.NewStyle().Foreground(styles.ColorSubtle).Render(fb.Session.SessionName)
		}
		circuit = lipgloss.NewStyle().Foreground(styles.ColorTeal).Render("  ⬡ " + fb.Session.CircuitShortName)
		statusPill = lipgloss.NewStyle().
			Foreground(styles.ColorSubtle).Background(styles.ColorBgCard).
			Padding(0, 1).Render("LAST RACE")
		serverPill = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#38bdf8")).Background(lipgloss.Color("#0c1a2e")).
			Padding(0, 1).Render("○ CONNECTED")
		if len(fb.Weathers) > 0 {
			wt := fb.Weathers[len(fb.Weathers)-1]
			weatherPill = fmt.Sprintf("  🌡 %.0f°C  💨 %.0fm/s", wt.AirTemperature, wt.WindSpeed)
		}

	case d.serverAlive:
		// Server connected, waiting for data.
		session = lipgloss.NewStyle().Foreground(styles.ColorSubtle).Render("Connecting to live timing…")
		serverPill = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#38bdf8")).Background(lipgloss.Color("#0c1a2e")).
			Padding(0, 1).Render("○ CONNECTED")

	default:
		// Server offline, showing historical data.
		if d.fallback != nil {
			fb := d.fallback
			session = styles.BoldWhite.Render(fb.Session.CountryName)
			if fb.Session.SessionName != "" {
				session += "  " + lipgloss.NewStyle().Foreground(styles.ColorSubtle).Render(fb.Session.SessionName)
			}
			circuit = lipgloss.NewStyle().Foreground(styles.ColorTeal).Render("  ⬡ " + fb.Session.CircuitShortName)
			if len(fb.Weathers) > 0 {
				wt := fb.Weathers[len(fb.Weathers)-1]
				weatherPill = fmt.Sprintf("  🌡 %.0f°C  💨 %.0fm/s", wt.AirTemperature, wt.WindSpeed)
			}
		}
		statusPill = lipgloss.NewStyle().
			Foreground(styles.ColorSubtle).Background(styles.ColorBgCard).
			Padding(0, 1).Render("HISTORICAL")
		serverPill = lipgloss.NewStyle().
			Foreground(styles.ColorOrange).Background(lipgloss.Color("#1c0d00")).
			Padding(0, 1).Render("○ SERVER OFFLINE")
	}

	right := lipgloss.JoinHorizontal(lipgloss.Center, weatherPill, "   ", statusPill, "  ", serverPill)
	rightWidth := lipgloss.Width(right)
	leftWidth := d.width - rightWidth - 2
	left := lipgloss.NewStyle().Width(leftWidth).Render(
		lipgloss.JoinHorizontal(lipgloss.Center, session, circuit, lapInfo),
	)

	bar := lipgloss.JoinHorizontal(lipgloss.Center, left, right)
	return lipgloss.NewStyle().
		Background(styles.ColorBgCard).
		Width(d.width).
		Padding(0, 1).
		Render(bar)
}

// liveStatusPill derives the session status label and colour from RCM + session status.
func (d *Dashboard2) liveStatusPill(status string, rcms []liveserver.RaceControlMsg) (string, lipgloss.Color) {
	for i := len(rcms) - 1; i >= 0; i-- {
		msg := strings.ToUpper(rcms[i].Message)
		if strings.Contains(msg, "SAFETY CAR DEPLOYED") {
			return "🚗 SC", styles.ColorYellow
		}
		if strings.Contains(msg, "VIRTUAL SAFETY CAR") && !strings.Contains(msg, "ENDING") {
			return "VSC", styles.ColorYellow
		}
		if strings.Contains(msg, "RED FLAG") {
			return "🔴 RED FLAG", styles.ColorF1Red
		}
		if strings.Contains(msg, "TRACK CLEAR") || strings.Contains(msg, "GREEN LIGHT") {
			break
		}
	}
	switch status {
	case "Started":
		return "🟢 LIVE", styles.ColorGreen
	case "Finished", "Ends":
		return "🏁 FINISHED", styles.ColorPurple
	case "Aborted":
		return "🔴 ABORTED", styles.ColorF1Red
	default:
		return "● " + status, styles.ColorSubtle
	}
}

// liveStatusPillStr renders the status label as a styled pill string.
func (d *Dashboard2) liveStatusPillStr(status string, rcms []liveserver.RaceControlMsg) string {
	label, color := d.liveStatusPill(status, rcms)
	return lipgloss.NewStyle().Foreground(color).Padding(0, 1).
		Background(lipgloss.Color("#111827")).Render(label)
}

// liveWeatherPill renders a compact weather summary.
func (d *Dashboard2) liveWeatherPill(wt liveserver.WeatherData) string {
	if wt.AirTemp == "" {
		return ""
	}
	rain := ""
	if wt.Rainfall != "" && wt.Rainfall != "0" && wt.Rainfall != "0.0" {
		rain = " 🌧"
	}
	return lipgloss.NewStyle().Foreground(styles.ColorTextDim).Render(
		fmt.Sprintf("🌡 %s°C  💨 %s m/s%s", wt.AirTemp, wt.WindSpeed, rain),
	)
}

// ── Main area ─────────────────────────────────────────────────────────────────

func (d *Dashboard2) renderMainArea() string {
	// reserve rows: 1 top-bar + 1 sep + 1 ticker + 1 footer = 4; add 2 buffer
	contentH := d.height - 6
	if contentH < 10 {
		contentH = 10
	}

	// Wider right panel on larger screens for better weather/RCM visibility.
	rightW := d.width * 36 / 100
	if d.width >= 200 {
		rightW = d.width * 32 / 100
	}
	if rightW < 44 {
		rightW = 44
	}
	if rightW > 80 {
		rightW = 80
	}
	leftW := d.width - rightW - 2 // 2-char gutter

	left := d.renderTimingPanel(leftW, contentH)
	gutter := lipgloss.NewStyle().Width(1).Height(contentH).Render("")
	right := d.renderSidePanel(rightW, contentH)

	return lipgloss.JoinHorizontal(lipgloss.Top, left, gutter, right)
}
