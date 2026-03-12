package dashboard

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/dk-a-dev/termf1/v2/internal/api/multiviewer"
	"github.com/dk-a-dev/termf1/v2/internal/ui/styles"
	"github.com/dk-a-dev/termf1/v2/internal/ui/views/common"
)

// ── Side panel (right) ────────────────────────────────────────────────────────

func (d *Dashboard2) renderSidePanel(w, h int) string {
	// Track map gets 42% of height, stats+radio get the rest.
	trackH := h * 42 / 100
	if trackH < 12 {
		trackH = 12
	}
	statsH := h - trackH - 1 // 1-row divider

	trackWidget := d.renderMiniMap(w, trackH)
	divider := styles.Divider.Render(strings.Repeat("─", w))
	statsWidget := d.renderWeatherAndRCM(w, statsH)

	return lipgloss.JoinVertical(lipgloss.Left, trackWidget, divider, statsWidget)
}

// ── Track map ─────────────────────────────────────────────────────────────────

func (d *Dashboard2) renderMiniMap(w, h int) string {
	title := styles.DimStyle.Render("TRACK MAP")
	if d.circuitName != "" {
		title = styles.DimStyle.Render("TRACK  ") +
			lipgloss.NewStyle().Foreground(styles.ColorTeal).Render(d.circuitName)
	}

	var mapLines []string
	if d.circuit != nil {
			mapLines = renderCircuitOnCanvas(d.circuit, d.liveCarPositions(), d.getCarColors(), w-4, h-3)
	} else {
		art := common.GetCircuitArt(d.circuitName)
		for _, l := range art {
			vis := l
			if len(vis) > w-4 {
				vis = vis[:w-4]
			}
			mapLines = append(mapLines, lipgloss.NewStyle().
				Foreground(lipgloss.Color("#4FC3F7")).Render(vis))
			if len(mapLines) >= h-3 {
				break
			}
		}
	}

	inner := title + "\n" + strings.Join(mapLines, "\n")
	return lipgloss.NewStyle().
		Width(w).Height(h).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.ColorBorder).
		Background(styles.ColorBgCard).
		Padding(0, 1).
		Render(inner)
}

// liveCarPositions returns car positions normalised to [0,1] from the live state.
func (d *Dashboard2) liveCarPositions() map[string][2]float64 {
	if d.liveState == nil || len(d.liveState.Positions) == 0 {
		return nil
	}
	var minX, maxX, minY, maxY float64
	first := true
	for _, p := range d.liveState.Positions {
		if first {
			minX, maxX = p.X, p.X
			minY, maxY = p.Y, p.Y
			first = false
		}
		if p.X < minX {
			minX = p.X
		}
		if p.X > maxX {
			maxX = p.X
		}
		if p.Y < minY {
			minY = p.Y
		}
		if p.Y > maxY {
			maxY = p.Y
		}
	}
	rangeX := maxX - minX
	rangeY := maxY - minY
	if rangeX == 0 || rangeY == 0 {
		return nil
	}
	out := make(map[string][2]float64, len(d.liveState.Positions))
	for num, p := range d.liveState.Positions {
		out[num] = [2]float64{
			(p.X - minX) / rangeX,
			(p.Y - minY) / rangeY,
		}
	}
	return out
}

// canvasCell is a rune + foreground colour for the circuit canvas.
type canvasCell struct {
	r   rune
	col string // hex colour, e.g. "#FF0000"; "" = space
}

const trackDotColor = "#4FC3F7"

