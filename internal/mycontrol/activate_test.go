package mycontrol

import (
	"fmt"
	"testing"

	mc "github.com/odch/aircraft-logbook/functions-go/flightsync/mocks/github.com/odch/go-mycontrol"
	"github.com/stretchr/testify/assert"
)

func TestActivateUser(t *testing.T) {
	mockClient := mc.NewMockClient(t)
	ok := true
	mockClient.EXPECT().GetToken().RunAndReturn(func() (string, error) {
		if ok {
			return "token", nil
		} else {
			return "", fmt.Errorf("invalid token")
		}
	})
	status, _, err := activateUser(mockClient, "valid-token")
	assert.Nil(t, err)
	assert.Equal(t, "success", status)

	ok = false
	status, _, err = activateUser(mockClient, "invalid-token")
	assert.NotNil(t, err)
	assert.Equal(t, "failure", status)

}
