// Package f1static fetches historical session data from the F1 live-timing
// static archive (livetiming.formula1.com/static/...).
//
// It is used as a tier-2 fallback: richer than OpenF1 (real lap times, sector
// times, tyre data) and always contains the last completed session through the
// SessionInfo.json discovery endpoint.
package f1static

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"time"

	"github.com/dk-a-dev/termf1/v2/internal/api/liveserver"
)

const baseURL = "https://livetiming.formula1.com/static/"

// Client fetches from the F1 static archive.
type Client struct {
	http *http.Client
}

func NewClient() *Client {
	return &Client{http: &http.Client{Timeout: 10 * time.Second}}
}

// ── raw JSON types ────────────────────────────────────────────────────────────

type sessionInfo struct {
	Meeting struct {
		Name    string `json:"Name"`
		Location string `json:"Location"`
		Country struct {
			Name string `json:"Name"`
			Code string `json:"Code"`
		} `json:"Country"`
		Circuit struct {
			ShortName string `json:"ShortName"`
		} `json:"Circuit"`
	} `json:"Meeting"`
	SessionStatus string `json:"SessionStatus"`
	Type          string `json:"Type"`
	Name          string `json:"Name"`
	Path          string `json:"Path"` // e.g. "2026/2026-03-08_Australian_Grand_Prix/2026-03-08_Race/"
}

type driverEntry struct {
	RacingNumber string `json:"RacingNumber"`
	BroadcastName string `json:"BroadcastName"`
	FullName     string `json:"FullName"`
	Tla          string `json:"Tla"`
	Line         int    `json:"Line"`
	TeamName     string `json:"TeamName"`
	TeamColour   string `json:"TeamColour"`
}

type sector struct {
	Value           string `json:"Value"`
	PreviousValue   string `json:"PreviousValue"`
	OverallFastest  bool   `json:"OverallFastest"`
	PersonalFastest bool   `json:"PersonalFastest"`
	Segments        []struct {
		Status int `json:"Status"`
	} `json:"Segments"`
}

type timingLine struct {
	RacingNumber string   `json:"RacingNumber"`
	Line         int      `json:"Line"`
	Position     string   `json:"Position"`
	GapToLeader  string   `json:"GapToLeader"`
	IntervalToPositionAhead struct {
		Value    string `json:"Value"`
		Catching bool   `json:"Catching"`
	} `json:"IntervalToPositionAhead"`
	LastLapTime struct {
		Value           string `json:"Value"`
		OverallFastest  bool   `json:"OverallFastest"`
		PersonalFastest bool   `json:"PersonalFastest"`
	} `json:"LastLapTime"`
	BestLapTime struct {
		Value string `json:"Value"`
		Lap   int    `json:"Lap"`
	} `json:"BestLapTime"`
	Sectors          []sector `json:"Sectors"`
	NumberOfLaps     int      `json:"NumberOfLaps"`
	NumberOfPitStops int      `json:"NumberOfPitStops"`
	Retired          bool     `json:"Retired"`
	InPit            bool     `json:"InPit"`
	Stopped          bool     `json:"Stopped"`
	ShowPosition     bool     `json:"ShowPosition"`
}

type appStint struct {
	Compound string `json:"Compound"`
	New      string `json:"New"`
	TotalLaps int   `json:"TotalLaps"`
	StartLaps int   `json:"StartLaps"`
}

type appLine struct {
	Stints []appStint `json:"Stints"`
}

type rcMsg struct {
	Utc      string `json:"Utc"`
	Lap      int    `json:"Lap"`
	Category string `json:"Category"`
	Message  string `json:"Message"`
	Flag     string `json:"Flag"`
	Scope    string `json:"Scope"`
}

// ── wrapper types (feed nests data under a single top-level key) ──────────────

type timingDataWrapper struct {
	Lines map[string]timingLine `json:"Lines"`
}

type appDataWrapper struct {
	Lines map[string]appLine `json:"Lines"`
}

type rcmWrapper struct {
	Messages []rcMsg `json:"Messages"`
}

type lapCountWrapper struct {
	CurrentLap int `json:"CurrentLap"`
	TotalLaps  int `json:"TotalLaps"`
}

type radioCapture struct {
	Utc          string `json:"Utc"`
	RacingNumber string `json:"RacingNumber"`
	Path         string `json:"Path"`
}

type teamRadioWrapper struct {
	Captures []radioCapture `json:"Captures"`
}

// ── HTTP helpers ──────────────────────────────────────────────────────────────

func (c *Client) getJSON(ctx context.Context, url string, out interface{}) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, url)
	}
	// F1 static files are UTF-8 with BOM — strip it.
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if len(body) >= 3 && body[0] == 0xEF && body[1] == 0xBB && body[2] == 0xBF {
		body = body[3:]
	}
	return json.Unmarshal(body, out)
}

// ── Public API ────────────────────────────────────────────────────────────────

