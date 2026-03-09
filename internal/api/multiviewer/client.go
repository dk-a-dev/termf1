// Package multiviewer provides a client for the Multiviewer F1 API,
// which exposes real circuit coordinate data (x/y GPS path arrays and
// corner annotations) scraped from the F1 timing system.
package multiviewer

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const baseURL = "https://api.multiviewer.app/api/v1"

// Circuit holds the track coordinate data returned by the Multiviewer API.
// X and Y are arrays of integer GPS-like coordinates forming the closed track
// outline (sampled ~8 m apart for a typical 5 km circuit → ~600 points).
type Circuit struct {
	CircuitKey          int      `json:"circuitKey"`
	CircuitName         string   `json:"circuitName"`
	Location            string   `json:"location"`
	MeetingName         string   `json:"meetingName"`
	MeetingOfficialName string   `json:"meetingOfficialName"`
	Rotation            float64  `json:"rotation"`
	X                   []int    `json:"x"`
	Y                   []int    `json:"y"`
	Corners             []Corner `json:"corners"`
	Year                int      `json:"year"`
}

// Corner is a labelled corner on the circuit with its track position.
type Corner struct {
	Number        int     `json:"number"`
	Angle         float64 `json:"angle"`
	Length        float64 `json:"length"`
	TrackPosition struct {
		X float64 `json:"x"`
		Y float64 `json:"y"`
	} `json:"trackPosition"`
}

// Client is a lightweight caching HTTP client for the Multiviewer API.
type Client struct {
	http  *http.Client
	cache map[int]*Circuit
}

// NewClient creates a new Multiviewer API client.
func NewClient() *Client {
	return &Client{
		http:  &http.Client{Timeout: 10 * time.Second},
		cache: make(map[int]*Circuit),
	}
}

// GetCircuit fetches circuit layout data for the given circuit key and year.
// Results are cached in-memory so subsequent calls for the same circuit are
// instantaneous.
func (c *Client) GetCircuit(ctx context.Context, circuitKey, year int) (*Circuit, error) {
	if cached, ok := c.cache[circuitKey]; ok {
		return cached, nil
	}
	url := fmt.Sprintf("%s/circuits/%d/%d", baseURL, circuitKey, year)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching circuit: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("circuit API returned HTTP %d", resp.StatusCode)
	}
	var circuit Circuit
	if err := json.NewDecoder(resp.Body).Decode(&circuit); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}
	c.cache[circuitKey] = &circuit
	return &circuit, nil
}

// CircuitKeyByName maps common circuit/location names to the Multiviewer API
// circuit key. Keys were verified by scanning the live API (IDs 1–300, 2026).
// The primary lookup for live sessions uses session.CircuitKey directly;
// this map is a convenience fallback for name-based lookups.
var CircuitKeyByName = map[string]int{
	// ── Active calendar circuits (verified from API scan) ──────────────────────
	"Melbourne":          10,  // Australian GP – Albert Park
	"Sakhir":             63,  // Bahrain GP – Bahrain International Circuit
	"Jeddah":             149, // Saudi Arabian GP – Jeddah Corniche Circuit
	"Shanghai":           49,  // Chinese GP – Shanghai International Circuit
	"Suzuka":             46,  // Japanese GP – Suzuka Circuit
	"Miami":              151, // Miami GP – Miami International Autodrome
	"Imola":              6,   // Emilia Romagna GP – Autodromo Enzo e Dino Ferrari
	"Monte Carlo":        22,  // Monaco GP – Circuit de Monaco
	"Monaco":             22,  // Monaco GP (alias)
	"Montreal":           23,  // Canadian GP – Circuit Gilles Villeneuve
	"Barcelona":          15,  // Spanish GP – Circuit de Barcelona-Catalunya
	"Catalunya":          15,  // Spanish GP (alias)
	"Spielberg":          19,  // Austrian GP – Red Bull Ring
	"Silverstone":        2,   // British GP – Silverstone Circuit
	"Budapest":           4,   // Hungarian GP – Hungaroring
	"Spa-Francorchamps":  7,   // Belgian GP – Circuit de Spa-Francorchamps
	"Spa":                7,   // Belgian GP (alias)
	"Zandvoort":          55,  // Dutch GP – Circuit Zandvoort
	"Monza":              39,  // Italian GP – Autodromo Nazionale di Monza
	"Baku":               144, // Azerbaijan GP – Baku City Circuit
	"Singapore":          61,  // Singapore GP – Marina Bay Street Circuit
	"Austin":             9,   // US GP – Circuit of the Americas
	"Mexico City":        65,  // Mexico City GP – Autódromo Hermanos Rodríguez
	"São Paulo":          14,  // Brazilian GP – Autodromo Jose Carlos Pace
	"Interlagos":         14,  // Brazilian GP (alias)
	"Las Vegas":          152, // Las Vegas GP – Las Vegas Street Circuit
	"Lusail":             150, // Qatar GP – Lusail International Circuit
	"Losail":             150, // Qatar GP (alias)
	"Yas Marina Circuit": 70,  // Abu Dhabi GP – Yas Marina Circuit
	"Yas Island":         70,  // Abu Dhabi GP (alias)
	// ── Historic/occasional circuits (not on current calendar) ────────────────
	"Paul Ricard":        28,  // French GP – Circuit Paul Ricard
	"Hockenheim":         34,  // German GP – Hockenheimring
	"Istanbul":           59,  // Turkish GP – Istanbul Park
	"Nürburgring":        72,  // German GP (Eifel) – Nürburgring
	"Sochi":              79,  // Russian GP – Sochi Autodrom
	"Mugello":            146, // Tuscany GP – Mugello Circuit
	"Portimão":           147, // Portuguese GP – Algarve International Circuit
	"Algarve":            147, // Portuguese GP (alias)
}
