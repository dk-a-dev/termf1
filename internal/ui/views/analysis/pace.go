package analysis

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"github.com/charmbracelet/lipgloss"
	"github.com/dk-a-dev/termf1/v2/internal/ui/styles"
	"github.com/dk-a-dev/termf1/v2/internal/ui/views/common"
)

// renderPaceChart draws a horizontal bar chart of each driver's best and
// median lap times relative to pace leader, sorted by best lap.
func (a *Analysis) renderPaceChart(w, h int) string {
	drMap := a.driverMap()

	type driverPace struct {
		num            int
		acronym, team  string
		best, median   float64
	}

	byDriver := map[int][]float64{}
	for _, lap := range a.laps {
		if lap.LapDuration > 0 && !lap.IsPitOutLap {
			byDriver[lap.DriverNumber] = append(byDriver[lap.DriverNumber], lap.LapDuration)
		}
	}

	var paceList []driverPace
	for num, times := range byDriver {
		if len(times) == 0 {
			continue
		}
		best := math.MaxFloat64
		for _, t := range times {
			if t < best {
				best = t
			}
		}
		sort.Float64s(times)
		// Filter to within 7 % of best to exclude safety-car distortion.
		var filtered []float64
		for _, t := range times {
			if t <= best*1.07 {
				filtered = append(filtered, t)
			}
		}
		if len(filtered) == 0 {
			filtered = times
		}
		med := filtered[len(filtered)/2]
		drv := drMap[num]
		paceList = append(paceList, driverPace{num, drv.NameAcronym, drv.TeamName, best, med})
	}
	if len(paceList) == 0 {
		return common.Centred(w, h, styles.DimStyle.Render("No pace data available"))
	}
	sort.Slice(paceList, func(i, j int) bool { return paceList[i].best < paceList[j].best })

	poleBest := paceList[0].best
	poleMed := math.MaxFloat64
	for _, p := range paceList {
		if p.median < poleMed {
			poleMed = p.median
		}
	}

	maxBestDelta, maxMedDelta := 0.001, 0.001
	for _, p := range paceList {
		if d := p.best - poleBest; d > maxBestDelta {
			maxBestDelta = d
		}
		if d := p.median - poleMed; d > maxMedDelta {
			maxMedDelta = d
		}
	}

	const (
		rankW  = 4
		labelW = 6
	)
	fixedW := rankW + labelW + 24
	barMaxW := (w - fixedW) / 2
	if barMaxW < 4 {
		barMaxW = 4
	}
	if barMaxW > 50 {
		barMaxW = 50
	}

	lines := []string{
		"",
		sectionTitle("RACE PACE",
			fmt.Sprintf("pole %s (%s)  ·  best median %s",
				formatLapSec(poleBest), paceList[0].acronym, formatLapSec(poleMed))),
		"",
		styles.DimStyle.Render(fmt.Sprintf("  %-4s %-6s %-10s %-10s  %-*s  %-*s",
			"POS", "DRV", "BEST LAP", "Δ POLE",
			barMaxW, "BEST LAP GAP →",
			barMaxW, "MEDIAN PACE GAP →")),
		"  " + styles.Divider.Render(safeRep("─", w-4)),
	}

	for rank, p := range paceList {
		teamCol := styles.TeamColor(p.team)
		acronym := p.acronym
		if acronym == "" {
			acronym = fmt.Sprintf("#%d", p.num)
		}

		bestDelta := p.best - poleBest
		if bestDelta < 0 {
			bestDelta = 0
		}
		medDelta := p.median - poleMed
		if medDelta < 0 {
			medDelta = 0
		}

		bestBarW := int(bestDelta / maxBestDelta * float64(barMaxW))
		if bestBarW > barMaxW {
			bestBarW = barMaxW
		}
		medBarW := int(medDelta / maxMedDelta * float64(barMaxW))
		if medBarW > barMaxW {
			medBarW = barMaxW
		}

		var rankColor lipgloss.Color
		switch rank {
		case 0:
			rankColor = lipgloss.Color("#FFD700")
		case 1:
			rankColor = lipgloss.Color("#C0C0C0")
		case 2:
			rankColor = lipgloss.Color("#CD7F32")
		default:
			rankColor = teamCol
		}

		rankStr := lipgloss.NewStyle().Foreground(rankColor).Bold(true).Render(fmt.Sprintf("P%-2d", rank+1))
		label := lipgloss.NewStyle().Foreground(teamCol).Bold(true).Width(labelW).Render(acronym)
		lapStr := lipgloss.NewStyle().Foreground(teamCol).Render(fmt.Sprintf("%-10s", formatLapSec(p.best)))

		var deltaStr string
		if bestDelta < 0.001 {
			deltaStr = fmt.Sprintf("%-10s", lipgloss.NewStyle().Foreground(styles.ColorGreen).Render("FASTEST"))
		} else {
			deltaStr = fmt.Sprintf("%-10s", styles.DimStyle.Render(fmt.Sprintf("+%.3fs", bestDelta)))
		}

		bestBar := hBar("█", "·", bestBarW, barMaxW, teamCol, styles.ColorMuted)
		medBar := hBar("░", "·", medBarW, barMaxW, styles.ColorSubtle, styles.ColorMuted)

		lines = append(lines, fmt.Sprintf("  %s  %s  %s  %s  %s  %s",
			rankStr, label, lapStr, deltaStr, bestBar, medBar))
	}

	lines = append(lines, "",
		styles.DimStyle.Render("  █ = best lap gap to pole   ░ = median pace gap"))
	return joinLines(lines)
}

func joinLines(lines []string) string {
	var b strings.Builder
	for i, l := range lines {
		if i > 0 {
			b.WriteByte('\n')
		}
		b.WriteString(l)
	}
	return b.String()
}