// FetchLastRace discovers the latest session via SessionInfo.json, then fetches
// timing/driver/tyre/weather/RCM data and converts it into a liveserver.State
// so the dashboard can render it without any layout changes.
func (c *Client) FetchLastRace(ctx context.Context) (*liveserver.State, error) {
	// 1. Discover session path.
	var info sessionInfo
	if err := c.getJSON(ctx, baseURL+"SessionInfo.json", &info); err != nil {
		return nil, fmt.Errorf("SessionInfo: %w", err)
	}
	base := baseURL + info.Path

	// 2. Fetch all needed feeds concurrently.
	var (
		driverMap  map[string]driverEntry
		timingW    timingDataWrapper
		appW       appDataWrapper
		weatherRaw liveserver.WeatherData
		rcW        rcmWrapper
		lapW       lapCountWrapper
		radioW     teamRadioWrapper
	)

	type result struct {
		name string
		err  error
	}
	ch := make(chan result, 7)

	fetch := func(name, url string, dest interface{}) {
		err := c.getJSON(ctx, url, dest)
		ch <- result{name, err}
	}

	go fetch("DriverList",    base+"DriverList.json",          &driverMap)
	go fetch("TimingData",    base+"TimingData.json",          &timingW)
	go fetch("TimingAppData", base+"TimingAppData.json",       &appW)
	go fetch("WeatherData",   base+"WeatherData.json",         &weatherRaw)
	go fetch("RCM",           base+"RaceControlMessages.json", &rcW)
	go fetch("LapCount",      base+"LapCount.json",            &lapW)
	go fetch("TeamRadio",     base+"TeamRadio.json",           &radioW)

	for i := 0; i < 7; i++ {
		r := <-ch
		if r.err != nil && r.name == "DriverList" {
			return nil, fmt.Errorf("%s: %w", r.name, r.err)
		}
		// other errors are non-fatal — partial data is fine
	}

	// 3. Convert to liveserver.State.
	state := &liveserver.State{
		UpdatedAt:     time.Now(),
		SessionStatus: info.SessionStatus,
		SessionInfo: liveserver.SessionInfo{},
		Weather:     weatherRaw,
	}
	state.SessionInfo.Meeting.Name = info.Meeting.Name
	state.SessionInfo.Meeting.Location = info.Meeting.Location
	state.SessionInfo.Meeting.Country.Name = info.Meeting.Country.Name
	state.SessionInfo.Meeting.Country.Code = info.Meeting.Country.Code
	state.SessionInfo.Meeting.Circuit.ShortName = info.Meeting.Circuit.ShortName
	state.SessionInfo.Name = info.Name
	state.SessionInfo.Type = info.Type

	// Drivers
	state.Drivers = make(map[string]*liveserver.DriverInfo, len(driverMap))
	for num, d := range driverMap {
		state.Drivers[num] = &liveserver.DriverInfo{
			RacingNumber:  d.RacingNumber,
			BroadcastName: d.BroadcastName,
			FullName:      d.FullName,
			Tla:           d.Tla,
			TeamName:      d.TeamName,
			TeamColour:    d.TeamColour,
		}
	}

	// Timing — round-trip sectors through JSON to satisfy liveserver's anonymous struct type.
	state.Timing = make(map[string]*liveserver.DriverTiming, len(timingW.Lines))
	for num, t := range timingW.Lines {
		dt := &liveserver.DriverTiming{
			RacingNumber:     t.RacingNumber,
			Line:             t.Line,
			GapToLeader:      t.GapToLeader,
			NumberOfLaps:     t.NumberOfLaps,
			NumberOfPitStops: t.NumberOfPitStops,
			Retired:          t.Retired,
			InPit:            t.InPit,
			Stopped:          t.Stopped,
			ShowPosition:     t.ShowPosition,
		}
		dt.IntervalToPositionAhead.Value = t.IntervalToPositionAhead.Value
		dt.IntervalToPositionAhead.Catching = t.IntervalToPositionAhead.Catching
		dt.LastLapTime.Value = t.LastLapTime.Value
		dt.LastLapTime.OverallFastest = t.LastLapTime.OverallFastest
		dt.LastLapTime.PersonalFastest = t.LastLapTime.PersonalFastest
		dt.BestLapTime.Value = t.BestLapTime.Value
		dt.BestLapTime.Lap = t.BestLapTime.Lap
		// Sectors: marshal local type → unmarshal into liveserver anonymous struct slice.
		if sBytes, err := json.Marshal(t.Sectors); err == nil {
			_ = json.Unmarshal(sBytes, &dt.Sectors)
		}
		state.Timing[num] = dt
	}

	// Tyre stints — use the last stint per driver.
	if len(appW.Lines) > 0 {
		state.Stints = make(map[string]*liveserver.TyreData, len(appW.Lines))
		for num, al := range appW.Lines {
			if len(al.Stints) == 0 {
				continue
			}
			// sort by StartLaps desc to get current stint
			sort.Slice(al.Stints, func(i, j int) bool {
				return al.Stints[i].StartLaps > al.Stints[j].StartLaps
			})
			last := al.Stints[0]
			state.Stints[num] = &liveserver.TyreData{
				Compound:  last.Compound,
				TotalLaps: last.TotalLaps,
				New:       last.New == "true",
			}
		}
	}

	// Race control messages
	for _, m := range rcW.Messages {
		state.RaceControlMessages = append(state.RaceControlMessages, liveserver.RaceControlMsg{
			Utc:      m.Utc,
			Lap:      m.Lap,
			Category: m.Category,
			Message:  m.Message,
			Flag:     m.Flag,
			Scope:    m.Scope,
		})
	}

	// LapCount
	state.LapCount = liveserver.LapCount{
		CurrentLap: lapW.CurrentLap,
		TotalLaps:  lapW.TotalLaps,
	}

	// TeamRadio — resolve full audio URLs
	for _, clip := range radioW.Captures {
		state.TeamRadio = append(state.TeamRadio, liveserver.TeamRadioCapture{
			Utc:          clip.Utc,
			RacingNumber: clip.RacingNumber,
			AudioURL:     baseURL + info.Path + clip.Path,
		})
	}

	return state, nil
}