// renderCircuitOnCanvas draws the GPS circuit outline and live car markers
// onto a canvas, returning one styled string per row.
// carColors maps car-number strings to hex team-colour strings.
func renderCircuitOnCanvas(
	circuit *multiviewer.Circuit,
	carPos map[string][2]float64,
	carColors map[string]string,
	w, h int,
) []string {
	if w < 4 || h < 2 {
		return nil
	}

	canvas := make([][]canvasCell, h)
	for i := range canvas {
		canvas[i] = make([]canvasCell, w)
		for j := range canvas[i] {
			canvas[i][j] = canvasCell{' ', ""}
		}
	}

	// Draw the track outline.
	if len(circuit.X) > 0 && len(circuit.Y) == len(circuit.X) {
		minX, maxX := circuit.X[0], circuit.X[0]
		minY, maxY := circuit.Y[0], circuit.Y[0]
		for _, x := range circuit.X {
			if x < minX {
				minX = x
			}
			if x > maxX {
				maxX = x
			}
		}
		for _, y := range circuit.Y {
			if y < minY {
				minY = y
			}
			if y > maxY {
				maxY = y
			}
		}
		rx := float64(maxX - minX)
		ry := float64(maxY - minY)
		if rx > 0 && ry > 0 {
			prev := [2]int{-1, -1}
			for i := range circuit.X {
				cx := int(float64(w-1) * float64(circuit.X[i]-minX) / rx)
				cy := int(float64(h-1) * (1.0 - float64(circuit.Y[i]-minY)/ry))
				if cx >= 0 && cx < w && cy >= 0 && cy < h {
					canvas[cy][cx] = canvasCell{'·', trackDotColor}
					if prev[0] >= 0 {
						drawLineOnCanvas(canvas, prev[0], prev[1], cx, cy, w, h)
					}
					prev = [2]int{cx, cy}
				}
			}
		}
	}

	// Overlay car positions with team-coloured number labels.
	for num, pos := range carPos {
		cx := int(pos[0] * float64(w-1))
		cy := int((1 - pos[1]) * float64(h-1))
		if cx < 0 || cx >= w || cy < 0 || cy >= h {
			continue
		}
		col := carColors[num]
		if col == "" {
			col = "#FFFFFF"
		} else if !strings.HasPrefix(col, "#") {
			col = "#" + col
		}
		label := num
		if len(label) > 2 {
			label = label[len(label)-2:]
		}
		for k, ch := range []rune(label) {
			if cx+k < w {
				canvas[cy][cx+k] = canvasCell{ch, col}
			}
		}
	}

	// Render each row by grouping consecutive same-colour runs.
	rows := make([]string, h)
	for i, row := range canvas {
		var sb strings.Builder
		j := 0
		for j < w {
			col := row[j].col
			start := j
			for j < w && row[j].col == col {
				j++
			}
			runes := make([]rune, j-start)
			for k, c := range row[start:j] {
				runes[k] = c.r
			}
			text := string(runes)
			if col == "" {
				sb.WriteString(text)
			} else {
				sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(col)).Render(text))
			}
		}
		rows[i] = sb.String()
	}
	return rows
}

// drawLineOnCanvas draws a Bresenham line of track dots on the canvas.
func drawLineOnCanvas(canvas [][]canvasCell, x0, y0, x1, y1, w, h int) {
	dx := abs(x1 - x0)
	dy := -abs(y1 - y0)
	sx := 1
	if x0 > x1 {
		sx = -1
	}
	sy := 1
	if y0 > y1 {
		sy = -1
	}
	err := dx + dy
	for {
		if x0 >= 0 && x0 < w && y0 >= 0 && y0 < h {
			canvas[y0][x0] = canvasCell{'·', trackDotColor}
		}
		if x0 == x1 && y0 == y1 {
			break
		}
		e2 := 2 * err
		if e2 >= dy {
			err += dy
			x0 += sx
		}
		if e2 <= dx {
			err += dx
			y0 += sy
		}
	}
}

