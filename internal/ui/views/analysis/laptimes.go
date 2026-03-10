package analysis

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/devkeshwani/termf1/internal/api/openf1"
	"github.com/devkeshwani/termf1/internal/ui/styles"
	"github.com/devkeshwani/termf1/internal/ui/views/common"
)

// ── Chart 2: 2-D Lap Time Line Chart ─────────────────────────────────────────

// renderLapLineChart draws a proper 2-D grid where each driver's lap times are
// plotted as coloured dots/lines, enabling direct driver-vs-driver comparison.
func (a *Analysis) renderLapLineChart(w, h int) string {
	drMap := a.driverMap()

	type pt struct{ lap int; t float64 }
	byDriver := map[int][]pt{}
	for _, l := range a.laps {
		if l.LapDuration > 0 && !l.IsPitOutLap {
			byDriver[l.DriverNumber] = append(byDriver[l.DriverNumber], pt{l.LapNumber, l.LapDuration})
		}
	}
	if len(byDriver) == 0 {
		return common.Centred(w, h, styles.DimStyle.Render("No lap time data"))
	}

	type drvData struct {
		num                int
		acronym, teamName  string
		pts                []pt
		best               float64
	}
	var drvs []drvData
	for num, pts := range byDriver {
		sort.Slice(pts, func(i, j int) bool { return pts[i].lap < pts[j].lap })
		best := math.MaxFloat64
		for _, p := range pts {
			if p.t < best {
				best = p.t
			}
		}
		drv := drMap[num]
		drvs = append(drvs, drvData{num, drv.NameAcronym, drv.TeamName, pts, best})
	}
	sort.Slice(drvs, func(i, j int) bool { return drvs[i].best < drvs[j].best })
	if len(drvs) > 15 {
		drvs = drvs[:15]
	}

	// Compute y range using p10–p90 of all lap times to exclude anomalies.
	var allT []float64
	for _, d := range drvs {
		for _, p := range d.pts {
			allT = append(allT, p.t)
		}
	}
	sort.Float64s(allT)
	n := len(allT)
	tMin := allT[n/10] - 0.3
	tMax := allT[n*9/10] + 0.3
	tRange := tMax - tMin
	if tRange < 1 {
		tMin -= 0.5
		tMax += 0.5
		tRange = 1
	}

	maxLap := 0
	for _, d := range drvs {
		for _, p := range d.pts {
			if p.lap > maxLap {
				maxLap = p.lap
			}
		}
	}
	if maxLap == 0 {
		maxLap = 1
	}

	const (
		yAxisW  = 9 // "1:23.456" width + space
		titleH  = 3
		xAxisH  = 3 // axis line + tick labels + "Lap →"
		legendH = 3
	)
	cW := w - yAxisW - 6
	cH := h - titleH - xAxisH - legendH
	if cW < 20 {
		cW = 20
	}
	if cH < 6 {
		cH = 6
	}

	// Allocate a 2-D grid of styled cells.
	type cell struct {
		r   rune
		col lipgloss.Color
		set bool
	}
	grid := make([][]cell, cH)
	for i := range grid {
		grid[i] = make([]cell, cW)
	}

	lapToCol := func(lap int) int {
		c := (lap - 1) * (cW - 1) / maxLap
		if c < 0 {
			c = 0
		}
		if c >= cW {
			c = cW - 1
		}
		return c
	}
	timeToRow := func(t float64) int {
		norm := (t - tMin) / tRange
		if norm < 0 {
			norm = 0
		}
		if norm > 1 {
			norm = 1
		}
		r := int(norm * float64(cH-1))
		if r < 0 {
			r = 0
		}
		if r >= cH {
			r = cH - 1
		}
		return r
	}

	// Plot each driver: draw connecting dots then place the data point on top.
	for di, d := range drvs {
		ch := dotChars[di%len(dotChars)]
		col := styles.TeamColor(d.teamName)

		prevC, prevR := -1, -1
		for _, p := range d.pts {
			if p.t < tMin || p.t > tMax {
				prevC = -1
				continue
			}
			c := lapToCol(p.lap)
			r := timeToRow(p.t)

			// Interpolate connector dots between previous and current point.
			if prevC >= 0 && c > prevC {
				for step := 1; step < c-prevC; step++ {
					ic := prevC + step
					ir := prevR + (r-prevR)*step/(c-prevC)
					if ic >= 0 && ic < cW && ir >= 0 && ir < cH {
						if !grid[ir][ic].set {
							grid[ir][ic] = cell{'·', col, true}
						}
					}
				}
			}
			// Place the actual data point (always overwrites connectors).
			if r >= 0 && r < cH && c >= 0 && c < cW {
				grid[r][c] = cell{ch, col, true}
			}
			prevC, prevR = c, r
		}
	}

	var sb strings.Builder
	sb.WriteString("\n")
	sb.WriteString(sectionTitle("LAP TIME COMPARISON",
		fmt.Sprintf("top %d drivers · laps 1–%d · %s – %s",
			len(drvs), maxLap, formatLapSec(tMin), formatLapSec(tMax))))
	sb.WriteString("\n\n")

	// Render rows with y-axis label.
	for row := 0; row < cH; row++ {
		t := tMin + float64(row)/float64(cH-1)*tRange
		yLabel := styles.DimStyle.Render(fmt.Sprintf("%-8s", formatLapSec(t)))
		sb.WriteString("  ")
		sb.WriteString(yLabel)
		sb.WriteString(styles.DimStyle.Render("│"))
		for col := 0; col < cW; col++ {
			c := grid[row][col]
			if c.set {
				sb.WriteString(lipgloss.NewStyle().Foreground(c.col).Render(string(c.r)))
			} else {
				sb.WriteRune(' ')
			}
		}
		sb.WriteRune('\n')
	}

	// X axis.
	sb.WriteString("  " + safeRep(" ", yAxisW) + styles.DimStyle.Render("└"+safeRep("─", cW)) + "\n")

	// X tick labels.
	ticks := make([]byte, cW+2)
	for i := range ticks {
		ticks[i] = ' '
	}
	for lap := 1; lap <= maxLap; lap += 10 {
		col := lapToCol(lap)
		s := fmt.Sprintf("%d", lap)
		for si, ch := range s {
			pos := col + si
			if pos < len(ticks) {
				ticks[pos] = byte(ch)
			}
		}
	}
	sb.WriteString("  " + safeRep(" ", yAxisW+1) + styles.DimStyle.Render(string(ticks)) + "\n")
	sb.WriteString("  " + safeRep(" ", yAxisW+1+cW/2) + styles.DimStyle.Render("Lap →") + "\n")

	// Legend.
	sb.WriteString("\n  ")
	for di, d := range drvs {
		ch := dotChars[di%len(dotChars)]
		col := styles.TeamColor(d.teamName)
		acronym := d.acronym
		if acronym == "" {
			acronym = fmt.Sprintf("#%d", d.num)
		}
		sb.WriteString(lipgloss.NewStyle().Foreground(col).Render(string(ch)+" "+acronym))
		sb.WriteString("  ")
	}
	sb.WriteString("\n" + styles.DimStyle.Render("  · = connector between consecutive laps   shape = driver"))
	return sb.String()
}

