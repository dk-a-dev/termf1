package dashboard

import (
	"math"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

// Init starts the spinner, kicks the first fetch, and starts spring ticks.
func (d *Dashboard2) Init() tea.Cmd {
	return tea.Batch(
		d.spin.Tick,
		d.fetchCmd(),
		d.springTickCmd(),
	)
}

// UpdateDash processes all Dashboard2-specific messages.
func (d *Dashboard2) UpdateDash(msg tea.Msg) (*Dashboard2, tea.Cmd) {
	switch msg := msg.(type) {

	case spinner.TickMsg:
		if d.loading {
			var cmd tea.Cmd
			d.spin, cmd = d.spin.Update(msg)
			return d, cmd
		}

	case dashLiveMsg:
		d.loading = false
		d.err = nil
		d.serverAlive = true
		d.liveState = msg.state
		d.liveRows = buildLiveRows(msg.state, d.liveRows)
		// Only overwrite rcTicker when live state has actual messages; otherwise
		// the off-season state would blank out the last-race messages we loaded.
		if live := liveRCMessages(msg.state, 20); len(live) > 0 {
			d.rcTicker = live
			d.rcmScroll = 0
		}

		cmds := []tea.Cmd{d.tickCmd(), d.springTickCmd()}

		// No active session — fetch last race to fill the timing panel.
		if len(d.liveRows) == 0 && len(d.fbRows) == 0 {
			cmds = append(cmds, fetchFallbackForIdle(d.of1Client))
		}

		circName := msg.state.SessionInfo.Meeting.Circuit.ShortName
		if circName != d.circuitName && circName != "" {
			d.circuitName = circName
			cmds = append(cmds, fetchCircuitForDash(d.mv, msg.state))
		}
		return d, tea.Batch(cmds...)

	case dashFallbackMsg:
		d.loading = false
		d.err = nil
		d.serverAlive = false
		d.fallbackFetched = true
		d.fallback = msg.payload
		d.fbRows = buildFallbackRows(msg.payload)
		if len(d.rcTicker) == 0 {
			d.rcTicker = fallbackRCMessages(msg.payload, 20)
		}
		circName := msg.payload.Session.CircuitShortName
		if circName != d.circuitName && circName != "" {
			d.circuitName = circName
			return d, tea.Batch(
				d.tickCmd(),
				d.springTickCmd(),
				fetchCircuitForDashFallback(d.mv, msg.payload),
			)
		}
		return d, tea.Batch(d.tickCmd(), d.springTickCmd())

	case dashIdleFallbackMsg:
		// Server is live but no active session — populate fallback panel.
		if msg.payload == nil {
			return d, nil
		}
		d.fallback = msg.payload
		d.fbRows = buildFallbackRows(msg.payload)
		d.rcTicker = fallbackRCMessages(msg.payload, 20)
		circName := msg.payload.Session.CircuitShortName
		if circName != d.circuitName && circName != "" {
			d.circuitName = circName
			return d, tea.Batch(fetchCircuitForDashFallback(d.mv, msg.payload), d.springTickCmd())
		}
		return d, d.springTickCmd()

	case dashErrMsg2:
		d.loading = false
		d.err = msg.err
		return d, d.tickCmd()

	case dashTickMsg2:
		d.loading = true
		return d, tea.Batch(d.spin.Tick, d.fetchCmd())

	case tea.KeyMsg:
		switch msg.String() {
		case "pgup", "k":
			max := len(d.rcTicker) - 1
			if d.rcmScroll < max {
				d.rcmScroll++
			}
		case "pgdown", "j":
			if d.rcmScroll > 0 {
				d.rcmScroll--
			}
		}
		return d, nil

	case dashSpringTickMsg:
		// Advance gap springs toward their targets.
		springing := false
		for i := range d.liveRows {
			r := &d.liveRows[i]
			if math.Abs(r.gapDisplay-r.gapTarget) > 0.001 {
				val, _ := r.gapSpring.Update(r.gapDisplay, 0, r.gapTarget)
				r.gapDisplay = val
				springing = true
			}
		}
		if springing {
			return d, d.springTickCmd()
		}
		return d, nil

	case dashCircuitMsg2:
		d.circuit = msg.circuit
		return d, nil
	}

	return d, nil
}
