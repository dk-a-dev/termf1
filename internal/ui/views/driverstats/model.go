package driverstats

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/bubbletea"
	"github.com/devkeshwani/termf1/internal/ui/views/common"
	"github.com/charmbracelet/lipgloss"
	"github.com/devkeshwani/termf1/internal/api/jolpica"
	"github.com/devkeshwani/termf1/internal/api/openf1"
	"github.com/devkeshwani/termf1/internal/ui/styles"
)

// ── Messages ──────────────────────────────────────────────────────────────────

type driverStatsDataMsg struct {
	session      openf1.Session
	drivers      []openf1.Driver
	laps         []openf1.Lap
	stints       []openf1.Stint
	pits         []openf1.Pit
	positions    []openf1.Position
	driverStandings []jolpica.DriverStanding
}
type driverStatsErrMsg struct{ err error }

// ── Per-driver computed profile ───────────────────────────────────────────────

type driverProfile struct {
	number      int
	acronym     string
	fullName    string
	team        string
	teamColor   string
	nationality string

	// session stats
	bestLap      float64
	avgLap       float64
	worstLap     float64
	lapCount     int
	pitStops     int
	avgS1        float64
	avgS2        float64
	avgS3        float64
	topSpeed     int
	position     int

	// championship
	champPoints string
	champPos    string
	champWins   string

	// lap history for sparkline
	lapHistory []float64
	// sector history
	s1History  []float64
	s2History  []float64
}

// ── Model ─────────────────────────────────────────────────────────────────────

type DriverStats struct {
	of1Client  *openf1.Client
	jolClient  *jolpica.Client
	width      int
	height     int
	loading    bool
	err        error
	profiles   []driverProfile
	viewport   viewport.Model
	spin       spinner.Model
	cursor     int // selected driver
	tab        int // 0=overview 1=lap analysis 2=sectors
}

func NewDriverStats(of1 *openf1.Client, joli *jolpica.Client) *DriverStats {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(styles.ColorF1Red)
	vp := viewport.New(80, 30)
	return &DriverStats{of1Client: of1, jolClient: joli, loading: true, spin: s, viewport: vp}
}

func (d *DriverStats) SetSize(w, h int) {
	d.width = w
	d.height = h
	d.viewport.Width = w
	d.viewport.Height = h - 6
}

func (d *DriverStats) Init() tea.Cmd {
	return tea.Batch(d.spin.Tick, fetchDriverStatsCmd(d.of1Client, d.jolClient))
}

func (d *DriverStats) UpdateDriverStats(msg tea.Msg) (*DriverStats, tea.Cmd) {
	switch msg := msg.(type) {
	case spinner.TickMsg:
		if d.loading {
			var cmd tea.Cmd
			d.spin, cmd = d.spin.Update(msg)
			return d, cmd
		}
	case driverStatsDataMsg:
		d.loading = false
		d.err = nil
		d.profiles = buildProfiles(msg)
		d.viewport.SetContent(d.buildContent())
	case driverStatsErrMsg:
		d.loading = false
		d.err = msg.err

	case tea.KeyMsg:
		switch msg.String() {
		case "left", "h":
			if d.cursor > 0 {
				d.cursor--
				d.viewport.SetContent(d.buildContent())
				d.viewport.GotoTop()
			}
			return d, nil
		case "right", "l":
			if d.cursor < len(d.profiles)-1 {
				d.cursor++
				d.viewport.SetContent(d.buildContent())
				d.viewport.GotoTop()
			}
			return d, nil
		case "t":
			d.tab = (d.tab + 1) % 3
			d.viewport.SetContent(d.buildContent())
			d.viewport.GotoTop()
			return d, nil
		}
	}

	var cmd tea.Cmd
	d.viewport, cmd = d.viewport.Update(msg)
	return d, cmd
}

