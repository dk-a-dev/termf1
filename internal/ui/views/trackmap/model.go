package trackmap

import (
	"fmt"
	"math"
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


// simDrivers provides a fake Australian GP leaderboard for simulation/demo.
var simDrivers = []struct {
	acronym string
	team    string
	color   string
	pos     int
}{
	{"NOR", "McLaren", "FF8000", 1},
	{"PIA", "McLaren", "FF8000", 2},
	{"RUS", "Mercedes", "27F4D2", 3},
	{"HAM", "Mercedes", "27F4D2", 4},
	{"VER", "Red Bull Racing", "3671C6", 5},
	{"LEC", "Ferrari", "E8002D", 6},
	{"SAI", "Ferrari", "E8002D", 7},
	{"ANT", "Mercedes", "27F4D2", 8},
	{"GAS", "Alpine", "FF87BC", 9},
	{"ALO", "Aston Martin", "229971", 10},
}

// simLapTimes for Australian GP simulation (seconds)
var simLapData = map[string]float64{
	"NOR": 81.931, "PIA": 82.140, "RUS": 82.325, "HAM": 82.498,
	"VER": 82.612, "LEC": 82.780, "SAI": 82.891, "ANT": 83.104,
	"GAS": 83.256, "ALO": 83.410,
}

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
	simMode   bool
	simFrame  int
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
		t.simFrame++
		if t.simMode {
			// faster tick for simulation animation
			return t, tea.Tick(2*time.Second, func(tt time.Time) tea.Msg {
				return trackMapTickMsg(tt)
			})
		}
		// live mode: refresh data from API
		return t, fetchTrackMapCmd(t.client)

	case tea.KeyMsg:
		switch msg.String() {
		case "s":
			t.simMode = !t.simMode
			if t.simMode {
				// Melbourne circuitKey = 10 (Australian GP simulation)
				circFetch := fetchCircuitByKeyCmd(t.mv, 10, time.Now().Year())
				animTick := tea.Tick(2*time.Second, func(tt time.Time) tea.Msg {
					return trackMapTickMsg(tt)
				})
				return t, tea.Batch(animTick, circFetch)
			}
			return t, fetchTrackMapCmd(t.client)
		case "r":
			t.loading = true
			return t, tea.Batch(t.spin.Tick, fetchTrackMapCmd(t.client))
		}
	}
	return t, nil
}

func (t *TrackMap) View() string {
	if t.loading {
		return common.Centred(t.width, t.height, t.spin.View()+" Loading track data…")
	}

	circuitName := t.session.CircuitShortName
	if circuitName == "" {
		circuitName = "default"
	}

	// Determine the data source
	var driverRows []driverDot
	hasLiveData := len(t.positions) > 0
	if t.simMode {
		driverRows = buildSimDots(t.simFrame)
		circuitName = "Melbourne" // Australian GP
	} else {
		driverRows = t.buildLiveDots()
	}

	title := styles.Title.Render(fmt.Sprintf(" 🗺  Track Map  —  %s %s",
		t.session.CountryName, t.session.SessionName))
	if t.simMode {
		title = styles.Title.Render(" 🗺  Track Map  —  Australian GP 2025  ") +
			lipgloss.NewStyle().Foreground(styles.ColorYellow).Bold(true).Render("[SIMULATION MODE]")
	}
	sep := styles.Divider.Render(strings.Repeat("─", t.width))

	hintParts := []string{"  s: simulation (AUS GP demo)  │  r: refresh"}
	if !t.simMode && !hasLiveData {
		hintParts = append(hintParts,
			lipgloss.NewStyle().Foreground(styles.ColorYellow).Render("  ⚠ No live position data"))
	} else if !t.simMode && hasLiveData {
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
		circuitStr := renderDynamicCircuit(t.circuit, leftW-6, mapH-4)
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

func buildSimDots(frame int) []driverDot {
	var rows []driverDot
	for _, d := range simDrivers {
		lt := simLapData[d.acronym]
		// Slightly vary lap times per frame for animation feel
		lt += math.Sin(float64(frame)*0.3+float64(d.pos)*0.5) * 0.1
		rows = append(rows, driverDot{
			acronym: d.acronym,
			pos:     d.pos,
			lapTime: lt,
			bestLap: simLapData[d.acronym],
			teamCol: d.color,
		})
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

	if t.simMode {
		lines = append(lines, "")
		lines = append(lines, lipgloss.NewStyle().Foreground(styles.ColorYellow).
			Render("  ★ Australian GP Final Classification"))
		lines = append(lines, styles.DimStyle.Render("  Simulated — Race ended 16 Mar 2025"))
	}

	return styles.Card.Render(strings.Join(lines, "\n"))
}

// ── command ───────────────────────────────────────────────────────────────────

// ── Dynamic circuit renderer ───────────────────────────────────────────────────

// renderDynamicCircuit converts real GPS-sampled track coordinates (from the
// Multiviewer API) into a terminal string using directional box-drawing chars.
// It applies the circuit's own rotation angle and corrects for the ~2:1
// terminal cell aspect ratio so the circuit looks proportionally correct.
func renderDynamicCircuit(c *multiviewer.Circuit, cols, rows int) string {
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
				grid[y1][x1] = cell{ch, true}
			}
			continue
		}
		for s := 0; s <= steps; s++ {
			px := x1 + dx*s/steps
			py := y1 + dy*s/steps
			if px >= 0 && px < cols && py >= 0 && py < rows {
				grid[py][px] = cell{ch, true}
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
	trackStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#4FC3F7"))
	cornerStyle := lipgloss.NewStyle().Foreground(styles.ColorYellow).Bold(true)

	var sb strings.Builder
	for r := 0; r < rows; r++ {
		for col := 0; col < cols; col++ {
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
			if grid[r][col].set {
				sb.WriteString(trackStyle.Render(string(grid[r][col].ch)))
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
