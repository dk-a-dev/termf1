package jolpica

// MRData is the root wrapper returned by all Jolpica (Ergast) endpoints.
type MRData struct {
	Series        string        `json:"series"`
	URL           string        `json:"url"`
	Limit         string        `json:"limit"`
	Offset        string        `json:"offset"`
	Total         string        `json:"total"`
	RaceTable     *RaceTable    `json:"RaceTable,omitempty"`
	StandingsTable *StandingsTable `json:"StandingsTable,omitempty"`
}

type Response struct {
	MRData MRData `json:"MRData"`
}

// --- Schedule ----------------------------------------------------------------

type RaceTable struct {
	Season string `json:"season"`
	Races  []Race `json:"Races"`
}

type Race struct {
	Season   string   `json:"season"`
	Round    string   `json:"round"`
	RaceName string   `json:"raceName"`
	Circuit  Circuit  `json:"Circuit"`
	Date     string   `json:"date"`
	Time     string   `json:"time,omitempty"`
	Sessions Sessions `json:"-"` // populated manually from flat fields
	// Flat session fields from API
	FirstPractice  *SessionTime `json:"FirstPractice,omitempty"`
	SecondPractice *SessionTime `json:"SecondPractice,omitempty"`
	ThirdPractice  *SessionTime `json:"ThirdPractice,omitempty"`
	Qualifying     *SessionTime `json:"Qualifying,omitempty"`
	Sprint         *SessionTime `json:"Sprint,omitempty"`
	SprintQualifying *SessionTime `json:"SprintQualifying,omitempty"`
}

type Circuit struct {
	CircuitID   string   `json:"circuitId"`
	URL         string   `json:"url"`
	CircuitName string   `json:"circuitName"`
	Location    Location `json:"Location"`
}

type Location struct {
	Lat      string `json:"lat"`
	Long     string `json:"long"`
	Locality string `json:"locality"`
	Country  string `json:"country"`
}

type SessionTime struct {
	Date string `json:"date"`
	Time string `json:"time"`
}

type Sessions struct {
	FP1        *SessionTime
	FP2        *SessionTime
	FP3        *SessionTime
	Qualifying *SessionTime
	Sprint     *SessionTime
	Race       *SessionTime
}

// --- Standings ---------------------------------------------------------------

type StandingsTable struct {
	Season        string          `json:"season"`
	StandingsLists []StandingsList `json:"StandingsLists"`
}

type StandingsList struct {
	Season             string               `json:"season"`
	Round              string               `json:"round"`
	DriverStandings    []DriverStanding     `json:"DriverStandings,omitempty"`
	ConstructorStandings []ConstructorStanding `json:"ConstructorStandings,omitempty"`
}

type DriverStanding struct {
	Position     string        `json:"position"`
	PositionText string        `json:"positionText"`
	Points       string        `json:"points"`
	Wins         string        `json:"wins"`
	Driver       DriverInfo    `json:"Driver"`
	Constructors []ConstructorInfo `json:"Constructors"`
}

type DriverInfo struct {
	DriverID        string `json:"driverId"`
	PermanentNumber string `json:"permanentNumber"`
	Code            string `json:"code"`
	GivenName       string `json:"givenName"`
	FamilyName      string `json:"familyName"`
	Nationality     string `json:"nationality"`
}

type ConstructorStanding struct {
	Position     string          `json:"position"`
	PositionText string          `json:"positionText"`
	Points       string          `json:"points"`
	Wins         string          `json:"wins"`
	Constructor  ConstructorInfo `json:"Constructor"`
}

type ConstructorInfo struct {
	ConstructorID string `json:"constructorId"`
	Name          string `json:"name"`
	Nationality   string `json:"nationality"`
}
