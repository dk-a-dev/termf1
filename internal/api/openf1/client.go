package openf1

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/dk-a-dev/termf1/v2/internal/api/cache"
)

const baseURL = "https://api.openf1.org/v1"

// Client is an HTTP client for the OpenF1 API.
type Client struct {
	http    *http.Client
	baseURL string
	cache   *cache.Cache
}

func NewClient() *Client {
	cc, _ := cache.New()
	return &Client{
		http:    &http.Client{Timeout: 15 * time.Second},
		baseURL: baseURL,
		cache:   cc,
	}
}

func (c *Client) get(ctx context.Context, path string, params map[string]string, result interface{}, ttl time.Duration) error {
	fetch := func() error {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
		if err != nil {
			return fmt.Errorf("building request: %w", err)
		}
		q := req.URL.Query()
		var rawExtras []string
		for k, v := range params {
			if strings.HasSuffix(k, ">") || strings.HasSuffix(k, "<") {
				rawExtras = append(rawExtras, k+v)
			} else {
				q.Set(k, v)
			}
		}
		rawQ := q.Encode()
		for _, extra := range rawExtras {
			if rawQ != "" {
				rawQ += "&"
			}
			rawQ += extra
		}
		req.URL.RawQuery = rawQ

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

	if c.cache != nil && ttl > 0 {
		cacheKey := c.baseURL + path + fmt.Sprintf("%v", params)
		return c.cache.Get(cacheKey, ttl, result, fetch)
	}

	return fetch()
}

func skParam(sessionKey int) map[string]string {
	return map[string]string{"session_key": fmt.Sprintf("%d", sessionKey)}
}

// GetLatestSession returns information about the current or most-recent session.
func (c *Client) GetLatestSession(ctx context.Context) (Session, error) {
	var sessions []Session
	// Very short TTL for "latest" resolution
	err := c.get(ctx, "/sessions", map[string]string{"session_key": "latest"}, &sessions, 1*time.Minute)
	if err != nil || len(sessions) == 0 {
		return Session{}, err
	}
	return sessions[len(sessions)-1], nil
}

func (c *Client) GetDrivers(ctx context.Context, sessionKey int) ([]Driver, error) {
	var out []Driver
	// Drivers are very static for a given session
	return out, c.get(ctx, "/drivers", skParam(sessionKey), &out, 24*time.Hour)
}

func (c *Client) GetIntervals(ctx context.Context, sessionKey int) ([]Interval, error) {
	var out []Interval
	return out, c.get(ctx, "/intervals", skParam(sessionKey), &out, 0)
}

func (c *Client) GetLaps(ctx context.Context, sessionKey int) ([]Lap, error) {
	var out []Lap
	return out, c.get(ctx, "/laps", skParam(sessionKey), &out, 5*time.Minute)
}

func (c *Client) GetPositions(ctx context.Context, sessionKey int) ([]Position, error) {
	var out []Position
	return out, c.get(ctx, "/position", skParam(sessionKey), &out, 0)
}

func (c *Client) GetWeather(ctx context.Context, sessionKey int) ([]Weather, error) {
	var out []Weather
	return out, c.get(ctx, "/weather", skParam(sessionKey), &out, 5*time.Minute)
}

func (c *Client) GetStints(ctx context.Context, sessionKey int) ([]Stint, error) {
	var out []Stint
	return out, c.get(ctx, "/stints", skParam(sessionKey), &out, 5*time.Minute)
}

func (c *Client) GetPits(ctx context.Context, sessionKey int) ([]Pit, error) {
	var out []Pit
	return out, c.get(ctx, "/pit", skParam(sessionKey), &out, 5*time.Minute)
}

func (c *Client) GetRaceControl(ctx context.Context, sessionKey int) ([]RaceControl, error) {
	var out []RaceControl
	return out, c.get(ctx, "/race_control", skParam(sessionKey), &out, 5*time.Minute)
}

func (c *Client) GetSessionResults(ctx context.Context, sessionKey int) ([]SessionResult, error) {
	var out []SessionResult
	return out, c.get(ctx, "/session_result", skParam(sessionKey), &out, 1*time.Minute)
}

// GetCarData returns per-sample car telemetry for a driver within an optional
// date range. dateGt / dateLt are ISO-8601 strings; pass empty strings to omit.
func (c *Client) GetCarData(ctx context.Context, sessionKey, driverNumber int, dateGt, dateLt string) ([]CarData, error) {
	params := map[string]string{
		"session_key":   fmt.Sprintf("%d", sessionKey),
		"driver_number": fmt.Sprintf("%d", driverNumber),
	}
	if dateGt != "" {
		params["date>"] = dateGt
	}
	if dateLt != "" {
		params["date<"] = dateLt
	}
	var out []CarData
	return out, c.get(ctx, "/car_data", params, &out, 0)
}

// GetCarLocations returns per-sample GPS location data for a driver within an
// optional date range. dateGt / dateLt are ISO-8601 strings; pass empty strings
// to omit. The /location endpoint does not support lap_number filtering.
func (c *Client) GetCarLocations(ctx context.Context, sessionKey, driverNumber int, dateGt, dateLt string) ([]CarLocation, error) {
	params := map[string]string{
		"session_key":   fmt.Sprintf("%d", sessionKey),
		"driver_number": fmt.Sprintf("%d", driverNumber),
	}
	if dateGt != "" {
		params["date>"] = dateGt
	}
	if dateLt != "" {
		params["date<"] = dateLt
	}
	var out []CarLocation
	return out, c.get(ctx, "/location", params, &out, 0)
}

// GetOvertakes returns a list of overtakes for the specified session.
// This is only available during races.
func (c *Client) GetOvertakes(ctx context.Context, sessionKey int) ([]Overtake, error) {
	var out []Overtake
	return out, c.get(ctx, "/overtakes", skParam(sessionKey), &out, 5*time.Minute)
}

// DashboardPayload aggregates all data needed to render the live timing screen.
type DashboardPayload struct {
	Session        Session
	Drivers        []Driver
	Laps           []Lap
	Intervals      []Interval
	SessionResults []SessionResult
	Stints         []Stint
	Pits           []Pit
	RaceControls   []RaceControl
	Weathers       []Weather
}

// FetchDashboard fetches all dashboard data sequentially with a small delay
// between requests to stay within OpenF1's rate limit (~3 req/s).
func (c *Client) FetchDashboard(ctx context.Context) (*DashboardPayload, error) {
	session, err := c.GetLatestSession(ctx)
	if err != nil {
		return nil, fmt.Errorf("session: %w", err)
	}
	sk := session.SessionKey
	p := &DashboardPayload{Session: session}

	// 350 ms between each call → ~2.8 req/s, safely under the 3/s limit.
	// Each call checks ctx so the whole thing can be cancelled.
	type call struct {
		name string
		fn   func() error
	}
	calls := []call{
		{"drivers", func() error {
			v, e := c.GetDrivers(ctx, sk); p.Drivers = v; return e
		}},
		{"results", func() error {
			v, e := c.GetSessionResults(ctx, sk); p.SessionResults = v; return e
		}},
		{"stints", func() error {
			v, e := c.GetStints(ctx, sk); p.Stints = v; return e
		}},
		{"pits", func() error {
			v, e := c.GetPits(ctx, sk); p.Pits = v; return e
		}},
		{"laps", func() error {
			v, e := c.GetLaps(ctx, sk); p.Laps = v; return e
		}},
		{"intervals", func() error {
			v, e := c.GetIntervals(ctx, sk); p.Intervals = v; return e
		}},
		{"weather", func() error {
			v, e := c.GetWeather(ctx, sk); p.Weathers = v; return e
		}},
		{"race_control", func() error {
			v, e := c.GetRaceControl(ctx, sk); p.RaceControls = v; return e
		}},
	}

	for i, c := range calls {
		if err := ctx.Err(); err != nil {
			break // context cancelled
		}
		_ = c.fn() // partial data on error is fine
		if i < len(calls)-1 {
			select {
			case <-ctx.Done():
				return p, nil
			case <-time.After(350 * time.Millisecond):
			}
		}
	}
	return p, nil
}
