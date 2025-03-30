package capzlog

import (
	"fmt"
	"testing"
	"time"

	"github.com/go-openapi/runtime"
	"github.com/odch/aircraft-logbook/functions-go/flightsync/internal/vuela"
	mock_runtime "github.com/odch/aircraft-logbook/functions-go/flightsync/mocks/github.com/go-openapi/runtime"
	"github.com/odch/aircraft-logbook/functions-go/flightsync/mocks/github.com/go-openapi/strfmt"
	esa "github.com/odch/aircraft-logbook/functions-go/flightsync/mocks/github.com/odch/go-capzlog/client/external_system_flights"
	"github.com/odch/go-capzlog/client/external_system_flights"
	"github.com/odch/go-capzlog/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/xorcare/pointer"
)

func setupFlightMock(t *testing.T, pf vuela.PilotFunction) *esa.MockClientService {
	mc := esa.NewMockClientService(t)
	mc.EXPECT().ExternalSystemFlightsPost(mock.Anything, mock.Anything, mock.Anything).RunAndReturn(func(esfpp *external_system_flights.ExternalSystemFlightsPostParams, caiw runtime.ClientAuthInfoWriter, co ...external_system_flights.ClientOption) (*external_system_flights.ExternalSystemFlightsPostOK, error) {
		cr := mock_runtime.NewMockClientRequest(t)
		reg := strfmt.NewMockRegistry(t)
		var result string
		cr.EXPECT().SetHeaderParam("SystemInstanceIdentifier", "SYSTEM_IDENTIFIER").Return(nil)
		cr.EXPECT().SetHeaderParam("Authorization", "Basic TOKEN_OK").Return(nil).Run(func(_a0 string, _a1 ...string) {
			result = ""
		}).Maybe()
		cr.EXPECT().SetHeaderParam("Authorization", "Basic TOKEN_MISSING_SUBSCRIPTION").Return(nil).Run(func(_a0 string, _a1 ...string) {
			result = "MissingRequiredSubscription"
		}).Maybe()
		cr.EXPECT().SetHeaderParam("Authorization", "Basic TOKEN_MISMATCH").Return(nil).Run(func(_a0 string, _a1 ...string) {
			result = "MismatchBetweenPilotTokenAndExternalSystemIdentifier"
		}).Maybe()
		cr.EXPECT().SetHeaderParam("Authorization", "Basic TOKEN_INVALID").Return(nil).Run(func(_a0 string, _a1 ...string) {
			result = "FAILURE"
		}).Maybe()

		err := caiw.AuthenticateRequest(cr, reg)
		assert.Nil(t, err)
		if result != "" {
			return nil, fmt.Errorf(result)
		}
		pp := external_system_flights.NewExternalSystemFlightsPostParams()

		pp.InputEntity = &models.ExternalSystemFlightInput{
			ExternalSystemUniqueID: pointer.String("EXTERNAL_ID"),
			FlightInputBase: models.FlightInputBase{
				AutoCalculateDayNightTime: true,
				Aircraft: &models.AircraftInput{
					Registration: pointer.String("HBTST"),
				},
				//PICName:       pointer.String("SELF"),
				PilotFunction: models.PilotFunctionsPIC.Pointer(),
				TypeOfFlight:  models.FlightTypesVFR.Pointer(),
				ArrivalAirport: &models.AirportInput{
					ICAOCode: pointer.String("LSZZ"),
				},
				DepartureAirport: &models.AirportInput{
					ICAOCode: pointer.String("LSXX"),
				},
				Markers:      make([]*models.Marker, 0),
				Date:         pointer.String("2023-01-02"),
				OffBlockTime: pointer.String("11:12"),
				OnBlockTime:  pointer.String("12:13"),
				TakeoffTime:  "11:15",
				LandingTime:  "12:10",
				Landings:     pointer.Int32(1),
			},
		}
		if pf == vuela.PilotFunctionPilot {
			pp.InputEntity.FlightInputBase.PilotFunction = models.PilotFunctionsPIC.Pointer()
		} else if pf == vuela.PilotFunctionStudent {
			pp.InputEntity.FlightInputBase.PilotFunction = models.PilotFunctionsDual.Pointer()
			pp.InputEntity.FlightInputBase.PICName = pointer.String("Instructor")

		} else if pf == vuela.PilotFunctionInstructor {
			mt := models.NewMarkerType("WithStudent")
			marker := models.Marker{
				Type:  mt,
				Value: "Pilot My",
			}
			pp.InputEntity.FlightInputBase.Markers = []*models.Marker{
				&marker,
			}
			pp.InputEntity.FlightInputBase.PilotFunction = models.PilotFunctionsInstructorOnPilotSeat.Pointer()
		}
		assert.Equal(t, pp, esfpp)
		return nil, nil
	})

	return mc
}

