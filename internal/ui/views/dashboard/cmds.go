package dashboard

import (
	"context"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/dk-a-dev/termf1/internal/api/liveserver"
	"github.com/dk-a-dev/termf1/internal/api/multiviewer"
	"github.com/dk-a-dev/termf1/internal/api/openf1"
)

// fetchCmd implements a three-tier data strategy:
//  1. Live server alive + has timing data  → dashLiveMsg   (live session)
//  2. Live server alive but idle/empty     → f1static last race → dashLiveMsg (server-alive indicator stays green)
//     falls back to OpenF1 if f1static fails
//  3. Live server offline                  → f1static last race → dashLiveMsg
//     falls back to OpenF1 (dashFallbackMsg) if f1static fails
func (d *Dashboard2) fetchCmd() tea.Cmd {
	ls  := d.lsClient
	f1s := d.f1s
	of1 := d.of1Client
	return func() tea.Msg {
		// ── Tier 1: probe live server ──────────────────────────────────────────
		probeCtx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
		defer cancel()

		serverUp := ls.IsAlive(probeCtx)
		if serverUp {
			state, err := ls.GetState(probeCtx)
			if err == nil && liveserver.IsLive(state) {
				return dashLiveMsg{state} // active session
			}
		}

		// ── Tier 2: F1 static archive (historical, richest data) ──────────────
		staticCtx, staticCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer staticCancel()
		if state, err := f1s.FetchLastRace(staticCtx); err == nil {
			return dashLiveMsg{state}
		}

		// ── Tier 3: OpenF1 fallback ────────────────────────────────────────────
		of1Ctx, of1Cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer of1Cancel()
		p, err := of1.FetchDashboard(of1Ctx)
		if err != nil {
			return dashErrMsg2{err}
		}
		if serverUp {
			return dashIdleFallbackMsg{p}
		}
		return dashFallbackMsg{p}
	}
}

// tickCmd schedules the next data refresh.
// When we're only showing historical data (server offline) we poll much less
// frequently — session results don't change between GPs.
func (d *Dashboard2) tickCmd() tea.Cmd {
	interval := time.Duration(d.refreshSec) * time.Second
	if !d.serverAlive && d.fallbackFetched {
		interval = 5 * time.Minute // historical data doesn't change; be kind to OpenF1
	}
	return tea.Tick(interval, func(t time.Time) tea.Msg {
		return dashTickMsg2(t)
	})
}

// springTickCmd schedules the next spring animation frame (~60 fps).
func (d *Dashboard2) springTickCmd() tea.Cmd {
	return tea.Tick(16*time.Millisecond, func(t time.Time) tea.Msg {
		return dashSpringTickMsg{}
	})
}

// fetchFallbackForIdle fetches the last race from OpenF1 to fill the timing
// panel when the server is connected but has no active session. It returns a
// dashIdleFallbackMsg so the server-alive indicator is not affected.
func fetchFallbackForIdle(of1 *openf1.Client) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
		p, err := of1.FetchDashboard(ctx)
		if err != nil {
			return nil
		}
		return dashIdleFallbackMsg{p}
	}
}

// fetchCircuitForDash fetches GPS circuit data for the current live session.
func fetchCircuitForDash(mv *multiviewer.Client, st *liveserver.State) tea.Cmd {
	return func() tea.Msg {
		if st == nil {
			return nil
		}
		year := time.Now().Year()
		circKey := 0
		if st.SessionInfo.Meeting.Circuit.ShortName != "" {
			circKey = multiviewer.CircuitKeyByName[st.SessionInfo.Meeting.Circuit.ShortName]
		}
		if circKey == 0 {
			return nil
		}
		c, err := mv.GetCircuit(context.Background(), circKey, year)
		if err != nil {
			return dashCircuitMsg2{}
		}
		return dashCircuitMsg2{c}
	}
}

// fetchCircuitForDashFallback fetches GPS circuit data for a historical session.
func fetchCircuitForDashFallback(mv *multiviewer.Client, p *openf1.DashboardPayload) tea.Cmd {
	return func() tea.Msg {
		if p == nil {
			return nil
		}
		year := p.Session.Year
		circKey := multiviewer.CircuitKeyByName[p.Session.CircuitShortName]
		if circKey == 0 {
			circKey = p.Session.CircuitKey
		}
		if circKey == 0 {
			return nil
		}
		c, err := mv.GetCircuit(context.Background(), circKey, year)
		if err != nil {
			return dashCircuitMsg2{}
		}
		return dashCircuitMsg2{c}
	}
}
