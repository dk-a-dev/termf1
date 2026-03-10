package trackmap

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbletea"
	"github.com/devkeshwani/termf1/internal/ui/views/common"
	"github.com/charmbracelet/lipgloss"
	"github.com/devkeshwani/termf1/internal/api/multiviewer"
	"github.com/devkeshwani/termf1/internal/api/openf1"
	"github.com/devkeshwani/termf1/internal/ui/styles"
)

// ── Messages ──────────────────────────────────────────────────────────────────

type trackMapDataMsg struct {
	session   openf1.Session
	drivers   []openf1.Driver
	positions []openf1.Position
	laps      []openf1.Lap
}
type trackMapErrMsg struct{ err error }
type trackMapTickMsg time.Time
type trackMapCircuitMsg struct{ circuit *multiviewer.Circuit }
type trackMapCircuitErrMsg struct{ err error }
type trackMapSpeedMsg struct {
	driverNum int
	carData   []openf1.CarData
	locations []openf1.CarLocation
}
type trackMapSpeedErrMsg struct{ err error }


// ── Model ─────────────────────────────────────────────────────────────────────

type TrackMap struct {
	client    *openf1.Client
	mv        *multiviewer.Client
	width     int
	height    int
	loading   bool
	err       error
	session   openf1.Session
	drivers   []openf1.Driver
	positions []openf1.Position
	laps      []openf1.Lap
	circuit   *multiviewer.Circuit // live circuit layout from Multiviewer API
	spin      spinner.Model
	view      int // 0=map 1=speed heatmap 2=sector table
	// speed heatmap state
	segmentSpeeds    []float64
	heatDriverNum    int    // currently selected driver number for heatmap
	heatDriverAcronym string
	heatLoading      bool
}

func NewTrackMap(client *openf1.Client) *TrackMap {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(styles.ColorF1Red)
	return &TrackMap{
		client:  client,
		mv:      multiviewer.NewClient(),
		loading: true,
		spin:    s,
	}
}

func (t *TrackMap) SetSize(w, h int) { t.width = w; t.height = h }

func (t *TrackMap) Init() tea.Cmd {
	return tea.Batch(t.spin.Tick, fetchTrackMapCmd(t.client))
}

func (t *TrackMap) UpdateTrackMap(msg tea.Msg) (*TrackMap, tea.Cmd) {
	switch msg := msg.(type) {
	case spinner.TickMsg:
		if t.loading {
			var cmd tea.Cmd
			t.spin, cmd = t.spin.Update(msg)
			return t, cmd
		}
	case trackMapDataMsg:
		t.loading = false
		t.err = nil
		t.session = msg.session
		t.drivers = msg.drivers
		t.positions = msg.positions
		t.laps = msg.laps
		// Use session.CircuitKey directly — OpenF1 circuit keys match Multiviewer API keys 1:1.
		year := time.Now().Year()
		circFetch := fetchCircuitByKeyCmd(t.mv, t.session.CircuitKey, year)
		tickCmd := tea.Tick(10*time.Second, func(tt time.Time) tea.Msg {
			return trackMapTickMsg(tt)
		})
		return t, tea.Batch(tickCmd, circFetch)

	case trackMapCircuitMsg:
		t.circuit = msg.circuit
		return t, nil

	case trackMapCircuitErrMsg:
		// Fallback to ASCII art — not fatal
		return t, nil
	case trackMapErrMsg:
		t.loading = false
		t.err = msg.err

	case trackMapTickMsg:
		// live mode: refresh data from API
		return t, fetchTrackMapCmd(t.client)

	case trackMapSpeedMsg:
		if msg.driverNum == t.heatDriverNum {
			t.heatLoading = false
			if t.circuit != nil {
				t.segmentSpeeds = buildSegmentSpeeds(t.circuit, msg.locations, msg.carData)
			}
		}
		return t, nil

	case trackMapSpeedErrMsg:
		t.heatLoading = false
		return t, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "r":
			t.loading = true
			return t, tea.Batch(t.spin.Tick, fetchTrackMapCmd(t.client))
		case "v":
			t.view = (t.view + 1) % 3
			// When entering speed heatmap view, fetch data for first driver if not loaded
			if t.view == 1 && len(t.segmentSpeeds) == 0 && !t.heatLoading {
				return t, t.triggerSpeedFetch()
			}
			return t, nil
		case "n":
			if t.view == 1 {
				return t, t.cycleHeatDriver(1)
			}
		case "p":
			if t.view == 1 {
				return t, t.cycleHeatDriver(-1)
			}
		}
	}
	return t, nil
}

