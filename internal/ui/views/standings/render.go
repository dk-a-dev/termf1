package standings

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/dk-a-dev/termf1/v2/internal/api/jolpica"
	"github.com/dk-a-dev/termf1/v2/internal/ui/styles"
)

// RenderStandings formats the driver and constructor standings into a single string.
func RenderStandings(drivers []jolpica.DriverStanding, constructors []jolpica.ConstructorStanding, width int) string {
	half := width / 2
	if half < 30 {
		return lipgloss.JoinVertical(lipgloss.Left,
			RenderDriverStandingsTable(drivers, width),
			"",
			RenderConstructorStandingsTable(constructors, width),
		)
	}

	left := RenderDriverStandingsTable(drivers, half-1)
	right := RenderConstructorStandingsTable(constructors, width-half-1)
	return lipgloss.JoinHorizontal(lipgloss.Top, left, " ", right)
}

func RenderDriverStandingsTable(drivers []jolpica.DriverStanding, w int) string {
	title := styles.Title.Render("  Driver Championship")
	sep := styles.Divider.Render(strings.Repeat("─", w))

	maxPts := 1.0
	for _, d := range drivers {
		if p, _ := strconv.ParseFloat(d.Points, 64); p > maxPts {
			maxPts = p
		}
	}

	barW := w - 28
	if barW < 4 {
		barW = 4
	}

	lines := []string{title, sep}
	for _, d := range drivers {
		pts, _ := strconv.ParseFloat(d.Points, 64)
		pos, _ := strconv.Atoi(d.Position)

		posStr := fmt.Sprintf("%2d", pos)
		code := d.Driver.Code
		if code == "" {
			code = d.Driver.FamilyName
			if len(code) > 3 {
				code = code[:3]
			}
		}

		teamCol := styles.ColorSubtle
		if len(d.Constructors) > 0 {
			teamCol = styles.TeamColor(d.Constructors[0].Name)
		}

		bar := styles.BarChart(pts, maxPts, barW, teamCol)
		ptsStr := fmt.Sprintf("%4.0f", pts)

		posC := lipgloss.NewStyle().Width(3).Foreground(posColor(pos)).Bold(pos <= 3).Render(posStr)
		codeC := lipgloss.NewStyle().Width(5).Bold(true).Foreground(styles.ColorText).Render(code)
		ptsC := lipgloss.NewStyle().Width(5).Align(lipgloss.Right).Foreground(styles.ColorYellow).Render(ptsStr)

		line := posC + codeC + bar + " " + ptsC
		lines = append(lines, line)
	}
	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func RenderConstructorStandingsTable(constructors []jolpica.ConstructorStanding, w int) string {
	title := styles.Title.Render("  Constructor Championship")
	sep := styles.Divider.Render(strings.Repeat("─", w))

	maxPts := 1.0
	for _, c := range constructors {
		if p, _ := strconv.ParseFloat(c.Points, 64); p > maxPts {
			maxPts = p
		}
	}

	barW := w - 28
	if barW < 4 {
		barW = 4
	}

	lines := []string{title, sep}
	for _, c := range constructors {
		pts, _ := strconv.ParseFloat(c.Points, 64)
		pos, _ := strconv.Atoi(c.Position)

		name := c.Constructor.Name
		if len(name) > 12 {
			name = name[:12]
		}

		teamCol := styles.TeamColor(c.Constructor.Name)
		bar := styles.BarChart(pts, maxPts, barW, teamCol)
		ptsStr := fmt.Sprintf("%4.0f", pts)

		posC := lipgloss.NewStyle().Width(3).Foreground(posColor(pos)).Bold(pos <= 3).Render(fmt.Sprintf("%2d", pos))
		nameC := lipgloss.NewStyle().Width(14).Bold(true).Foreground(styles.ColorText).Render(name)
		ptsC := lipgloss.NewStyle().Width(5).Align(lipgloss.Right).Foreground(styles.ColorYellow).Render(ptsStr)

		line := posC + nameC + bar + " " + ptsC
		lines = append(lines, line)
	}
	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func posColor(pos int) lipgloss.Color {
	switch pos {
	case 1:
		return lipgloss.Color("#FFD700")
	case 2:
		return lipgloss.Color("#C0C0C0")
	case 3:
		return lipgloss.Color("#CD7F32")
	default:
		return lipgloss.Color("#666666")
	}
}
