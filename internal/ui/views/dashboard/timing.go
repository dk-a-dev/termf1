package dashboard

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/devkeshwani/termf1/internal/ui/styles"
)

// ── Timing panel (left) ───────────────────────────────────────────────────────

func (d *Dashboard2) renderTimingPanel(w, h int) string {
	if d.serverAlive && len(d.liveRows) > 0 {
		return d.renderLiveTiming(w, h)
	}
	// Server alive but idle, or server offline — show last race if available.
	if len(d.fbRows) > 0 {
		return d.renderFallbackTiming(w, h)
	}
	label := "Waiting for data…"
	if !d.serverAlive {
		label = "No live session · showing last race results when available"
	}
	return lipgloss.NewStyle().
		Width(w).Height(h).
		Padding(2, 2).
		Render(styles.DimStyle.Render(label))
}

// ── Live timing ───────────────────────────────────────────────────────────────

func (d *Dashboard2) renderLiveTiming(w, h int) string {
	hdr := liveTimingHeader()
	sep := styles.Divider.Render(strings.Repeat("─", w-2))
	lines := []string{hdr, sep}

	maxW := lipgloss.NewStyle().MaxWidth(w - 1)
	for i := range d.liveRows {
		lines = append(lines, maxW.Render(renderLiveRow(&d.liveRows[i])))
		if i >= h-6 {
			break
		}
	}

	modeLabel := lipgloss.NewStyle().Foreground(styles.ColorGreen).Render(" ● LIVE TIMING")
	if d.loading {
		modeLabel = styles.DimStyle.Render(d.spin.View() + " refreshing…")
	}
	lines = append(lines, "", modeLabel)

	content := lipgloss.JoinVertical(lipgloss.Left, lines...)
	return lipgloss.NewStyle().Width(w).MaxWidth(w).Height(h).Render(content)
}

// Column widths for the F1-style live timing grid.
// Total: 4+8+7+7+9+9+10+10+12+12+12+5 = 105 — fits a ~105-col left panel.
const (
	colPos    = 4
	colDriver = 8  // " NOR " (5) + "DRS" or "   " (3) = 8
	colTyre   = 7  // "(H)12 " compound+age
	colLap    = 7  // "L58 P2"
	colGap    = 9
	colInt    = 9
	colLast   = 10
	colBest   = 10
	colSector = 12 // "29.312 ████" time(7)+bar(4)+pad(1) = 12
	colSpeed  = 5
)

func liveTimingHeader() string {
	hdrs := []struct{ s string; w int }{
		{"POS", colPos},
		{"DRIVER", colDriver},
		{"TYR", colTyre},
		{"LAP", colLap},
		{"    GAP", colGap},
		{" INTERVAL", colInt},
		{" LAST LAP", colLast},
		{" BEST LAP", colBest},
		{"   S1", colSector},
		{"   S2", colSector},
		{"   S3", colSector},
		{" SPD", colSpeed},
	}
	cells := make([]string, len(hdrs))
	for i, h := range hdrs {
		cells[i] = lipgloss.NewStyle().Width(h.w).
			Foreground(styles.ColorSubtle).Bold(true).Render(h.s)
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, cells...)
}

// renderSegBar returns a compact coloured segment bar (up to 4 chars wide).
// Uses segment statuses if available; falls back to sector-level colour.
func renderSegBar(segs []int, overallFastest, personalFastest bool) string {
	const barWidth = 4
	if len(segs) == 0 {
		// Fallback: solid bar from sector status.
		if overallFastest {
			return lipgloss.NewStyle().Foreground(styles.ColorPurple).Render("████")
		}
		if personalFastest {
			return lipgloss.NewStyle().Foreground(styles.ColorGreen).Render("████")
		}
		return lipgloss.NewStyle().Foreground(styles.ColorMuted).Render("────")
	}
	// Use up to barWidth segments for display.
	show := segs
	if len(show) > barWidth {
		show = show[len(show)-barWidth:]
	}
	var sb strings.Builder
	for _, s := range show {
		switch s {
		case 2051: // overall fastest (purple)
			sb.WriteString(lipgloss.NewStyle().Foreground(styles.ColorPurple).Render("█"))
		case 2049: // currently overall fastest (green)
			sb.WriteString(lipgloss.NewStyle().Foreground(styles.ColorGreen).Render("█"))
		case 2048: // personal best (yellow)
			sb.WriteString(lipgloss.NewStyle().Foreground(styles.ColorYellow).Render("█"))
		case 64: // active segment
			sb.WriteString(lipgloss.NewStyle().Foreground(styles.ColorTextDim).Render("▌"))
		default:
			sb.WriteString(lipgloss.NewStyle().Foreground(styles.ColorMuted).Render("─"))
		}
	}
	// Pad to barWidth chars.
	for i := len(show); i < barWidth; i++ {
		sb.WriteString(" ")
	}
	return sb.String()
}

