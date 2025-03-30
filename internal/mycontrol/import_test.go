package mycontrol

import (
	"testing"
	"time"

	"github.com/odch/aircraft-logbook/functions-go/flightsync/internal/vuela"
	mc "github.com/odch/aircraft-logbook/functions-go/flightsync/mocks/github.com/odch/go-mycontrol"
	"github.com/odch/go-mycontrol"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/xorcare/pointer"
)

func TestImportOK(t *testing.T) {
	mockClient := mc.NewMockClient(t)
	mockClient.EXPECT().GetToken().Return("token", nil).Once()
	mockClient.EXPECT().AddFlight(mock.Anything).RunAndReturn(func(f *mycontrol.Flight) (*mycontrol.Flight, error) {

		df := &mycontrol.Flight{
			Aircraft: mycontrol.Aircraft{
				Registration: "HBTST",
			},
			Departure: mycontrol.ArrDep{
				Time: mycontrol.Time("2023-01-02T11:12:00Z"),
				Place: mycontrol.Place{
					Name: "LSXX",
				},
			},
			Arrival: mycontrol.ArrDep{
				Time: mycontrol.Time("2023-01-02T12:13:00Z"),
				Place: mycontrol.Place{
					Name: "LSZZ",
				},
			},
			PIC: "SELF",
			Landings: mycontrol.Landings{
				Day: pointer.Int(1),
			},
		}
		assert.Equal(t, df, f)
		return nil, nil
	})

	authToken := "token"
	id := "flight1"
	registration := "HBTST"
	pf := vuela.PilotFunctionPilot
	f := getRecord(pf)
	err := runImport(mockClient, authToken, id, registration, pf, f)
	assert.Nil(t, err)
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
