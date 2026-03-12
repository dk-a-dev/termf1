package schedule

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/dk-a-dev/termf1/internal/api/jolpica"
	"github.com/dk-a-dev/termf1/internal/ui/styles"
)

// RenderScheduleCards formats all race cards grouped by month.
func RenderScheduleCards(races []jolpica.Race, now time.Time, width int, cursor int) string {
	cardW := width - 4
	if cardW < 40 {
		cardW = 40
	}

	type monthGroup struct {
		month string
		cards []string
	}

	groups := map[string]*monthGroup{}
	order := []string{}

	for i, race := range races {
		raceTime := ParseDate(race.Date, race.Time)
		monthKey := "Unknown"
		if !raceTime.IsZero() {
			monthKey = raceTime.Format("January 2006")
		}

		if _, exists := groups[monthKey]; !exists {
			groups[monthKey] = &monthGroup{month: monthKey}
			order = append(order, monthKey)
		}
		
		isSelected := i == cursor && cursor >= 0
		groups[monthKey].cards = append(groups[monthKey].cards, RenderRaceCard(i, race, now, cardW, isSelected, IsNextRaceIdx(i, races, now)))
	}

	var sb strings.Builder
	for _, month := range order {
		g := groups[month]
		monthHdr := lipgloss.NewStyle().
			Bold(true).
			Foreground(styles.ColorYellow).
			Background(lipgloss.Color("#1a1a1a")).
			Padding(0, 2).
			Width(width).
			Render("  " + g.month)
		sb.WriteString(monthHdr + "\n")
		for _, card := range g.cards {
			sb.WriteString(card + "\n")
		}
	}
	return sb.String()
}

func RenderRaceCard(idx int, race jolpica.Race, now time.Time, w int, isSelected, isNext bool) string {
	raceTime := ParseDate(race.Date, race.Time)
	isPast := !raceTime.IsZero() && raceTime.Before(now)

	var statusStr string
	switch {
	case isPast:
		statusStr = lipgloss.NewStyle().Foreground(styles.ColorMuted).Render("  ✓ PAST")
	case isNext:
		statusStr = lipgloss.NewStyle().Foreground(styles.ColorF1Red).Bold(true).Render("  ► NEXT RACE")
	default:
		statusStr = lipgloss.NewStyle().Foreground(styles.ColorSubtle).Render("  upcoming")
	}

	roundStr := lipgloss.NewStyle().Foreground(styles.ColorSubtle).Render(fmt.Sprintf("R%s", race.Round))
	nameStr := lipgloss.NewStyle().Bold(true).Foreground(styles.ColorF1White).Render(race.RaceName)
	headerLine := roundStr + "  " + nameStr + statusStr

	locStr := styles.DimStyle.Render(fmt.Sprintf("  📍 %s  ·  %s, %s",
		race.Circuit.CircuitName, race.Circuit.Location.Locality, race.Circuit.Location.Country))

	dateStr := ""
	if !raceTime.IsZero() {
		switch {
		case isNext:
			dateStr = lipgloss.NewStyle().Foreground(styles.ColorYellow).
				Render(fmt.Sprintf("  📅 %s  (%s)",
					raceTime.Format("Mon 2 Jan 2006 15:04 UTC"), CountdownStr(raceTime, now)))
		case isPast:
			dateStr = styles.DimStyle.Render("  📅 " + raceTime.Format("Mon 2 Jan 2006"))
		default:
			dateStr = lipgloss.NewStyle().Foreground(styles.ColorTextDim).
				Render("  📅 " + raceTime.Format("Mon 2 Jan 2006 15:04 UTC"))
		}
	}

	lines := []string{headerLine, locStr, dateStr}

	// Show session detail when this card is selected or is the next race
	if isSelected || isNext {
		sessLines := SessionDetailLines(race, now)
		if sessLines != "" {
			lines = append(lines, "")
			lines = append(lines, strings.Split(sessLines, "\n")...)
		}
	}

	body := strings.Join(lines, "\n")
	if isPast {
		body = lipgloss.NewStyle().Foreground(styles.ColorSubtle).Render(body)
	}

	borderCol := styles.ColorBorder
	switch {
	case isNext:
		borderCol = styles.ColorF1Red
	case isSelected && !isPast:
		borderCol = styles.ColorBlue
	case isPast:
		borderCol = styles.ColorMuted
	}

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderCol).
		Padding(0, 1).
		Width(w).
		Render(body)
}
