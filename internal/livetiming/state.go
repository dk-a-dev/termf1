// Package livetiming subscribes to the F1 live-timing SignalR feed at
// livetiming.formula1.com and aggregates the stream into a single State
// struct that can be queried via a REST/SSE HTTP server.
package livetiming

import (
	"encoding/json"
	"sync"
	"time"
)

// State is the live aggregated snapshot of all timing topics.
// All fields are updated by the apply* methods which hold the write lock.
type State struct {
	mu sync.RWMutex

	UpdatedAt time.Time `json:"updated_at"`

	SessionStatus string `json:"session_status"` // "Started", "Finished", etc.
	SessionInfo   SessionInfo `json:"session_info"`

	LapCount struct {
		CurrentLap int `json:"CurrentLap"`
		TotalLaps  int `json:"TotalLaps"`
	} `json:"lap_count"`

	Timing   map[string]*DriverTiming `json:"timing"`   // keyed by racing number
	Drivers  map[string]*DriverInfo   `json:"drivers"`  // keyed by racing number
	CarData  map[string]*CarData      `json:"car_data"` // keyed by racing number
	Positions map[string]*Position    `json:"positions"`
	RaceControlMessages []RaceControlMessage `json:"race_control_messages"`
	Weather  WeatherData `json:"weather"`
	Stints   map[string]*TyreStint `json:"stints"` // keyed by racing number

	// rcmSeen deduplicates messages by Utc timestamp; not serialised.
	rcmSeen map[string]struct{}
}

// RLock acquires a read lock on the state (for external callers).
func (s *State) RLock() { s.mu.RLock() }

// RUnlock releases a read lock on the state (for external callers).
func (s *State) RUnlock() { s.mu.RUnlock() }

// NewState returns an initialised State.
func NewState() *State {
	return &State{
		Timing:    make(map[string]*DriverTiming),
		Drivers:   make(map[string]*DriverInfo),
		CarData:   make(map[string]*CarData),
		Positions: make(map[string]*Position),
		Stints:    make(map[string]*TyreStint),
		rcmSeen:   make(map[string]struct{}),
	}
}

// Snapshot returns a JSON-serialisable copy of the state (deep copy via
// marshal/unmarshal is simple enough given the update frequency).
func (s *State) Snapshot() []byte {
	s.mu.RLock()
	defer s.mu.RUnlock()
	b, _ := json.Marshal(s)
	return b
}

// applyTimingData merges a TimingData patch into the state.
// The F1 feed sends partial updates so we only overwrite fields that are
// present in the patch.
func (s *State) applyTimingData(raw json.RawMessage) {
	var patch struct {
		Lines map[string]json.RawMessage `json:"Lines"`
	}
	if err := json.Unmarshal(raw, &patch); err != nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for num, line := range patch.Lines {
		if existing, ok := s.Timing[num]; ok {
			// merge: unmarshal patch on top of existing
			_ = json.Unmarshal(line, existing)
		} else {
			var dt DriverTiming
			_ = json.Unmarshal(line, &dt)
			s.Timing[num] = &dt
		}
	}
	s.UpdatedAt = time.Now()
}

// applyDriverList merges a DriverList patch.
func (s *State) applyDriverList(raw json.RawMessage) {
	var patch map[string]json.RawMessage
	if err := json.Unmarshal(raw, &patch); err != nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for num, info := range patch {
		if existing, ok := s.Drivers[num]; ok {
			_ = json.Unmarshal(info, existing)
		} else {
			var di DriverInfo
			_ = json.Unmarshal(info, &di)
			s.Drivers[num] = &di
		}
	}
	s.UpdatedAt = time.Now()
}

