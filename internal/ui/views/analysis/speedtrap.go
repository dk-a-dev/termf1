package analysis

import (
	"fmt"
	"sort"

	"github.com/charmbracelet/lipgloss"
	"github.com/dk-a-dev/termf1/v2/internal/ui/styles"
	"github.com/dk-a-dev/termf1/v2/internal/ui/views/common"
)

// renderSpeedChart draws a horizontal bar chart of peak speeds at three
// sensor points: ST (finish straight), I1, and I2.
func (a *Analysis) renderSpeedChart(w, h int) string {
	drMap := a.driverMap()

	maxST := map[int]int{}
	maxI1 := map[int]int{}
	maxI2 := map[int]int{}
	for _, lap := range a.laps {
		if lap.StSpeed > maxST[lap.DriverNumber] {
			maxST[lap.DriverNumber] = lap.StSpeed
		}
		if lap.I1Speed > maxI1[lap.DriverNumber] {
			maxI1[lap.DriverNumber] = lap.I1Speed
		}
		if lap.I2Speed > maxI2[lap.DriverNumber] {
			maxI2[lap.DriverNumber] = lap.I2Speed
		}
	}
	if len(maxST) == 0 {
		return common.Centred(w, h, styles.DimStyle.Render("No speed trap data available"))
	}

	type entry struct {
		num            int
		acronym, team  string
		st, i1, i2     int
	}
	seen := map[int]bool{}
	var entries []entry
	for num := range maxST {
		if seen[num] {
			continue
		}
		seen[num] = true
		drv := drMap[num]
		entries = append(entries, entry{num, drv.NameAcronym, drv.TeamName,
			maxST[num], maxI1[num], maxI2[num]})
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].st > entries[j].st })

	globalMax := 0
	for _, e := range entries {
		if e.st > globalMax {
			globalMax = e.st
		}
	}
	if globalMax == 0 {
		globalMax = 400
	}

	const labelW = 6
	const statsW = 28
	barW := w - labelW - statsW - 10
	if barW < 10 {
		barW = 10
	}
	if barW > 60 {
		barW = 60
	}

	lines := []string{
		"",
		sectionTitle("SPEED TRAP",
			fmt.Sprintf("peak: %d km/h (%s)  ·  ST = finish straight  I1/I2 = intermediate",
				globalMax, entries[0].acronym)),
		"",
		styles.DimStyle.Render(fmt.Sprintf("  %-4s %-6s  %-*s  ST     I1     I2",
			"POS", "DRV", barW, "SPEED TRAP BAR (ST)")),
		"  " + styles.Divider.Render(safeRep("─", w-4)),
	}

	for rank, e := range entries {
		teamCol := styles.TeamColor(e.team)
		acronym := e.acronym
		if acronym == "" {
			acronym = fmt.Sprintf("#%d", e.num)
		}

		filled := e.st * barW / globalMax
		if filled > barW {
			filled = barW
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
		bar := hBar("█", "·", filled, barW, teamCol, styles.ColorMuted)

		stStr := lipgloss.NewStyle().Foreground(rankColor).Bold(rank == 0).Render(fmt.Sprintf("%4d", e.st))
		i1Str := styles.DimStyle.Render(fmt.Sprintf("%4d", e.i1))
		i2Str := styles.DimStyle.Render(fmt.Sprintf("%4d", e.i2))

		lines = append(lines,
			fmt.Sprintf("  %s  %s  %s  %s   %s   %s", rankStr, label, bar, stStr, i1Str, i2Str))
	}

	lines = append(lines, "", styles.DimStyle.Render("  km/h"))
	return joinLines(lines)
}
