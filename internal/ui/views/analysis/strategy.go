package analysis

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/devkeshwani/termf1/internal/ui/styles"
	"github.com/devkeshwani/termf1/internal/ui/views/common"
)

// renderStrategy draws the tyre-strategy Gantt chart.
func (a *Analysis) renderStrategy(w, h int) string {
	drMap := a.driverMap()

	// Group stints by driver and find max lap.
	stintByDriver := map[int][]stintEntry{}
	maxLap := 0
	for _, st := range a.stints {
		stintByDriver[st.DriverNumber] = append(stintByDriver[st.DriverNumber], stintEntry{
			lapStart: st.LapStart, lapEnd: st.LapEnd, compound: st.Compound,
		})
		if st.LapEnd > maxLap {
			maxLap = st.LapEnd
		}
	}
	if maxLap == 0 {
		return common.Centred(w, h, styles.DimStyle.Render("No tyre stint data available"))
	}

	type drvInfo struct {
		num, lastLap        int
		acronym, teamName   string
	}
	seen := map[int]bool{}
	var drivers []drvInfo
	for num, sts := range stintByDriver {
		if seen[num] {
			continue
		}
		seen[num] = true
		last := 0
		for _, s := range sts {
			if s.lapEnd > last {
				last = s.lapEnd
			}
		}
		drv := drMap[num]
		drivers = append(drivers, drvInfo{num, last, drv.NameAcronym, drv.TeamName})
	}
	// Only include drivers who have at least one stint with a known compound.
	var filteredDrivers []drvInfo
	for _, drv := range drivers {
		hasData := false
		for _, st := range stintByDriver[drv.num] {
			if st.lapStart > 0 && strings.TrimSpace(st.compound) != "" {
				hasData = true
				break
			}
		}
		if hasData {
			filteredDrivers = append(filteredDrivers, drv)
		}
	}
	if len(filteredDrivers) == 0 {
		filteredDrivers = drivers // fallback: show all if none pass filter
	}
	drivers = filteredDrivers
	sort.Slice(drivers, func(i, j int) bool { return drivers[i].lastLap > drivers[j].lastLap })

	const labelW = 6
	barW := w - labelW - 6
	if barW < 10 {
		barW = 10
	}

	legend := "  " +
		lipgloss.NewStyle().Foreground(styles.ColorTyreSoft).Render("██ S") + "  " +
		lipgloss.NewStyle().Foreground(styles.ColorTyreMedium).Render("██ M") + "  " +
		lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).Render("██ H") + "  " +
		lipgloss.NewStyle().Foreground(styles.ColorTyreInter).Render("██ I") + "  " +
		lipgloss.NewStyle().Foreground(styles.ColorTyreWet).Render("██ W")

	lines := []string{
		"",
		sectionTitle("TYRE STRATEGY", fmt.Sprintf("%d drivers  ·  %d laps", len(drivers), maxLap)),
		legend,
		"  " + styles.Divider.Render(safeRep("─", w-4)),
		"",
	}

	for _, drv := range drivers {
		col := styles.TeamColor(drv.teamName)
		acronym := drv.acronym
		if acronym == "" {
			acronym = fmt.Sprintf("#%d", drv.num)
		}
		label := lipgloss.NewStyle().Foreground(col).Bold(true).Width(labelW).Render(acronym)

		bar := make([]string, barW)
		for i := range bar {
			bar[i] = styles.DimStyle.Render("·")
		}

		sts := stintByDriver[drv.num]
		sort.Slice(sts, func(i, j int) bool { return sts[i].lapStart < sts[j].lapStart })
		for _, st := range sts {
			start := (st.lapStart - 1) * barW / maxLap
			end := st.lapEnd * barW / maxLap
			if start < 0 {
				start = 0
			}
			if start >= barW {
				start = barW - 1
			}
			if end > barW {
				end = barW
			}
			blk := compoundBlock(st.compound)
			for i := start; i < end && i < barW; i++ {
				bar[i] = blk
			}
		}
		lines = append(lines, "  "+label+" "+strings.Join(bar, ""))
	}

	// Axis ticks
	var tickBuf strings.Builder
	tickBuf.WriteString(safeRep(" ", labelW+2))
	c := 0
	for c < barW {
		lap := c * maxLap / barW
		if lap%10 == 0 {
			s := fmt.Sprintf("%-5d", lap+1)
			if len(s) > barW-c {
				s = s[:barW-c]
			}
			tickBuf.WriteString(s)
			c += len(s)
		} else {
			tickBuf.WriteRune(' ')
			c++
		}
	}
	lines = append(lines, "",
		styles.DimStyle.Render(tickBuf.String()),
		styles.DimStyle.Render(safeRep(" ", labelW+2)+"Lap →"))
	return strings.Join(lines, "\n")
}

// stintEntry is the minimal data needed for the strategy chart.
type stintEntry struct {
	lapStart, lapEnd int
	compound         string
}