// applyCarData merges the decompressed CarData.z entries.
// The CarData topic sends: {"Entries":[{"Cars":{"1":{...},...},"Utc":"..."},...]}
func (s *State) applyCarData(raw json.RawMessage) {
	var payload struct {
		Entries []struct {
			Utc  string                     `json:"Utc"`
			Cars map[string]json.RawMessage `json:"Cars"`
		} `json:"Entries"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return
	}
	if len(payload.Entries) == 0 {
		return
	}
	latest := payload.Entries[len(payload.Entries)-1]
	s.mu.Lock()
	defer s.mu.Unlock()
	for num, carRaw := range latest.Cars {
		var cd CarData
		_ = json.Unmarshal(carRaw, &cd)
		cd.Utc = latest.Utc
		s.CarData[num] = &cd
	}
	s.UpdatedAt = time.Now()
}

// applyPosition merges the decompressed Position.z entries.
// The Position topic sends: {"Position":[{"Timestamp":"...","Entries":{"1":{...},...}},...]}
func (s *State) applyPosition(raw json.RawMessage) {
	var payload struct {
		Position []struct {
			Timestamp string                     `json:"Timestamp"`
			Entries   map[string]json.RawMessage `json:"Entries"`
		} `json:"Position"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return
	}
	if len(payload.Position) == 0 {
		return
	}
	latest := payload.Position[len(payload.Position)-1]
	s.mu.Lock()
	defer s.mu.Unlock()
	for num, posRaw := range latest.Entries {
		var p Position
		_ = json.Unmarshal(posRaw, &p)
		s.Positions[num] = &p
	}
	s.UpdatedAt = time.Now()
}

// applyRaceControlMessages appends new RCM entries.
func (s *State) applyRaceControlMessages(raw json.RawMessage) {
	var payload struct {
		Messages map[string]json.RawMessage `json:"Messages"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, msgRaw := range payload.Messages {
		var m RaceControlMessage
		if err := json.Unmarshal(msgRaw, &m); err == nil {
			if _, seen := s.rcmSeen[m.Utc]; !seen {
				s.rcmSeen[m.Utc] = struct{}{}
				s.RaceControlMessages = append(s.RaceControlMessages, m)
			}
		}
	}
	s.UpdatedAt = time.Now()
}

// applyWeatherData overwrites the weather snapshot.
func (s *State) applyWeatherData(raw json.RawMessage) {
	s.mu.Lock()
	defer s.mu.Unlock()
	_ = json.Unmarshal(raw, &s.Weather)
	s.UpdatedAt = time.Now()
}

// applySessionStatus overwrites the session status string.
func (s *State) applySessionStatus(raw json.RawMessage) {
	var payload struct {
		Status string `json:"Status"`
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := json.Unmarshal(raw, &payload); err == nil {
		s.SessionStatus = payload.Status
	}
	s.UpdatedAt = time.Now()
}

// applyLapCount updates lap count.
func (s *State) applyLapCount(raw json.RawMessage) {
	s.mu.Lock()
	defer s.mu.Unlock()
	_ = json.Unmarshal(raw, &s.LapCount)
	s.UpdatedAt = time.Now()
}

// applySessionInfo updates session metadata.
func (s *State) applySessionInfo(raw json.RawMessage) {
	s.mu.Lock()
	defer s.mu.Unlock()
	_ = json.Unmarshal(raw, &s.SessionInfo)
	s.UpdatedAt = time.Now()
}

// applyTimingAppData merges tyre stint information from the TimingAppData topic.
// The feed sends: {"Lines":{"<num>":{"Stints":[{"Compound":"HARD","TotalLaps":23,...},...]}}}
func (s *State) applyTimingAppData(raw json.RawMessage) {
	var patch struct {
		Lines map[string]struct {
			Stints []TyreStint `json:"Stints"`
		} `json:"Lines"`
	}
	if err := json.Unmarshal(raw, &patch); err != nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for num, line := range patch.Lines {
		if len(line.Stints) == 0 {
			continue
		}
		// Last stint = current tyre.
		current := line.Stints[len(line.Stints)-1]
		s.Stints[num] = &current
	}
	s.UpdatedAt = time.Now()
}