func (t *TrackMap) View() string {
	if t.loading {
		return common.Centred(t.width, t.height, t.spin.View()+" Loading track data…")
	}

	title := styles.Title.Render(fmt.Sprintf(" 🗺  Track Map  —  %s %s",
		t.session.CountryName, t.session.SessionName))
	sep := styles.Divider.Render(strings.Repeat("─", t.width))

	// Route to appropriate sub-view.
	if t.view == 1 {
		hint := styles.FooterStyle.Render("  v: sector table  │  n/p: driver  │  r: refresh")
		content := t.renderSpeedHeatmap()
		return lipgloss.JoinVertical(lipgloss.Left, title, sep, "", content, "", hint)
	}
	if t.view == 2 {
		hint := styles.FooterStyle.Render("  v: track map  │  r: refresh")
		content := t.renderHeatmap()
		return lipgloss.JoinVertical(lipgloss.Left, title, sep, "", content, "", hint)
	}

	circuitName := t.session.CircuitShortName
	if circuitName == "" {
		circuitName = "default"
	}

	driverRows := t.buildLiveDots()
	hasLiveData := len(t.positions) > 0

	hintParts := []string{"  v: speed heatmap  │  r: refresh"}
	if !hasLiveData {
		hintParts = append(hintParts,
			lipgloss.NewStyle().Foreground(styles.ColorYellow).Render("  ⚠ No live position data"))
	} else {
		hintParts = append(hintParts,
			lipgloss.NewStyle().Foreground(styles.ColorGreen).Render("  ● LIVE"))
	}
	if t.circuit != nil {
		hintParts = append(hintParts,
			styles.DimStyle.Render(fmt.Sprintf("  │  Circuit: %s (%d corners)", t.circuit.CircuitName, len(t.circuit.Corners))))
	}
	hint := styles.FooterStyle.Render(strings.Join(hintParts, ""))

	leftW := t.width/2 - 2
	rightW := t.width - leftW - 3
	mapH := t.height - 7
	if mapH < 12 {
		mapH = 12
	}

	// Map panel: use real coordinate rendering when available, ASCII art otherwise
	var mapInner string
	if t.circuit != nil {
		cTitle := lipgloss.NewStyle().Bold(true).Foreground(styles.ColorSubtle).
			Render(fmt.Sprintf(" %s  ·  %d corners", t.circuit.CircuitName, len(t.circuit.Corners)))
		circuitStr := renderDynamicCircuit(t.circuit, leftW-6, mapH-4, nil)
		mapInner = cTitle + "\n" + circuitStr
	} else {
		art := common.GetCircuitArt(circuitName)
		cTitle := lipgloss.NewStyle().Bold(true).Foreground(styles.ColorSubtle).Render(" Circuit Layout")
		var artLines []string
		for _, line := range art {
			artLines = append(artLines, lipgloss.NewStyle().Foreground(lipgloss.Color("#4FC3F7")).Render(line))
		}
		mapInner = cTitle + "\n\n" + strings.Join(artLines, "\n")
	}
	mapPanel := styles.Card.Width(leftW).Render(mapInner)

	legend := t.renderLegend(driverRows)
	legendPanel := lipgloss.NewStyle().Width(rightW).Render(legend)
	body := lipgloss.JoinHorizontal(lipgloss.Top, mapPanel, "  ", legendPanel)

	return lipgloss.JoinVertical(lipgloss.Left,
		title, sep, "",
		body, "",
		hint,
	)
}

// ── rendering helpers ─────────────────────────────────────────────────────────

type driverDot struct {
	acronym  string
	pos      int
	lapTime  float64
	bestLap  float64
	teamCol  string
	compound string
	tyreAge  int
	pitCount int
}

func (t *TrackMap) buildLiveDots() []driverDot {
	driverMap := make(map[int]openf1.Driver)
	for _, d := range t.drivers {
		driverMap[d.DriverNumber] = d
	}
	posMap := make(map[int]int)
	for _, p := range t.positions {
		posMap[p.DriverNumber] = p.Position
	}
	bestLap := make(map[int]float64)
	lastLap := make(map[int]float64)
	for _, l := range t.laps {
		if l.LapDuration > 0 {
			if cur, ok := bestLap[l.DriverNumber]; !ok || l.LapDuration < cur {
				bestLap[l.DriverNumber] = l.LapDuration
			}
		}
		lastLap[l.DriverNumber] = l.LapDuration
	}
	var rows []driverDot
	for num, drv := range driverMap {
		rows = append(rows, driverDot{
			acronym: drv.NameAcronym,
			pos:     posMap[num],
			lapTime: lastLap[num],
			bestLap: bestLap[num],
			teamCol: drv.TeamColour,
		})
	}
	// sort by position
	for i := 0; i < len(rows); i++ {
		for j := i + 1; j < len(rows); j++ {
			if rows[j].pos < rows[i].pos || rows[i].pos == 0 {
				rows[i], rows[j] = rows[j], rows[i]
			}
		}
	}
	return rows
}

func (t *TrackMap) renderLegend(rows []driverDot) string {
	title := lipgloss.NewStyle().Bold(true).Foreground(styles.ColorSubtle).Render("Current Order")
	sep := styles.Divider.Render(strings.Repeat("─", 36))

	lines := []string{title, sep}
	// Column headers
	hdr := lipgloss.NewStyle().Foreground(styles.ColorMuted).Render(
		fmt.Sprintf("  %-4s %-5s %-10s %-10s", "POS", "DRV", "LAST LAP", "BEST LAP"))
	lines = append(lines, hdr)

	for _, row := range rows {
		if row.pos == 0 {
			continue
		}
		col := lipgloss.Color("#" + row.teamCol)
		if row.teamCol == "" {
			col = styles.ColorSubtle
		}
		badge := lipgloss.NewStyle().
			Background(col).
			Foreground(lipgloss.Color("#000000")).
			Bold(true).
			Render(" " + row.acronym + " ")

		posStr := lipgloss.NewStyle().Width(3).Foreground(common.PosColor(row.pos)).Bold(row.pos <= 3).
			Render(fmt.Sprintf("%2d", row.pos))

		lastStr := common.FormatDuration(row.lapTime)
		bestStr := common.FormatDuration(row.bestLap)

		// Colour last lap purple if it equals best lap
		lastCol := styles.ColorText
		if row.bestLap > 0 && math.Abs(row.lapTime-row.bestLap) < 0.001 {
			lastCol = styles.ColorPurple
		}

		lastR := lipgloss.NewStyle().Width(10).Foreground(lastCol).Render(lastStr)
		bestR := lipgloss.NewStyle().Width(10).Foreground(styles.ColorSubtle).Render(bestStr)

		line := fmt.Sprintf("  %s %s  %s%s", posStr, badge, lastR, bestR)
		lines = append(lines, line)
	}

	return styles.Card.Render(strings.Join(lines, "\n"))
}

