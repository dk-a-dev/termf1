package dashboard

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/dk-a-dev/termf1/v2/internal/ui/styles"
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
// Global col widths for the live timing panel.
// Total: 4+8+11+9+9+6+6+6+6+8+4 = 77+ cols.
const (
	colPos     = 3
	colDriver  = 8
	colLapTime = 11
	colGap     = 9
	colInt     = 9
	colS1      = 6
	colS2      = 6
	colS3      = 6
	colChg     = 6
	colTyre    = 8
	colPit     = 4
)

func liveTimingHeader() string {
	h := lipgloss.NewStyle().
		Bold(true).
		Foreground(styles.ColorTextDim).
		Background(styles.ColorBgHeader).
		Padding(0, 1)

	return lipgloss.JoinHorizontal(lipgloss.Top,
		h.Width(colPos).Render("POS"),
		h.Width(colDriver).Render("DRIVER"),
		h.Width(colLapTime).Render("LAP TIME"),
		h.Width(colGap).Render("GAP"),
		h.Width(colInt).Render("INTERVAL"),
		h.Width(colS1).Align(lipgloss.Center).Render("S1"),
		h.Width(colS2).Align(lipgloss.Center).Render("S2"),
		h.Width(colS3).Align(lipgloss.Center).Render("S3"),
		h.Width(colChg).Align(lipgloss.Center).Render("CHG"),
		h.Width(colTyre).Render("TYRE"),
		h.Width(colPit).Render("PIT"),
	)
}



func renderLiveRow(r *liveRow) string {
	// 1. Position Rank
	posColor := styles.ColorText
	if r.Pos <= 3 && r.Pos > 0 {
		posColor = styles.ColorYellow
	}
	posCell := lipgloss.NewStyle().Width(colPos).
		Foreground(posColor).Bold(r.Pos <= 3).
		Render(fmt.Sprintf("%2d", r.Pos))

	// 2. Driver badge
	teamCol := styles.TeamColor(r.TeamName)
	if r.TeamColour != "" {
		teamCol = lipgloss.Color("#" + r.TeamColour)
	}
	displayTla := r.Tla
	if displayTla == "" {
		displayTla = "???"
	}
	badge := lipgloss.NewStyle().
		Background(teamCol).Foreground(lipgloss.Color("#000000")).Bold(true).
		Render(" " + fmt.Sprintf("%-3s", displayTla) + " ") // 5 cols
	drsLabel := "   "
	if r.DRS >= 10 {
		drsLabel = lipgloss.NewStyle().
			Background(styles.ColorGreen).Foreground(lipgloss.Color("#000000")).Bold(true).
			Render("DRS") // 3 cols
	}
	driverCell := lipgloss.NewStyle().Width(colDriver).
		Render(lipgloss.JoinHorizontal(lipgloss.Top, badge, drsLabel))

	// 3. Lap Time
	ltColor := styles.ColorText
	if r.LastFastest {
		ltColor = styles.ColorPurple
	} else if r.LastPersonal {
		ltColor = styles.ColorGreen
	}
	lapTimeCell := lipgloss.NewStyle().Width(colLapTime).
		Foreground(ltColor).
		Render(r.LastLap)

	// 4. Gap & Interval
	gapCell := lipgloss.NewStyle().Width(colGap).
		Foreground(styles.ColorTextDim).
		Render(r.GapDisplay())

	intColor := styles.ColorTextDim
	if r.IntervalCatching {
		intColor = styles.ColorGreen
	}
	intCell := lipgloss.NewStyle().Width(colInt).
		Foreground(intColor).
		Render(r.Interval)

	// 5. Sectors (Status Bars)
	s1 := renderSectorStatus(r.S1Segs, r.S1Fast, r.S1Personal)
	s2 := renderSectorStatus(r.S2Segs, r.S2Fast, r.S2Personal)
	s3 := renderSectorStatus(r.S3Segs, r.S3Fast, r.S3Personal)

	// 6. Position Change (CHG)
	chgColor := styles.ColorTextDim
	if strings.Contains(r.PosDelta, "↑") {
		chgColor = styles.ColorGreen
	} else if strings.Contains(r.PosDelta, "↓") {
		chgColor = styles.ColorF1Red
	}
	chgCell := lipgloss.NewStyle().Width(colChg).
		Align(lipgloss.Center).
		Foreground(chgColor).
		Render(r.PosDelta)

	// 7. Tyre (Age + Compound)
	tc := styles.TyreColor(r.Compound)
	ageStr := fmt.Sprintf("%2d", r.TyreAge)
	compBadge := " "
	if r.Compound != "" {
		label := r.Compound
		if len(label) > 1 {
			label = label[:1]
		}
		compBadge = lipgloss.NewStyle().
			Bold(true).Padding(0, 1).
			Foreground(lipgloss.Color("#000000")).
			Background(tc).
			Render(label)
	}
	tyreCell := lipgloss.NewStyle().Width(colTyre).
		Render(ageStr + " " + compBadge)

	// 8. Pit
	pitStr := ""
	if r.PitCount > 0 {
		pitStr = fmt.Sprintf("%d", r.PitCount)
	}
	pitCell := lipgloss.NewStyle().Width(colPit).
		Align(lipgloss.Center).
		Foreground(styles.ColorTextDim).
		Render(pitStr)

	return lipgloss.JoinHorizontal(lipgloss.Top,
		posCell,
		driverCell,
		lapTimeCell,
		gapCell,
		intCell,
		lipgloss.NewStyle().Width(colS1).Render(s1),
		lipgloss.NewStyle().Width(colS2).Render(s2),
		lipgloss.NewStyle().Width(colS3).Render(s3),
		chgCell,
		tyreCell,
		pitCell,
	)
}

func renderSectorStatus(segs []int, overall, personal bool) string {
	if len(segs) == 0 {
		if overall {
			return lipgloss.NewStyle().Foreground(styles.ColorPurple).Render("▬▬▬")
		}
		if personal {
			return lipgloss.NewStyle().Foreground(styles.ColorGreen).Render("▬▬▬")
		}
		return styles.DimStyle.Render("────")
	}
	// Pick the strongest status from segments
	status := 0
	for _, s := range segs {
		if s > status {
			status = s
		}
	}
	
	// Professional horizontal bar look: ▬▬▬
	char := "▬▬▬"
	col := styles.ColorTextDim
	switch status {
	case 2048: // Personal Best
		col = styles.ColorGreen
	case 2049, 2051: // Overall Fastest
		col = styles.ColorPurple
	case 2052: // Yellow flag in segment
		col = styles.ColorYellow
	case 0: // In progress
		col = styles.ColorTextDim
		char = "..."
	}
	
	return lipgloss.NewStyle().Foreground(col).Render(char)
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
