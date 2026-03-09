package openf1

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

const baseURL = "https://api.openf1.org/v1"

// Client is an HTTP client for the OpenF1 API.
type Client struct {
	http    *http.Client
	baseURL string
}

func NewClient() *Client {
	return &Client{
		http:    &http.Client{Timeout: 15 * time.Second},
		baseURL: baseURL,
	}
}

func (c *Client) get(ctx context.Context, path string, params map[string]string, result interface{}) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return fmt.Errorf("building request: %w", err)
	}
	q := req.URL.Query()
	for k, v := range params {
		q.Set(k, v)
	}
	req.URL.RawQuery = q.Encode()

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API returned HTTP %d for %s", resp.StatusCode, path)
	}
	return json.NewDecoder(resp.Body).Decode(result)
}

func skParam(sessionKey int) map[string]string {
	return map[string]string{"session_key": fmt.Sprintf("%d", sessionKey)}
}

// GetLatestSession returns information about the current or most-recent session.
func (c *Client) GetLatestSession(ctx context.Context) (Session, error) {
	var sessions []Session
	err := c.get(ctx, "/sessions", map[string]string{"session_key": "latest"}, &sessions)
	if err != nil || len(sessions) == 0 {
		return Session{}, err
	}
	return sessions[len(sessions)-1], nil
}

func (c *Client) GetDrivers(ctx context.Context, sessionKey int) ([]Driver, error) {
	var out []Driver
	return out, c.get(ctx, "/drivers", skParam(sessionKey), &out)
}

func (c *Client) GetIntervals(ctx context.Context, sessionKey int) ([]Interval, error) {
	var out []Interval
	return out, c.get(ctx, "/intervals", skParam(sessionKey), &out)
}

func (c *Client) GetLaps(ctx context.Context, sessionKey int) ([]Lap, error) {
	var out []Lap
	return out, c.get(ctx, "/laps", skParam(sessionKey), &out)
}

func (c *Client) GetPositions(ctx context.Context, sessionKey int) ([]Position, error) {
	var out []Position
	return out, c.get(ctx, "/position", skParam(sessionKey), &out)
}

func (c *Client) GetWeather(ctx context.Context, sessionKey int) ([]Weather, error) {
	var out []Weather
	return out, c.get(ctx, "/weather", skParam(sessionKey), &out)
}

func (c *Client) GetStints(ctx context.Context, sessionKey int) ([]Stint, error) {
	var out []Stint
	return out, c.get(ctx, "/stints", skParam(sessionKey), &out)
}

func (c *Client) GetPits(ctx context.Context, sessionKey int) ([]Pit, error) {
	var out []Pit
	return out, c.get(ctx, "/pit", skParam(sessionKey), &out)
}

func (c *Client) GetRaceControl(ctx context.Context, sessionKey int) ([]RaceControl, error) {
	var out []RaceControl
	return out, c.get(ctx, "/race_control", skParam(sessionKey), &out)
}

// DashboardPayload aggregates all data needed to render the live timing screen.
type DashboardPayload struct {
	Session      Session
	Drivers      []Driver
	Laps         []Lap
	Intervals    []Interval
	Positions    []Position
	Stints       []Stint
	Pits         []Pit
	RaceControls []RaceControl
	Weathers     []Weather
}

// FetchDashboard fetches all dashboard data concurrently.
func (c *Client) FetchDashboard(ctx context.Context) (*DashboardPayload, error) {
	session, err := c.GetLatestSession(ctx)
	if err != nil {
		return nil, fmt.Errorf("session: %w", err)
	}
	sk := session.SessionKey

	type result struct {
		key string
		val interface{}
		err error
	}
	ch := make(chan result, 8)
	var wg sync.WaitGroup

	fetch := func(key string, fn func() (interface{}, error)) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			v, e := fn()
			ch <- result{key, v, e}
		}()
	}

	fetch("drivers", func() (interface{}, error) { return c.GetDrivers(ctx, sk) })
	fetch("laps", func() (interface{}, error) { return c.GetLaps(ctx, sk) })
	fetch("intervals", func() (interface{}, error) { return c.GetIntervals(ctx, sk) })
	fetch("positions", func() (interface{}, error) { return c.GetPositions(ctx, sk) })
	fetch("stints", func() (interface{}, error) { return c.GetStints(ctx, sk) })
	fetch("pits", func() (interface{}, error) { return c.GetPits(ctx, sk) })
	fetch("race_control", func() (interface{}, error) { return c.GetRaceControl(ctx, sk) })
	fetch("weather", func() (interface{}, error) { return c.GetWeather(ctx, sk) })

	go func() { wg.Wait(); close(ch) }()

	p := &DashboardPayload{Session: session}
	for r := range ch {
		if r.err != nil {
			continue // partial data is fine; skip failed endpoint
		}
		switch r.key {
		case "drivers":
			p.Drivers = r.val.([]Driver)
		case "laps":
			p.Laps = r.val.([]Lap)
		case "intervals":
			p.Intervals = r.val.([]Interval)
		case "positions":
			p.Positions = r.val.([]Position)
		case "stints":
			p.Stints = r.val.([]Stint)
		case "pits":
			p.Pits = r.val.([]Pit)
		case "race_control":
			p.RaceControls = r.val.([]RaceControl)
		case "weather":
			p.Weathers = r.val.([]Weather)
		}
	}
	return p, nil
}
