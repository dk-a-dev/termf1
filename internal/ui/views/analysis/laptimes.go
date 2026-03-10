package analysis

import (
    "fmt"
    "math"
    "sort"
    "strings"

    "github.com/charmbracelet/lipgloss"
    "github.com/devkeshwani/termf1/internal/api/openf1"
    "github.com/devkeshwani/termf1/internal/ui/styles"
    "github.com/devkeshwani/termf1/internal/ui/views/common"
)

// ── Chart 2: Lap Time Line Chart ──────────────────────────────────────────────

// renderLapLineChart draws a 2-D grid for top-8 drivers.
// Each driver uses a single filled circle ● in their team colour.
// Consecutive lap points are joined with ─ connectors.
// Pit stops are marked with a vertical │.
func (a *Analysis) renderLapLineChart(w, h int) string {
    drMap := a.driverMap()

    type pt struct {
        lap   int
        t     float64
        isPit bool
    }
    byDriver := map[int][]pt{}
    pitLaps := map[int]map[int]bool{}
    for _, l := range a.laps {
        if l.LapDuration > 0 {
            byDriver[l.DriverNumber] = append(byDriver[l.DriverNumber], pt{
                lap:   l.LapNumber,
                t:     l.LapDuration,
                isPit: l.IsPitOutLap,
            })
        }
        if l.IsPitOutLap {
            if pitLaps[l.DriverNumber] == nil {
                pitLaps[l.DriverNumber] = map[int]bool{}
            }
            pitLaps[l.DriverNumber][l.LapNumber] = true
        }
    }
    if len(byDriver) == 0 {
        return common.Centred(w, h, styles.DimStyle.Render("No lap time data"))
    }

    type drvData struct {
        num               int
        acronym, teamName string
        pts               []pt
        best              float64
    }
    var drvs []drvData
    for num, pts := range byDriver {
        sort.Slice(pts, func(i, j int) bool { return pts[i].lap < pts[j].lap })
        best := math.MaxFloat64
        for _, p := range pts {
            if !p.isPit && p.t < best {
                best = p.t
            }
        }
        if best == math.MaxFloat64 {
            continue
        }
        drv := drMap[num]
        drvs = append(drvs, drvData{num, drv.NameAcronym, drv.TeamName, pts, best})
    }
    sort.Slice(drvs, func(i, j int) bool { return drvs[i].best < drvs[j].best })
    if len(drvs) > 8 {
        drvs = drvs[:8] // top 8 for clarity
    }

    // Y range: use p5–p95 of clean laps to exclude outliers.
    var allT []float64
    for _, d := range drvs {
        for _, p := range d.pts {
            if !p.isPit {
                allT = append(allT, p.t)
            }
        }
    }
    sort.Float64s(allT)
    n := len(allT)
    tMin := allT[n/20] - 0.2
    tMax := allT[n*19/20] + 0.5
    tRange := tMax - tMin
    if tRange < 1 {
        tMin -= 0.3
        tMax += 0.7
        tRange = 1
    }

    maxLap := 0
    for _, d := range drvs {
        for _, p := range d.pts {
            if p.lap > maxLap {
                maxLap = p.lap
            }
        }
    }
    if maxLap == 0 {
        maxLap = 1
    }

    const (
        yAxisW  = 9  // "1:23.456 " width
        titleH  = 4
        xAxisH  = 3
        legendH = 4
    )
    cW := w - yAxisW - 4
    cH := h - titleH - xAxisH - legendH
    if cW < 20 {
        cW = 20
    }
    if cH < 8 {
        cH = 8
    }

    type cell struct {
        r   rune
        col lipgloss.Color
        set bool
    }
    grid := make([][]cell, cH)
    for i := range grid {
        grid[i] = make([]cell, cW)
    }

    lapToCol := func(lap int) int {
        c := (lap - 1) * (cW - 1) / maxLap
        if c < 0 { c = 0 }
        if c >= cW { c = cW - 1 }
        return c
    }
    timeToRow := func(t float64) int {
        norm := (t - tMin) / tRange
        if norm < 0 { norm = 0 }
        if norm > 1 { norm = 1 }
        r := int(norm * float64(cH-1))
        if r < 0 { r = 0 }
        if r >= cH { r = cH - 1 }
        return r
    }

    // Draw each driver's trace.
    for _, d := range drvs {
        col := styles.TeamColor(d.teamName)
        prevC, prevR := -1, -1
        for _, p := range d.pts {
            if p.t < tMin || p.t > tMax {
                prevC = -1
                continue
            }
            c := lapToCol(p.lap)
            r := timeToRow(p.t)

            // Draw horizontal connector (─) between previous and current point.
            if prevC >= 0 && c > prevC+1 {
                midR := (prevR + r) / 2
                for ic := prevC + 1; ic < c; ic++ {
                    if ic >= 0 && ic < cW {
                        ir := prevR + (r-prevR)*(ic-prevC)/(c-prevC)
                        if ir >= 0 && ir < cH && !grid[ir][ic].set {
                            grid[ir][ic] = cell{'─', col, true}
                        }
                        // also fill midpoint connector
                        if midR >= 0 && midR < cH && !grid[midR][ic].set {
                            _ = midR // suppress unused
                        }
                    }
                }
            }

            // Place data point — ● for normal laps, P for pit entry.
            sym := '●'
            if p.isPit {
                sym = 'p'
            }
            if pits, ok := pitLaps[d.num]; ok && pits[p.lap] {
                sym = 'P'
            }
            if r >= 0 && r < cH && c >= 0 && c < cW {
                grid[r][c] = cell{sym, col, true}
            }
            prevC, prevR = c, r
        }
    }

    var sb strings.Builder
    sb.WriteString("\n")
    sb.WriteString(sectionTitle("LAP TIME COMPARISON",
        fmt.Sprintf("top %d drivers · laps 1–%d · %s – %s",
            len(drvs), maxLap, formatLapSec(tMin), formatLapSec(tMax))))
    sb.WriteString("\n\n")

    for row := 0; row < cH; row++ {
        t := tMin + float64(row)/float64(cH-1)*tRange
        yLabel := styles.DimStyle.Render(fmt.Sprintf("%-8s", formatLapSec(t)))
        sb.WriteString("  ")
        sb.WriteString(yLabel)
        sb.WriteString(styles.DimStyle.Render("│"))
        for col := 0; col < cW; col++ {
            c := grid[row][col]
            if c.set {
                sb.WriteString(lipgloss.NewStyle().Foreground(c.col).Render(string(c.r)))
            } else {
                sb.WriteRune(' ')
            }
        }
        sb.WriteRune('\n')
    }

    sb.WriteString("  " + safeRep(" ", yAxisW) + styles.DimStyle.Render("└"+safeRep("─", cW)) + "\n")

    // X tick labels.
    ticks := make([]byte, cW+2)
    for i := range ticks {
        ticks[i] = ' '
    }
    for lap := 1; lap <= maxLap; lap += 10 {
        col := lapToCol(lap)
        s := fmt.Sprintf("%d", lap)
        for si, ch := range s {
            pos := col + si
            if pos < len(ticks) {
                ticks[pos] = byte(ch)
            }
        }
    }
    sb.WriteString("  " + safeRep(" ", yAxisW+1) + styles.DimStyle.Render(string(ticks)) + "\n")
    sb.WriteString("  " + safeRep(" ", yAxisW+1+cW/2) + styles.DimStyle.Render("Lap →") + "\n")

    // Legend — driver colour squares.
    sb.WriteString("\n  ")
    for _, d := range drvs {
        col := styles.TeamColor(d.teamName)
        acronym := d.acronym
        if acronym == "" {
            acronym = fmt.Sprintf("#%d", d.num)
        }
        sb.WriteString(lipgloss.NewStyle().Foreground(col).Render("● "+acronym))
        sb.WriteString("  ")
    }
    sb.WriteString("\n")
    sb.WriteString(styles.DimStyle.Render("  ─ = lap connector   P = pit stop entry   +/h ←→/l scroll   r refresh"))
    return sb.String()
}