// ── heatmap ───────────────────────────────────────────────────────────────────

// heatColor maps a value in [minV, maxV] to a colour gradient:
// best (low for times, high for speeds) → green, worst → red.
func heatColor(val, minV, maxV float64, lowerIsBetter bool) lipgloss.Color {
	if maxV == minV {
		return lipgloss.Color("#4ADE80")
	}
	var t float64
	if lowerIsBetter {
		t = (val - minV) / (maxV - minV) // 0=best(green) 1=worst(red)
	} else {
		t = (maxV - val) / (maxV - minV) // 0=best(green) 1=worst(red)
	}
	if t < 0 {
		t = 0
	}
	if t > 1 {
		t = 1
	}
	// Interpolate R,G,B: green(74,222,128) → yellow(250,204,21) → red(239,68,68)
	var r, g, b float64
	if t < 0.5 {
		tt := t * 2
		r = 74 + (250-74)*tt
		g = 222 + (204-222)*tt
		b = 128 + (21-128)*tt
	} else {
		tt := (t - 0.5) * 2
		r = 250 + (239-250)*tt
		g = 204 + (68-204)*tt
		b = 21 + (68-21)*tt
	}
	return lipgloss.Color(fmt.Sprintf("#%02X%02X%02X", int(r), int(g), int(b)))
}

