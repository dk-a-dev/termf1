package openf1

// Session represents a single F1 session (FP1, FP2, FP3, Q, Sprint, Race).
type Session struct {
	CircuitKey       int    `json:"circuit_key"`
	CircuitShortName string `json:"circuit_short_name"`
	CountryCode      string `json:"country_code"`
	CountryName      string `json:"country_name"`
	DateEnd          string `json:"date_end"`
	DateStart        string `json:"date_start"`
	GmtOffset        string `json:"gmt_offset"`
	Location         string `json:"location"`
	MeetingKey       int    `json:"meeting_key"`
	SessionKey       int    `json:"session_key"`
	SessionName      string `json:"session_name"`
	SessionType      string `json:"session_type"`
	Year             int    `json:"year"`
}

// Driver contains static driver metadata for a session.
type Driver struct {
	BroadcastName string `json:"broadcast_name"`
	CountryCode   string `json:"country_code"`
	DriverNumber  int    `json:"driver_number"`
	FirstName     string `json:"first_name"`
	FullName      string `json:"full_name"`
	HeadshotURL   string `json:"headshot_url"`
	LastName      string `json:"last_name"`
	MeetingKey    int    `json:"meeting_key"`
	NameAcronym   string `json:"name_acronym"`
	SessionKey    int    `json:"session_key"`
	TeamColour    string `json:"team_colour"`
	TeamName      string `json:"team_name"`
}

// Interval holds gap/interval data. gap_to_leader and interval can be
// a float64, the string "+1 LAP", or null (leader) – so we use interface{}.
type Interval struct {
	Date         string      `json:"date"`
	DriverNumber int         `json:"driver_number"`
	GapToLeader  interface{} `json:"gap_to_leader"`
	Interval     interface{} `json:"interval"`
	MeetingKey   int         `json:"meeting_key"`
	SessionKey   int         `json:"session_key"`
}

// Lap holds lap-timing data for a single completed lap.
type Lap struct {
	DateStart        string  `json:"date_start"`
	DriverNumber     int     `json:"driver_number"`
	DurationSector1  float64 `json:"duration_sector_1"`
	DurationSector2  float64 `json:"duration_sector_2"`
	DurationSector3  float64 `json:"duration_sector_3"`
	I1Speed          int     `json:"i1_speed"`
	I2Speed          int     `json:"i2_speed"`
	IsPitOutLap      bool    `json:"is_pit_out_lap"`
	LapDuration      float64 `json:"lap_duration"`
	LapNumber        int     `json:"lap_number"`
	MeetingKey       int     `json:"meeting_key"`
	SessionKey       int     `json:"session_key"`
	StSpeed          int     `json:"st_speed"`
}

// Position represents a driver's current race/session position.
type Position struct {
	Date         string `json:"date"`
	DriverNumber int    `json:"driver_number"`
	MeetingKey   int    `json:"meeting_key"`
	Position     int    `json:"position"`
	SessionKey   int    `json:"session_key"`
}

// Weather holds a single weather sample.
type Weather struct {
	AirTemperature   float64 `json:"air_temperature"`
	Date             string  `json:"date"`
	Humidity         float64 `json:"humidity"`
	MeetingKey       int     `json:"meeting_key"`
	Pressure         float64 `json:"pressure"`
	Rainfall         int     `json:"rainfall"`
	SessionKey       int     `json:"session_key"`
	TrackTemperature float64 `json:"track_temperature"`
	WindDirection    int     `json:"wind_direction"`
	WindSpeed        float64 `json:"wind_speed"`
}

// Stint describes a tyre stint for a driver.
type Stint struct {
	Compound       string `json:"compound"`
	DriverNumber   int    `json:"driver_number"`
	LapEnd         int    `json:"lap_end"`
	LapStart       int    `json:"lap_start"`
	MeetingKey     int    `json:"meeting_key"`
	SessionKey     int    `json:"session_key"`
	StintNumber    int    `json:"stint_number"`
	TyreAgeAtStart int    `json:"tyre_age_at_start"`
}

// Pit records a single pit stop event.
type Pit struct {
	Date         string   `json:"date"`
	DriverNumber int      `json:"driver_number"`
	LaneDuration float64  `json:"lane_duration"`
	LapNumber    int      `json:"lap_number"`
	MeetingKey   int      `json:"meeting_key"`
	PitDuration  *float64 `json:"pit_duration"`
	SessionKey   int      `json:"session_key"`
	StopDuration float64  `json:"stop_duration"`
}

// SessionResult stores the final result for one driver in a session.
type SessionResult struct {
	DNF          bool        `json:"dnf"`
	DNS          bool        `json:"dns"`
	DSQ          bool        `json:"dsq"`
	DriverNumber int         `json:"driver_number"`
	Duration     interface{} `json:"duration"` // float64 (race time or best lap), or array for quali
	GapToLeader  interface{} `json:"gap_to_leader"` // float64 or "+N LAP(S)"
	NumberOfLaps int         `json:"number_of_laps"`
	MeetingKey   int         `json:"meeting_key"`
	Position     int         `json:"position"`
	SessionKey   int         `json:"session_key"`
}

// RaceControl stores a race-control broadcast message.
type RaceControl struct {
	Category     string      `json:"category"`
	Date         string      `json:"date"`
	DriverNumber interface{} `json:"driver_number"`
	Flag         string      `json:"flag"`
	LapNumber    int         `json:"lap_number"`
	MeetingKey   int         `json:"meeting_key"`
	Message      string      `json:"message"`
	Scope        string      `json:"scope"`
	Sector       interface{} `json:"sector"`
	SessionKey   int         `json:"session_key"`
}

// CarData holds per-sample car telemetry (speed, throttle, brake, RPM, gear, DRS).
type CarData struct {
	Brake        int    `json:"brake"`
	Date         string `json:"date"`
	Drs          int    `json:"drs"`
	DriverNumber int    `json:"driver_number"`
	Gear         int    `json:"n_gear"`
	MeetingKey   int    `json:"meeting_key"`
	RPM          int    `json:"rpm"`
	SessionKey   int    `json:"session_key"`
	Speed        int    `json:"speed"`
	Throttle     int    `json:"throttle"`
}

// CarLocation holds the on-track X/Y/Z telemetry position for a car.
type CarLocation struct {
	Date         string `json:"date"`
	DriverNumber int    `json:"driver_number"`
	MeetingKey   int    `json:"meeting_key"`
	SessionKey   int    `json:"session_key"`
	X            int    `json:"x"`
	Y            int    `json:"y"`
	Z            int    `json:"z"`
}

// Overtake records an incident where one driver passes another.
type Overtake struct {
	Date                   string `json:"date"`
	MeetingKey             int    `json:"meeting_key"`
	OvertakenDriverNumber  int    `json:"overtaken_driver_number"`
	OvertakingDriverNumber int    `json:"overtaking_driver_number"`
	Position               int    `json:"position"`
	SessionKey             int    `json:"session_key"`
}
