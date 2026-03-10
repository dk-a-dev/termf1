// Package common provides small UI utilities shared across all view subpackages.
package common

import (
	"context"
	"fmt"
	"math"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/devkeshwani/termf1/internal/ui/styles"
)

// Centred places s in the centre of a w×h terminal region.
func Centred(w, h int, s string) string {
	if h < 1 {
		h = 1
	}
	topPad := (h - 1) / 2
	return strings.Repeat("\n", topPad) +
		lipgloss.NewStyle().Width(w).Align(lipgloss.Center).Render(s)
}

// ContextBG returns a background context for fetch commands.
func ContextBG() context.Context { return context.Background() }

// PosColor returns medal colour for pos 1–3, text colour otherwise.
func PosColor(pos int) lipgloss.Color {
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

// FormatDuration formats seconds as "M:SS.mmm" (or "SS.mmm" when < 1 min).
func FormatDuration(secs float64) string {
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

// FormatSector formats a sector time as "SS.mmm".
func FormatSector(secs float64) string {
	if secs <= 0 {
		return "—"
	}
	return fmt.Sprintf("%.3f", secs)
}

// FormatLap returns a formatted lap time and its display colour.
func FormatLap(last, best float64) (string, lipgloss.Color) {
	s := FormatDuration(last)
	if last <= 0 {
		return s, styles.ColorSubtle
	}
	if best > 0 && math.Abs(last-best) < 0.001 {
		return s, styles.ColorPurple
	}
	return s, styles.ColorText
}

// GetCircuitArt returns the ASCII art lines for the named circuit (or default).
func GetCircuitArt(name string) []string {
	if art, ok := circuitArt[name]; ok {
		return art
	}
	return circuitArt["default"]
}

var circuitArt = map[string][]string{
	"Melbourne": {
		"                                              ",
		"           ╭──────╮                           ",
		"          ╭╯      │  ╭────╮                  ",
		"         ╭╯       ╰──╯    │                  ",
		"        ╭╯               ╭╯                  ",
		"       ╭╯              ╭─╯                   ",
		"      ╭╯           ╭───╯                     ",
		"     ╭╯         ╭──╯                         ",
		"     │        ╭─╯                            ",
		"     │       ╭╯                              ",
		"     │      ╭╯   ╭╮                          ",
		"     │     ╭╯  ╭─╯╰──╮                       ",
		"     ╰─────╯  │      ╰─────────╮             ",
		"              ╰╮               │             ",
		"               ╰───────────────╯             ",
		"                                              ",
	},
	"Monza": {
		"                                              ",
		"        ╭────────────────────╮               ",
		"        │ ╭──────────────╮   │               ",
		"        │ │              │   │               ",
		"        │ │  ╭────────╮  │   │               ",
		"        │ │  │        │  │   │               ",
		"        │ ╰──╯        ╰──╯   │               ",
		"        │                    │               ",
		"        ╰────────────────────╯               ",
		"                                              ",
	},
	"Silverstone": {
		"                                              ",
		"    ╭──────────────────╮                     ",
		"   ╭╯                  ╰╮                    ",
		"  ╭╯    ╭──────╮        ╰╮                   ",
		"  │    ╭╯      ╰╮        │                   ",
		"  │   ╭╯         ╰╮      │                   ",
		"  │   │            ╰─────╯                   ",
		"  ╰───╯                                      ",
		"                                              ",
	},
	"Monaco": {
		"                                              ",
		"  ╭──────────────────────╮                   ",
		"  │  ╭─────────────────╮ │                   ",
		"  │  │ ╭─╮   ╭──╮      │ │                   ",
		"  │  │ │ ╰───╯  ╰────╮ │ │                   ",
		"  │  │ │             │ │ │                   ",
		"  │  ╰─╯             ╰─╯ │                   ",
		"  ╰────────────────────╯                     ",
		"                                              ",
	},
	"Spa": {
		"                                              ",
		"      ╭─────────────────────────╮            ",
		"     ╭╯                         │            ",
		"    ╭╯   ╭──────────────╮       │            ",
		"   ╭╯    │              ╰──╮    │            ",
		"  ╭╯     │                 ╰────╯            ",
		"  │      │                                   ",
		"  ╰──────╯                                   ",
		"                                              ",
	},
	"default": {
		"                                             ",
		"         ╭───────────────────────╮          ",
		"        ╭╯                       ╰╮         ",
		"       ╭╯     ╭───────────╮       ╰╮        ",
		"      ╭╯      │           │        ╰╮       ",
		"     ╭╯       │           │         │       ",
		"     │        ╰─────╮     │         │       ",
		"     │          ╭───╯     │         │       ",
		"     ╰──────────╯         ╰─────────╯       ",
		"                                             ",
	},
}