func renderLiveRow(r *liveRow) string {
	// 1. Position + change indicator
	posStr := fmt.Sprintf("%2d", r.Pos)
	posDelta := ""
	if r.PrevPos > 0 && r.PrevPos != r.Pos {
		if r.Pos < r.PrevPos {
			posDelta = lipgloss.NewStyle().Foreground(styles.ColorGreen).Render("▲")
		} else {
			posDelta = lipgloss.NewStyle().Foreground(styles.ColorF1Red).Render("▼")
		}
	}
	posCell := lipgloss.NewStyle().Width(colPos).
		Foreground(posColor(r.Pos)).Bold(r.Pos <= 3).
		Render(posStr + posDelta)

	// 2. Driver badge (team colour background) + DRS label
	teamCol := styles.TeamColor(r.TeamName)
	if r.TeamColour != "" {
		teamCol = lipgloss.Color("#" + r.TeamColour)
	}
	badge := lipgloss.NewStyle().
		Background(teamCol).Foreground(lipgloss.Color("#000000")).Bold(true).
		Render(" " + r.Tla + " ") // 5 visual cols
	drsLabel := "   "             // 3 spaces
	if r.DRS >= 10 {
		drsLabel = lipgloss.NewStyle().
			Background(styles.ColorGreen).Foreground(lipgloss.Color("#000000")).Bold(true).
			Render("DRS") // 3 visual cols
	}
	// badge(5) + drsLabel(3) = 8 = colDriver — join then constrain
	driverCell := lipgloss.NewStyle().Width(colDriver).
		Render(lipgloss.JoinHorizontal(lipgloss.Top, badge, drsLabel))

	// 3. Tyre compound + age
	tc := styles.TyreColor(r.Compound)
	tl := styles.TyreLabel(r.Compound)
	tyreStr := ""
	if r.Compound != "" {
		ageStr := fmt.Sprintf("%d", r.TyreAge)
		newPip := ""
		if r.TyreNew {
			newPip = lipgloss.NewStyle().Foreground(styles.ColorGreen).Render("*")
		}
		tyreStr = lipgloss.NewStyle().Foreground(tc).Bold(true).Render("("+tl+")") +
			styles.DimStyle.Render(ageStr) + newPip
	} else {
		tyreStr = styles.DimStyle.Render("  — ")
	}
	tyreCell := lipgloss.NewStyle().Width(colTyre).Render(tyreStr)

	// 4. Lap + pit count
	lapStr := ""
	if r.Laps > 0 {
		lapStr = fmt.Sprintf("L%2d", r.Laps)
	}
	pitStr := ""
	if r.Retired {
		pitStr = lipgloss.NewStyle().Foreground(styles.ColorF1Red).Render("OUT")
	} else if r.InPit {
		pitStr = lipgloss.NewStyle().Foreground(styles.ColorOrange).Render("PIT")
	} else if r.PitCount > 0 {
		pitStr = lipgloss.NewStyle().Foreground(styles.ColorSubtle).Render(fmt.Sprintf("P%d", r.PitCount))
	}
	lapCombined := lapStr
	if pitStr != "" {
		lapCombined += pitStr
	}
	lapCell := lipgloss.NewStyle().Width(colLap).
		Foreground(styles.ColorTextDim).Render(lapCombined)

	// 5. Gap (spring-animated when numeric)
	gapStr := r.GapToLeader
	if r.gapTarget > 0 && r.gapStr == "" {
		gapStr = fmt.Sprintf("+%.3f", r.gapDisplay)
	}
	gapCell := fixedRight(gapStr, colGap, styles.ColorText)

	// 6. Interval
	intColor := styles.ColorTextDim
	if r.IntervalCatching {
		intColor = styles.ColorGreen
	}
	intCell := fixedRight(r.Interval, colInt, intColor)

	// 7. Last lap
	lastColor := styles.ColorText
	if r.LastFastest {
		lastColor = styles.ColorPurple
	} else if r.LastPersonal {
		lastColor = styles.ColorGreen
	}
	lastCell := fixedRight(r.LastLap, colLast, lastColor)

	// 8. Best lap
	bestCell := fixedRight(r.BestLap, colBest, styles.ColorSubtle)

	// 9–11. Sectors with mini-segment bars
	s1Cell := renderSectorCell(r.S1, r.S1Segs, r.S1Fast, r.S1Personal)
	s2Cell := renderSectorCell(r.S2, r.S2Segs, r.S2Fast, r.S2Personal)
	s3Cell := renderSectorCell(r.S3, r.S3Segs, r.S3Fast, r.S3Personal)

	// 12. Speed
	spdCell := fixedRight(r.SpeedST, colSpeed, styles.ColorTextDim)

	return lipgloss.JoinHorizontal(lipgloss.Top,
		posCell, driverCell, tyreCell, lapCell,
		gapCell, intCell, lastCell, bestCell,
		s1Cell, s2Cell, s3Cell, spdCell,
	)
}

