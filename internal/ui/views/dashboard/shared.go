// Package views contains all TUI view models.
// This file holds shared formatting helpers and the OpenF1 DriverRow type
// used by both the historical fallback panel and other views.
package dashboard

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/dk-a-dev/termf1/v2/internal/api/openf1"
	"github.com/dk-a-dev/termf1/v2/internal/ui/styles"
)

// ── DriverRow: processed view-model for one car (OpenF1 data) ────────────────

type DriverRow struct {
	Position    int
	Number      int
	Acronym     string
	TeamName    string
	TeamColor   string
	GapToLeader string
	Interval    string
	LastLap     float64
	BestLap     float64
	Sector1     float64
	Sector2     float64
	Sector3     float64
	SpeedTrap   int
	Compound    string
	TyreAge     int
	PitCount    int
	LapNumber   int
	IsPitOutLap bool
	DNF         bool
}

// ── Data helpers ──────────────────────────────────────────────────────────────

// buildRows builds sorted DriverRow slice from an OpenF1 DashboardPayload.
func buildRows(p *openf1.DashboardPayload) []DriverRow {
	if p == nil {
		return nil
	}

	driverMap := make(map[int]openf1.Driver, len(p.Drivers))
	for _, d := range p.Drivers {
		driverMap[d.DriverNumber] = d
	}

	// session_result is the authoritative source: position, gap, laps, DNF/DNS/DSQ.
	type resultEntry struct {
		position    int
		gapToLeader interface{}
		numberOfLaps int
		dnf, dns, dsq bool
	}
	resultMap := make(map[int]resultEntry, len(p.SessionResults))
	for _, sr := range p.SessionResults {
		resultMap[sr.DriverNumber] = resultEntry{
			position:     sr.Position,
			gapToLeader:  sr.GapToLeader,
			numberOfLaps: sr.NumberOfLaps,
			dnf:          sr.DNF,
			dns:          sr.DNS,
			dsq:          sr.DSQ,
		}
	}

	// Keep only the last interval sample per driver as fallback gap source.
	gapMap := make(map[int]openf1.Interval)
	for _, iv := range p.Intervals {
		gapMap[iv.DriverNumber] = iv
	}

	bestLapMap := make(map[int]float64)
	lastLapMap := make(map[int]openf1.Lap)
	lapNumMap := make(map[int]int)
	for _, l := range p.Laps {
		if l.LapDuration > 0 {
			if cur, ok := bestLapMap[l.DriverNumber]; !ok || l.LapDuration < cur {
				bestLapMap[l.DriverNumber] = l.LapDuration
			}
		}
		if l.LapNumber > lapNumMap[l.DriverNumber] {
			lapNumMap[l.DriverNumber] = l.LapNumber
			lastLapMap[l.DriverNumber] = l
		}
	}

	stintMap := make(map[int]openf1.Stint)
	for _, st := range p.Stints {
		if cur, ok := stintMap[st.DriverNumber]; !ok || st.StintNumber > cur.StintNumber {
			stintMap[st.DriverNumber] = st
		}
	}

	pitCount := make(map[int]int)
	for _, pit := range p.Pits {
		pitCount[pit.DriverNumber]++
	}

	rows := make([]DriverRow, 0, len(driverMap))
	for num, drv := range driverMap {
		re := resultMap[num]
		iv := gapMap[num]
		lastLap := lastLapMap[num]
		stint := stintMap[num]

		// Determine gap string: prefer session_result, fall back to intervals.
		gapStr := formatGap(re.gapToLeader)
		if gapStr == "" || gapStr == "–" {
			gapStr = formatGap(iv.GapToLeader)
		}

		// Tyre age: stint age at start + laps since stint began.
		tyreAge := stint.TyreAgeAtStart
		lapsThisSession := lapNumMap[num]
		if re.numberOfLaps > 0 {
			lapsThisSession = re.numberOfLaps
		}
		if lapsThisSession > 0 && stint.LapStart > 0 {
			tyreAge = stint.TyreAgeAtStart + (lapsThisSession - stint.LapStart + 1)
		}

		row := DriverRow{
			Position:    re.position,
			Number:      num,
			Acronym:     drv.NameAcronym,
			TeamName:    drv.TeamName,
			TeamColor:   drv.TeamColour,
			GapToLeader: gapStr,
			Interval:    formatGap(iv.Interval),
			LastLap:     lastLap.LapDuration,
			BestLap:     bestLapMap[num],
			Sector1:     lastLap.DurationSector1,
			Sector2:     lastLap.DurationSector2,
			Sector3:     lastLap.DurationSector3,
			SpeedTrap:   lastLap.StSpeed,
			Compound:    stint.Compound,
			TyreAge:     tyreAge,
			PitCount:    pitCount[num],
			LapNumber:   lapsThisSession,
			IsPitOutLap: lastLap.IsPitOutLap,
			DNF:         re.dnf || re.dns || re.dsq || re.position == 0,
		}
		rows = append(rows, row)
	}

	// Assign positions:
	// 1. Classified finishers — use session_result.position (already set).
	// 2. Unclassified DNF/DNS/DSQ (position=0) — place after classified,
	//    ordered by laps completed descending (more laps = higher placed).
	// 3. Drivers absent from session_result entirely (position=0, DNF=false)
	//    — place last.
	allMissing := true
	maxClassified := 0
	for _, r := range rows {
		if r.Position > 0 {
			allMissing = false
			if r.Position > maxClassified {
				maxClassified = r.Position
			}
		}
	}

	if allMissing && len(rows) > 0 {
		assignPositionsFromIntervals(rows, gapMap)
	} else {
		// Collect all unclassified drivers (position=0) and sort by laps desc.
		// DNF flag is for display only — don't use it for ordering.
		type unclassified struct {
			idx  int
			laps int
		}
		var unc []unclassified
		for i, r := range rows {
			if r.Position == 0 {
				unc = append(unc, unclassified{idx: i, laps: r.LapNumber})
			}
		}
		sort.Slice(unc, func(i, j int) bool {
			if unc[i].laps != unc[j].laps {
				return unc[i].laps > unc[j].laps
			}
			return rows[unc[i].idx].Number < rows[unc[j].idx].Number
		})
		rank := maxClassified + 1
		for _, u := range unc {
			rows[u.idx].Position = rank
			rank++
		}
	}

	sort.Slice(rows, func(i, j int) bool {
		if rows[i].Position == rows[j].Position {
			return rows[i].Number < rows[j].Number
		}
		return rows[i].Position < rows[j].Position
	})
	return rows
}