func getRecord(pf vuela.PilotFunction) *vuela.FlightLogRecord {
	f := &vuela.FlightLogRecord{
		Pilot: vuela.UserRef{
			LastName:  "Pilot",
			FirstName: "My",
			Id:        "something",
		},
		Instructor: vuela.UserRef{
			FirstName: "Mister",
			LastName:  "Instructor",
		},
		DepartureAerodrome: vuela.Aerodrome{
			Identification: "LSXX",
		},
		DestinationAerodrome: vuela.Aerodrome{
			Identification: "LSZZ",
		},
		BlockOffTime: time.Date(2023, 1, 2, 11, 12, 0, 0, time.UTC),
		BlockOnTime:  time.Date(2023, 1, 2, 12, 13, 0, 0, time.UTC),
		TakeOffTime:  time.Date(2023, 1, 2, 11, 15, 0, 0, time.UTC),
		LandingTime:  time.Date(2023, 1, 2, 12, 10, 0, 0, time.UTC),
		Landings:     1,
	}
	if pf == vuela.PilotFunctionStudent {
		f.Instructor.Id = "isset"
	}
	return f
}
func TestImportPilotPIC(t *testing.T) {
	pf := vuela.PilotFunctionPilot

	mc := setupFlightMock(t, pf)
	c.ExternalSystemFlights = mc

	authToken := "TOKEN_OK"
	systemInstanceIdentifier := "SYSTEM_IDENTIFIER"
	id := "EXTERNAL_ID"
	registration := "HBTST"

	f := getRecord(pf)
	err := Import(authToken, systemInstanceIdentifier, id, registration, pf, f)
	assert.Nil(t, err)
}

func TestImportStudent(t *testing.T) {
	pf := vuela.PilotFunctionStudent

	mc := setupFlightMock(t, pf)
	c.ExternalSystemFlights = mc

	authToken := "TOKEN_OK"
	systemInstanceIdentifier := "SYSTEM_IDENTIFIER"
	id := "EXTERNAL_ID"
	registration := "HBTST"

	f := getRecord(pf)
	pf = vuela.PilotFunctionPilot
	err := Import(authToken, systemInstanceIdentifier, id, registration, pf, f)
	assert.Nil(t, err)
}
func TestImportInstructor(t *testing.T) {
	pf := vuela.PilotFunctionInstructor

	mc := setupFlightMock(t, pf)
	c.ExternalSystemFlights = mc

	authToken := "TOKEN_OK"
	systemInstanceIdentifier := "SYSTEM_IDENTIFIER"
	id := "EXTERNAL_ID"
	registration := "HBTST"

	f := getRecord(pf)
	err := Import(authToken, systemInstanceIdentifier, id, registration, pf, f)
	assert.Nil(t, err)
}

func TestImportFail(t *testing.T) {
	pf := vuela.PilotFunctionPilot

	mc := setupFlightMock(t, pf)
	c.ExternalSystemFlights = mc

	authToken := "TOKEN_INVALID"
	systemInstanceIdentifier := "SYSTEM_IDENTIFIER"
	id := "EXTERNAL_ID"
	registration := "HBTST"

	f := getRecord(pf)

	err := Import(authToken, systemInstanceIdentifier, id, registration, pf, f)
	serr := fmt.Errorf("FAILURE")
	assert.Equal(t, serr, err)
}

func TestImportFailMismatch(t *testing.T) {
	pf := vuela.PilotFunctionPilot

	mc := setupFlightMock(t, pf)
	c.ExternalSystemFlights = mc

	authToken := "TOKEN_MISMATCH"
	systemInstanceIdentifier := "SYSTEM_IDENTIFIER"
	id := "EXTERNAL_ID"
	registration := "HBTST"

	f := getRecord(pf)

	err := Import(authToken, systemInstanceIdentifier, id, registration, pf, f)
	serr := fmt.Errorf("MismatchBetweenPilotTokenAndExternalSystemIdentifier")
	assert.Equal(t, serr, err)
}

func TestImportFailMissingSubscription(t *testing.T) {
	pf := vuela.PilotFunctionPilot

	mc := setupFlightMock(t, pf)
	c.ExternalSystemFlights = mc

	authToken := "TOKEN_MISSING_SUBSCRIPTION"
	systemInstanceIdentifier := "SYSTEM_IDENTIFIER"
	id := "EXTERNAL_ID"
	registration := "HBTST"

	f := getRecord(pf)

	err := Import(authToken, systemInstanceIdentifier, id, registration, pf, f)
	serr := fmt.Errorf("MissingRequiredSubscription")
	assert.Equal(t, serr, err)
}
