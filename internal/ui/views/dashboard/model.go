package dashboard

import (
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/harmonica"
	"github.com/charmbracelet/lipgloss"
	"github.com/devkeshwani/termf1/internal/api/f1static"
	"github.com/devkeshwani/termf1/internal/api/jolpica"
	"github.com/devkeshwani/termf1/internal/api/liveserver"
	"github.com/devkeshwani/termf1/internal/api/multiviewer"
	"github.com/devkeshwani/termf1/internal/api/openf1"
	"github.com/devkeshwani/termf1/internal/ui/styles"
)

// ── Messages ──────────────────────────────────────────────────────────────────

type dashLiveMsg struct{ state *liveserver.State }
type dashFallbackMsg struct{ payload *openf1.DashboardPayload }
type dashIdleFallbackMsg struct{ payload *openf1.DashboardPayload } // server alive but no live session
type dashErrMsg2 struct{ err error }
type dashTickMsg2 time.Time
type dashSpringTickMsg struct{}
type dashCircuitMsg2 struct{ circuit *multiviewer.Circuit }

// ── Row types ─────────────────────────────────────────────────────────────────

// liveRow is the view-model for one car in the live timing panel.
type liveRow struct {
	Pos            int
	PrevPos        int // position on last update (0 = same/unknown)
	Number         string
	Tla            string
	TeamName       string
	TeamColour     string
	GapToLeader    string
	Interval       string
	IntervalCatching bool
	LastLap        string
	LastFastest    bool
	LastPersonal   bool
	BestLap        string
	S1, S2, S3             string
	S1Prev, S2Prev, S3Prev string
	S1Fast, S2Fast, S3Fast bool
	S1Personal, S2Personal, S3Personal bool
	S1Segs, S2Segs, S3Segs []int // segment statuses (0=off,2048=personal,2049=overall fastest)
	SpeedST  string
	InPit    bool
	PitCount int
	Laps     int
	Retired  bool
	Compound string // tyre compound: SOFT/MEDIUM/HARD/INTERMEDIATE/WET
	TyreAge  int    // laps on current tyre
	TyreNew  bool
	DRS      int    // 10/12/14 = open; 0/8 = closed/eligible

	// spring animation for gap
	gapSpring  harmonica.Spring
	gapDisplay float64
	gapTarget  float64
	gapStr     string // non-numeric gap (e.g. "LEADER", "+1 LAP")
}

// fallbackRow is the view-model for one car in the historical results panel.
type fallbackRow struct {
	Pos         int
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
	DNF         bool
}

// ── Dashboard2 ────────────────────────────────────────────────────────────────

// Dashboard2 is the unified dashboard: live timing OR historical fallback,
// GPS track map, weather, and RCM widgets.
type Dashboard2 struct {
	// clients
	lsClient   *liveserver.Client
	of1Client  *openf1.Client
	f1s        *f1static.Client
	joli       *jolpica.Client
	mv         *multiviewer.Client

	// sizing
	width  int
	height int

	// state
	loading     bool
	err         error
	serverAlive bool
	liveState   *liveserver.State
	fallback    *openf1.DashboardPayload
	liveRows    []liveRow
	fbRows      []fallbackRow
	rcTicker    []string // last N RCM messages
	tickerPos   int      // bottom ticker horizontal scroll offset
	rcmScroll   int      // RCM widget scroll (lines from bottom)

	// track map
	circuit     *multiviewer.Circuit
	circuitName string

	// spinner
	spin spinner.Model

	// refresh
	refreshSec     int
	fallbackFetched bool // true once we've loaded historical data (don't re-fetch every tick)
}

// NewDashboard2 constructs a Dashboard2 ready to be included in the app model.
func NewDashboard2(
	lsClient *liveserver.Client,
	of1 *openf1.Client,
	joli *jolpica.Client,
	mv *multiviewer.Client,
	refreshSec int,
) *Dashboard2 {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(styles.ColorF1Red)
	return &Dashboard2{
		lsClient:   lsClient,
		of1Client:  of1,
		f1s:        f1static.NewClient(),
		joli:       joli,
		mv:         mv,
		refreshSec: refreshSec,
		loading:    true,
		spin:       s,
	}
}

// SetSize updates the terminal dimensions used for layout.
func (d *Dashboard2) SetSize(w, h int) {
	d.width = w
	d.height = h
}