// ── Chart 3: Per-driver sparklines ───────────────────────────────────────────

// renderLapSparklines draws one row per driver.
// Each lap column uses a 3-tier block based on pace relative to that driver's best:
//   █  bright team colour  → fast  (≤ 102 % of best)
//   ▄  normal team colour  → avg   (102–107 %)
//   ░  dim team colour     → slow  (107–115 %)
//   ·  subtle              → pit-out lap
//   ' '                    → pit-in / no data
func (a *Analysis) renderLapSparklines(w, h int) string {
    drMap := a.driverMap()

    byDriver := map[int][]openf1.Lap{}
    for _, lap := range a.laps {
        if lap.LapDuration > 0 {
            byDriver[lap.DriverNumber] = append(byDriver[lap.DriverNumber], lap)
        }
    }

    type dl struct {
        num           int
        laps          []openf1.Lap
        best, avg     float64
        acronym, team string
    }
    var driverList []dl
    for num, laps := range byDriver {
        sort.Slice(laps, func(i, j int) bool { return laps[i].LapNumber < laps[j].LapNumber })
        best := math.MaxFloat64
        total, count := 0.0, 0
        for _, l := range laps {
            if l.LapDuration > 0 && !l.IsPitOutLap {
                if l.LapDuration < best {
                    best = l.LapDuration
                }
                total += l.LapDuration
                count++
            }
        }
        if best == math.MaxFloat64 {
            continue
        }
        avg := 0.0
        if count > 0 {
            avg = total / float64(count)
        }
        drv := drMap[num]
        driverList = append(driverList, dl{num, laps, best, avg, drv.NameAcronym, drv.TeamName})
    }
    sort.Slice(driverList, func(i, j int) bool { return driverList[i].best < driverList[j].best })
    if len(driverList) == 0 {
        return common.Centred(w, h, styles.DimStyle.Render("No lap time data"))
    }

    maxLapNum := 0
    for _, d := range driverList {
        for _, l := range d.laps {
            if l.LapNumber > maxLapNum {
                maxLapNum = l.LapNumber
            }
        }
    }
    if maxLapNum == 0 {
        maxLapNum = 1
    }

    const (
        rankW  = 4
        labelW = 5
        statW  = 34
    )
    sparkW := w - rankW - labelW - statW - 6
    if sparkW < 10 {
        sparkW = 10
    }
    if sparkW > maxLapNum {
        sparkW = maxLapNum
    }

    globalMin := math.MaxFloat64
    for _, d := range driverList {
        if d.best < globalMin {
            globalMin = d.best
        }
    }

    lines := []string{
        "",
        sectionTitle("LAP TIME SPARKLINES",
            fmt.Sprintf("%d drivers · %d laps · range %s – %s",
                len(driverList), maxLapNum,
                formatLapSec(globalMin), formatLapSec(driverList[len(driverList)-1].avg))),
        "  " + styles.Divider.Render(safeRep("─", w-4)),
        "",
    }

    for rank, d := range driverList {
        teamCol := styles.TeamColor(d.team)
        dimCol := lipgloss.Color(dimHex(string(teamCol)))
        acronym := d.acronym
        if acronym == "" {
            acronym = fmt.Sprintf("#%d", d.num)
        }
        rankStr := styles.DimStyle.Render(fmt.Sprintf("P%-2d", rank+1))
        label := lipgloss.NewStyle().Foreground(teamCol).Bold(true).Width(labelW).Render(acronym)

        // Build per-column arrays (one char per lap or grouped laps).
        lapMap := make(map[int]openf1.Lap, len(d.laps))
        for _, l := range d.laps {
            lapMap[l.LapNumber] = l
        }

        var spark strings.Builder
        for col := 0; col < sparkW; col++ {
            lap := 1 + col*maxLapNum/sparkW
            l, ok := lapMap[lap]
            if !ok {
                spark.WriteRune(' ')
                continue
            }
            if l.LapDuration <= 0 {
                spark.WriteRune(' ')
                continue
            }
            if l.IsPitOutLap {
                spark.WriteString(styles.DimStyle.Render("·"))
                continue
            }
            ratio := l.LapDuration / d.best
            switch {
            case ratio <= 1.02:
                spark.WriteString(lipgloss.NewStyle().Foreground(teamCol).Render("█"))
            case ratio <= 1.07:
                spark.WriteString(lipgloss.NewStyle().Foreground(teamCol).Render("▄"))
            case ratio <= 1.15:
                spark.WriteString(lipgloss.NewStyle().Foreground(dimCol).Render("░"))
            default:
                spark.WriteString(styles.DimStyle.Render("_"))
            }
        }

        var bestColor lipgloss.Color
        if rank == 0 {
            bestColor = lipgloss.Color("#FFD700")
        } else {
            bestColor = teamCol
        }
        stats := lipgloss.NewStyle().Foreground(bestColor).Bold(rank == 0).Render(formatLapSec(d.best)) +
            styles.DimStyle.Render("  avg "+formatLapSec(d.avg))
        lines = append(lines, "  "+rankStr+" "+label+" "+spark.String()+"  "+stats)
    }

    lines = append(lines, "",
        styles.DimStyle.Render(safeRep(" ", rankW+labelW+4)+
            "█ fast (≤102%)  ▄ avg (102–107%)  ░ slow  · pit-out  ' ' pit-in"),
        styles.DimStyle.Render(safeRep(" ", rankW+labelW+4)+
            fmt.Sprintf("Lap 1 ←──────────────────────────────────────→ %d", maxLapNum)))
    return strings.Join(lines, "\n")
}

// dimHex returns a darker shade of a hex colour string for "dim" rendering.
// Falls back to #555555 for non-hex inputs.
func dimHex(hex string) string {
    if len(hex) == 7 && hex[0] == '#' {
        var r, g, b int
        fmt.Sscanf(hex[1:], "%02x%02x%02x", &r, &g, &b)
        r = r * 55 / 100
        g = g * 55 / 100
        b = b * 55 / 100
        return fmt.Sprintf("#%02x%02x%02x", r, g, b)
    }
    return "#555555"
}