// getCarColors returns a map from car number to hex team colour for map rendering.
func (d *Dashboard2) getCarColors() map[string]string {
	if d.liveState == nil || len(d.liveState.Drivers) == 0 {
		return nil
	}
	out := make(map[string]string, len(d.liveState.Drivers))
	for num, drv := range d.liveState.Drivers {
		col := drv.TeamColour
		if col != "" && !strings.HasPrefix(col, "#") {
			col = "#" + col
		}
		out[num] = col
	}
	return out
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// ── Weather + RCM ─────────────────────────────────────────────────────────────

func (d *Dashboard2) renderWeatherAndRCM(w, h int) string {
	weatherH := 7
	remaining := h - weatherH

	hasRadio := d.liveState != nil && len(d.liveState.TeamRadio) > 0
	if !hasRadio {
		rcmH := remaining
		if rcmH < 4 {
			rcmH = 4
		}
		return lipgloss.JoinVertical(lipgloss.Left,
			d.renderWeatherWidget(w, weatherH),
			d.renderRCMWidget(w, rcmH),
		)
	}

	radioH := 7
	rcmH := remaining - radioH
	if rcmH < 4 {
		rcmH = 4
	}
	return lipgloss.JoinVertical(lipgloss.Left,
		d.renderWeatherWidget(w, weatherH),
		d.renderRCMWidget(w, rcmH),
		d.renderTeamRadioWidget(w, radioH),
	)
}

func (d *Dashboard2) renderWeatherWidget(w, h int) string {
	var air, track, humid, wind, rain string

	if d.serverAlive && d.liveState != nil && d.liveState.Weather.AirTemp != "" {
		wt := d.liveState.Weather
		air = wt.AirTemp
		track = wt.TrackTemp
		humid = wt.Humidity
		wind = wt.WindSpeed + " m/s"
		if wt.Rainfall != "" && wt.Rainfall != "0" && wt.Rainfall != "0.0" {
			rain = lipgloss.NewStyle().Foreground(styles.ColorBlue).Render("🌧 Rain")
		} else {
			rain = lipgloss.NewStyle().Foreground(styles.ColorGreen).Render("Dry")
		}
	} else if d.fallback != nil && len(d.fallback.Weathers) > 0 {
		wt := d.fallback.Weathers[len(d.fallback.Weathers)-1]
		air = fmt.Sprintf("%.1f", wt.AirTemperature)
		track = fmt.Sprintf("%.1f", wt.TrackTemperature)
		humid = fmt.Sprintf("%.0f%%", wt.Humidity)
		wind = fmt.Sprintf("%.1f m/s", wt.WindSpeed)
		if wt.Rainfall > 0 {
			rain = lipgloss.NewStyle().Foreground(styles.ColorBlue).Render("🌧 Rain")
		} else {
			rain = lipgloss.NewStyle().Foreground(styles.ColorGreen).Render("Dry")
		}
	} else {
		return styles.Card.Width(w).Height(h).Render(styles.DimStyle.Render("No weather data"))
	}

	rows := []string{
		styles.BoldWhite.Render("⛅ WEATHER"),
		fmt.Sprintf("  Air    %s°C", lipgloss.NewStyle().Foreground(styles.ColorOrange).Render(air)),
		fmt.Sprintf("  Track  %s°C", lipgloss.NewStyle().Foreground(styles.ColorF1Red).Render(track)),
		fmt.Sprintf("  Humid  %s", humid),
		fmt.Sprintf("  Wind   %s", wind),
		fmt.Sprintf("  Rain   %s", rain),
	}
	return lipgloss.NewStyle().
		Width(w).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.ColorBorder).
		Background(styles.ColorBgCard).
		Padding(0, 1).
		Render(strings.Join(rows, "\n"))
}

