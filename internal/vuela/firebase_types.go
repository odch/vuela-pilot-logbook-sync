package vuela

import (
	"time"

	"cloud.google.com/go/firestore"
)

type PilotFunction int

const PilotFunctionInvalid PilotFunction = 0
const PilotFunctionPilot PilotFunction = 1
const PilotFunctionStudent PilotFunction = 2
const PilotFunctionInstructor PilotFunction = 3

type FlightLogRecord struct {
	TakeOffTime           time.Time
	LandingTime           time.Time
	BlockOffTime          time.Time
	BlockOnTime           time.Time
	Counters              Counters
	Landings              int
	FuelType              string
	FuelUplift            float64
	Nature                string
	DepartureAerodrome    Aerodrome
	DestinationAerodrome  Aerodrome
	Pilot                 UserRef
	Instructor            UserRef
	Remarks               string
	Deleted               bool
	Correction            bool
	PilotLogbookSync      *firestore.DocumentRef `json:"pilotLogbookSync"`
	InstructorLogbookSync *firestore.DocumentRef `json:"instructorLogbookSync"`
}

type PilotLogbookSync struct {
	Type        string                 `json:"type"`
	Message     string                 `json:"message"`
	Status      string                 `json:"status"`
	Description string                 `json:"description"`
	Timestamp   *time.Time             `json:"timestamp"`
	User        *firestore.DocumentRef `json:"user"`
	Flight      *firestore.DocumentRef `json:"flight"`
}

type User struct {
	PilotLogbookSync struct {
		ApiKey string `json:"apiKey"`
		Status string `json:"status"`
		Type   string `json:"type"`
	} `json:"pilotLogbookSync"`
}

type UserRef struct {
	Id        string
	Nr        string
	LastName  string
	FirstName string
}

type Aerodrome struct {
	Identification string
	Name           string
}

type Counters struct {
	FlightHours Counter
}

type Counter struct {
	Start int
	End   int
}

type Aircraft struct {
	Registration string
}