func (d *DriverStats) View() string {
	if d.loading && len(d.profiles) == 0 {
		return common.Centred(d.width, d.height, d.spin.View()+" Loading driver stats…")
	}
	if d.err != nil && len(d.profiles) == 0 {
		return common.Centred(d.width, d.height, styles.ErrorStyle.Render("⚠  "+d.err.Error()))
	}

	title := styles.Title.Render(" 📊 Driver Statistics")
	sep := styles.Divider.Render(strings.Repeat("─", d.width))

	// Driver selector strip
	selector := d.renderSelector()
	selSep := styles.Divider.Render(strings.Repeat("─", d.width))

	// Tab bar
	tabs := d.renderTabs()

	hint := styles.FooterStyle.Render(
		"  ←/→ driver  │  t switch tab  │  ↑↓/PgUp/PgDn scroll",
	)

	d.viewport.SetContent(d.buildContent())

	return lipgloss.JoinVertical(lipgloss.Left,
		title, sep,
		selector, selSep,
		tabs,
		d.viewport.View(),
		hint,
	)
}

// ── content builders ──────────────────────────────────────────────────────────

func (d *DriverStats) renderSelector() string {
	var chips []string
	visible := 12 // how many driver chips to show
	start := d.cursor - visible/2
	if start < 0 {
		start = 0
	}
	end := start + visible
	if end > len(d.profiles) {
		end = len(d.profiles)
	}

	for i := start; i < end; i++ {
		p := d.profiles[i]
		col := lipgloss.Color("#" + p.teamColor)
		if p.teamColor == "" {
			col = styles.ColorSubtle
		}
		if i == d.cursor {
			chips = append(chips, lipgloss.NewStyle().
				Background(col).
				Foreground(lipgloss.Color("#000")).
				Bold(true).
				Padding(0, 1).
				Render(" "+p.acronym+" "))
		} else {
			chips = append(chips, lipgloss.NewStyle().
				Foreground(col).
				Padding(0, 1).
				Render(p.acronym))
		}
	}
	return "  " + strings.Join(chips, " ")
}

func (d *DriverStats) renderTabs() string {
	tabs := []string{"Overview", "Lap Analysis", "Sector Breakdown"}
	var rendered []string
	for i, t := range tabs {
		if i == d.tab {
			rendered = append(rendered, styles.ActiveTab.Render(" "+t+" "))
		} else {
			rendered = append(rendered, styles.InactiveTab.Render(" "+t+" "))
		}
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, rendered...)
}

func (d *DriverStats) buildContent() string {
	if len(d.profiles) == 0 {
		return styles.DimStyle.Render("No data available.")
	}
	if d.cursor >= len(d.profiles) {
		d.cursor = len(d.profiles) - 1
	}
	p := d.profiles[d.cursor]

	switch d.tab {
	case 0:
		return d.renderOverview(p)
	case 1:
		return d.renderLapAnalysis(p)
	case 2:
		return d.renderSectors(p)
	}
	return ""
}

// ── Overview tab ──────────────────────────────────────────────────────────────

func (d *DriverStats) renderOverview(p driverProfile) string {
	col := lipgloss.Color("#" + p.teamColor)
	if p.teamColor == "" {
		col = styles.ColorSubtle
	}

	badge := lipgloss.NewStyle().
		Background(col).
		Foreground(lipgloss.Color("#000")).
		Bold(true).
		Padding(0, 2).
		Render(" " + p.acronym + " ")

	name := lipgloss.NewStyle().Bold(true).Foreground(styles.ColorF1White).Render(p.fullName)
	team := lipgloss.NewStyle().Foreground(col).Render(p.team)
	nat := styles.DimStyle.Render("  " + p.nationality)

	header := lipgloss.JoinHorizontal(lipgloss.Top, badge, "  ", name, "  ", team, nat)

	// Stats grid
	cardW := (d.width - 8) / 3
	if cardW < 18 {
		cardW = 18
	}

	statCards := []string{
		d.statCard("Championship", p.champPos+"  "+p.champPoints+" pts", styles.ColorYellow, cardW),
		d.statCard("Wins", p.champWins, styles.ColorGreen, cardW),
		d.statCard("Pit Stops", fmt.Sprintf("%d", p.pitStops), styles.ColorOrange, cardW),
	}
	row1 := lipgloss.JoinHorizontal(lipgloss.Top, statCards...)

	lapCards := []string{
		d.statCard("Best Lap", common.FormatDuration(p.bestLap), styles.ColorPurple, cardW),
		d.statCard("Avg Lap", common.FormatDuration(p.avgLap), styles.ColorText, cardW),
		d.statCard("Lap Count", fmt.Sprintf("%d", p.lapCount), styles.ColorTeal, cardW),
	}
	row2 := lipgloss.JoinHorizontal(lipgloss.Top, lapCards...)

	// Top speed
	topSpeedCard := d.statCard("Top Speed (km/h)", fmt.Sprintf("%d", p.topSpeed), styles.ColorBlue, cardW)

	// Lap time sparkline
	sparkTitle := lipgloss.NewStyle().Bold(true).Foreground(styles.ColorSubtle).Render("  Lap Time Trend")
	spark := d.lapSparkline(p, d.width-6)

	return lipgloss.JoinVertical(lipgloss.Left,
		header, "",
		"  "+row1, "  "+row2,
		"  "+topSpeedCard, "",
		sparkTitle, spark, "",
	)
}

