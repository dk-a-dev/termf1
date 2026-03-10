// Package liveserver provides a client for the termf1-server REST API.
// When the server is unavailable it transparently falls back to OpenF1.
package liveserver

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const defaultAddr = "http://localhost:8765"

// State mirrors the JSON returned by GET /state on termf1-server.
type State struct {
	UpdatedAt     time.Time              `json:"updated_at"`
	SessionStatus string                 `json:"session_status"`
	SessionInfo   SessionInfo            `json:"session_info"`
	LapCount      LapCount               `json:"lap_count"`
	Timing        map[string]*DriverTiming `json:"timing"`
	Drivers       map[string]*DriverInfo   `json:"drivers"`
	CarData       map[string]*CarData      `json:"car_data"`
	Positions     map[string]*Position     `json:"positions"`
	RaceControlMessages []RaceControlMsg   `json:"race_control_messages"`
	Weather       WeatherData              `json:"weather"`
	Stints        map[string]*TyreData     `json:"stints"`
	TeamRadio     []TeamRadioCapture       `json:"team_radio,omitempty"`
}

// TeamRadioCapture represents one team radio clip.
type TeamRadioCapture struct {
	Utc          string `json:"utc"`
	RacingNumber string `json:"racing_number"`
	AudioURL     string `json:"audio_url"` // full https:// URL to the MP3
}

type SessionInfo struct {
	Meeting struct {
		Name     string `json:"Name"`
		Location string `json:"Location"`
		Country  struct {
			Name string `json:"Name"`
			Code string `json:"Code"`
		} `json:"Country"`
		Circuit struct {
			ShortName string `json:"ShortName"`
		} `json:"Circuit"`
	} `json:"Meeting"`
	Name      string `json:"Name"`
	Type      string `json:"Type"`
	StartDate string `json:"StartDate"`
	EndDate   string `json:"EndDate"`
}

type LapCount struct {
	CurrentLap int `json:"CurrentLap"`
	TotalLaps  int `json:"TotalLaps"`
}

type DriverTiming struct {
	RacingNumber string `json:"RacingNumber"`
	Line         int    `json:"Line"`
	GapToLeader  string `json:"GapToLeader"`
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
	Sectors []struct {
		Value           string `json:"Value"`
		OverallFastest  bool   `json:"OverallFastest"`
		PersonalFastest bool   `json:"PersonalFastest"`
		PreviousValue   string `json:"PreviousValue"`
		Segments        []struct {
			Status int `json:"Status"`
		} `json:"Segments"`
	} `json:"Sectors"`
	Speeds struct {
		ST struct{ Value string `json:"Value"` } `json:"ST"`
	} `json:"Speeds"`
	InPit            bool `json:"InPit"`
	Stopped          bool `json:"Stopped"`
	NumberOfLaps     int  `json:"NumberOfLaps"`
	NumberOfPitStops int  `json:"NumberOfPitStops"`
	Retired          bool `json:"Retired"`
	ShowPosition     bool `json:"ShowPosition"`
}

type DriverInfo struct {
	RacingNumber  string `json:"RacingNumber"`
	BroadcastName string `json:"BroadcastName"`
	FullName      string `json:"FullName"`
	Tla           string `json:"Tla"`
	TeamName      string `json:"TeamName"`
	TeamColour    string `json:"TeamColour"`
}

type CarData struct {
	Channels struct {
		RPM      int `json:"0"`
		Speed    int `json:"2"`
		Gear     int `json:"3"`
		Throttle int `json:"4"`
		Brake    int `json:"5"`
		DRS      int `json:"45"`
	} `json:"Channels"`
	Utc string `json:"Utc"`
}

type Position struct {
	Status string  `json:"Status"`
	X      float64 `json:"X"`
	Y      float64 `json:"Y"`
	Z      float64 `json:"Z"`
}

type RaceControlMsg struct {
	Utc      string `json:"Utc"`
	Lap      int    `json:"Lap"`
	Category string `json:"Category"`
	Message  string `json:"Message"`
	Flag     string `json:"Flag"`
	Scope    string `json:"Scope"`
}

// TyreData holds the current tyre stint info for one driver.
type TyreData struct {
	Compound  string `json:"Compound"`
	TotalLaps int    `json:"TotalLaps"`
	New       bool   `json:"New"`
}

type WeatherData struct {
	AirTemp       string `json:"AirTemp"`
	Humidity      string `json:"Humidity"`
	Pressure      string `json:"Pressure"`
	Rainfall      string `json:"Rainfall"`
	TrackTemp     string `json:"TrackTemp"`
	WindDirection string `json:"WindDirection"`
	WindSpeed     string `json:"WindSpeed"`
}

// Client is an HTTP client for termf1-server.
type Client struct {
	addr string
	http *http.Client
}

// New returns a Client pointing at addr (default: http://localhost:8765).
func New(addr string) *Client {
	if addr == "" {
		addr = defaultAddr
	}
	return &Client{
		addr: addr,
		http: &http.Client{Timeout: 3 * time.Second},
	}
}

// IsAlive returns true if the server is reachable.
func (c *Client) IsAlive(ctx context.Context) bool {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.addr+"/health", nil)
	if err != nil {
		return false
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return false
	}
	resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

// GetState fetches the full state snapshot.
func (c *Client) GetState(ctx context.Context) (*State, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.addr+"/state", nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("termf1-server unreachable: %w", err)
	}
	defer resp.Body.Close()
	var s State
	if err := json.NewDecoder(resp.Body).Decode(&s); err != nil {
		return nil, err
	}
	return &s, nil
}

// IsLive returns true if the state contains active session data
// (timing or session status is set and non-empty).
func IsLive(s *State) bool {
	if s == nil {
		return false
	}
	if s.SessionStatus != "" && s.SessionStatus != "Inactive" {
		return true
	}
	if len(s.Timing) > 0 {
		return true
	}
	return false
}