// renderHeatmap builds a sector/speed heat table for all drivers using lap data.
func (t *TrackMap) renderHeatmap() string {
	type driverMetrics struct {
		acronym string
		teamCol string
		bestS1  float64
		bestS2  float64
		bestS3  float64
		avgS1   float64
		avgS2   float64
		avgS3   float64
		bestI1  int
		bestI2  int
		bestST  int
		bestLap float64
	}

	// Aggregate per driver.
	type acc struct {
		sumS1, sumS2, sumS3   float64
		countS1, countS2, countS3 int
		bestS1, bestS2, bestS3    float64
		bestI1, bestI2, bestST    int
		bestLap                   float64
	}
	driverMap := make(map[int]openf1.Driver)
	for _, d := range t.drivers {
		driverMap[d.DriverNumber] = d
	}

	accs := make(map[int]*acc)
	for _, d := range t.drivers {
		accs[d.DriverNumber] = &acc{
			bestS1: math.MaxFloat64, bestS2: math.MaxFloat64, bestS3: math.MaxFloat64,
			bestLap: math.MaxFloat64,
		}
	}
	for _, lap := range t.laps {
		a, ok := accs[lap.DriverNumber]
		if !ok {
			continue
		}
		if lap.LapDuration > 0 && lap.LapDuration < a.bestLap {
			a.bestLap = lap.LapDuration
		}
		if lap.DurationSector1 > 0 {
			if lap.DurationSector1 < a.bestS1 {
				a.bestS1 = lap.DurationSector1
			}
			a.sumS1 += lap.DurationSector1
			a.countS1++
		}
		if lap.DurationSector2 > 0 {
			if lap.DurationSector2 < a.bestS2 {
				a.bestS2 = lap.DurationSector2
			}
			a.sumS2 += lap.DurationSector2
			a.countS2++
		}
		if lap.DurationSector3 > 0 {
			if lap.DurationSector3 < a.bestS3 {
				a.bestS3 = lap.DurationSector3
			}
			a.sumS3 += lap.DurationSector3
			a.countS3++
		}
		if lap.I1Speed > a.bestI1 {
			a.bestI1 = lap.I1Speed
		}
		if lap.I2Speed > a.bestI2 {
			a.bestI2 = lap.I2Speed
		}
		if lap.StSpeed > a.bestST {
			a.bestST = lap.StSpeed
		}
	}

	// Build slice, skip drivers with no data.
	var metrics []driverMetrics
	for num, drv := range driverMap {
		a := accs[num]
		if a == nil || a.bestLap == math.MaxFloat64 {
			continue
		}
		dm := driverMetrics{
			acronym: drv.NameAcronym,
			teamCol: drv.TeamColour,
			bestLap: a.bestLap,
			bestI1:  a.bestI1,
			bestI2:  a.bestI2,
			bestST:  a.bestST,
		}
		if a.bestS1 < math.MaxFloat64 {
			dm.bestS1 = a.bestS1
		}
		if a.bestS2 < math.MaxFloat64 {
			dm.bestS2 = a.bestS2
		}
		if a.bestS3 < math.MaxFloat64 {
			dm.bestS3 = a.bestS3
		}
		if a.countS1 > 0 {
			dm.avgS1 = a.sumS1 / float64(a.countS1)
		}
		if a.countS2 > 0 {
			dm.avgS2 = a.sumS2 / float64(a.countS2)
		}
		if a.countS3 > 0 {
			dm.avgS3 = a.sumS3 / float64(a.countS3)
		}
		metrics = append(metrics, dm)
	}
	if len(metrics) == 0 {
		return styles.Card.Render(
			lipgloss.NewStyle().Foreground(styles.ColorMuted).Render(
				"  No lap data available yet. Check back once laps are being timed."))
	}

	// Sort by best lap.
	for i := 0; i < len(metrics); i++ {
		for j := i + 1; j < len(metrics); j++ {
			if metrics[j].bestLap < metrics[i].bestLap {
				metrics[i], metrics[j] = metrics[j], metrics[i]
			}
		}
	}

	// Compute column min/max for colour scaling.
	colMin := func(vals []float64) float64 {
		m := math.MaxFloat64
		for _, v := range vals {
			if v > 0 && v < m {
				m = v
			}
		}
		if m == math.MaxFloat64 {
			return 0
		}
		return m
	}
	colMax := func(vals []float64) float64 {
		m := -math.MaxFloat64
		for _, v := range vals {
			if v > m {
				m = v
			}
		}
		if m == -math.MaxFloat64 {
			return 0
		}
		return m
	}

	s1vals, s2vals, s3vals := make([]float64, len(metrics)), make([]float64, len(metrics)), make([]float64, len(metrics))
	i1vals, i2vals, stvals := make([]float64, len(metrics)), make([]float64, len(metrics)), make([]float64, len(metrics))
	lapvals := make([]float64, len(metrics))
	for i, m := range metrics {
		s1vals[i], s2vals[i], s3vals[i] = m.bestS1, m.bestS2, m.bestS3
		i1vals[i], i2vals[i], stvals[i] = float64(m.bestI1), float64(m.bestI2), float64(m.bestST)
		lapvals[i] = m.bestLap
	}
	minS1, maxS1 := colMin(s1vals), colMax(s1vals)
	minS2, maxS2 := colMin(s2vals), colMax(s2vals)
	minS3, maxS3 := colMin(s3vals), colMax(s3vals)
	minI1, maxI1 := colMin(i1vals), colMax(i1vals)
	minI2, maxI2 := colMin(i2vals), colMax(i2vals)
	minST, maxST := colMin(stvals), colMax(stvals)
	minLap, maxLap := colMin(lapvals), colMax(lapvals)

	fmtS := func(v float64) string {
		if v <= 0 {
			return "  —  "
		}
		return fmt.Sprintf("%5.3f", v)
	}
	fmtSpd := func(v int) string {
		if v == 0 {
			return " — "
		}
		return fmt.Sprintf("%3d", v)
	}

	cell := func(text string, col lipgloss.Color, w int) string {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("#000000")).
			Background(col).
			Width(w).
			Align(lipgloss.Center).
			Bold(true).
			Render(text)
	}

	muted := styles.ColorMuted
	hdrStyle := lipgloss.NewStyle().Foreground(muted).Bold(true)

	hdr := hdrStyle.Render("  DRV  │") +
		hdrStyle.Width(8).Align(lipgloss.Center).Render("BEST LAP") + hdrStyle.Render("│") +
		hdrStyle.Width(7).Align(lipgloss.Center).Render(" S1 ") + hdrStyle.Render("│") +
		hdrStyle.Width(7).Align(lipgloss.Center).Render(" S2 ") + hdrStyle.Render("│") +
		hdrStyle.Width(7).Align(lipgloss.Center).Render(" S3 ") + hdrStyle.Render("│") +
		hdrStyle.Width(5).Align(lipgloss.Center).Render("I1") + hdrStyle.Render("│") +
		hdrStyle.Width(5).Align(lipgloss.Center).Render("I2") + hdrStyle.Render("│") +
		hdrStyle.Width(5).Align(lipgloss.Center).Render(" ST") + hdrStyle.Render("│")

	sep := styles.DimStyle.Render(strings.Repeat("─", t.width-6))
	lines := []string{
		lipgloss.NewStyle().Bold(true).Foreground(styles.ColorSubtle).Render("  Speed & Sector Heat Map"),
		"",
		hdr,
		sep,
	}

	legend := "  " +
		lipgloss.NewStyle().Background(lipgloss.Color("#4ADE80")).Foreground(lipgloss.Color("#000")).Render(" BEST ") +
		"  " +
		lipgloss.NewStyle().Background(lipgloss.Color("#FACC15")).Foreground(lipgloss.Color("#000")).Render("  MID ") +
		"  " +
		lipgloss.NewStyle().Background(lipgloss.Color("#EF4444")).Foreground(lipgloss.Color("#000")).Render(" WORST")

	for i, m := range metrics {
		col := lipgloss.Color("#" + m.teamCol)
		if m.teamCol == "" {
			col = styles.ColorSubtle
		}
		badge := lipgloss.NewStyle().Background(col).Foreground(lipgloss.Color("#000")).
			Bold(true).Render(" " + m.acronym + " ")

		posStr := lipgloss.NewStyle().Foreground(common.PosColor(i + 1)).Render(fmt.Sprintf("%2d", i+1))

		lapCell := cell(common.FormatDuration(m.bestLap), heatColor(m.bestLap, minLap, maxLap, true), 8)

		var s1Cell, s2Cell, s3Cell, i1Cell, i2Cell, stCell string
		if m.bestS1 > 0 {
			s1Cell = cell(fmtS(m.bestS1), heatColor(m.bestS1, minS1, maxS1, true), 7)
		} else {
			s1Cell = lipgloss.NewStyle().Width(7).Align(lipgloss.Center).Foreground(muted).Render("—")
		}
		if m.bestS2 > 0 {
			s2Cell = cell(fmtS(m.bestS2), heatColor(m.bestS2, minS2, maxS2, true), 7)
		} else {
			s2Cell = lipgloss.NewStyle().Width(7).Align(lipgloss.Center).Foreground(muted).Render("—")
		}
		if m.bestS3 > 0 {
			s3Cell = cell(fmtS(m.bestS3), heatColor(m.bestS3, minS3, maxS3, true), 7)
		} else {
			s3Cell = lipgloss.NewStyle().Width(7).Align(lipgloss.Center).Foreground(muted).Render("—")
		}
		if m.bestI1 > 0 {
			i1Cell = cell(fmtSpd(m.bestI1), heatColor(float64(m.bestI1), minI1, maxI1, false), 5)
		} else {
			i1Cell = lipgloss.NewStyle().Width(5).Align(lipgloss.Center).Foreground(muted).Render("—")
		}
		if m.bestI2 > 0 {
			i2Cell = cell(fmtSpd(m.bestI2), heatColor(float64(m.bestI2), minI2, maxI2, false), 5)
		} else {
			i2Cell = lipgloss.NewStyle().Width(5).Align(lipgloss.Center).Foreground(muted).Render("—")
		}
		if m.bestST > 0 {
			stCell = cell(fmtSpd(m.bestST), heatColor(float64(m.bestST), minST, maxST, false), 5)
		} else {
			stCell = lipgloss.NewStyle().Width(5).Align(lipgloss.Center).Foreground(muted).Render("—")
		}

		sep2 := styles.DimStyle.Render("│")
		line := fmt.Sprintf("  %s %s  %s%s%s%s%s%s%s%s%s%s%s%s%s",
			posStr, badge, lapCell,
			sep2, s1Cell,
			sep2, s2Cell,
			sep2, s3Cell,
			sep2, i1Cell,
			sep2, i2Cell,
			sep2, stCell)
		lines = append(lines, line)
	}

	lines = append(lines, sep, legend)
	return strings.Join(lines, "\n")
}

