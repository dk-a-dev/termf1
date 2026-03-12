package schedule

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/dk-a-dev/termf1/v2/internal/api/jolpica"
	"github.com/dk-a-dev/termf1/v2/internal/ui/styles"
)

func SessionDetailLines(race jolpica.Race, now time.Time) string {
	type sessEntry struct {
		label string
		s     *jolpica.SessionTime
	}
	entries := []sessEntry{
		{"FP1      ", race.FirstPractice},
		{"FP2      ", race.SecondPractice},
		{"FP3      ", race.ThirdPractice},
		{"Sprint Q ", race.SprintQualifying},
		{"Sprint   ", race.Sprint},
		{"Qualifying", race.Qualifying},
	}
	var parts []string
	for _, e := range entries {
		if e.s == nil {
			continue
		}
		t := ParseDate(e.s.Date, e.s.Time)
		if t.IsZero() {
			continue
		}
		dot := "○"
		col := styles.ColorSubtle
		if t.Before(now) {
			dot = "●"
			col = styles.ColorMuted
		}
		parts = append(parts, lipgloss.NewStyle().Foreground(col).Render(
			fmt.Sprintf("    %s %-10s  %s", dot, e.label, t.Format("Mon 2 Jan  15:04 UTC")),
		))
	}
	return strings.Join(parts, "\n")
}

func ParseDate(date, timeStr string) time.Time {
	if date == "" {
		return time.Time{}
	}
	combined := date
	if timeStr != "" && !strings.Contains(date, "T") {
		combined = date + "T" + timeStr
	}
	formats := []string{
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
		"2006-01-02",
	}
	for _, f := range formats {
		if t, err := time.Parse(f, combined); err == nil {
			return t
		}
	}
	if t, err := time.Parse("2006-01-02", date); err == nil {
		return t
	}
	return time.Time{}
}

func IsNextRaceIdx(idx int, races []jolpica.Race, now time.Time) bool {
	for i, r := range races {
		t := ParseDate(r.Date, r.Time)
		if !t.IsZero() && t.After(now) {
			return i == idx
		}
	}
	return false
}

func NextRaceIdx(races []jolpica.Race, now time.Time) int {
	for i, r := range races {
		t := ParseDate(r.Date, r.Time)
		if !t.IsZero() && t.After(now) {
			return i
		}
	}
	return 0
}

func CountdownStr(future, now time.Time) string {
	d := future.Sub(now)
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	if days > 0 {
		return fmt.Sprintf("in %dd %dh", days, hours)
	}
	return fmt.Sprintf("in %dh %dm", int(d.Hours()), int(d.Minutes())%60)
}

func RaceYear(races []jolpica.Race) string {
	if len(races) == 0 {
		return ""
	}
	return races[0].Season
}
