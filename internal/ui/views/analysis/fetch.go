package analysis

import (
	"context"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/dk-a-dev/termf1/v2/internal/api/openf1"
)

// FetchCmd returns a Cmd that fetches session data sequentially from OpenF1,
// with 350 ms delays between requests to respect the 3 req/s rate limit.
func FetchCmd(of1 *openf1.Client) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		sess, err := of1.GetLatestSession(ctx)
		if err != nil {
			return ErrMsg{err}
		}

		// Sequential fetches — concurrent calls hit OpenF1's 3 req/s limit,
		// causing empty responses that make charts appear blank.
		drivers, _ := of1.GetDrivers(ctx, sess.SessionKey)
		time.Sleep(350 * time.Millisecond)

		laps, _ := of1.GetLaps(ctx, sess.SessionKey)
		time.Sleep(350 * time.Millisecond)

		stints, _ := of1.GetStints(ctx, sess.SessionKey)
		time.Sleep(350 * time.Millisecond)

		positions, _ := of1.GetPositions(ctx, sess.SessionKey)
		time.Sleep(350 * time.Millisecond)

		pits, _ := of1.GetPits(ctx, sess.SessionKey)

		return DataMsg{
			Session:   sess,
			Drivers:   drivers,
			Laps:      laps,
			Stints:    stints,
			Positions: positions,
			Pits:      pits,
		}
	}
}