// triggerSpeedFetch picks the default (best-lap) driver and issues a fetch command.
func (t *TrackMap) triggerSpeedFetch() tea.Cmd {
	if len(t.drivers) == 0 {
		return nil
	}
	// Sort drivers alphabetically and pick first if no driver selected yet.
	sorted := make([]openf1.Driver, len(t.drivers))
	copy(sorted, t.drivers)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].NameAcronym < sorted[j].NameAcronym })
	if t.heatDriverNum == 0 {
		t.heatDriverNum = sorted[0].DriverNumber
		t.heatDriverAcronym = sorted[0].NameAcronym
	}
	t.heatLoading = true
	t.segmentSpeeds = nil
	lap := t.bestLapForDriver(t.heatDriverNum)
	return fetchSpeedDataCmd(t.client, t.session.SessionKey, t.heatDriverNum, lap)
}

// cycleHeatDriver advances (delta=+1) or retreats (delta=-1) the selected driver.
func (t *TrackMap) cycleHeatDriver(delta int) tea.Cmd {
	if len(t.drivers) == 0 {
		return nil
	}
	sorted := make([]openf1.Driver, len(t.drivers))
	copy(sorted, t.drivers)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].NameAcronym < sorted[j].NameAcronym })

	cur := 0
	for i, d := range sorted {
		if d.DriverNumber == t.heatDriverNum {
			cur = i
			break
		}
	}
	next := (cur + delta + len(sorted)) % len(sorted)
	t.heatDriverNum = sorted[next].DriverNumber
	t.heatDriverAcronym = sorted[next].NameAcronym
	t.heatLoading = true
	t.segmentSpeeds = nil
	lap := t.bestLapForDriver(t.heatDriverNum)
	return fetchSpeedDataCmd(t.client, t.session.SessionKey, t.heatDriverNum, lap)
}

// bestLapForDriver returns the full Lap record for the driver's fastest lap.
// Returns a zero Lap if no timed lap exists.
func (t *TrackMap) bestLapForDriver(driverNum int) openf1.Lap {
	var best openf1.Lap
	bestTime := math.MaxFloat64
	for _, lap := range t.laps {
		if lap.DriverNumber == driverNum && lap.LapDuration > 0 && lap.LapDuration < bestTime {
			bestTime = lap.LapDuration
			best = lap
		}
	}
	return best
}

// lapDateRange computes the (dateGt, dateLt) strings for OpenF1 date filtering
// based on a lap's DateStart and LapDuration.
func lapDateRange(lap openf1.Lap) (string, string) {
	if lap.DateStart == "" || lap.LapDuration <= 0 {
		return "", ""
	}
	formats := []string{
		"2006-01-02T15:04:05.000000+00:00",
		"2006-01-02T15:04:05.000000Z",
		"2006-01-02T15:04:05.000000",
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02T15:04:05.000",
		"2006-01-02T15:04:05",
	}
	var start time.Time
	for _, f := range formats {
		if pt, err := time.Parse(f, lap.DateStart); err == nil {
			start = pt
			break
		}
	}
	if start.IsZero() {
		return "", ""
	}
	end := start.Add(time.Duration(float64(time.Second) * lap.LapDuration))
	// Add a 1-second buffer on the end to ensure the final sample is captured.
	end = end.Add(time.Second)
	const layout = "2006-01-02T15:04:05.000"
	return start.UTC().Format(layout), end.UTC().Format(layout)
}

// ── command ───────────────────────────────────────────────────────────────────

// ── Dynamic circuit renderer ───────────────────────────────────────────────────

