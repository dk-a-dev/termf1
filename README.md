# termf1-v1 🏎

A full-featured **Formula 1 terminal dashboard** built with Go and [Charm](https://charm.sh/) — standings, race calendar, circuit track maps, driver statistics, live weather, and an AI chat powered by Groq.

---

## Views

| # | Tab | What you get |
|---|-----|--------------|
| 1 | **Dashboard** *(WIP)* | Live timing — being rebuilt on top of the official F1 live-timing protocol for real sector splits, telemetry, and race control. |
| 2 | **Standings** | Driver & Constructor championship standings with team-coloured proportional bar charts. |
| 3 | **Schedule** | Full season calendar grouped by month. Cursor-navigate with `j`/`k`. Next race is auto-highlighted. |
| 4 | **Weather** | Air temp, track temp, humidity, pressure, wind speed/direction, rainfall + sparkline trend charts. |
| 5 | **Ask AI** | Chat with Groq's `compound-beta` model (has live web search) — race strategy, regulations, history, anything F1. |
| 6 | **Track Map** | Real circuit outlines rendered from GPS coordinate data ([Multiviewer API](https://api.multiviewer.app)). Corner numbers overlaid. Press `s` to replay the Australian GP 2025 in simulation mode. |
| 7 | **Driver Stats** | Per-driver lap-time sparklines, best/avg/worst laps, sector trends, lap histogram, pit stop counts, championship position — three sub-tabs. |

---

## Install

```bash
git clone https://github.com/devkeshwani/termf1
cd termf1
cp .env.example .env   # fill in your GROQ_API_KEY
source .env
make run
```

Or install the binary globally:

```bash
make install
termf1
```

## Requirements

- Go 1.22+
- A 256-colour / true-colour terminal (iTerm2, Ghostty, kitty, WezTerm, etc.)
- [Groq API key](https://console.groq.com) — free tier is sufficient for the Ask AI tab

## Configuration

Copy `.env.example` to `.env` and set:

```env
GROQ_API_KEY=your_groq_api_key_here   # required for Ask AI
GROQ_MODEL=compound-beta              # default; supports any Groq model
REFRESH_RATE=5                        # seconds between live-data polls
```

## Keybindings

| Key | Action |
|-----|--------|
| `1`–`7` | Jump to tab |
| `Tab` / `Shift+Tab` | Cycle tabs |
| `↑` `↓` / `j` `k` | Scroll / move cursor |
| `r` | Refresh current view |
| `q` / `Ctrl+C` | Quit |
| **Schedule** | |
| `Enter` / `Space` | Expand race session detail |
| **Track Map** | |
| `s` | Toggle Australian GP simulation |
| **Driver Stats** | |
| `←` `→` / `h` `l` | Previous / next driver |
| `t` | Switch sub-tab (Overview → Lap Analysis → Sectors) |
| **Ask AI** | |
| `Enter` | Send message |
| `Esc` | Blur input (enables viewport scroll) |
| `Ctrl+L` | Clear chat history |

## Data sources

| Source | Used for |
|--------|----------|
| [OpenF1 API](https://openf1.org) | Session data, laps, positions, stints, pits, weather |
| [Jolpica / Ergast API](https://jolpi.ca) | Championship standings, full season calendar |
| [Multiviewer API](https://api.multiviewer.app) | Real GPS circuit coordinate data for track maps |
| [Groq API](https://groq.com) | AI chat with live web search (`compound-beta`) |

## Roadmap

- [ ] **Live timing backend** — custom server reverse-engineered from `livetiming.formula1.com` SignalR feed: real-time sector splits, per-car telemetry (speed, throttle, brake, gear), animated track positions, race control messages, commentary
- [ ] Animated driver dots on track map from live X/Y positions
- [ ] Team radio clips (audio stream URLs from `TeamRadio` topic)
- [ ] On-track battle detection & gap trend sparklines

## Project layout

```
termf1/
├── main.go
├── internal/
│   ├── config/             – env var loading (GROQ_API_KEY, GROQ_MODEL, REFRESH_RATE)
│   ├── api/
│   │   ├── openf1/         – OpenF1 REST client & typed models
│   │   ├── jolpica/        – Jolpica/Ergast client & models
│   │   ├── groq/           – Groq chat completions client
│   │   └── multiviewer/    – Multiviewer circuit coordinate client + verified circuit key map
│   └── ui/
│       ├── app.go          – root Bubbletea model, tab navigation, header/footer
│       ├── styles/         – Lipgloss colour palette, shared styles, team/tyre colour helpers
│       └── views/
│           ├── dashboard.go   – [WIP] live timing table
│           ├── standings.go   – championship standings + bar charts
│           ├── schedule.go    – scrollable calendar with month grouping
│           ├── weather.go     – weather cards + sparklines
│           ├── chat.go        – Groq AI chat with viewport history
│           ├── trackmap.go    – real circuit renderer + simulation mode
│           └── driverstats.go – per-driver stats, graphs, sector breakdown
```

Note: i am currently re-writing the dashboard and will ensure track-maps , weather, player-data and audio come on same one page, the project v1 is vibecoded v2 will be very structured.

