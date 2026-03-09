package views

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/devkeshwani/termf1/internal/api/openf1"
	"github.com/devkeshwani/termf1/internal/ui/styles"
)

// ── Messages ──────────────────────────────────────────────────────────────────

type dashDataMsg struct{ payload *openf1.DashboardPayload }
type dashErrMsg struct{ err error }
type dashTickMsg time.Time

// ── DriverRow: the processed view-model for one car ──────────────────────────

type DriverRow struct {
	Position    int
	Number      int
	Acronym     string
	TeamName    string
	TeamColor   string
	GapToLeader string
	Interval    string
	LastLap     float64
	BestLap     float64
	Sector1     float64
	Sector2     float64
	Sector3     float64
	SpeedTrap   int
	Compound    string
	TyreAge     int
	PitCount    int
	LapNumber   int
	IsPitOutLap bool
	DNF         bool
}

// ── Dashboard model ───────────────────────────────────────────────────────────

type Dashboard struct {
	client      *openf1.Client
	refreshSec  int
	width       int
	height      int
	loading     bool
	err         error
	payload     *openf1.DashboardPayload
	rows        []DriverRow
	rcMessages  []string
	weather     openf1.Weather
	spin        spinner.Model
}

func NewDashboard(client *openf1.Client, refreshSec int) *Dashboard {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(styles.ColorF1Red)
	return &Dashboard{
		client:     client,
		refreshSec: refreshSec,
		loading:    true,
		spin:       s,
	}
}

func (d *Dashboard) SetSize(w, h int) {
	d.width = w
	d.height = h
}

func (d *Dashboard) Init() tea.Cmd {
	return tea.Batch(d.spin.Tick, fetchDashCmd(d.client))
}

func (d *Dashboard) UpdateDash(msg tea.Msg) (*Dashboard, tea.Cmd) {
	switch msg := msg.(type) {
	case spinner.TickMsg:
		if d.loading {
			var cmd tea.Cmd
			d.spin, cmd = d.spin.Update(msg)
			return d, cmd
		}

	case dashDataMsg:
		d.loading = false
		d.err = nil
		d.payload = msg.payload
		d.rows = buildRows(msg.payload)
		rcCount := (d.height - 20) / 2
		if rcCount < 5 {
			rcCount = 5
		}
		if rcCount > 12 {
			rcCount = 12
		}
		d.rcMessages = lastRCMessages(msg.payload.RaceControls, rcCount)
		if len(msg.payload.Weathers) > 0 {
			d.weather = msg.payload.Weathers[len(msg.payload.Weathers)-1]
		}
		return d, tickCmd(d.refreshSec)

	case dashErrMsg:
		d.loading = false
		d.err = msg.err
		return d, tickCmd(d.refreshSec)

	case dashTickMsg:
		d.loading = true
		return d, tea.Batch(d.spin.Tick, fetchDashCmd(d.client))
	}
	return d, nil
}

func (d *Dashboard) View() string {
	if d.loading && d.payload == nil {
		return centred(d.width, d.height,
			d.spin.View()+" Connecting to OpenF1…")
	}
	if d.err != nil && d.payload == nil {
		return centred(d.width, d.height,
			styles.ErrorStyle.Render("⚠  "+d.err.Error()))
	}

	var parts []string

	// Session header bar
	parts = append(parts, d.sessionBar())

	// Dynamic side panel width: ~34% of terminal, clamped
	sideW := d.width * 34 / 100
	if sideW < 46 {
		sideW = 46
	}
	if sideW > 72 {
		sideW = 72
	}

	// Timing table + side panel
	tableStr := d.timingTable(d.width - sideW - 2)
	sideStr := d.sidePanel(sideW - 2)

	tableW := d.width - sideW - 2
	if tableW < 50 {
		tableW = d.width
		sideStr = ""
	}

	tablePanel := lipgloss.NewStyle().Width(tableW).Render(tableStr)
	if sideStr != "" {
		row := lipgloss.JoinHorizontal(lipgloss.Top, tablePanel, " ", sideStr)
		parts = append(parts, row)
	} else {
		parts = append(parts, tablePanel)
	}

	if d.loading {
		parts = append(parts, styles.DimStyle.Render(d.spin.View()+" refreshing…"))
	}

	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}

// ── session info bar ──────────────────────────────────────────────────────────

