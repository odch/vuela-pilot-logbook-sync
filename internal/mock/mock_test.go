package mock

import (
	"testing"

	"github.com/odch/aircraft-logbook/functions-go/flightsync/internal/vuela"
	"github.com/stretchr/testify/assert"
)

func TestActivateUser(t *testing.T) {
	status, _, err := ActivateUser("mock")
	assert.Nil(t, err)
	assert.Equal(t, "success", status)

	_, _, err = ActivateUser("mock-fail-activate")
	assert.NotNil(t, err)
}

func TestImport(t *testing.T) {
	f := &vuela.FlightLogRecord{}

	// default -> success
	assert.Nil(t, Import("mock", "id", "HBXYZ", vuela.PilotFunctionPilot, f))

	// generic import failure
	assert.Error(t, Import("mock-fail-import", "id", "HBXYZ", vuela.PilotFunctionPilot, f))

	// auth failure must carry the token the dispatcher keys on
	err := Import("mock-fail-auth", "id", "HBXYZ", vuela.PilotFunctionPilot, f)
	assert.ErrorContains(t, err, "MissingRequiredSubscription")
}
