package livetiming

// This file contains all F1 data model types received from the SignalR feed.
// The structs here mirror the JSON shapes broadcast by livetiming.formula1.com.

// DriverTiming holds the latest timing line for one driver.
type DriverTiming struct {
	RacingNumber string `json:"RacingNumber"`
	Line         int    `json:"Line"` // grid position
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
	Sectors  []SectorTiming `json:"Sectors"`
	Speeds   struct {
		I1 SpeedTrap `json:"I1"`
		I2 SpeedTrap `json:"I2"`
		FL SpeedTrap `json:"FL"`
		ST SpeedTrap `json:"ST"`
	} `json:"Speeds"`
	InPit            bool `json:"InPit"`
	PitOut           bool `json:"PitOut"`
	Stopped          bool `json:"Stopped"`
	Status           int  `json:"Status"`
	NumberOfLaps     int  `json:"NumberOfLaps"`
	NumberOfPitStops int  `json:"NumberOfPitStops"`
	KnockedOut       bool `json:"KnockedOut"`
	Retired          bool `json:"Retired"`
	ShowPosition     bool `json:"ShowPosition"`
}

// SectorSegment holds the status of a single mini-sector (also called segment).
// Status values: 0=inactive, 64=active, 2048=personal best, 2049=overall fastest, 2051=fastest overall.
type SectorSegment struct {
	Status int `json:"Status"`
}

// SectorTiming holds mini-sector and segment data for one sector.
type SectorTiming struct {
	Stopped         bool            `json:"Stopped"`
	Value           string          `json:"Value"`
	OverallFastest  bool            `json:"OverallFastest"`
	PersonalFastest bool            `json:"PersonalFastest"`
	PreviousValue   string          `json:"PreviousValue"`
	Segments        []SectorSegment `json:"Segments"`
}

// SpeedTrap holds a speed trap value.
type SpeedTrap struct {
	Value           string `json:"Value"`
	Status          int    `json:"Status"`
	OverallFastest  bool   `json:"OverallFastest"`
	PersonalFastest bool   `json:"PersonalFastest"`
}

// DriverInfo holds broadcast name, team and colours from the DriverList topic.
type DriverInfo struct {
	RacingNumber  string `json:"RacingNumber"`
	BroadcastName string `json:"BroadcastName"`
	FullName      string `json:"FullName"`
	Tla           string `json:"Tla"` // 3-letter acronym
	TeamName      string `json:"TeamName"`
	TeamColour    string `json:"TeamColour"`
}

// CarData holds the latest telemetry sample for one car.
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

// Position holds the latest GPS position for one car.
type Position struct {
	Status string  `json:"Status"`
	X      float64 `json:"X"`
	Y      float64 `json:"Y"`
	Z      float64 `json:"Z"`
}

// RaceControlMessage is one message from Race Control.
type RaceControlMessage struct {
	Utc      string `json:"Utc"`
	Lap      int    `json:"Lap"`
	Category string `json:"Category"`
	Message  string `json:"Message"`
	Flag     string `json:"Flag"`
	Scope    string `json:"Scope"`
	Sector   int    `json:"Sector"`
}

// TyreStint holds the current tyre data for one driver derived from TimingAppData.
type TyreStint struct {
	Compound        string `json:"Compound"`   // "SOFT", "MEDIUM", "HARD", "INTERMEDIATE", "WET"
	New             bool   `json:"New"`
	TotalLaps       int    `json:"TotalLaps"`  // laps done on this tyre
	StartLaps       int    `json:"StartLaps"`
}

// WeatherData holds the latest weather sample.
type WeatherData struct {
	AirTemp       string `json:"AirTemp"`
	Humidity      string `json:"Humidity"`
	Pressure      string `json:"Pressure"`
	Rainfall      string `json:"Rainfall"`
	TrackTemp     string `json:"TrackTemp"`
	WindDirection string `json:"WindDirection"`
	WindSpeed     string `json:"WindSpeed"`
}

// SessionInfo holds top-level session metadata.
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
	Name      string `json:"Name"` // "Race", "Qualifying", etc.
	Type      string `json:"Type"`
	StartDate string `json:"StartDate"`
	EndDate   string `json:"EndDate"`
	GmtOffset string `json:"GmtOffset"`
}