func (d *Dashboard2) renderRCMWidget(w, h int) string {
	title := styles.BoldWhite.Render("📡 RACE CONTROL")
	lines := []string{title, ""}

	msgs := d.rcTicker
	maxLines := h - 4
	if maxLines < 1 {
		maxLines = 1
	}

	// clamp scroll so we can't scroll past the oldest message
	scroll := d.rcmScroll
	if scroll > len(msgs)-1 {
		scroll = len(msgs) - 1
	}
	if scroll < 0 {
		scroll = 0
	}

	// Render newest-first, offset by scroll
	startIdx := len(msgs) - 1 - scroll
	shown := 0
	for i := startIdx; i >= 0 && shown < maxLines; i-- {
		wrapped := wordWrapMulti(msgs[i], w-5)
		for _, l := range strings.Split(wrapped, "\n") {
			col := styles.ColorTextDim
			upper := strings.ToUpper(msgs[i])
			if strings.Contains(upper, "RED FLAG") {
				col = styles.ColorF1Red
			} else if strings.Contains(upper, "SAFETY CAR") || strings.Contains(upper, "VSC") {
				col = styles.ColorYellow
			} else if strings.Contains(upper, "DRS") {
				col = styles.ColorGreen
			}
			lines = append(lines, lipgloss.NewStyle().Foreground(col).Render("  "+l))
			shown++
			if shown >= maxLines {
				break
			}
		}
	}
	if len(msgs) == 0 {
		lines = append(lines, styles.DimStyle.Render("  No messages"))
	} else if scroll > 0 {
		// scroll indicator
		lines = append(lines, styles.DimStyle.Render(fmt.Sprintf("  ↑ %d more  (k/pgup scroll)", scroll)))
	}

	return lipgloss.NewStyle().
		Width(w).Height(h).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.ColorBorder).
		Background(styles.ColorBgCard).
		Padding(0, 1).
		Render(strings.Join(lines, "\n"))
}

// ── Team Radio ────────────────────────────────────────────────────────────────

func (d *Dashboard2) renderTeamRadioWidget(w, h int) string {
	title := styles.BoldWhite.Render("📻 TEAM RADIO")
	lines := []string{title, ""}

	if d.liveState == nil || len(d.liveState.TeamRadio) == 0 {
		lines = append(lines, styles.DimStyle.Render("  No radio clips"))
		return lipgloss.NewStyle().
			Width(w).Height(h).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(styles.ColorBorder).
			Background(styles.ColorBgCard).
			Padding(0, 1).
			Render(strings.Join(lines, "\n"))
	}

	clips := d.liveState.TeamRadio
	maxShow := h - 4
	if maxShow < 1 {
		maxShow = 1
	}

	// Newest first
	for i := len(clips) - 1; i >= 0 && len(lines)-2 < maxShow; i-- {
		clip := clips[i]

		tla := clip.RacingNumber
		if d.liveState.Drivers != nil {
			if drv, ok := d.liveState.Drivers[clip.RacingNumber]; ok && drv.Tla != "" {
				tla = drv.Tla
			}
		}

		// "2025-03-16T03:16:04.01Z" → "03:16"
		timeStr := clip.Utc
		if len(clip.Utc) >= 16 {
			timeStr = clip.Utc[11:16]
		}

		line := fmt.Sprintf("  %s  %s #%s",
			styles.DimStyle.Render(timeStr),
			lipgloss.NewStyle().Foreground(styles.ColorTeal).Bold(true).Render(tla),
			clip.RacingNumber,
		)
		lines = append(lines, line)
	}

	return lipgloss.NewStyle().
		Width(w).Height(h).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.ColorBorder).
		Background(styles.ColorBgCard).
		Padding(0, 1).
		Render(strings.Join(lines, "\n"))
}

// ── Bottom ticker ─────────────────────────────────────────────────────────────

func (d *Dashboard2) renderTicker() string {
	msg := " No race control messages"
	if len(d.rcTicker) > 0 {
		msg = "📡 " + d.rcTicker[len(d.rcTicker)-1]
	}
	// Clamp to terminal width so it never wraps.
	maxW := d.width - 2
	if maxW < 1 {
		maxW = 1
	}
	runes := []rune(msg)
	if len(runes) > maxW {
		runes = runes[:maxW-1]
		msg = string(runes) + "…"
	}
	return lipgloss.NewStyle().
		Background(styles.ColorBgCard).
		Foreground(styles.ColorTextDim).
		Width(d.width).
		MaxWidth(d.width).
		Padding(0, 1).
		Render(msg)
}
