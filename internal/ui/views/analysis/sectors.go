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

// renderSectorChart shows each driver's best sector splits with delta to the
// overall fastest sector and a theoretical best lap (sum of personal bests).
func (a *Analysis) renderSectorChart(w, h int) string {
	drMap := a.driverMap()

	type secBests struct {
		num            int
		acronym, team  string
		bestS1, bestS2, bestS3 float64
	}

	bySec := map[int]*secBests{}
	for _, lap := range a.laps {
		if lap.LapDuration <= 0 || lap.IsPitOutLap {
			continue
		}
		b, ok := bySec[lap.DriverNumber]
		if !ok {
			drv := drMap[lap.DriverNumber]
			b = &secBests{num: lap.DriverNumber, acronym: drv.NameAcronym, team: drv.TeamName,
				bestS1: math.MaxFloat64, bestS2: math.MaxFloat64, bestS3: math.MaxFloat64}
			bySec[lap.DriverNumber] = b
		}
		if lap.DurationSector1 > 0 && lap.DurationSector1 < b.bestS1 {
			b.bestS1 = lap.DurationSector1
		}
		if lap.DurationSector2 > 0 && lap.DurationSector2 < b.bestS2 {
			b.bestS2 = lap.DurationSector2
		}
		if lap.DurationSector3 > 0 && lap.DurationSector3 < b.bestS3 {
			b.bestS3 = lap.DurationSector3
		}
	}

	var list []secBests
	for _, b := range bySec {
		if b.bestS1 != math.MaxFloat64 {
			list = append(list, *b)
		}
	}
	if len(list) == 0 {
		return common.Centred(w, h, styles.DimStyle.Render("No sector data available"))
	}
	sort.Slice(list, func(i, j int) bool {
		si := noInf(list[i].bestS1) + noInf(list[i].bestS2) + noInf(list[i].bestS3)
		sj := noInf(list[j].bestS1) + noInf(list[j].bestS2) + noInf(list[j].bestS3)
		return si < sj
	})

	besS1, besS2, besS3 := math.MaxFloat64, math.MaxFloat64, math.MaxFloat64
	for _, b := range list {
		if b.bestS1 < besS1 {
			besS1 = b.bestS1
		}
		if b.bestS2 < besS2 {
			besS2 = b.bestS2
		}
		if b.bestS3 < besS3 {
			besS3 = b.bestS3
		}
	}

	const labelW = 6
	secW := (w - labelW - 14) / 3
	if secW < 16 {
		secW = 16
	}
	barW := secW - 13
	if barW < 3 {
		barW = 3
	}

	lines := []string{
		"",
		sectionTitle("SECTOR ANALYSIS", "best sector times  ·  ★ = overall fastest sector"),
		"",
		styles.DimStyle.Render(fmt.Sprintf("  %-4s %-6s  %-*s  %-*s  %-*s",
			"POS", "DRV", secW, "SECTOR 1", secW, "SECTOR 2", secW, "SECTOR 3")),
		"  " + styles.Divider.Render(safeRep("─", w-4)),
	}

	renderSec := func(val, best float64, teamCol lipgloss.Color) string {
		if val == math.MaxFloat64 {
			return safeRep(" ", secW)
		}
		delta := val - best
		timeStr := fmt.Sprintf("%.3f", val)
		var col lipgloss.Color
		var tag string
		if delta < 0.001 {
			col = styles.ColorPurple
			tag = lipgloss.NewStyle().Foreground(styles.ColorPurple).Render("★")
		} else {
			col = teamCol
			tag = styles.DimStyle.Render(fmt.Sprintf("+%.3f", delta))
		}
		norm := 1.0 - delta/3.0
		if norm < 0 {
			norm = 0
		}
		if norm > 1 {
			norm = 1
		}
		filled := int(norm * float64(barW))
		bar := lipgloss.NewStyle().Foreground(col).Render(safeRep("█", filled)) +
			lipgloss.NewStyle().Foreground(styles.ColorMuted).Render(safeRep("·", barW-filled))
		return lipgloss.NewStyle().Foreground(col).Render(timeStr) + " " + tag + " " + bar
	}

	for rank, b := range list {
		teamCol := styles.TeamColor(b.team)
		acronym := b.acronym
		if acronym == "" {
			acronym = fmt.Sprintf("#%d", b.num)
		}
		rankStr := styles.DimStyle.Render(fmt.Sprintf("P%-2d", rank+1))
		label := lipgloss.NewStyle().Foreground(teamCol).Bold(true).Width(labelW).Render(acronym)

		theoBest := noInf(b.bestS1) + noInf(b.bestS2) + noInf(b.bestS3)
		theoStr := styles.DimStyle.Render(fmt.Sprintf("  [%s]", formatLapSec(theoBest)))

		lines = append(lines,
			fmt.Sprintf("  %s %s  %s  %s  %s%s",
				rankStr, label,
				renderSec(b.bestS1, besS1, teamCol),
				renderSec(b.bestS2, besS2, teamCol),
				renderSec(b.bestS3, besS3, teamCol),
				theoStr))
	}

	lines = append(lines, "",
		styles.DimStyle.Render("  [time] = theoretical best lap  (sum of each driver's best sectors)"))
	return strings.Join(lines, "\n")
}