// renderDynamicCircuit converts real GPS-sampled track coordinates (from the
// Multiviewer API) into a terminal string using directional box-drawing chars.
// It applies the circuit's own rotation angle and corrects for the ~2:1
// terminal cell aspect ratio so the circuit looks proportionally correct.
func renderDynamicCircuit(c *multiviewer.Circuit, cols, rows int, segmentColors []lipgloss.Color) string {
	if len(c.X) == 0 || len(c.X) != len(c.Y) {
		return styles.DimStyle.Render("  No coordinate data")
	}
	n := len(c.X)

	// 1. Apply rotation + terminal aspect correction
	rad := float64(c.Rotation) * math.Pi / 180.0
	cosR, sinR := math.Cos(rad), math.Sin(rad)

	type pt struct{ x, y float64 }
	rotated := make([]pt, n)
	for i := 0; i < n; i++ {
		x := float64(c.X[i])
		y := float64(c.Y[i])
		rotated[i] = pt{
			x: x*cosR - y*sinR,
			y: (x*sinR + y*cosR) * 0.45,
		}
	}

	// 2. Find bounding box
	minX, maxX := rotated[0].x, rotated[0].x
	minY, maxY := rotated[0].y, rotated[0].y
	for _, p := range rotated[1:] {
		if p.x < minX { minX = p.x }
		if p.x > maxX { maxX = p.x }
		if p.y < minY { minY = p.y }
		if p.y > maxY { maxY = p.y }
	}
	rangeX, rangeY := maxX-minX, maxY-minY
	if rangeX == 0 { rangeX = 1 }
	if rangeY == 0 { rangeY = 1 }

	// 3. Uniform scale to fill available space, centred
	scaleX := float64(cols-3) / rangeX
	scaleY := float64(rows-2) / rangeY
	scale := math.Min(scaleX, scaleY)
	offX := (float64(cols) - rangeX*scale) / 2.0
	offY := (float64(rows) - rangeY*scale) / 2.0

	toCol := func(x float64) int { return int((x-minX)*scale + offX) }
	toRow := func(y float64) int { return int((y-minY)*scale + offY) }

	// 4. Grid of directional chars
	type cell struct {
		ch  rune
		set bool
		col lipgloss.Color
	}
	grid := make([][]cell, rows)
	for i := range grid {
		grid[i] = make([]cell, cols)
	}

	dirChar := func(dx, dy int) rune {
		adx, ady := dx, dy
		if adx < 0 { adx = -adx }
		if ady < 0 { ady = -ady }
		if adx == 0 && ady == 0 { return '·' }
		if ady*10 < adx*3  { return '─' } // < ~17° from horizontal
		if adx*10 < ady*3  { return '│' } // < ~17° from vertical
		if (dx > 0) == (dy > 0) { return '╲' }
		return '╱'
	}

	for i := 0; i < n; i++ {
		// Per-segment colour: fall back to the classic light blue when no heat data.
		var segCol lipgloss.Color
		if segmentColors != nil && i < len(segmentColors) {
			segCol = segmentColors[i]
		} else {
			segCol = lipgloss.Color("#4FC3F7")
		}
		j := (i + 1) % n
		x1, y1 := toCol(rotated[i].x), toRow(rotated[i].y)
		x2, y2 := toCol(rotated[j].x), toRow(rotated[j].y)
		dx, dy := x2-x1, y2-y1
		adx, ady := dx, dy
		if adx < 0 { adx = -adx }
		if ady < 0 { ady = -ady }
		ch := dirChar(dx, dy)
		steps := adx
		if ady > adx { steps = ady }
		if steps == 0 {
			if x1 >= 0 && x1 < cols && y1 >= 0 && y1 < rows {
				grid[y1][x1] = cell{ch: ch, set: true, col: segCol}
			}
			continue
		}
		for s := 0; s <= steps; s++ {
			px := x1 + dx*s/steps
			py := y1 + dy*s/steps
			if px >= 0 && px < cols && py >= 0 && py < rows {
				grid[py][px] = cell{ch: ch, set: true, col: segCol}
			}
		}
	}

	// 5. Corner number overlay: find nearest empty cell to each corner
	cornerAt := make(map[int]map[int]int)
	for _, corner := range c.Corners {
		rx := corner.TrackPosition.X*cosR - corner.TrackPosition.Y*sinR
		ry := (corner.TrackPosition.X*sinR + corner.TrackPosition.Y*cosR) * 0.45
		baseCol := toCol(rx)
		baseRow := toRow(ry)
		placed := false
	outer:
		for delta := 0; delta <= 2; delta++ {
			for dr := -delta; dr <= delta; dr++ {
				for dc := -delta; dc <= delta; dc++ {
					adr, adc := dr, dc
					if adr < 0 { adr = -adr }
					if adc < 0 { adc = -adc }
					if adr+adc != delta { continue }
					nr, nc := baseRow+dr, baseCol+dc
					if nr < 0 || nr >= rows || nc < 0 || nc >= cols { continue }
					if grid[nr][nc].set { continue }
					if _, ok := cornerAt[nr]; !ok { cornerAt[nr] = make(map[int]int) }
					if _, taken := cornerAt[nr][nc]; !taken {
						cornerAt[nr][nc] = corner.Number
						placed = true
						break outer
					}
				}
			}
		}
		if !placed && baseRow >= 0 && baseRow < rows && baseCol >= 0 && baseCol < cols {
			if _, ok := cornerAt[baseRow]; !ok { cornerAt[baseRow] = make(map[int]int) }
			cornerAt[baseRow][baseCol] = corner.Number
		}
	}

	// 6. Render
	// When no segment colors provided, skip corner numbers (they clutter heatmap)
	showCorners := segmentColors == nil
	cornerStyle := lipgloss.NewStyle().Foreground(styles.ColorYellow).Bold(true)

	var sb strings.Builder
	for r := 0; r < rows; r++ {
		for col := 0; col < cols; col++ {
			if showCorners {
				if rowMap, ok := cornerAt[r]; ok {
					if num, ok2 := rowMap[col]; ok2 {
						if num <= 9 {
							sb.WriteString(cornerStyle.Render(fmt.Sprintf("%d", num)))
						} else {
							sb.WriteString(cornerStyle.Render("+"))
						}
						continue
					}
				}
			}
			if grid[r][col].set {
				celStyle := lipgloss.NewStyle().Foreground(grid[r][col].col)
				sb.WriteString(celStyle.Render(string(grid[r][col].ch)))
			} else {
				sb.WriteRune(' ')
			}
		}
		sb.WriteRune('\n')
	}
	return sb.String()
}