func (d *DriverStats) statCard(label, value string, col lipgloss.Color, w int) string {
	l := styles.DimStyle.Render(label)
	v := lipgloss.NewStyle().Bold(true).Foreground(col).Render(value)
	return styles.Card.Width(w).Render(l + "\n" + v)
}

func (d *DriverStats) lapSparkline(p driverProfile, w int) string {
	if len(p.lapHistory) < 2 {
		return styles.DimStyle.Render("  Not enough lap data")
	}
	// Find min/max excluding 0
	minV, maxV := 9999.0, 0.0
	for _, v := range p.lapHistory {
		if v > 0 && v < minV {
			minV = v
		}
		if v > maxV {
			maxV = v
		}
	}
	diff := maxV - minV
	if diff == 0 {
		diff = 1
	}

	levels := []string{"▁", "▂", "▃", "▄", "▅", "▆", "▇", "█"}
	nLevels := len(levels)

	var spark strings.Builder
	for i, v := range p.lapHistory {
		if v <= 0 {
			spark.WriteString(styles.DimStyle.Render("?"))
			continue
		}
		// Invert: faster = taller bar
		norm := 1.0 - (v-minV)/diff
		idx := int(norm*float64(nLevels-1) + 0.5)
		if idx < 0 {
			idx = 0
		}
		if idx >= nLevels {
			idx = nLevels - 1
		}
		// Colour: purple for best, orange for slowest, green for rest
		col := styles.ColorGreen
		if v == p.bestLap {
			col = styles.ColorPurple
		} else if i > 0 && v > p.avgLap*1.01 {
			col = styles.ColorOrange
		}
		spark.WriteString(lipgloss.NewStyle().Foreground(col).Render(levels[idx]))
	}

	legend := styles.DimStyle.Render(fmt.Sprintf("  Best:%s  Avg:%s  Worst:%s  (%d laps)",
		common.FormatDuration(p.bestLap), common.FormatDuration(p.avgLap), common.FormatDuration(p.worstLap), p.lapCount))

	return "  " + spark.String() + "\n" + legend
}

// ── Lap Analysis tab ──────────────────────────────────────────────────────────