func (d *Dashboard) sessionBar() string {
	if d.payload == nil {
		return ""
	}
	s := d.payload.Session

	country := styles.BoldWhite.Render(s.CountryName)
	sessionName := lipgloss.NewStyle().Foreground(styles.ColorSubtle).Render(s.SessionName)

	// Current lap info
	lapInfo := ""
	maxLap := 0
	for _, l := range d.payload.Laps {
		if l.LapNumber > maxLap {
			maxLap = l.LapNumber
		}
	}
	if maxLap > 0 {
		lapInfo = lipgloss.NewStyle().Foreground(styles.ColorTeal).
			Render(fmt.Sprintf("  Lap %d", maxLap))
	}

	// Track status from latest race control
	statusLabel, statusColor := d.trackStatus()
	flag := lipgloss.NewStyle().Foreground(statusColor).Bold(true).Render("● " + statusLabel)

	// DRS info from session
	drsInfo := ""
	if s.SessionType == "Race" {
		drsInfo = lipgloss.NewStyle().Foreground(styles.ColorGreen).Render("  DRS ✓")
	}

	parts := []string{country, " ", sessionName, lapInfo, drsInfo, "   ", flag}
	return lipgloss.NewStyle().
		Background(styles.ColorBgCard).
		Width(d.width).
		Padding(0, 1).
		Render(strings.Join(parts, ""))
}

// trackStatus derives the current track status from the most recent race-control message.
func (d *Dashboard) trackStatus() (string, lipgloss.Color) {
	if len(d.rcMessages) == 0 {
		return "LIVE", styles.ColorGreen
	}
	latest := strings.ToUpper(d.rcMessages[0])
	switch {
	case strings.Contains(latest, "SAFETY CAR DEPLOYED") || strings.Contains(latest, "SAFETY CAR IN"):
		return "🚗 SAFETY CAR", styles.ColorYellow
	case strings.Contains(latest, "VIRTUAL SAFETY CAR"):
		return "VSC", styles.ColorYellow
	case strings.Contains(latest, "RED FLAG"):
		return "🔴 RED FLAG", styles.ColorF1Red
	case strings.Contains(latest, "YELLOW") && strings.Contains(latest, "SECTOR"):
		return "⚠ YELLOW", styles.ColorYellow
	case strings.Contains(latest, "CHEQUERED"):
		return "🏁 FINISHED", styles.ColorPurple
	case strings.Contains(latest, "TRACK CLEAR") || strings.Contains(latest, "GREEN"):
		return "🟢 TRACK CLEAR", styles.ColorGreen
	default:
		return "● LIVE", styles.ColorGreen
	}
}

// ── timing table ─────────────────────────────────────────────────────────────

// timingCols defines the fixed columns. Dynamic width ones are handled in timingTable().
var timingCols = []struct {
	header string
	width  int
}{
	{"POS", 4},
	{"DRIVER", 8},
	{"TYR", 5},
	{"LAP", 4},
	{"     GAP", 12},
	{" INTERVAL", 12},
	{" LAST LAP", 12},
	{" BEST LAP", 12},
	{"    S1", 8},
	{"    S2", 8},
	{"    S3", 8},
	{" SPEED", 7},
	{"PIT", 4},
}