// assignPositionsFromIntervals ranks rows 1–N using their last known gap to
// leader when the /position endpoint returned nothing.
//
// Ranking rules:
//  1. Leader: gap is nil, 0.0, or empty string
//  2. Drivers with a numeric gap, sorted ascending
//  3. Lapped drivers ("+N LAP"), sorted by lap count then numeric gap
func assignPositionsFromIntervals(rows []DriverRow, gapMap map[int]openf1.Interval) {
	type entry struct {
		idx     int
		gapF    float64  // numeric gap (0 = leader)
		lapped  bool
		lapDiff int      // number of laps behind
	}
	entries := make([]entry, len(rows))
	for i, r := range rows {
		e := entry{idx: i}
		raw := gapMap[r.Number].GapToLeader
		switch v := raw.(type) {
		case nil:
			// leader
		case float64:
			e.gapF = v
		case string:
			if v == "" {
				// leader
			} else if strings.HasPrefix(v, "+") && strings.Contains(v, "LAP") {
				e.lapped = true
				fmt.Sscanf(v, "+%d", &e.lapDiff)
			} else {
				fmt.Sscanf(strings.TrimPrefix(v, "+"), "%f", &e.gapF)
			}
		}
		entries[i] = e
	}
	sort.Slice(entries, func(i, j int) bool {
		a, b := entries[i], entries[j]
		if a.lapped != b.lapped {
			return !a.lapped // on-lead-lap first
		}
		if a.lapped {
			if a.lapDiff != b.lapDiff {
				return a.lapDiff < b.lapDiff
			}
			return a.gapF < b.gapF
		}
		return a.gapF < b.gapF
	})
	for rank, e := range entries {
		rows[e.idx].Position = rank + 1
	}
}

// lastRCMessages returns the last n race control messages in reverse order
// (newest first).
func lastRCMessages(rcs []openf1.RaceControl, n int) []string {
	if len(rcs) == 0 {
		return nil
	}
	start := len(rcs) - n
	if start < 0 {
		start = 0
	}
	out := make([]string, 0, n)
	for _, rc := range rcs[start:] {
		out = append(out, rc.Message)
	}
	for i, j := 0, len(out)-1; i < j; i, j = i+1, j-1 {
		out[i], out[j] = out[j], out[i]
	}
	return out
}

// ── Formatting helpers ────────────────────────────────────────────────────────

func formatGap(v interface{}) string {
	if v == nil {
		return "LEADER"
	}
	switch val := v.(type) {
	case float64:
		if val == 0 {
			return "LEADER"
		}
		return fmt.Sprintf("+%.3f", val)
	case string:
		return val
	}
	return "—"
}

func formatDuration(secs float64) string {
	if secs <= 0 {
		return "—"
	}
	mins := int(secs) / 60
	rem := secs - float64(mins*60)
	if mins > 0 {
		return fmt.Sprintf("%d:%06.3f", mins, rem)
	}
	return fmt.Sprintf("%.3f", secs)
}

func formatSector(secs float64) string {
	if secs <= 0 {
		return "—"
	}
	return fmt.Sprintf("%.3f", secs)
}

func formatLap(last, best float64) (string, lipgloss.Color) {
	s := formatDuration(last)
	if last <= 0 {
		return s, styles.ColorSubtle
	}
	if best > 0 && math.Abs(last-best) < 0.001 {
		return s, styles.ColorPurple
	}
	return s, styles.ColorText
}

func posColor(pos int) lipgloss.Color {
	switch pos {
	case 1:
		return styles.ColorYellow
	case 2:
		return lipgloss.Color("#C0C0C0")
	case 3:
		return lipgloss.Color("#CD7F32")
	default:
		return styles.ColorText
	}
}

func fixedRight(s string, width int, color lipgloss.Color) string {
	return lipgloss.NewStyle().Width(width).Align(lipgloss.Right).Foreground(color).Render(s)
}

func wrapText(s string, width int) string {
	if width <= 0 || len(s) <= width {
		return s
	}
	return s[:width-1] + "…"
}

func wordWrapMulti(s string, width int) string {
	if width <= 0 || len(s) <= width {
		return s
	}
	var lines []string
	words := strings.Fields(s)
	var current strings.Builder
	for _, w := range words {
		if current.Len() > 0 && current.Len()+1+len(w) > width {
			lines = append(lines, current.String())
			current.Reset()
		}
		if current.Len() > 0 {
			current.WriteByte(' ')
		}
		current.WriteString(w)
	}
	if current.Len() > 0 {
		lines = append(lines, current.String())
	}
	return strings.Join(lines, "\n")
}