func (d *DriverStats) renderLapAnalysis(p driverProfile) string {
	title := lipgloss.NewStyle().Bold(true).Foreground(styles.ColorSubtle).Render("  Lap-by-Lap Breakdown")

	if len(p.lapHistory) == 0 {
		return lipgloss.JoinVertical(lipgloss.Left, title, "",
			styles.DimStyle.Render("  No lap data available."))
	}

	barW := d.width - 30
	if barW < 10 {
		barW = 10
	}

	// Find range
	minL, maxL := p.bestLap, p.worstLap
	if minL <= 0 {
		minL = p.avgLap
	}
	rng := maxL - minL
	if rng == 0 {
		rng = 1
	}

	lines := []string{title, ""}
	lapNumW := lipgloss.NewStyle().Width(5).Foreground(styles.ColorMuted).Render("LAP ")
	timeW := lipgloss.NewStyle().Width(10).Foreground(styles.ColorMuted).Render("TIME      ")
	deltaW := lipgloss.NewStyle().Width(8).Foreground(styles.ColorMuted).Render("DELTA   ")
	hdr := lapNumW + timeW + deltaW
	lines = append(lines, "  "+hdr)
	lines = append(lines, "  "+styles.Divider.Render(strings.Repeat("─", barW+25)))

	for i, lt := range p.lapHistory {
		lapNum := i + 1
		if lt <= 0 {
			lines = append(lines, styles.DimStyle.Render(fmt.Sprintf("  %3d  —", lapNum)))
			continue
		}

		// Bar representing how close to best lap (taller = faster)
		norm := 1.0 - (lt-minL)/rng
		if norm < 0 {
			norm = 0
		}
		filled := int(norm * float64(barW))
		bar := lipgloss.NewStyle().Foreground(barColor(lt, p.bestLap, p.avgLap)).
			Render(strings.Repeat("█", filled) + strings.Repeat("░", barW-filled))

		delta := lt - p.bestLap
		deltaStr := fmt.Sprintf("+%.3f", delta)
		if lt == p.bestLap {
			deltaStr = "best  "
		}

		lapNumStr := lipgloss.NewStyle().Width(4).Foreground(styles.ColorSubtle).
			Render(fmt.Sprintf("%3d", lapNum))
		timeStr := lipgloss.NewStyle().Width(9).Foreground(lapTimeColor(lt, p.bestLap, p.avgLap)).
			Render(common.FormatDuration(lt))
		deltaColored := lipgloss.NewStyle().Width(8).Foreground(barColor(lt, p.bestLap, p.avgLap)).
			Render(deltaStr)

		lines = append(lines, "  "+lapNumStr+" "+timeStr+" "+deltaColored+" "+bar)
	}
	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

// ── Sector Breakdown tab ──────────────────────────────────────────────────────

func (d *DriverStats) renderSectors(p driverProfile) string {
	title := lipgloss.NewStyle().Bold(true).Foreground(styles.ColorSubtle).Render("  Sector Performance")

	lines := []string{title, ""}

	cardW := (d.width - 8) / 3
	if cardW < 18 {
		cardW = 18
	}

	s1Card := d.statCard("Avg Sector 1", common.FormatSector(p.avgS1), styles.ColorGreen, cardW)
	s2Card := d.statCard("Avg Sector 2", common.FormatSector(p.avgS2), styles.ColorTeal, cardW)
	s3Card := d.statCard("Avg Sector 3", common.FormatSector(p.avgS3), styles.ColorYellow, cardW)
	row := lipgloss.JoinHorizontal(lipgloss.Top, s1Card, s2Card, s3Card)
	lines = append(lines, "  "+row, "")

	// S1 sparkline
	if len(p.s1History) >= 2 {
		lines = append(lines, d.sectorSparkLine("Sector 1", p.s1History, styles.ColorGreen))
		lines = append(lines, d.sectorSparkLine("Sector 2", p.s2History, styles.ColorTeal))
	} else {
		lines = append(lines, styles.DimStyle.Render("  Not enough sector data."))
	}

	// Distribution histogram for lap times
	if len(p.lapHistory) >= 3 {
		lines = append(lines, "", d.lapHistogram(p))
	}

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func (d *DriverStats) sectorSparkLine(label string, data []float64, col lipgloss.Color) string {
	minV, maxV := 9999.0, 0.0
	for _, v := range data {
		if v > 0 && v < minV {
			minV = v
		}
		if v > maxV {
			maxV = v
		}
	}
	diff := maxV - minV
	if diff == 0 {
		diff = 0.001
	}
	levels := []string{"▁", "▂", "▃", "▄", "▅", "▆", "▇", "█"}
	var spark strings.Builder
	for _, v := range data {
		if v <= 0 {
			spark.WriteString(" ")
			continue
		}
		norm := 1.0 - (v-minV)/diff
		idx := int(norm*float64(len(levels)-1) + 0.5)
		if idx < 0 {
			idx = 0
		}
		if idx >= len(levels) {
			idx = len(levels) - 1
		}
		spark.WriteString(lipgloss.NewStyle().Foreground(col).Render(levels[idx]))
	}
	labelStr := lipgloss.NewStyle().Width(12).Foreground(styles.ColorSubtle).Render(label)
	minStr := styles.DimStyle.Render(fmt.Sprintf("%.3f", minV))
	maxStr := lipgloss.NewStyle().Foreground(col).Render(fmt.Sprintf("%.3f", maxV))
	return "  " + labelStr + " " + minStr + " " + spark.String() + " " + maxStr
}

// lapHistogram renders a simple ASCII histogram of lap time distribution.
func (d *DriverStats) lapHistogram(p driverProfile) string {
	if len(p.lapHistory) < 3 {
		return ""
	}
	minL, maxL := p.worstLap, p.bestLap
	if minL-maxL == 0 {
		return ""
	}
	bins := 10
	counts := make([]int, bins)
	rng := p.worstLap - p.bestLap
	for _, v := range p.lapHistory {
		if v <= 0 {
			continue
		}
		idx := int((v-p.bestLap)/rng*float64(bins-1) + 0.5)
		if idx < 0 {
			idx = 0
		}
		if idx >= bins {
			idx = bins - 1
		}
		counts[idx]++
	}
	maxC := 1
	for _, c := range counts {
		if c > maxC {
			maxC = c
		}
	}

	title := lipgloss.NewStyle().Bold(true).Foreground(styles.ColorSubtle).Render("  Lap Time Distribution")
	lines := []string{title}
	barH := 6
	// Render columns top-down
	for row := barH; row >= 1; row-- {
		var sb strings.Builder
		sb.WriteString("  ")
		for _, c := range counts {
			filled := int(float64(c) / float64(maxC) * float64(barH))
			if filled >= row {
				sb.WriteString(lipgloss.NewStyle().Foreground(styles.ColorTeal).Render("██ "))
			} else {
				sb.WriteString("   ")
			}
		}
		lines = append(lines, sb.String())
	}
	// X axis labels
	var xAxis strings.Builder
	xAxis.WriteString("  ")
	for i := 0; i < bins; i++ {
		v := minL + (maxL-minL)*float64(i)/float64(bins-1)
		if i == 0 || i == bins-1 {
			xAxis.WriteString(lipgloss.NewStyle().Foreground(styles.ColorMuted).Render(fmt.Sprintf("%-3.0f", v)))
		} else {
			xAxis.WriteString("   ")
		}
	}
	lines = append(lines, styles.Divider.Render("  "+strings.Repeat("─", bins*3)))
	lines = append(lines, xAxis.String())
	lines = append(lines, styles.DimStyle.Render("  seconds →"))
	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

// ── colour helpers ─────────────────────────────────────────────────────────────

func barColor(lt, best, avg float64) lipgloss.Color {
	if lt == best {
		return styles.ColorPurple
	}
	if lt <= avg {
		return styles.ColorGreen
	}
	if lt <= avg*1.005 {
		return styles.ColorYellow
	}
	return styles.ColorOrange
}

func lapTimeColor(lt, best, avg float64) lipgloss.Color {
	if lt == best {
		return styles.ColorPurple
	}
	if lt <= avg {
		return styles.ColorText
	}
	return styles.ColorTextDim
}

// ── data processing ────────────────────────────────────────────────────────────

func buildProfiles(msg driverStatsDataMsg) []driverProfile {
	// Build lookup maps
	driverMap := make(map[int]openf1.Driver)
	for _, d := range msg.drivers {
		driverMap[d.DriverNumber] = d
	}
	posMap := make(map[int]int)
	for _, p := range msg.positions {
		posMap[p.DriverNumber] = p.Position
	}

	// Championship lookup
	champMap := make(map[string]jolpica.DriverStanding)
	for _, ds := range msg.driverStandings {
		champMap[ds.Driver.Code] = ds
	}

	// Lap data per driver (ordered by lap number)
	type lapRec struct {
		num  int
		lap  openf1.Lap
	}
	lapsByDriver := make(map[int][]lapRec)
	for _, l := range msg.laps {
		lapsByDriver[l.DriverNumber] = append(lapsByDriver[l.DriverNumber], lapRec{l.LapNumber, l})
	}
	for num := range lapsByDriver {
		recs := lapsByDriver[num]
		sort.Slice(recs, func(i, j int) bool { return recs[i].num < recs[j].num })
		lapsByDriver[num] = recs
	}

	// Pit counts
	pitCount := make(map[int]int)
	for _, pit := range msg.pits {
		pitCount[pit.DriverNumber]++
	}

	profiles := make([]driverProfile, 0, len(driverMap))
	for num, drv := range driverMap {
		recs := lapsByDriver[num]
		var lapHist []float64
		var s1Hist, s2Hist []float64
		best, worst, sum := 0.0, 0.0, 0.0
		count := 0
		sumS1, sumS2, sumS3 := 0.0, 0.0, 0.0
		cntS := 0
		topSpeed := 0
		for _, r := range recs {
			lap := r.lap
			if lap.LapDuration > 0 {
				lapHist = append(lapHist, lap.LapDuration)
				sum += lap.LapDuration
				count++
				if best == 0 || lap.LapDuration < best {
					best = lap.LapDuration
				}
				if lap.LapDuration > worst {
					worst = lap.LapDuration
				}
			}
			if lap.DurationSector1 > 0 && lap.DurationSector2 > 0 {
				s1Hist = append(s1Hist, lap.DurationSector1)
				s2Hist = append(s2Hist, lap.DurationSector2)
				sumS1 += lap.DurationSector1
				sumS2 += lap.DurationSector2
				sumS3 += lap.DurationSector3
				cntS++
			}
			if lap.StSpeed > topSpeed {
				topSpeed = lap.StSpeed
			}
		}
		avg := 0.0
		if count > 0 {
			avg = sum / float64(count)
		}
		avgS1, avgS2, avgS3 := 0.0, 0.0, 0.0
		if cntS > 0 {
			avgS1 = sumS1 / float64(cntS)
			avgS2 = sumS2 / float64(cntS)
			avgS3 = sumS3 / float64(cntS)
		}

		standing := champMap[drv.NameAcronym]
		// Try full name lookup too
		if standing.Driver.Code == "" {
			for _, ds := range msg.driverStandings {
				if ds.Driver.Code == drv.NameAcronym {
					standing = ds
					break
				}
			}
		}

		profiles = append(profiles, driverProfile{
			number:      num,
			acronym:     drv.NameAcronym,
			fullName:    drv.FullName,
			team:        drv.TeamName,
			teamColor:   drv.TeamColour,
			nationality: drv.CountryCode,
			bestLap:     best,
			avgLap:      avg,
			worstLap:    worst,
			lapCount:    count,
			pitStops:    pitCount[num],
			avgS1:       avgS1,
			avgS2:       avgS2,
			avgS3:       avgS3,
			topSpeed:    topSpeed,
			position:    posMap[num],
			champPoints: standing.Points,
			champPos:    standing.Position,
			champWins:   standing.Wins,
			lapHistory:  lapHist,
			s1History:   s1Hist,
			s2History:   s2Hist,
		})
	}

	// Sort by position
	sort.Slice(profiles, func(i, j int) bool {
		pi, pj := profiles[i].position, profiles[j].position
		if pi == 0 {
			return false
		}
		if pj == 0 {
			return true
		}
		return pi < pj
	})

	// If no championship data, try sorting by champPos
	if len(profiles) > 0 && profiles[0].champPos != "" {
		sort.SliceStable(profiles, func(i, j int) bool {
			pi, _ := strconv.Atoi(profiles[i].champPos)
			pj, _ := strconv.Atoi(profiles[j].champPos)
			if pi == 0 {
				return false
			}
			if pj == 0 {
				return true
			}
			return pi < pj
		})
	}

	return profiles
}

// ── command ───────────────────────────────────────────────────────────────────

func fetchDriverStatsCmd(of1 *openf1.Client, joli *jolpica.Client) tea.Cmd {
	return func() tea.Msg {
		ctx := common.ContextBG()
		session, err := of1.GetLatestSession(ctx)
		if err != nil {
			return driverStatsErrMsg{err}
		}
		sk := session.SessionKey

		drivers, _ := of1.GetDrivers(ctx, sk)
		laps, _ := of1.GetLaps(ctx, sk)
		stints, _ := of1.GetStints(ctx, sk)
		pits, _ := of1.GetPits(ctx, sk)
		positions, _ := of1.GetPositions(ctx, sk)
		driverStandings, _ := joli.GetDriverStandings(ctx)

		return driverStatsDataMsg{
			session:         session,
			drivers:         drivers,
			laps:            laps,
			stints:          stints,
			pits:            pits,
			positions:       positions,
			driverStandings: driverStandings,
		}
	}
}
