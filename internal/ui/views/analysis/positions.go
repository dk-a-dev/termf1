package analysis

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/dk-a-dev/termf1/internal/api/openf1"
	"github.com/dk-a-dev/termf1/internal/ui/styles"
	"github.com/dk-a-dev/termf1/internal/ui/views/common"
)

// posLineStyles cycles through distinct line patterns per driver.
// Fields: dot marker, horizontal connector, diagonal connector.
var posLineStyles = []struct {
	dot  rune
	hori rune
	diag rune
}{
	{'●', '─', '╱'},
	{'◆', '╌', '╱'},
	{'▲', '┄', '╱'},
	{'■', '━', '╲'},
	{'◉', '╍', '╱'},
	{'○', '·', '/'},
	{'▷', '─', '\\'},
	{'◻', '╌', '/'},
	{'△', '┄', '\\'},
	{'☆', '·', '/'},
}

// ── Chart 7: Position Changes ─────────────────────────────────────────────────

// renderPositionChart draws a 2-D grid of position vs lap for all drivers.
// Each driver is plotted with their team colour and a unique line style.
// A right sidebar shows each driver's final position anchored at their row.
func (a *Analysis) renderPositionChart(w, h int) string {
	drMap := a.driverMap()

	// Build per-driver (lap → position) map using real /position data.
	// For each lap we find the driver's position at the moment that lap ended
	// (lap.DateStart + lap.LapDuration), matching FastF1's drv_laps['Position'].
	type lapPos struct {
		lap int
		pos int
	}
	byDriver := make(map[int][]lapPos)

	// Parse and group position entries by driver, sorted by time.
	type posEntry struct {
		t   time.Time
		pos int
	}
	posByDriver := make(map[int][]posEntry)
	for _, p := range a.positions {
		t, err := time.Parse(time.RFC3339Nano, p.Date)
		if err != nil {
			continue
		}
		posByDriver[p.DriverNumber] = append(posByDriver[p.DriverNumber], posEntry{t, p.Position})
	}
	for drv := range posByDriver {
		sort.Slice(posByDriver[drv], func(i, j int) bool {
			return posByDriver[drv][i].t.Before(posByDriver[drv][j].t)
		})
	}

	// Group laps by driver.
	lapsByDriver := make(map[int][]openf1.Lap)
	for _, l := range a.laps {
		if l.LapDuration > 0 && l.DateStart != "" {
			lapsByDriver[l.DriverNumber] = append(lapsByDriver[l.DriverNumber], l)
		}
	}
	if len(lapsByDriver) == 0 {
		return common.Centred(w, h, styles.DimStyle.Render("No lap data for position chart"))
	}

	maxLap := 0
	for drvNum, drvLaps := range lapsByDriver {
		entries := posByDriver[drvNum]
		if len(entries) == 0 {
			continue
		}
		sort.Slice(drvLaps, func(i, j int) bool { return drvLaps[i].LapNumber < drvLaps[j].LapNumber })
		for _, lap := range drvLaps {
			lapStart, err := time.Parse(time.RFC3339Nano, lap.DateStart)
			if err != nil {
				continue
			}
			lapEnd := lapStart.Add(time.Duration(lap.LapDuration * float64(time.Second)))
			// Last position entry at or before lap end.
			pos := 0
			for _, e := range entries {
				if !e.t.After(lapEnd) {
					pos = e.pos
				} else {
					break
				}
			}
			if pos > 0 {
				byDriver[drvNum] = append(byDriver[drvNum], lapPos{lap.LapNumber, pos})
				if lap.LapNumber > maxLap {
					maxLap = lap.LapNumber
				}
			}
		}
	}

	if len(byDriver) == 0 {
		return common.Centred(w, h, styles.DimStyle.Render("No position data"))
	}

	// Sort drivers by their final position.
	type drvInfo struct {
		num      int
		acronym  string
		team     string
		col      lipgloss.Color
		laps     []lapPos
		finalP   int
		styleIdx int
	}
	var drvs []drvInfo
	for num, lps := range byDriver {
		drv := drMap[num]
		col := lipgloss.Color("#888888")
		if drv.TeamColour != "" {
			col = lipgloss.Color("#" + drv.TeamColour)
		}
		finalP := lps[len(lps)-1].pos
		drvs = append(drvs, drvInfo{
			num:     num,
			acronym: drv.NameAcronym,
			team:    drv.TeamName,
			col:     col,
			laps:    lps,
			finalP:  finalP,
		})
	}
	sort.Slice(drvs, func(i, j int) bool { return drvs[i].finalP < drvs[j].finalP })

	// Assign line style indices — drivers on the same team get different styles.
	teamStyleCount := make(map[string]int)
	for i := range drvs {
		cnt := teamStyleCount[drvs[i].team]
		drvs[i].styleIdx = (i + cnt) % len(posLineStyles)
		teamStyleCount[drvs[i].team]++
	}

	// ── Layout ────────────────────────────────────────────────────────────────
	const axisW    = 4  // "20 │"
	const sidebarW = 13 // right sidebar width
	plotW := w - axisW - sidebarW - 1
	plotH := h - 4 // top title + bottom axis label

	maxPos := len(drvs)
	if maxPos < 20 {
		maxPos = 20
	}
	if plotW < 10 {
		plotW = 10
	}
	if plotH < 5 {
		plotH = 5
	}

	// Resolve (lap, pos) → pixel col/row.
	toCol := func(lap int) int {
		if maxLap <= 1 {
			return 0
		}
		return int(float64(lap-1) / float64(maxLap-1) * float64(plotW-1))
	}
	toRow := func(pos int) int {
		if maxPos <= 1 {
			return 0
		}
		return int(float64(pos-1) / float64(maxPos-1) * float64(plotH-1))
	}

	// Build RGB grid for each cell.
	type pixel struct {
		ch  rune
		col lipgloss.Color
		set bool
	}
	grid := make([][]pixel, plotH)
	for i := range grid {
		grid[i] = make([]pixel, plotW)
	}

	absInt := func(x int) int {
		if x < 0 {
			return -x
		}
		return x
	}

	// Build sidebar map: grid row → driver entry anchored at final position row.
	type sideEntry struct {
		acronym  string
		col      lipgloss.Color
		styleIdx int
		finalP   int
	}
	sideRow := make(map[int]*sideEntry)
	for i := range drvs {
		row := toRow(drvs[i].finalP)
		for sideRow[row] != nil && row < plotH-1 {
			row++
		}
		d := &drvs[i]
		sideRow[row] = &sideEntry{d.acronym, d.col, d.styleIdx, d.finalP}
	}

	// Draw each driver as connected line segments using their assigned style.
	for _, drv := range drvs {
		st := posLineStyles[drv.styleIdx]
		pts := drv.laps
		for i := 0; i < len(pts); i++ {
			c := toCol(pts[i].lap)
			r := toRow(pts[i].pos)
			if r >= 0 && r < plotH && c >= 0 && c < plotW {
				grid[r][c] = pixel{st.dot, drv.col, true}
			}
			if i+1 < len(pts) {
				c2 := toCol(pts[i+1].lap)
				r2 := toRow(pts[i+1].pos)
				dx, dy := c2-c, r2-r
				steps := absInt(dx)
				if absInt(dy) > steps {
					steps = absInt(dy)
				}
				if steps == 0 {
					continue
				}
				for s := 1; s < steps; s++ {
					pc := c + dx*s/steps
					pr := r + dy*s/steps
					if pr >= 0 && pr < plotH && pc >= 0 && pc < plotW && !grid[pr][pc].set {
						adx, ady := absInt(dx), absInt(dy)
						ch := st.hori
						if ady*10 < adx*3 {
							ch = st.hori
						} else if adx*10 < ady*3 {
							ch = '│'
						} else {
							ch = st.diag
						}
						grid[pr][pc] = pixel{ch, drv.col, true}
					}
				}
			}
		}
	}

	// Y-axis tick labels: map pixel row → label for positions 1,5,10,15,20.
	// Precompute using toRow so each label appears exactly once at the correct row.
	yTickRows := make(map[int]string)
	for _, pos := range []int{1, 5, 10, 15, 20} {
		if pos <= maxPos {
			r := toRow(pos)
			label := fmt.Sprintf("%2d", pos)
			yTickRows[r] = label
		}
	}

	axisStyle := lipgloss.NewStyle().Foreground(styles.ColorMuted)
	sepStyle := lipgloss.NewStyle().Foreground(styles.ColorMuted)

	var sb strings.Builder
	sb.WriteString(lipgloss.NewStyle().Bold(true).Foreground(styles.ColorSubtle).
		Render(fmt.Sprintf("  Position Changes  ·  %s %s %d", a.session.CountryName, a.session.SessionName, a.session.Year)))
	sb.WriteRune('\n')

	for row := 0; row < plotH; row++ {
		if label, ok := yTickRows[row]; ok {
			sb.WriteString(axisStyle.Render(label + " │"))
		} else {
			sb.WriteString(axisStyle.Render("   │"))
		}

		for col := 0; col < plotW; col++ {
			px := grid[row][col]
			if px.set {
				sb.WriteString(lipgloss.NewStyle().Foreground(px.col).Render(string(px.ch)))
			} else {
				sb.WriteRune(' ')
			}
		}
		sb.WriteString(sepStyle.Render("│"))

		// Right sidebar: driver tag anchored at their final position row.
		if se, ok := sideRow[row]; ok {
			st := posLineStyles[se.styleIdx]
			posLabel := fmt.Sprintf("P%-2d", se.finalP)
			lineInd := string(st.dot) + string(st.hori)
			entry := fmt.Sprintf(" %-3s %-3s %s", posLabel, se.acronym, lineInd)
			sb.WriteString(lipgloss.NewStyle().Foreground(se.col).Render(entry))
		}
		sb.WriteRune('\n')
	}

	// X-axis — tick every 5 laps, always include lap 1 and maxLap.
	sb.WriteString(axisStyle.Render("   └" + strings.Repeat("─", plotW+4)))
	sb.WriteRune('\n')

	// Build label row: place lap numbers at their pixel column.
	labelRow := make([]byte, plotW)
	for i := range labelRow {
		labelRow[i] = ' '
	}
	step := 5
	if maxLap > 50 {
		step = 10
	}
	tickLaps := []int{1}
	for l := step; l < maxLap; l += step {
		tickLaps = append(tickLaps, l)
	}
	tickLaps = append(tickLaps, maxLap)
	for _, lap := range tickLaps {
		col := toCol(lap)
		s := fmt.Sprintf("%d", lap)
		for si, ch := range []byte(s) {
			pos := col + si
			if pos >= 0 && pos < len(labelRow) {
				labelRow[pos] = ch
			}
		}
	}
	sb.WriteString("    " + axisStyle.Render(string(labelRow)))
	sb.WriteRune('\n')
	sb.WriteString(axisStyle.Render(fmt.Sprintf("    %*sLap →", plotW/2, "")))
	sb.WriteRune('\n')

	return sb.String()
}

// ── helpers ───────────────────────────────────────────────────────────────────

func teamColorFromDrivers(teamName string, drivers []openf1.Driver) lipgloss.Color {
	for _, d := range drivers {
		if d.TeamName == teamName && d.TeamColour != "" {
			return lipgloss.Color("#" + d.TeamColour)
		}
	}
	return lipgloss.Color("#888888")
}