// ── commands ───────────────────────────────────────────────────────────────────

func fetchTrackMapCmd(client *openf1.Client) tea.Cmd {
	return func() tea.Msg {
		ctx := common.ContextBG()
		session, err := client.GetLatestSession(ctx)
		if err != nil {
			return trackMapErrMsg{err}
		}
		drivers, _ := client.GetDrivers(ctx, session.SessionKey)
		positions, _ := client.GetPositions(ctx, session.SessionKey)
		laps, _ := client.GetLaps(ctx, session.SessionKey)
		return trackMapDataMsg{session, drivers, positions, laps}
	}
}

// fetchCircuitByKeyCmd fetches a circuit directly by its integer key (which
// OpenF1 exposes as session.CircuitKey and matches the Multiviewer API).
func fetchCircuitByKeyCmd(mv *multiviewer.Client, circuitKey, year int) tea.Cmd {
	return func() tea.Msg {
		if circuitKey <= 0 {
			return trackMapCircuitErrMsg{fmt.Errorf("invalid circuit key %d", circuitKey)}
		}
		ctx := common.ContextBG()
		c, err := mv.GetCircuit(ctx, circuitKey, year)
		if err != nil {
			// Try previous year — layouts rarely change
			c, err = mv.GetCircuit(ctx, circuitKey, year-1)
			if err != nil {
				return trackMapCircuitErrMsg{err}
			}
		}
		return trackMapCircuitMsg{c}
	}
}

// fetchSpeedDataCmd fetches car telemetry + location for one driver's best lap
// using date-range filtering (OpenF1 /location and /car_data do not support
// lap_number; we derive the date window from Lap.DateStart + LapDuration).
func fetchSpeedDataCmd(client *openf1.Client, sessionKey, driverNumber int, lap openf1.Lap) tea.Cmd {
	return func() tea.Msg {
		ctx := common.ContextBG()
		dateGt, dateLt := lapDateRange(lap)
		locs, locErr := client.GetCarLocations(ctx, sessionKey, driverNumber, dateGt, dateLt)
		if locErr != nil {
			return trackMapSpeedErrMsg{locErr}
		}
		cd, cdErr := client.GetCarData(ctx, sessionKey, driverNumber, dateGt, dateLt)
		if cdErr != nil {
			return trackMapSpeedErrMsg{cdErr}
		}
		return trackMapSpeedMsg{driverNum: driverNumber, carData: cd, locations: locs}
	}
}

// ── speed heatmap helpers ─────────────────────────────────────────────────────

// plasmaColor maps t ∈ [0,1] to the matplotlib Plasma colormap.
// t=0 → slowest (dark purple), t=1 → fastest (bright yellow).
func plasmaColor(t float64) lipgloss.Color {
	if t < 0 {
		t = 0
	}
	if t > 1 {
		t = 1
	}
	// Five-stop plasma LUT (matches matplotlib's 'plasma' colormap)
	type stop struct{ r, g, b float64 }
	stops := [5]stop{
		{13, 8, 135},    // 0.00 – deep indigo
		{126, 3, 168},   // 0.25 – purple
		{204, 71, 120},  // 0.50 – pink-red
		{248, 148, 65},  // 0.75 – orange
		{240, 249, 33},  // 1.00 – yellow
	}
	scaled := t * 4
	idx := int(scaled)
	if idx >= 4 {
		idx = 3
	}
	frac := scaled - float64(idx)
	a, b := stops[idx], stops[idx+1]
	r := a.r + (b.r-a.r)*frac
	g := a.g + (b.g-a.g)*frac
	bl := a.b + (b.b-a.b)*frac
	return lipgloss.Color(fmt.Sprintf("#%02X%02X%02X", int(r), int(g), int(bl)))
}

// parseDateMs converts an OpenF1 ISO-8601 date string to Unix milliseconds.
func parseDateMs(s string) int64 {
	for _, f := range []string{
		"2006-01-02T15:04:05.000000",
		"2006-01-02T15:04:05.000",
		"2006-01-02T15:04:05",
		time.RFC3339Nano,
		time.RFC3339,
	} {
		if t, err := time.Parse(f, s); err == nil {
			return t.UnixMilli()
		}
	}
	return 0
}