// ── Chart 3: Per-driver sparklines ───────────────────────────────────────────

func (a *Analysis) renderLapSparklines(w, h int) string {
	drMap := a.driverMap()

	byDriver := map[int][]openf1.Lap{}
	for _, lap := range a.laps {
		if lap.LapDuration > 0 {
			byDriver[lap.DriverNumber] = append(byDriver[lap.DriverNumber], lap)
		}
	}

	type dl struct {
		num            int
		laps           []openf1.Lap
		best, avg      float64
		acronym, team  string
	}
	var driverList []dl
	for num, laps := range byDriver {
		sort.Slice(laps, func(i, j int) bool { return laps[i].LapNumber < laps[j].LapNumber })
		best := math.MaxFloat64
		total, count := 0.0, 0
		for _, l := range laps {
			if l.LapDuration > 0 && !l.IsPitOutLap {
				if l.LapDuration < best {
					best = l.LapDuration
				}
				total += l.LapDuration
				count++
			}
		}
		if best == math.MaxFloat64 {
			continue
		}
		avg := 0.0
		if count > 0 {
			avg = total / float64(count)
		}
		drv := drMap[num]
		driverList = append(driverList, dl{num, laps, best, avg, drv.NameAcronym, drv.TeamName})
	}
	sort.Slice(driverList, func(i, j int) bool { return driverList[i].best < driverList[j].best })
	if len(driverList) == 0 {
		return common.Centred(w, h, styles.DimStyle.Render("No lap time data"))
	}

	globalMin, globalMax := math.MaxFloat64, 0.0
	for _, d := range driverList {
		for _, l := range d.laps {
			if l.LapDuration > 0 && !l.IsPitOutLap && l.LapDuration < d.best*1.15 {
				if l.LapDuration < globalMin {
					globalMin = l.LapDuration
				}
				if l.LapDuration > globalMax {
					globalMax = l.LapDuration
				}
			}
		}
	}
	spread := globalMax - globalMin
	if spread < 0.1 {
		spread = 0.1
	}

	maxLapNum := 0
	for _, d := range driverList {
		for _, l := range d.laps {
			if l.LapNumber > maxLapNum {
				maxLapNum = l.LapNumber
			}
		}
	}
	if maxLapNum == 0 {
		maxLapNum = 1
	}

	const (
		rankW  = 4
		labelW = 5
		statW  = 34
	)
	sparkW := w - rankW - labelW - statW - 8
	if sparkW < 10 {
		sparkW = 10
	}

	lines := []string{
		"",
		sectionTitle("LAP TIME SPARKLINES",
			fmt.Sprintf("%d drivers · %d laps · range %s – %s",
				len(driverList), maxLapNum,
				formatLapSec(globalMin), formatLapSec(globalMax))),
		"  " + styles.Divider.Render(safeRep("─", w-4)),
		"",
	}

	for rank, d := range driverList {
		teamCol := styles.TeamColor(d.team)
		acronym := d.acronym
		if acronym == "" {
			acronym = fmt.Sprintf("#%d", d.num)
		}
		rankStr := styles.DimStyle.Render(fmt.Sprintf("P%-2d", rank+1))
		label := lipgloss.NewStyle().Foreground(teamCol).Bold(true).Width(labelW).Render(acronym)

		colLaps := make([]float64, sparkW)
		for i := range colLaps {
			colLaps[i] = -1
		}
		for _, l := range d.laps {
			if l.LapDuration <= 0 || l.IsPitOutLap || l.LapDuration > d.best*1.15 {
				continue
			}
			col := (l.LapNumber - 1) * sparkW / maxLapNum
			if col < 0 {
				col = 0
			}
			if col >= sparkW {
				col = sparkW - 1
			}
			if colLaps[col] < 0 || l.LapDuration < colLaps[col] {
				colLaps[col] = l.LapDuration
			}
		}

		var spark strings.Builder
		for _, v := range colLaps {
			if v < 0 {
				spark.WriteRune(' ')
				continue
			}
			norm := (globalMax - v) / spread
			if norm < 0 {
				norm = 0
			}
			if norm > 1 {
				norm = 1
			}
			idx := int(norm * 7)
			if idx > 7 {
				idx = 7
			}
			spark.WriteRune(sparkBlocks[idx])
		}

		var bestColor lipgloss.Color
		if rank == 0 {
			bestColor = lipgloss.Color("#FFD700")
		} else {
			bestColor = teamCol
		}
		stats := lipgloss.NewStyle().Foreground(bestColor).Render(formatLapSec(d.best)) +
			styles.DimStyle.Render("  avg "+formatLapSec(d.avg))
		sparkStr := lipgloss.NewStyle().Foreground(teamCol).Render(spark.String())
		lines = append(lines, "  "+rankStr+" "+label+" "+sparkStr+"  "+stats)
	}

	lines = append(lines, "",
		styles.DimStyle.Render(safeRep(" ", rankW+labelW+4)+"▁ slow  ▇ fast  blank = pit/outlap"),
		styles.DimStyle.Render(safeRep(" ", rankW+labelW+4)+
			fmt.Sprintf("Lap 1 ←────────────────────────→ %d", maxLapNum)))
	return strings.Join(lines, "\n")
}
