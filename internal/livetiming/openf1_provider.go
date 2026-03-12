package livetiming

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/dk-a-dev/termf1/internal/api/openf1"
)

// OpenF1Provider implements StreamProvider by periodically polling the OpenF1 API
// and translating its DashboardPayload to the unified State format.
type OpenF1Provider struct {
	state  *State
	client *openf1.Client
	logger *log.Logger
}

func NewOpenF1Provider(state *State, client *openf1.Client, logger *log.Logger) *OpenF1Provider {
	return &OpenF1Provider{
		state:  state,
		client: client,
		logger: logger,
	}
}

func (p *OpenF1Provider) Run(ctx context.Context) error {
	p.logger.Println("[openf1] starting polling fallback")
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	if err := p.fetchAndApply(ctx); err != nil {
		p.logger.Printf("[openf1] initial fetch failed: %v", err)
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			if err := p.fetchAndApply(ctx); err != nil {
				p.logger.Printf("[openf1] fetch failed: %v", err)
			}
		}
	}
}

func (p *OpenF1Provider) fetchAndApply(ctx context.Context) error {
	payload, err := p.client.FetchDashboard(ctx)
	if err != nil {
		return err
	}

	p.state.mu.Lock()
	defer p.state.mu.Unlock()

	p.state.SessionInfo.Meeting.Name = payload.Session.CountryName
	p.state.SessionInfo.Meeting.Circuit.ShortName = payload.Session.CircuitShortName
	p.state.SessionInfo.Name = payload.Session.SessionName
	p.state.SessionStatus = "Playback"

	if p.state.Drivers == nil {
		p.state.Drivers = make(map[string]*DriverInfo)
	}
	if p.state.Timing == nil {
		p.state.Timing = make(map[string]*DriverTiming)
	}
	if p.state.CarData == nil {
		p.state.CarData = make(map[string]*CarData)
	}
	if p.state.Stints == nil {
		p.state.Stints = make(map[string]*TyreStint)
	}

	for _, rawD := range payload.Drivers {
		numStr := strconv.Itoa(rawD.DriverNumber)
		p.state.Drivers[numStr] = &DriverInfo{
			RacingNumber:  numStr,
			BroadcastName: rawD.NameAcronym,
			FullName:      rawD.FullName,
			Tla:           rawD.NameAcronym,
			TeamName:      rawD.TeamName,
			TeamColour:    rawD.TeamColour,
		}
	}

	bestLapMap := make(map[int]float64)
	lastLapMap := make(map[int]openf1.Lap)
	lapNumMap := make(map[int]int)
	
	for _, l := range payload.Laps {
		if l.LapDuration > 0 {
			if cur, ok := bestLapMap[l.DriverNumber]; !ok || l.LapDuration < cur {
				bestLapMap[l.DriverNumber] = l.LapDuration
			}
		}
		if l.LapNumber > lapNumMap[l.DriverNumber] {
			lapNumMap[l.DriverNumber] = l.LapNumber
			lastLapMap[l.DriverNumber] = l
		}
	}

	gapMap := make(map[int]openf1.Interval)
	for _, iv := range payload.Intervals {
		gapMap[iv.DriverNumber] = iv
	}
	
	pitCount := make(map[int]int)
	for _, pit := range payload.Pits {
		pitCount[pit.DriverNumber]++
	}

	for _, sr := range payload.SessionResults {
		numStr := strconv.Itoa(sr.DriverNumber)
		
		iv := gapMap[sr.DriverNumber]
		bestLap := bestLapMap[sr.DriverNumber]
		lastLap := lastLapMap[sr.DriverNumber]
		
		gapStr := formatVal(sr.GapToLeader)
		if gapStr == "" || gapStr == "–" {
			gapStr = formatVal(iv.GapToLeader)
		}

		timing := &DriverTiming{
			RacingNumber: numStr,
			Line:         sr.Position,
			GapToLeader:  gapStr,
			NumberOfLaps: sr.NumberOfLaps,
			Retired:      sr.DNF || sr.DNS || sr.DSQ,
			NumberOfPitStops: pitCount[sr.DriverNumber],
		}
		timing.IntervalToPositionAhead.Value = formatVal(iv.Interval)
		
		if bestLap > 0 {
			timing.BestLapTime.Value = fmt.Sprintf("%.3f", bestLap)
		}
		if lastLap.LapDuration > 0 {
			timing.LastLapTime.Value = fmt.Sprintf("%.3f", lastLap.LapDuration)
		}
		timing.Speeds.ST.Value = strconv.Itoa(lastLap.StSpeed)

		timing.Sectors = make([]SectorTiming, 3)
		
		timing.Sectors[0].Value = fmt.Sprintf("%.3f", lastLap.DurationSector1)
		timing.Sectors[1].Value = fmt.Sprintf("%.3f", lastLap.DurationSector2)
		timing.Sectors[2].Value = fmt.Sprintf("%.3f", lastLap.DurationSector3)

		p.state.Timing[numStr] = timing
	}

	for _, st := range payload.Stints {
		numStr := strconv.Itoa(st.DriverNumber)
		p.state.Stints[numStr] = &TyreStint{
			Compound:  strings.ToUpper(st.Compound),
			TotalLaps: st.TyreAgeAtStart - st.LapStart + 1, // rough approximation
			New:       false,
		}
	}

	if len(payload.Weathers) > 0 {
		w := payload.Weathers[len(payload.Weathers)-1]
		p.state.Weather.AirTemp = fmt.Sprintf("%.1f", w.AirTemperature)
		p.state.Weather.WindSpeed = fmt.Sprintf("%.1f", w.WindSpeed)
		p.state.Weather.Rainfall = fmt.Sprintf("%d", w.Rainfall)
	}

	var msgs []RaceControlMessage
	for _, m := range payload.RaceControls {
		msgs = append(msgs, RaceControlMessage{
			Message: m.Message,
		})
	}
	p.state.RaceControlMessages = msgs

	p.state.UpdatedAt = time.Now()
	p.logger.Printf("[openf1] state fully populated from payload. Updates applied.")

	return nil
}

func formatVal(v interface{}) string {
	if v == nil {
		return "LEADER"
	}
	switch val := v.(type) {
	case float64:
		if val == 0 {
			return "LEADER"
		}
		return fmt.Sprintf("+%.3f", val)
	case string:
		return val
	}
	return "—"
}
