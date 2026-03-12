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

// ── Chart 9: Pit Stop Duration ────────────────────────────────────────────────

// retirementThreshold: stops longer than this are treated as pit lane retirements,
// not normal service stops (e.g. parked in garage, VSC red flag stays).
const retirementThreshold = 120.0

// pitDurTierColor returns a colour that maps duration tiers fast→slow.
func pitDurTierColor(dur float64) lipgloss.Color {
	switch {
	case dur < 19.0:
		return styles.ColorGreen              // fast service
	case dur < 22.0:
		return lipgloss.Color("#A3E635")      // lime – within normal range
	case dur < 28.0:
		return styles.ColorOrange             // noticeably slow
	default:
		return styles.ColorF1Red              // critical / very slow
	}
}

// renderPitStopChart draws a ranked bar chart of all individual pit stop
// durations (fastest → slowest). Pit-lane retirements (>120 s) are excluded
// from the scale and listed separately at the bottom.
func (a *Analysis) renderPitStopChart(w, h int) string {
	if len(a.pits) == 0 {
		return common.Centred(w, h, styles.DimStyle.Render("No pit stop data for this session"))
	}

	drMap := a.driverMap()

	// Build stints lookup: (driverNum, lapStart) → compound.
	// A pit on lap N means the fresh tyre stint starts on lap N+1.
	type stintKey struct{ driver, lap int }
	stintCompound := make(map[stintKey]string)
	for _, st := range a.stints {
		stintCompound[stintKey{st.DriverNumber, st.LapStart}] = st.Compound
	}

	type stopInfo struct {
		driverNum  int
		acronym    string
		teamCol    lipgloss.Color
		lap        int
		duration   float64
		compound   string // tyre fitted (may be empty if stints data missing)
		stopNum    int    // 1st, 2nd, 3rd stop for this driver
		totalStops int    // total pit stops by this driver (excl. retirements)
	}

	// Pre-pass: count how many valid stops each driver has (excl. retirements),
	// and assign a chronological stop number based on lap order.
	driverStopCount := make(map[int]int)         // driverNum → total normal stop count
	driverStopNum   := make(map[[2]int]int)       // [driverNum, lap] → stop #
	// First, tally normal stops per driver in lap order.
	type rawStop struct {
		driverNum int
		lap       int
		duration  float64
	}
	var rawAll []rawStop
	for _, p := range a.pits {
		dur := p.LaneDuration
		if dur == 0 && p.PitDuration != nil {
			dur = *p.PitDuration
		}
		if dur == 0 || dur > retirementThreshold {
			continue
		}
		rawAll = append(rawAll, rawStop{p.DriverNumber, p.LapNumber, dur})
	}
	sort.Slice(rawAll, func(i, j int) bool {
		if rawAll[i].driverNum != rawAll[j].driverNum {
			return rawAll[i].driverNum < rawAll[j].driverNum
		}
		return rawAll[i].lap < rawAll[j].lap
	})
	seqCounter := make(map[int]int)
	for _, r := range rawAll {
		driverStopCount[r.driverNum]++
		seqCounter[r.driverNum]++
		driverStopNum[[2]int{r.driverNum, r.lap}] = seqCounter[r.driverNum]
	}

	var normal []stopInfo   // genuine pit stops ≤ retirementThreshold
	var retired []stopInfo  // parked in pit lane / very long stays

	for _, p := range a.pits {
		dur := p.LaneDuration
		if dur == 0 && p.PitDuration != nil {
			dur = *p.PitDuration
		}
		if dur == 0 {
			continue
		}

		drv := drMap[p.DriverNumber]
		acronym := drv.NameAcronym
		if acronym == "" {
			acronym = fmt.Sprintf("%d", p.DriverNumber)
		}
		col := lipgloss.Color("#888888")
		if drv.TeamColour != "" {
			col = lipgloss.Color("#" + drv.TeamColour)
		}
		compound := stintCompound[stintKey{p.DriverNumber, p.LapNumber + 1}]

		stopN := driverStopNum[[2]int{p.DriverNumber, p.LapNumber}]
		si := stopInfo{
			driverNum:  p.DriverNumber,
			acronym:    acronym,
			teamCol:    col,
			lap:        p.LapNumber,
			duration:   dur,
			compound:   compound,
			stopNum:    stopN,
			totalStops: driverStopCount[p.DriverNumber],
		}
		if dur > retirementThreshold {
			retired = append(retired, si)
		} else {
			normal = append(normal, si)
		}
	}

	if len(normal) == 0 && len(retired) == 0 {
		return common.Centred(w, h, styles.DimStyle.Render("No pit stop duration data available"))
	}

	// Sort normal stops fastest → slowest for the ranked view.
	sort.Slice(normal, func(i, j int) bool {
		return normal[i].duration < normal[j].duration
	})

	// Compute stats.
	minDur := math.MaxFloat64
	maxDur := 0.0
	var sum float64
	for _, s := range normal {
		if s.duration < minDur {
			minDur = s.duration
		}
		if s.duration > maxDur {
			maxDur = s.duration
		}
		sum += s.duration
	}
	if len(normal) == 0 {
		minDur = 0
		maxDur = 30
	}
	var avgDur float64
	if len(normal) > 0 {
		avgDur = sum / float64(len(normal))
	}

	// Bar scale: start at a fixed lower bound so fast stops still produce a
	// visible bar. We anchor at 15 s (minimum possible pit lane transit).
	const scaleMin = 15.0
	scaleMax := maxDur + 0.5

	// Layout widths.
	//   "  [DRVR] 1/2 Lnn  " = 2 + 5 + 5 + 5 = 17 chars of prefix
	//   " XX.Xs " = 7 chars of suffix
	const prefixW = 17
	const suffixW = 8
	barMaxW := w - prefixW - suffixW
	if barMaxW < 20 {
		barMaxW = 20
	}

	barW := func(dur float64) int {
		bw := int((dur - scaleMin) / (scaleMax - scaleMin) * float64(barMaxW))
		if bw < 1 {
			bw = 1
		}
		if bw > barMaxW {
			bw = barMaxW
		}
		return bw
	}

	barColor := func(s stopInfo) lipgloss.Color {
		if s.compound != "" {
			return compoundColor(s.compound)
		}
		return pitDurTierColor(s.duration)
	}

	axisStyle := lipgloss.NewStyle().Foreground(styles.ColorMuted)
	dimStyle := styles.DimStyle
	greenS := lipgloss.NewStyle().Foreground(styles.ColorGreen).Bold(true)
	dimS := lipgloss.NewStyle().Foreground(styles.ColorTextDim)

	var sb strings.Builder

	// ── Title ────────────────────────────────────────────────────────────────
	sb.WriteString(lipgloss.NewStyle().Bold(true).Foreground(styles.ColorSubtle).
		Render(fmt.Sprintf("  PIT STOP DURATIONS  ·  %s %s %d",
			a.session.CountryName, a.session.SessionName, a.session.Year)))
	sb.WriteString("\n\n")

	// ── Stats header ─────────────────────────────────────────────────────────
	if len(normal) > 0 {
		fast := normal[0]
		fastLabel := greenS.Render(fmt.Sprintf("%s %.1fs (L%d)", fast.acronym, fast.duration, fast.lap))
		avgLabel := dimS.Render(fmt.Sprintf("avg %.1fs", avgDur))
		cntLabel := dimStyle.Render(fmt.Sprintf("%d stops", len(normal)))
		retLabel := ""
		if len(retired) > 0 {
			var names []string
			for _, r := range retired {
				mins := int(r.duration) / 60
				names = append(names, fmt.Sprintf("%s(%dm)", r.acronym, mins))
			}
			retLabel = "   " + lipgloss.NewStyle().Foreground(styles.ColorF1Red).
				Render("⚑ parked in pits: "+strings.Join(names, ", "))
		}
		sb.WriteString(fmt.Sprintf("  Fastest: %s   %s   %s%s\n\n", fastLabel, avgLabel, cntLabel, retLabel))
	}

	// ── Axis tick labels ─────────────────────────────────────────────────────
	pad := strings.Repeat(" ", prefixW)
	tickRow := make([]byte, barMaxW)
	for i := range tickRow {
		tickRow[i] = ' '
	}
	// Place a tick label every 2 seconds from scaleMin+1 to scaleMax.
	for t := math.Ceil(scaleMin); t <= scaleMax; t += 2 {
		col := int((t-scaleMin)/(scaleMax-scaleMin)*float64(barMaxW-1))
		lbl := fmt.Sprintf("%.0fs", t)
		for li, ch := range []byte(lbl) {
			if col+li < len(tickRow) {
				tickRow[col+li] = ch
			}
		}
	}
	sb.WriteString(pad + dimStyle.Render(string(tickRow)) + "\n")
	sb.WriteString(pad + axisStyle.Render(strings.Repeat("┬", barMaxW)) + "\n")

	// ── Stop rows (fastest → slowest) ────────────────────────────────────────
	avgCol := int((avgDur-scaleMin)/(scaleMax-scaleMin)*float64(barMaxW-1))

	for _, s := range normal {
		// Driver badge: coloured background, white-on-team-colour text.
		badge := lipgloss.NewStyle().
			Background(s.teamCol).
			Foreground(lipgloss.Color("#000000")).
			Bold(true).
			Render(fmt.Sprintf(" %-3s ", s.acronym))

		// Stop indicator: "1/2" dim if only stop, brighter if multi-stop driver.
		stopTag := fmt.Sprintf("%d/%d", s.stopNum, s.totalStops)
		var stopTagStr string
		if s.totalStops > 1 {
			stopTagStr = lipgloss.NewStyle().Foreground(styles.ColorTextDim).Render(stopTag) + " "
		} else {
			stopTagStr = dimStyle.Render(stopTag) + " "
		}

		lapStr := dimS.Render(fmt.Sprintf("L%-2d ", s.lap))

		bw := barW(s.duration)
		col := barColor(s)

		// Build bar rune slice so we can embed the avg marker.
		barRunes := []rune(strings.Repeat("█", bw))

		// Overlay average marker "┊" on the bar where the average falls.
		if avgCol > 0 && avgCol < bw {
			barRunes[avgCol] = '┊'
		}

		bar := lipgloss.NewStyle().Foreground(col).Render(string(barRunes))

		// Avg marker to the right of the bar (if avg > this stop's bar end).
		avgMarker := ""
		if avgCol >= bw && avgCol < barMaxW {
			space := strings.Repeat(" ", avgCol-bw)
			avgMarker = space + lipgloss.NewStyle().Foreground(styles.ColorMuted).Render("┊")
		}

		durStr := lipgloss.NewStyle().Foreground(col).Bold(true).
			Render(fmt.Sprintf(" %.1fs", s.duration))

		// Compound suffix.
		cmpStr := ""
		if s.compound != "" {
			cmpStr = " " + lipgloss.NewStyle().Foreground(compoundColor(s.compound)).
				Render(strings.ToUpper(s.compound[:1]))
		}

		sb.WriteString(fmt.Sprintf("  %s%s%s%s%s%s%s\n",
			badge, stopTagStr, lapStr, bar, avgMarker, durStr, cmpStr))
	}

	// ── X-axis baseline ──────────────────────────────────────────────────────
	sb.WriteString(pad + axisStyle.Render(strings.Repeat("─", barMaxW)) + "\n")

	// avg marker label below the axis.
	if avgDur > 0 {
		markerLabelRow := make([]byte, barMaxW)
		for i := range markerLabelRow {
			markerLabelRow[i] = ' '
		}
		lbl := fmt.Sprintf("avg %.1fs", avgDur)
		start := avgCol - len(lbl)/2
		if start < 0 {
			start = 0
		}
		for li, ch := range []byte(lbl) {
			if start+li < len(markerLabelRow) {
				markerLabelRow[start+li] = ch
			}
		}
		sb.WriteString(pad + dimStyle.Render(string(markerLabelRow)) + "\n")
	}

	// ── Pit-lane retirements ─────────────────────────────────────────────────
	if len(retired) > 0 {
		sb.WriteString("\n")
		redS := lipgloss.NewStyle().Foreground(styles.ColorF1Red)
		sb.WriteString("  " + redS.Render("⚑  Parked / long stays in pit lane  (excluded from scale)") + "\n")
		for _, r := range retired {
			badge := lipgloss.NewStyle().
				Background(r.teamCol).
				Foreground(lipgloss.Color("#000000")).
				Bold(true).
				Render(fmt.Sprintf(" %-3s ", r.acronym))
			mins := int(r.duration) / 60
			secs := int(r.duration) % 60
			durLabel := fmt.Sprintf("%dm %02ds", mins, secs)
			sb.WriteString(fmt.Sprintf("  %s  L%-2d  %s\n",
				badge, r.lap, redS.Render(durLabel)))
		}
	}

	// ── Legend ───────────────────────────────────────────────────────────────
	sb.WriteString("\n  ")
	for _, entry := range []struct {
		col   lipgloss.Color
		label string
	}{
		{styles.ColorGreen, "< 19s fast"},
		{lipgloss.Color("#A3E635"), "19–22s normal"},
		{styles.ColorOrange, "22–28s slow"},
		{styles.ColorF1Red, "> 28s critical"},
	} {
		sb.WriteString(lipgloss.NewStyle().Foreground(entry.col).Render("█ "))
		sb.WriteString(dimStyle.Render(entry.label + "   "))
	}
	sb.WriteString(dimStyle.Render("  ┊ = field average   · all times = total pit lane duration"))
	sb.WriteRune('\n')

	return sb.String()
}
