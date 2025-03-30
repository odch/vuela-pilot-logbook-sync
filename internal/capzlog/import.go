package capzlog

import (
	"fmt"
	"log"
	"strings"

	"github.com/odch/aircraft-logbook/functions-go/flightsync/internal/vuela"
	"github.com/odch/go-capzlog/client/external_system_flights"
	"github.com/odch/go-capzlog/models"
)

func doPostFlight(authToken, systemInstanceIdentifier, id string, f models.FlightInputBase) error {
	bt := BearerBasicToken(authToken, systemInstanceIdentifier)

	i := &models.ExternalSystemFlightInput{
		ExternalSystemUniqueID: &id,
		FlightInputBase:        f,
	}
	flight := external_system_flights.NewExternalSystemFlightsPostParams().WithInputEntity(i)
	_, err := c.ExternalSystemFlights.ExternalSystemFlightsPost(flight, bt)
	if e, ok := err.(*external_system_flights.ExternalSystemFlightsPostBadRequest); ok {
		return fmt.Errorf("capzlog: %s", e.Payload)
	}
	return err
}

func capzlogMapTOF(s string) *models.FlightTypes {
	var ft models.FlightTypes

	switch strings.ToLower(s) {
	case "vp":
		ft = models.FlightTypesVFR
	case "vs":
		ft = models.FlightTypesVFR
	case "ip":
		ft = models.FlightTypesIFR
	case "is":
		ft = models.FlightTypesIFR
	case "z":
		ft = models.FlightTypesZ
	case "y":
		ft = models.FlightTypesY
	case "":
		ft = models.FlightTypesVFR
	default:
		log.Printf("unknown nature of flight: %s, sending VFR", s)
		ft = models.FlightTypesVFR
	}
	if ft != "" {
		return &ft
	}
	return nil
}

func capzlogMapFlight(id, registration string, pf vuela.PilotFunction, fr *vuela.FlightLogRecord) (*models.FlightInputBase, error) {
	markers := make([]*models.Marker, 0)
	pfunc := models.PilotFunctionsPIC
	pname := &fr.Pilot.LastName
	if pf == vuela.PilotFunctionPilot {
		if fr.Instructor.Id != "" {
			pname = &fr.Instructor.LastName
			pfunc = models.PilotFunctionsDual
		}
	} else if pf == vuela.PilotFunctionInstructor {
		pname = &fr.Instructor.LastName
		pfunc = models.PilotFunctionsInstructorOnPilotSeat
		withStudent := models.NewMarkerType(models.MarkerTypeWithStudent)
		markers = append(markers, &models.Marker{
			Type:  withStudent,
			Value: fmt.Sprintf("%s %s", fr.Pilot.LastName, fr.Pilot.FirstName),
		})
	} else { //
		// error, invalid sync
		log.Printf("user not found for sync %s (pilot: %s, instructor: %s)", id, fr.Pilot.Id, fr.Instructor.Id)

		return nil, fmt.Errorf("user not found for sync %s (pilot: %s, instructor: %s)", id, fr.Pilot.Id, fr.Instructor.Id)
	}

	//	_ = pname
	bofft := fr.BlockOffTime.UTC()
	boffds := bofft.Format("2006-01-02")
	boffts := bofft.Format("15:04")

	bont := fr.BlockOnTime.UTC()
	bonts := bont.Format("15:04")

	tofft := fr.TakeOffTime.UTC()
	toffts := tofft.Format("15:04")

	ldgt := fr.LandingTime.UTC()
	ldgts := ldgt.Format("15:04")

	ldgs := int32(fr.Landings)

	return &models.FlightInputBase{
		Aircraft: &models.AircraftInput{
			Registration: &registration,
		},
		// AircraftCounterStart:      fr.,
		// AircraftCounterEnd:        1,
		AreTimesLocal:             false,
		AutoCalculateDayNightTime: true,
		DepartureAirport:          &models.AirportInput{ICAOCode: &fr.DepartureAerodrome.Identification, Name: fr.DepartureAerodrome.Name},
		ArrivalAirport:            &models.AirportInput{ICAOCode: &fr.DestinationAerodrome.Identification, Name: fr.DestinationAerodrome.Name},
		Date:                      &boffds,
		LandingTime:               ldgts,
		Landings:                  &ldgs,
		OffBlockTime:              &boffts,
		OnBlockTime:               &bonts,
		TakeoffTime:               toffts,
		TypeOfFlight:              capzlogMapTOF(fr.Nature),
		PICName:                   pname,
		PilotFunction:             &pfunc,
		Markers:                   markers,
	}, nil
}

func Import(authToken, systemInstanceIdentifier, id, registration string, pf vuela.PilotFunction, f *vuela.FlightLogRecord) error {
	b, err := capzlogMapFlight(id, registration, pf, f)
	if err != nil {
		return err
	}

	if f.Deleted {
		return fmt.Errorf("delete not supported yet")
	}

	return doPostFlight(authToken, systemInstanceIdentifier, id, *b)
}
