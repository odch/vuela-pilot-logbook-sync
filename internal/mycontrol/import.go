package mycontrol

import (
	"fmt"
	"log"
	"time"

	"github.com/odch/aircraft-logbook/functions-go/flightsync/internal/vuela"
	mc "github.com/odch/go-mycontrol"
)

func postFlight(c mc.Client, apiToken, id string, f *mc.Flight) error {
	_, err := c.GetToken()
	if err != nil {
		return err
	}
	_, err = c.AddFlight(f)
	return err

}

func mapFlight(id, registration string, pf vuela.PilotFunction, fr *vuela.FlightLogRecord) (*mc.Flight, error) {
	var pname string
	pname = "SELF"

	if pf == vuela.PilotFunctionPilot {
		if fr.Instructor.Id != "" {
			pname = fr.Instructor.LastName
			// Function DUAL
		}
		// else PIC+SELF
	} else if pf == vuela.PilotFunctionInstructor {
		// Function Instructor
	} else { //
		// error, invalid sync
		log.Printf("user not found for sync %s (pilot: %s, instructor: %s)", id, fr.Pilot.Id, fr.Instructor.Id)

		return nil, fmt.Errorf("user not found for sync %s (pilot: %s, instructor: %s)", id, fr.Pilot.Id, fr.Instructor.Id)
	}

	depTime := fr.BlockOffTime.UTC()
	arrTime := fr.BlockOnTime.UTC()

	flightNew := mc.Flight{}
	flightNew.PIC = pname
	flightNew.Aircraft.Registration = registration
	flightNew.Departure.Time = mc.Time(depTime.Format(time.RFC3339))
	flightNew.Departure.Place.Name = fr.DepartureAerodrome.Identification
	flightNew.Arrival.Place.Name = fr.DestinationAerodrome.Identification
	flightNew.Arrival.Time = mc.Time(arrTime.Format(time.RFC3339))
	ldg := fr.Landings
	flightNew.Landings.Day = &ldg

	return &flightNew, nil
}

func runImport(c mc.Client, authToken, id, registration string, pf vuela.PilotFunction, f *vuela.FlightLogRecord) error {
	if f.Deleted {
		return fmt.Errorf("delete not supported yet")
	}

	b, err := mapFlight(id, registration, pf, f)
	if err != nil {
		return err
	}

	return postFlight(c, authToken, id, b)
}

func Import(authToken, id, registration string, pf vuela.PilotFunction, f *vuela.FlightLogRecord) error {
	c := mc.NewClient(authToken)
	return runImport(c, authToken, id, registration, pf, f)
}