func (d *Dashboard) timingTable(availW int) string {
	headerCells := make([]string, len(timingCols))
	for i, col := range timingCols {
		headerCells[i] = lipgloss.NewStyle().
			Width(col.width).
			Foreground(styles.ColorSubtle).
			Bold(true).
			Render(col.header)
	}
	header := lipgloss.JoinHorizontal(lipgloss.Top, headerCells...)
	sep := styles.Divider.Render(strings.Repeat("─", colsWidth()))

	lines := []string{header, sep}
	for _, row := range d.rows {
		lines = append(lines, d.renderRow(row))
	}

	if len(d.rows) == 0 && d.payload != nil {
		lines = append(lines, styles.DimStyle.Padding(1, 2).Render("No timing data available for this session."))
	}

	_ = availW // reserved for responsive column stretching
	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func colsWidth() int {
	w := 0
	for _, c := range timingCols {
		w += c.width
	}
	return w
}

func (d *Dashboard) renderRow(row DriverRow) string {
	// Position
	posStr := fmt.Sprintf("%2d", row.Position)
	posCell := lipgloss.NewStyle().Width(timingCols[0].width).
		Foreground(posColor(row.Position)).Bold(row.Position <= 3).
		Render(posStr)

	// Driver badge with team colour
	teamCol := styles.TeamColor(row.TeamName)
	if row.TeamColor != "" {
		teamCol = lipgloss.Color("#" + row.TeamColor)
	}
	badge := lipgloss.NewStyle().
		Background(teamCol).
		Foreground(lipgloss.Color("#000000")).
		Bold(true).
		Padding(0, 0).
		Render(" " + row.Acronym + " ")
	driverCell := lipgloss.NewStyle().Width(timingCols[1].width).Render(badge)

	// Tyre badge
	tc := styles.TyreColor(row.Compound)
	tl := styles.TyreLabel(row.Compound)
	age := ""
	if row.TyreAge > 0 {
		age = fmt.Sprintf("%d", row.TyreAge)
	}
	tyreStr := lipgloss.NewStyle().Foreground(tc).Bold(true).Render(tl) + styles.DimStyle.Render(age)
	tyreCell := lipgloss.NewStyle().Width(timingCols[2].width).Render(tyreStr)

	// Lap number
	lapStr := styles.DimStyle.Render("-")
	if row.LapNumber > 0 {
		lapStr = lipgloss.NewStyle().Foreground(styles.ColorSubtle).Render(fmt.Sprintf("%3d", row.LapNumber))
	}
	lapCell := lipgloss.NewStyle().Width(timingCols[3].width).Render(lapStr)

	// Gap & interval
	gapCell := fixedRight(row.GapToLeader, timingCols[4].width, styles.ColorText)
	intCell := fixedRight(row.Interval, timingCols[5].width, styles.ColorTextDim)

	// Last lap
	lastStr, lastColor := formatLap(row.LastLap, row.BestLap)
	lastCell := fixedRight(lastStr, timingCols[6].width, lastColor)

	// Best lap
	bestStr := formatDuration(row.BestLap)
	bestCell := fixedRight(bestStr, timingCols[7].width, styles.ColorSubtle)

	// Sector 1, 2, 3
	s1Cell := fixedRight(formatSector(row.Sector1), timingCols[8].width, styles.ColorTextDim)
	s2Cell := fixedRight(formatSector(row.Sector2), timingCols[9].width, styles.ColorTextDim)
	s3Cell := fixedRight(formatSector(row.Sector3), timingCols[10].width, styles.ColorTextDim)

	// Speed trap
	speedStr := "-"
	if row.SpeedTrap > 0 {
		speedStr = fmt.Sprintf("%d", row.SpeedTrap)
	}
	speedCell := fixedRight(speedStr, timingCols[11].width,
		func() lipgloss.Color {
			if row.SpeedTrap > 330 {
				return styles.ColorPurple
			}
			if row.SpeedTrap > 310 {
				return styles.ColorGreen
			}
			return styles.ColorTextDim
		}())

	// Pits
	pitStr := ""
	if row.PitCount > 0 {
		pitStr = fmt.Sprintf("P%d", row.PitCount)
	}
	pitCell := lipgloss.NewStyle().Width(timingCols[12].width).
		Foreground(styles.ColorOrange).Render(pitStr)

	line := lipgloss.JoinHorizontal(lipgloss.Top,
		posCell, driverCell, tyreCell, lapCell,
		gapCell, intCell, lastCell, bestCell,
		s1Cell, s2Cell, s3Cell, speedCell, pitCell,
	)

	// Highlight leader row
	if row.Position == 1 {
		line = lipgloss.NewStyle().
			Background(lipgloss.Color("#1A1A2E")).
			Render(line)
	}
	return line
}

// ── side panel (weather + race control + mini circuit) ───────────────────────

func (d *Dashboard) sidePanel(w int) string {
	weatherBox := d.weatherWidget(w)
	rcBox := d.rcWidget(w)
	miniMap := d.miniTrackWidget(w)

	return lipgloss.JoinVertical(lipgloss.Left, weatherBox, "", rcBox, "", miniMap)
}

func (d *Dashboard) miniTrackWidget(w int) string {
	circuitName := "default"
	if d.payload != nil && d.payload.Session.CircuitShortName != "" {
		circuitName = d.payload.Session.CircuitShortName
	}
	art := getCircuitArt(circuitName)
	title := styles.DimStyle.Render("Circuit: ") + lipgloss.NewStyle().Foreground(styles.ColorSubtle).Render(circuitName)
	lines := []string{title, ""}
	for _, line := range art {
		// trim to fit side panel width
		visible := line
		if len(line) > w-4 {
			visible = line[:w-4]
		}
		lines = append(lines, lipgloss.NewStyle().Foreground(lipgloss.Color("#4FC3F7")).Render(visible))
	}
	return styles.Card.Width(w).Render(strings.Join(lines, "\n"))
}

func (d *Dashboard) weatherWidget(w int) string {
	wt := d.weather
	if wt.Date == "" {
		return styles.Card.Width(w).Render(styles.DimStyle.Render("No weather data"))
	}

	rain := "Dry"
	if wt.Rainfall > 0 {
		rain = lipgloss.NewStyle().Foreground(styles.ColorBlue).Render("Rain")
	}

	lines := []string{
		styles.BoldWhite.Render("⛅ Weather"),
		"",
		fmt.Sprintf("  Air    %s%.1f°C", lipgloss.NewStyle().Foreground(styles.ColorOrange).Render(""), wt.AirTemperature),
		fmt.Sprintf("  Track  %s%.1f°C", lipgloss.NewStyle().Foreground(styles.ColorF1Red).Render(""), wt.TrackTemperature),
		fmt.Sprintf("  Humid  %.0f%%", wt.Humidity),
		fmt.Sprintf("  Press  %.1f hPa", wt.Pressure),
		fmt.Sprintf("  Wind   %.1fm/s %s", wt.WindSpeed, styles.WindDirection(wt.WindDirection)),
		fmt.Sprintf("  Rain   %s", rain),
	}

	return styles.Card.Width(w).Render(strings.Join(lines, "\n"))
}

func (d *Dashboard) rcWidget(w int) string {
	if len(d.rcMessages) == 0 {
		return styles.Card.Width(w).Render(styles.DimStyle.Render("No race control messages"))
	}

	lines := []string{styles.BoldWhite.Render("📡 Race Control"), ""}
	for _, m := range d.rcMessages {
		// wrap at panel width minus padding
		wrapped := wordWrapMulti(m, w-4)
		for _, line := range strings.Split(wrapped, "\n") {
			lines = append(lines, lipgloss.NewStyle().Foreground(styles.ColorTextDim).Render("  "+line))
		}
	}
	return styles.Card.Width(w).Render(strings.Join(lines, "\n"))
}

// wordWrapMulti wraps text into multiple lines at word boundaries.
func wordWrapMulti(s string, width int) string {
	if width <= 0 || len(s) <= width {
		return s
	}
	var lines []string
	words := strings.Fields(s)
	var current strings.Builder
	for _, w := range words {
		if current.Len() > 0 && current.Len()+1+len(w) > width {
			lines = append(lines, current.String())
			current.Reset()
		}
		if current.Len() > 0 {
			current.WriteByte(' ')
		}
		current.WriteString(w)
	}
	if current.Len() > 0 {
		lines = append(lines, current.String())
	}
	return strings.Join(lines, "\n")
}

// ── data processing ────────────────────────────────────────────────────────────

func buildRows(p *openf1.DashboardPayload) []DriverRow {
	if p == nil {
		return nil
	}

	// Build driver map
	driverMap := make(map[int]openf1.Driver, len(p.Drivers))
	for _, d := range p.Drivers {
		driverMap[d.DriverNumber] = d
	}

	// Latest position per driver
	posMap := make(map[int]int)
	for _, pos := range p.Positions {
		posMap[pos.DriverNumber] = pos.Position
	}

	// Latest interval per driver
	gapMap := make(map[int]openf1.Interval)
	for _, iv := range p.Intervals {
		gapMap[iv.DriverNumber] = iv
	}

	// Best lap per driver (min lap_duration > 0)
	bestLapMap := make(map[int]float64)
	lastLapMap := make(map[int]openf1.Lap)
	lapNumMap := make(map[int]int)
	for _, l := range p.Laps {
		if l.LapDuration > 0 {
			if cur, ok := bestLapMap[l.DriverNumber]; !ok || l.LapDuration < cur {
				bestLapMap[l.DriverNumber] = l.LapDuration
			}
		}
		if l.LapNumber > lapNumMap[l.DriverNumber] {
			lapNumMap[l.DriverNumber] = l.LapNumber
			lastLapMap[l.DriverNumber] = l
		}
	}

	// Current stint per driver
	stintMap := make(map[int]openf1.Stint)
	for _, st := range p.Stints {
		if cur, ok := stintMap[st.DriverNumber]; !ok || st.StintNumber > cur.StintNumber {
			stintMap[st.DriverNumber] = st
		}
	}

	// Pit count per driver
	pitCount := make(map[int]int)
	for _, pit := range p.Pits {
		pitCount[pit.DriverNumber]++
	}

	// Assemble rows
	rows := make([]DriverRow, 0, len(driverMap))
	for num, drv := range driverMap {
		pos := posMap[num]
		if pos == 0 {
			pos = 99
		}

		iv := gapMap[num]
		lastLap := lastLapMap[num]
		stint := stintMap[num]

		tyreAge := stint.TyreAgeAtStart
		if lapNumMap[num] > 0 && stint.LapStart > 0 {
			tyreAge = stint.TyreAgeAtStart + (lapNumMap[num] - stint.LapStart + 1)
		}

		row := DriverRow{
			Position:    pos,
			Number:      num,
			Acronym:     drv.NameAcronym,
			TeamName:    drv.TeamName,
			TeamColor:   drv.TeamColour,
			GapToLeader: formatGap(iv.GapToLeader),
			Interval:    formatGap(iv.Interval),
			LastLap:     lastLap.LapDuration,
			BestLap:     bestLapMap[num],
			Sector1:     lastLap.DurationSector1,
			Sector2:     lastLap.DurationSector2,
			Sector3:     lastLap.DurationSector3,				SpeedTrap:   lastLap.StSpeed,			Compound:    stint.Compound,
			TyreAge:     tyreAge,
			PitCount:    pitCount[num],
			LapNumber:   lapNumMap[num],
			IsPitOutLap: lastLap.IsPitOutLap,
		}
		rows = append(rows, row)
	}

	sort.Slice(rows, func(i, j int) bool {
		return rows[i].Position < rows[j].Position
	})
	return rows
}

func lastRCMessages(rcs []openf1.RaceControl, n int) []string {
	if len(rcs) == 0 {
		return nil
	}
	start := len(rcs) - n
	if start < 0 {
		start = 0
	}
	out := make([]string, 0, n)
	for _, rc := range rcs[start:] {
		out = append(out, rc.Message)
	}
	// Reverse so newest is first
	for i, j := 0, len(out)-1; i < j; i, j = i+1, j-1 {
		out[i], out[j] = out[j], out[i]
	}
	return out
}

// ── commands ──────────────────────────────────────────────────────────────────

func fetchDashCmd(client *openf1.Client) tea.Cmd {
	return func() tea.Msg {
		p, err := client.FetchDashboard(contextBG())
		if err != nil {
			return dashErrMsg{err}
		}
		return dashDataMsg{p}
	}
}

func tickCmd(seconds int) tea.Cmd {
	return tea.Tick(time.Duration(seconds)*time.Second, func(t time.Time) tea.Msg {
		return dashTickMsg(t)
	})
}

// ── formatting helpers ────────────────────────────────────────────────────────

func formatGap(v interface{}) string {
	if v == nil {
		return "LEADER"
	}
	switch val := v.(type) {
	case float64:
		if val == 0 {
			return "LEADER"
		}
		return fmt.Sprintf("+%.3f", val)
	case string:
		return val
	}
	return "—"
}

func formatDuration(secs float64) string {
	if secs <= 0 {
		return "—"
	}
	mins := int(secs) / 60
	rem := secs - float64(mins*60)
	if mins > 0 {
		return fmt.Sprintf("%d:%06.3f", mins, rem)
	}
	return fmt.Sprintf("%.3f", secs)
}

func formatSector(secs float64) string {
	if secs <= 0 {
		return "—"
	}
	return fmt.Sprintf("%.3f", secs)
}

func formatLap(last, best float64) (string, lipgloss.Color) {
	s := formatDuration(last)
	if last <= 0 {
		return s, styles.ColorSubtle
	}
	if best > 0 && math.Abs(last-best) < 0.001 {
		return s, styles.ColorPurple // personal/session best
	}
	return s, styles.ColorText
}

func posColor(pos int) lipgloss.Color {
	switch pos {
	case 1:
		return styles.ColorYellow
	case 2:
		return lipgloss.Color("#C0C0C0")
	case 3:
		return lipgloss.Color("#CD7F32")
	default:
		return styles.ColorText
	}
}

func fixedRight(s string, width int, color lipgloss.Color) string {
	return lipgloss.NewStyle().Width(width).Align(lipgloss.Right).Foreground(color).Render(s)
}

func wrapText(s string, width int) string {
	if width <= 0 || len(s) <= width {
		return s
	}
	return s[:width-1] + "…"
}

func centred(w, h int, s string) string {
	if h < 1 {
		h = 1
	}
	topPad := (h - 1) / 2
	return strings.Repeat("\n", topPad) +
		lipgloss.NewStyle().Width(w).Align(lipgloss.Center).Render(s)
}