// buildSegmentSpeeds maps car telemetry onto circuit segments, returning one
// average speed (km/h) per circuit coordinate index. Zero means no data.
func buildSegmentSpeeds(c *multiviewer.Circuit, locs []openf1.CarLocation, carData []openf1.CarData) []float64 {
	n := len(c.X)
	if n == 0 || len(locs) == 0 || len(carData) == 0 {
		return nil
	}

	// Build a sorted (ms, speed) slice from car_data for binary search.
	type tspeed struct {
		ms    int64
		speed int
	}
	speeds := make([]tspeed, 0, len(carData))
	for _, cd := range carData {
		speeds = append(speeds, tspeed{parseDateMs(cd.Date), cd.Speed})
	}
	sort.Slice(speeds, func(i, j int) bool { return speeds[i].ms < speeds[j].ms })

	// For each location sample, find the nearest car_data speed.
	sumSpeeds := make([]float64, n)
	count := make([]int, n)

	for _, loc := range locs {
		ms := parseDateMs(loc.Date)
		// Binary search for nearest speed sample.
		idx := sort.Search(len(speeds), func(i int) bool { return speeds[i].ms >= ms })
		var spd int
		switch {
		case idx == 0:
			spd = speeds[0].speed
		case idx >= len(speeds):
			spd = speeds[len(speeds)-1].speed
		default:
			if ms-speeds[idx-1].ms <= speeds[idx].ms-ms {
				spd = speeds[idx-1].speed
			} else {
				spd = speeds[idx].speed
			}
		}

		// Find nearest circuit segment (raw X/Y, same coordinate system).
		minD2 := math.MaxFloat64
		best := 0
		lx, ly := float64(loc.X), float64(loc.Y)
		for i := 0; i < n; i++ {
			dx := lx - float64(c.X[i])
			dy := ly - float64(c.Y[i])
			if d2 := dx*dx + dy*dy; d2 < minD2 {
				minD2 = d2
				best = i
			}
		}
		sumSpeeds[best] += float64(spd)
		count[best]++
	}

	result := make([]float64, n)
	for i := 0; i < n; i++ {
		if count[i] > 0 {
			result[i] = sumSpeeds[i] / float64(count[i])
		}
	}

	// Interpolate gaps: up to 10 passes to fill zero segments.
	for pass := 0; pass < 10; pass++ {
		changed := false
		for i := 0; i < n; i++ {
			if result[i] == 0 {
				prev, next := (i-1+n)%n, (i+1)%n
				sum, cnt := result[prev]+result[next], 0
				if result[prev] > 0 {
					cnt++
				}
				if result[next] > 0 {
					cnt++
				}
				if cnt > 0 {
					result[i] = sum / float64(cnt)
					changed = true
				}
			}
		}
		if !changed {
			break
		}
	}
	return result
}

// renderSpeedHeatmap renders the circuit outline colored by speed (plasma gradient).
func (t *TrackMap) renderSpeedHeatmap() string {
	if t.circuit == nil {
		return styles.Card.Render(
			lipgloss.NewStyle().Foreground(styles.ColorMuted).Render(
				"  Waiting for circuit data…"))
	}

	driverInfo := ""
	if t.heatDriverAcronym != "" {
		col := lipgloss.Color("#FFFFFF")
		for _, d := range t.drivers {
			if d.DriverNumber == t.heatDriverNum && d.TeamColour != "" {
				col = lipgloss.Color("#" + d.TeamColour)
				break
			}
		}
		badge := lipgloss.NewStyle().Background(col).Foreground(lipgloss.Color("#000000")).
			Bold(true).Render(" " + t.heatDriverAcronym + " ")
		driverInfo = "  " + badge + "  " +
			styles.DimStyle.Render("n: next driver  p: prev driver")
	}

	if t.heatLoading {
		content := lipgloss.NewStyle().Foreground(styles.ColorMuted).
			Render(fmt.Sprintf("  %s Fetching telemetry for %s…", t.spin.View(), t.heatDriverAcronym))
		return lipgloss.JoinVertical(lipgloss.Left, driverInfo, "", content)
	}

	if len(t.segmentSpeeds) == 0 {
		noData := styles.Card.Render(
			lipgloss.NewStyle().Foreground(styles.ColorMuted).Render(
				"  No telemetry available. Try pressing n to select another driver."))
		return lipgloss.JoinVertical(lipgloss.Left, driverInfo, "", noData)
	}

	// Find speed range for plasma normalisation.
	minSpd, maxSpd := math.MaxFloat64, -math.MaxFloat64
	for _, s := range t.segmentSpeeds {
		if s > 0 {
			if s < minSpd {
				minSpd = s
			}
			if s > maxSpd {
				maxSpd = s
			}
		}
	}
	if minSpd == math.MaxFloat64 {
		minSpd = 0
	}
	if maxSpd <= minSpd {
		maxSpd = minSpd + 1
	}

	// Build per-segment plasma colours.
	segColors := make([]lipgloss.Color, len(t.segmentSpeeds))
	for i, s := range t.segmentSpeeds {
		tv := 0.0
		if s > 0 {
			tv = (s - minSpd) / (maxSpd - minSpd)
		}
		segColors[i] = plasmaColor(tv)
	}

	mapW := t.width - 4
	mapH := t.height - 10
	if mapH < 12 {
		mapH = 12
	}

	circuit := renderDynamicCircuit(t.circuit, mapW, mapH, segColors)

	// Plasma legend bar.
	barW := 50
	if t.width < 60 {
		barW = t.width - 10
	}
	if barW < 10 {
		barW = 10
	}
	var barSB strings.Builder
	barSB.WriteString("  ")
	for i := 0; i < barW; i++ {
		tv := float64(i) / float64(barW-1)
		barSB.WriteString(lipgloss.NewStyle().Foreground(plasmaColor(tv)).Render("█"))
	}
	barSB.WriteString(fmt.Sprintf("\n  %-8s", fmt.Sprintf("%d km/h", int(minSpd))))
	mid := int((minSpd + maxSpd) / 2)
	midPad := barW/2 - 4
	if midPad > 0 {
		barSB.WriteString(strings.Repeat(" ", midPad))
	}
	barSB.WriteString(styles.DimStyle.Render(fmt.Sprintf("%d", mid)))
	barSB.WriteString(fmt.Sprintf("%s%d km/h",
		strings.Repeat(" ", barW/2-4),
		int(maxSpd)))

	return lipgloss.JoinVertical(lipgloss.Left,
		driverInfo,
		circuit,
		barSB.String(),
	)
}
