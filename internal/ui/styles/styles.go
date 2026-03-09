package styles

import "github.com/charmbracelet/lipgloss"

// ── Palette ──────────────────────────────────────────────────────────────────

var (
	ColorBg         = lipgloss.Color("#0D0D0D")
	ColorBgCard     = lipgloss.Color("#111827")
	ColorBgHeader   = lipgloss.Color("#0A0A0A")
	ColorBorder     = lipgloss.Color("#1F2937")
	ColorF1Red      = lipgloss.Color("#E8002D")
	ColorF1White    = lipgloss.Color("#FFFFFF")
	ColorMuted      = lipgloss.Color("#4B5563")
	ColorSubtle     = lipgloss.Color("#6B7280")
	ColorText       = lipgloss.Color("#D1D5DB")
	ColorTextDim    = lipgloss.Color("#9CA3AF")
	ColorGreen      = lipgloss.Color("#22C55E")
	ColorPurple     = lipgloss.Color("#A855F7")
	ColorYellow     = lipgloss.Color("#EAB308")
	ColorOrange     = lipgloss.Color("#F97316")
	ColorBlue       = lipgloss.Color("#3B82F6")
	ColorTeal       = lipgloss.Color("#14B8A6")

	// Tyre compound colours
	ColorTyreSoft   = lipgloss.Color("#E8002D")
	ColorTyreMedium = lipgloss.Color("#FDE047")
	ColorTyreHard   = lipgloss.Color("#E5E7EB")
	ColorTyreInter  = lipgloss.Color("#4ADE80")
	ColorTyreWet    = lipgloss.Color("#60A5FA")

	// F1 team colours (2024/2025 grid)
	TeamColors = map[string]lipgloss.Color{
		"red bull racing":  lipgloss.Color("#3671C6"),
		"red bull":         lipgloss.Color("#3671C6"),
		"ferrari":          lipgloss.Color("#E8002D"),
		"mercedes":         lipgloss.Color("#27F4D2"),
		"mclaren":          lipgloss.Color("#FF8000"),
		"aston martin":     lipgloss.Color("#229971"),
		"williams":         lipgloss.Color("#64C4FF"),
		"alpine":           lipgloss.Color("#FF87BC"),
		"haas":             lipgloss.Color("#B6BABD"),
		"haas f1 team":     lipgloss.Color("#B6BABD"),
		"rb":               lipgloss.Color("#6692FF"),
		"racing bulls":     lipgloss.Color("#6692FF"),
		"kick sauber":      lipgloss.Color("#52E252"),
		"sauber":           lipgloss.Color("#52E252"),
		"audi":             lipgloss.Color("#E8E800"),
	}
)

// TeamColor looks up a team colour by name (case-insensitive substring),
// falling back to white if not found.
func TeamColor(teamName string) lipgloss.Color {
	lower := toLower(teamName)
	for k, v := range TeamColors {
		if contains(lower, k) {
			return v
		}
	}
	return ColorText
}

// TyreColor returns the canonical colour for a tyre compound string.
func TyreColor(compound string) lipgloss.Color {
	switch compound {
	case "SOFT":
		return ColorTyreSoft
	case "MEDIUM":
		return ColorTyreMedium
	case "HARD":
		return ColorTyreHard
	case "INTERMEDIATE":
		return ColorTyreInter
	case "WET":
		return ColorTyreWet
	default:
		return ColorSubtle
	}
}

// TyreLabel returns a short compound label.
func TyreLabel(compound string) string {
	switch compound {
	case "SOFT":
		return "S"
	case "MEDIUM":
		return "M"
	case "HARD":
		return "H"
	case "INTERMEDIATE":
		return "I"
	case "WET":
		return "W"
	default:
		return "?"
	}
}

// ── Shared Styles ─────────────────────────────────────────────────────────────

var (
	Base = lipgloss.NewStyle().
		Background(ColorBg).
		Foreground(ColorText)

	Title = lipgloss.NewStyle().
		Bold(true).
		Foreground(ColorF1Red)

	SubTitle = lipgloss.NewStyle().
		Foreground(ColorSubtle)

	ActiveTab = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(ColorF1Red).
		Padding(0, 2)

	InactiveTab = lipgloss.NewStyle().
		Foreground(ColorSubtle).
		Padding(0, 2)

	Card = lipgloss.NewStyle().
		Background(ColorBgCard).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorBorder).
		Padding(0, 1)

	Header = lipgloss.NewStyle().
		Bold(true).
		Foreground(ColorTextDim).
		Background(ColorBgHeader)

	Divider = lipgloss.NewStyle().
		Foreground(ColorBorder)

	ErrorStyle = lipgloss.NewStyle().
		Foreground(ColorF1Red).
		Bold(true)

	InfoStyle = lipgloss.NewStyle().
		Foreground(ColorBlue)

	SuccessStyle = lipgloss.NewStyle().
		Foreground(ColorGreen)

	DimStyle = lipgloss.NewStyle().
		Foreground(ColorSubtle)

	BoldWhite = lipgloss.NewStyle().
		Bold(true).
		Foreground(ColorF1White)

	FooterStyle = lipgloss.NewStyle().
		Foreground(ColorMuted).
		Background(ColorBgHeader)
)

// BarChart renders a simple horizontal bar proportional to value/max.
func BarChart(value, max float64, width int, color lipgloss.Color) string {
	if max == 0 {
		return ""
	}
	filled := int(float64(width) * value / max)
	if filled > width {
		filled = width
	}
	bar := ""
	for i := 0; i < filled; i++ {
		bar += "█"
	}
	for i := filled; i < width; i++ {
		bar += "░"
	}
	return lipgloss.NewStyle().Foreground(color).Render(bar)
}

// WindDirection converts degrees to a compass label.
func WindDirection(deg int) string {
	dirs := []string{"N", "NE", "E", "SE", "S", "SW", "W", "NW"}
	idx := int(float64(deg)/45.0+0.5) % 8
	return dirs[idx]
}

// helpers (avoid importing strings to keep package light)
func toLower(s string) string {
	b := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 32
		}
		b[i] = c
	}
	return string(b)
}

func contains(s, sub string) bool {
	if len(sub) > len(s) {
		return false
	}
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
