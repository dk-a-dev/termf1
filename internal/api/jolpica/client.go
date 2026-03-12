package jolpica

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/dk-a-dev/termf1/internal/api/cache"
)

const baseURL = "https://api.jolpi.ca/ergast/f1"

// Client is an HTTP client for the Jolpica (Ergast-compatible) F1 API.
type Client struct {
	http    *http.Client
	baseURL string
	cache   *cache.Cache
}

func NewClient() *Client {
	cc, _ := cache.New() // Ignore error, if cache fails we just won't cache (Wait, actually we need to handle cache being nil)
	return &Client{
		http:    &http.Client{Timeout: 15 * time.Second},
		baseURL: baseURL,
		cache:   cc,
	}
}

func (c *Client) get(ctx context.Context, path string, result interface{}) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return fmt.Errorf("building request: %w", err)
	}
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

// GetDriverStandings returns the current season driver championship standings.
func (c *Client) GetDriverStandings(ctx context.Context) ([]DriverStanding, error) {
	var r Response
	fetch := func() error {
		return c.get(ctx, "/current/driverStandings.json", &r)
	}
	
	var err error
	if c.cache != nil {
		err = c.cache.Get(c.baseURL+"/current/driverStandings.json", 5*time.Minute, &r, fetch)
	} else {
		err = fetch()
	}
	
	if err != nil {
		return nil, err
	}
	if r.MRData.StandingsTable == nil || len(r.MRData.StandingsTable.StandingsLists) == 0 {
		return nil, nil
	}
	return r.MRData.StandingsTable.StandingsLists[0].DriverStandings, nil
}

// GetConstructorStandings returns the current season constructor standings.
func (c *Client) GetConstructorStandings(ctx context.Context) ([]ConstructorStanding, error) {
	var r Response
	fetch := func() error {
		return c.get(ctx, "/current/constructorStandings.json", &r)
	}

	var err error
	if c.cache != nil {
		err = c.cache.Get(c.baseURL+"/current/constructorStandings.json", 5*time.Minute, &r, fetch)
	} else {
		err = fetch()
	}

	if err != nil {
		return nil, err
	}
	if r.MRData.StandingsTable == nil || len(r.MRData.StandingsTable.StandingsLists) == 0 {
		return nil, nil
	}
	return r.MRData.StandingsTable.StandingsLists[0].ConstructorStandings, nil
}

// GetSchedule returns all races in the current season.
func (c *Client) GetSchedule(ctx context.Context) ([]Race, error) {
	var r Response
	fetch := func() error {
		return c.get(ctx, "/current.json", &r)
	}

	var err error
	if c.cache != nil {
		err = c.cache.Get(c.baseURL+"/current.json", 24*time.Hour, &r, fetch)
	} else {
		err = fetch()
	}

	if err != nil {
		return nil, err
	}
	if r.MRData.RaceTable == nil {
		return nil, nil
	}
	return r.MRData.RaceTable.Races, nil
}
