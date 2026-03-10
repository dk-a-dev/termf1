package analysis

import (
	"fmt"
	"math"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/devkeshwani/termf1/internal/api/openf1"
	"github.com/devkeshwani/termf1/internal/ui/styles"
	"github.com/devkeshwani/termf1/internal/ui/views/common"
)

// ── Chart dispatcher ──────────────────────────────────────────────────────────

func (a *Analysis) renderChart(w, h int) string {
	if len(a.laps) == 0 && len(a.stints) == 0 {
		return common.Centred(w, h, styles.DimStyle.Render("No data available for this session"))
	}
	switch a.chart {
	case ChartStrategy:
		return a.renderStrategy(w, h)
	case ChartLapLine:
		return a.renderLapLineChart(w, h)
	case ChartLapSpark:
		return a.renderLapSparklines(w, h)
	case ChartPace:
		return a.renderPaceChart(w, h)
	case ChartSectors:
		return a.renderSectorChart(w, h)
	case ChartSpeed:
		return a.renderSpeedChart(w, h)
	case ChartPositions:
		return a.renderPositionChart(w, h)
	case ChartTeamPace:
		return a.renderTeamPaceChart(w, h)
	}
	return ""
}

// ── Shared helpers ────────────────────────────────────────────────────────────

// safeRep is strings.Repeat guarded against negative n.
func safeRep(s string, n int) string {
	if n <= 0 {
		return ""
	}
	return strings.Repeat(s, n)
}

// formatLapSec converts a float seconds value to "M:SS.mmm".
func formatLapSec(sec float64) string {
	if sec <= 0 {
		return "–"
	}
	m := int(sec) / 60
	s := sec - float64(m*60)
	return fmt.Sprintf("%d:%06.3f", m, s)
}

// noInf replaces MaxFloat64 with 0 for summation safety.
func noInf(v float64) float64 {
	if v == math.MaxFloat64 {
		return 0
	}
	return v
}

// sectionTitle renders a bold section heading with a dim subtitle.
func sectionTitle(title, sub string) string {
	return styles.BoldWhite.Render("  "+title) + "   " + styles.DimStyle.Render(sub)
}

// hBar draws a horizontal bar of `filled` chars followed by `empty` chars.
func hBar(filled, empty string, barW, maxW int, col, emptyCol lipgloss.Color) string {
	if barW < 0 {
		barW = 0
	}
	if barW > maxW {
		barW = maxW
	}
	return lipgloss.NewStyle().Foreground(col).Render(safeRep(filled, barW)) +
		lipgloss.NewStyle().Foreground(emptyCol).Render(safeRep(empty, maxW-barW))
}

// compoundColor maps a tyre compound string to its canonical colour.
// Handles full names (SOFT/MEDIUM/HARD/INTERMEDIATE/WET) and short codes (S/M/H/I/W).
func compoundColor(s string) lipgloss.Color {
	switch strings.ToUpper(strings.TrimSpace(s)) {
	case "SOFT", "S":
		return styles.ColorTyreSoft
	case "MEDIUM", "M":
		return styles.ColorTyreMedium
	case "HARD", "H":
		return lipgloss.Color("#FFFFFF") // true white for HARD
	case "INTERMEDIATE", "I", "INTER":
		return styles.ColorTyreInter
	case "WET", "W":
		return styles.ColorTyreWet
	default:
		return lipgloss.Color("#888888") // mid-grey for truly unknown
	}
}

// compoundBlock returns a single █ in the compound's colour.
func compoundBlock(c string) string {
	return lipgloss.NewStyle().Foreground(compoundColor(c)).Render("█")
}

// driverMap builds a DriverNumber → Driver lookup from the loaded drivers slice.
func (a *Analysis) driverMap() map[int]openf1.Driver {
	m := make(map[int]openf1.Driver, len(a.drivers))
	for _, d := range a.drivers {
		m[d.DriverNumber] = d
	}
	return m
}

// dotChars is the per-driver shape palette used on the 2D line chart.
var dotChars = []rune{'●', '◆', '▲', '■', '◉', '★', '◈', '◯', '▷', '◻', '○', '▪', '△', '☆', '◇'}

// sparkBlocks are the 8-level Unicode block chars used in sparklines.
var sparkBlocks = []rune{'▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}
