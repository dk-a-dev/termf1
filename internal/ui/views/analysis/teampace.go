package analysis

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/dk-a-dev/termf1/internal/ui/styles"
	"github.com/dk-a-dev/termf1/internal/ui/views/common"
)

// ── Chart 8: Team Pace Comparison ────────────────────────────────────────────
//
// Renders a terminal box-plot for each team showing the distribution
// of non-pit-out lap times (whiskers = 5th/95th pct, box = 25th/75th, median line).
// Teams are sorted fastest to slowest by median lap time (mirrors seaborn example).

func (a *Analysis) renderTeamPaceChart(w, h int) string {
	drMap := a.driverMap()

	// Group lap times by team, skipping pit-out laps and safety-car outliers
	// (> 107% of session fastest lap).
	teamLaps := make(map[string][]float64)
	teamColour := make(map[string]lipgloss.Color)

	var sessionFastest float64 = math.MaxFloat64
	for _, l := range a.laps {
		if l.LapDuration > 0 && !l.IsPitOutLap {
			if l.LapDuration < sessionFastest {
				sessionFastest = l.LapDuration
			}
		}
	}
	cutoff := sessionFastest * 1.07

	for _, l := range a.laps {
		if l.LapDuration <= 0 || l.IsPitOutLap || l.LapDuration > cutoff {
			continue
		}
		drv, ok := drMap[l.DriverNumber]
		if !ok || drv.TeamName == "" {
			continue
		}
		teamLaps[drv.TeamName] = append(teamLaps[drv.TeamName], l.LapDuration)
		if drv.TeamColour != "" {
			teamColour[drv.TeamName] = lipgloss.Color("#" + drv.TeamColour)
		}
	}

	if len(teamLaps) == 0 {
		return common.Centred(w, h, styles.DimStyle.Render("No lap data for team pace"))
	}

	// Compute stats per team.
	type teamStats struct {
		name            string
		col             lipgloss.Color
		median          float64
		q1, q3          float64
		whiskerLo, whiskerHi float64
		min, max        float64
	}

	var teams []teamStats
	for name, lts := range teamLaps {
		sort.Float64s(lts)
		n := len(lts)

		pct := func(p float64) float64 {
			idx := p * float64(n-1)
			lo := int(idx)
			fr := idx - float64(lo)
			if lo+1 >= n {
				return lts[lo]
			}
			return lts[lo]*(1-fr) + lts[lo+1]*fr
		}

		col := teamColour[name]
		if col == "" {
			col = lipgloss.Color("#888888")
		}

		q1 := pct(0.25)
		med := pct(0.50)
		q3 := pct(0.75)
		iqr := q3 - q1

		// Tukey fences.
		wLo, wHi := q1-1.5*iqr, q3+1.5*iqr
		// Clamp to actual data range.
		if wLo < lts[0] {
			wLo = lts[0]
		}
		if wHi > lts[n-1] {
			wHi = lts[n-1]
		}

		teams = append(teams, teamStats{
			name:      name,
			col:       col,
			median:    med,
			q1:        q1,
			q3:        q3,
			whiskerLo: wLo,
			whiskerHi: wHi,
			min:       lts[0],
			max:       lts[n-1],
		})
	}

	// Sort by median ascending (fastest first).
	sort.Slice(teams, func(i, j int) bool { return teams[i].median < teams[j].median })

	// ── Scale ─────────────────────────────────────────────────────────────────
	globalMin, globalMax := math.MaxFloat64, -math.MaxFloat64
	for _, t := range teams {
		if t.whiskerLo < globalMin {
			globalMin = t.whiskerLo
		}
		if t.whiskerHi > globalMax {
			globalMax = t.whiskerHi
		}
	}
	if globalMax <= globalMin {
		globalMax = globalMin + 1
	}
	padding := (globalMax - globalMin) * 0.05
	globalMin -= padding
	globalMax += padding

	// Available width for the chart area (after left label).
	labelW := 16
	axisLabelW := 8 // e.g. "1:35.000"
	barAreaW := w - labelW - axisLabelW - 2
	if barAreaW < 20 {
		barAreaW = 20
	}

	toX := func(sec float64) int {
		return int((sec - globalMin) / (globalMax - globalMin) * float64(barAreaW-1))
	}

	// ── Drawing helpers ───────────────────────────────────────────────────────
	drawBox := func(col lipgloss.Color, x1, xMed, x2, xWlo, xWhi int) string {
		line := make([]rune, barAreaW)
		for i := range line {
			line[i] = ' '
		}

		// Whisker lo ─ to box.
		for x := xWlo; x < x1 && x < barAreaW; x++ {
			if x >= 0 {
				line[x] = '─'
			}
		}
		// Box interior Q1–Q3.
		for x := x1; x <= x2 && x < barAreaW; x++ {
			if x >= 0 {
				if x == x1 || x == x2 {
					line[x] = '│'
				} else if x == xMed {
					line[x] = '┃' // median
				} else {
					line[x] = '▓'
				}
			}
		}
		// Box to whisker hi.
		for x := x2 + 1; x <= xWhi && x < barAreaW; x++ {
			if x >= 0 {
				line[x] = '─'
			}
		}

		return lipgloss.NewStyle().Foreground(col).Render(string(line))
	}

	// ── Render ────────────────────────────────────────────────────────────────
	title := lipgloss.NewStyle().Bold(true).Foreground(styles.ColorSubtle).
		Render(fmt.Sprintf("  Team Pace  ·  %s %s %d  (quicklaps only, 107%% cut)",
			a.session.CountryName, a.session.SessionName, a.session.Year))

	var sb strings.Builder
	sb.WriteString(title + "\n\n")

	muted := lipgloss.NewStyle().Foreground(styles.ColorMuted)

	for _, t := range teams {
		x1 := toX(t.q1)
		x2 := toX(t.q3)
		xMed := toX(t.median)
		xWlo := toX(t.whiskerLo)
		xWhi := toX(t.whiskerHi)

		// Clamp to bounds.
		clamp := func(x int) int {
			if x < 0 {
				return 0
			}
			if x >= barAreaW {
				return barAreaW - 1
			}
			return x
		}
		x1, x2, xMed = clamp(x1), clamp(x2), clamp(xMed)
		xWlo, xWhi = clamp(xWlo), clamp(xWhi)

		// Team label.
		label := t.name
		if len(label) > labelW-1 {
			label = label[:labelW-1]
		}
		labelR := lipgloss.NewStyle().Width(labelW).Foreground(t.col).Bold(true).Render(label)

		// Box plot row.
		box := drawBox(t.col, x1, xMed, x2, xWlo, xWhi)

		// Median value annotation.
		medStr := muted.Render(" " + formatLapSec(t.median))

		sb.WriteString(labelR + box + medStr + "\n")
		sb.WriteString(strings.Repeat(" ", labelW) +
			muted.Render(fmt.Sprintf("med %-8s  IQR %s–%s",
				formatLapSec(t.median), formatLapSec(t.q1), formatLapSec(t.q3))) + "\n\n")
	}

	// X-axis ruler.
	rulerParts := strings.Repeat(" ", labelW)
	nTicks := 5
	for i := 0; i <= nTicks; i++ {
		sec := globalMin + (globalMax-globalMin)*float64(i)/float64(nTicks)
		tick := formatLapSec(sec)
		padded := fmt.Sprintf("%-*s", barAreaW/nTicks, tick)
		rulerParts += padded
	}
	sb.WriteString(lipgloss.NewStyle().Foreground(styles.ColorMuted).Render(rulerParts))
	sb.WriteString("\n")

	// Legend explanation.
	sb.WriteString("\n  ")
	sb.WriteString(muted.Render("│"))
	sb.WriteString(" Q1   ")
	sb.WriteString(lipgloss.NewStyle().Foreground(styles.ColorSubtle).Render("┃"))
	sb.WriteString(" median   ")
	sb.WriteString(muted.Render("│"))
	sb.WriteString(" Q3   ")
	sb.WriteString(muted.Render("─── whiskers (1.5×IQR)"))

	return sb.String()
}