// renderSectorCell renders a sector time + mini segment bar in colSector chars.
// Layout: time(7) + " " + bar(4) = 12 = colSector
func renderSectorCell(val string, segs []int, overallFastest, personalFastest bool) string {
	col := sectorColor(overallFastest, personalFastest)
	timeStr := lipgloss.NewStyle().Width(7).Foreground(col).Render(val)
	bar := renderSegBar(segs, overallFastest, personalFastest)
	inner := timeStr + bar // 7 + 4 = 11, padded to colSector(12) below
	return lipgloss.NewStyle().Width(colSector).Render(inner)
}

func sectorColor(fastest, personal bool) lipgloss.Color {
	if fastest {
		return styles.ColorPurple
	}
	if personal {
		return styles.ColorGreen
	}
	return styles.ColorTextDim
}

// ── Fallback (historical) timing ──────────────────────────────────────────────

func (d *Dashboard2) renderFallbackTiming(w, h int) string {
	title := styles.BoldWhite.Render("LAST RACE RESULTS") + "  " +
		styles.DimStyle.Render("(no live session)")

	hdr := d.fbTimingHeader()
	sep := styles.Divider.Render(strings.Repeat("─", w-2))
	lines := []string{title, "", hdr, sep}

	for i, row := range d.fbRows {
		lines = append(lines, d.renderFbRow(row))
		if i >= h-8 {
			break
		}
	}
	content := lipgloss.JoinVertical(lipgloss.Left, lines...)
	return lipgloss.NewStyle().Width(w).Height(h).Render(content)
}

var fbTimingCols = []struct {
	hdr   string
	width int
}{
	{"POS", 4},
	{"DRIVER", 9},
	{"TYR", 5},
	{"LAP", 5},
	{"     GAP", 11},
	{" LAST LAP", 11},
	{" BEST LAP", 11},
	{"    S1", 8},
	{"    S2", 8},
	{"    S3", 8},
	{"PIT", 4},
}

func (d *Dashboard2) fbTimingHeader() string {
	cells := make([]string, len(fbTimingCols))
	for i, col := range fbTimingCols {
		cells[i] = lipgloss.NewStyle().Width(col.width).
			Foreground(styles.ColorSubtle).Bold(true).Render(col.hdr)
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, cells...)
}

func (d *Dashboard2) renderFbRow(row fallbackRow) string {
	posCell := lipgloss.NewStyle().Width(fbTimingCols[0].width).
		Foreground(posColor(row.Pos)).Bold(row.Pos <= 3).
		Render(fmt.Sprintf("%2d", row.Pos))

	teamCol := styles.TeamColor(row.TeamName)
	if row.TeamColor != "" {
		teamCol = lipgloss.Color("#" + row.TeamColor)
	}
	badge := lipgloss.NewStyle().Background(teamCol).
		Foreground(lipgloss.Color("#000000")).Bold(true).
		Render(" " + row.Acronym + " ")
	driverCell := lipgloss.NewStyle().Width(fbTimingCols[1].width).Render(badge)

	tc := styles.TyreColor(row.Compound)
	tl := styles.TyreLabel(row.Compound)
	ageStr := ""
	if row.TyreAge > 0 {
		ageStr = fmt.Sprintf("%d", row.TyreAge)
	}
	tyreCell := lipgloss.NewStyle().Width(fbTimingCols[2].width).Render(
		lipgloss.NewStyle().Foreground(tc).Bold(true).Render(tl) +
			styles.DimStyle.Render(ageStr),
	)

	lapCell := lipgloss.NewStyle().Width(fbTimingCols[3].width).
		Foreground(styles.ColorSubtle).Render(fmt.Sprintf("%3d", row.LapNumber))

	gapCell := fixedRight(row.GapToLeader, fbTimingCols[4].width, styles.ColorText)
	if row.DNF {
		gapCell = lipgloss.NewStyle().Width(fbTimingCols[4].width).Align(lipgloss.Right).
			Foreground(styles.ColorF1Red).Render("DNF")
	}
	lastCell := fixedRight(formatDuration(row.LastLap), fbTimingCols[5].width, styles.ColorText)
	bestCell := fixedRight(formatDuration(row.BestLap), fbTimingCols[6].width, styles.ColorSubtle)
	s1Cell := fixedRight(formatSector(row.Sector1), fbTimingCols[7].width, styles.ColorTextDim)
	s2Cell := fixedRight(formatSector(row.Sector2), fbTimingCols[8].width, styles.ColorTextDim)
	s3Cell := fixedRight(formatSector(row.Sector3), fbTimingCols[9].width, styles.ColorTextDim)

	pitStr := ""
	if row.PitCount > 0 {
		pitStr = fmt.Sprintf("P%d", row.PitCount)
	}
	pitCell := lipgloss.NewStyle().Width(fbTimingCols[10].width).
		Foreground(styles.ColorOrange).Render(pitStr)

	return lipgloss.JoinHorizontal(lipgloss.Top,
		posCell, driverCell, tyreCell, lapCell,
		gapCell, lastCell, bestCell, s1Cell, s2Cell, s3Cell, pitCell,
	)
}
