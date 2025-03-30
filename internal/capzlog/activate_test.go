package capzlog

import (
	"fmt"
	"testing"

	"github.com/go-openapi/runtime"
	mock_runtime "github.com/odch/aircraft-logbook/functions-go/flightsync/mocks/github.com/go-openapi/runtime"

	"github.com/odch/aircraft-logbook/functions-go/flightsync/mocks/github.com/go-openapi/strfmt"
	esa "github.com/odch/aircraft-logbook/functions-go/flightsync/mocks/github.com/odch/go-capzlog/client/external_system_activation"
	"github.com/odch/go-capzlog/client/external_system_activation"
	"github.com/odch/go-capzlog/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupMock(t *testing.T) *esa.MockClientService {
	mc := esa.NewMockClientService(t)
	mc.EXPECT().ExternalSystemActivationPost(mock.Anything, mock.Anything).RunAndReturn(func(esapp *external_system_activation.ExternalSystemActivationPostParams, caiw runtime.ClientAuthInfoWriter, co ...external_system_activation.ClientOption) (*external_system_activation.ExternalSystemActivationPostOK, error) {
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
			result = "InvalidAuthenticationTokenType"
		}).Maybe()

		err := caiw.AuthenticateRequest(cr, reg)
		assert.Nil(t, err)
		if result != "" {
			return nil, fmt.Errorf(result)
		}
		act := external_system_activation.NewExternalSystemActivationPostParams()
		b := &externalSystemActivation{
			ActivationToken: "SYSTEM_IDENTIFIER",
		}
		ie := models.ExternalSystemConnectionActivationInput{
			ExternalSystemConnectionActivationBase: b,
		}
		act.SetInputEntity(&ie)
		assert.Equal(t, act, esapp)
		return nil, nil
	})
	return mc
}

func TestActivateUserOK(t *testing.T) {
	c.ExternalSystemActivation = setupMock(t)
	authToken := "TOKEN_OK"
	systemIdentifier := "SYSTEM_IDENTIFIER"
	status, _, err := activateUser(authToken, systemIdentifier)
	assert.Nil(t, err)
	assert.Equal(t, "success", status)
}

func TestActivateUserMissingRequiredSubscription(t *testing.T) {
	c.ExternalSystemActivation = setupMock(t)
	authToken := "TOKEN_MISSING_SUBSCRIPTION"
	systemIdentifier := "SYSTEM_IDENTIFIER"
	status, _, err := activateUser(authToken, systemIdentifier)
	assert.Nil(t, err)
	assert.Equal(t, "account_failure", status)
}

func TestActivateUserMismatchBetweenPilotTokenAndExternalSystemIdentifier(t *testing.T) {
	c.ExternalSystemActivation = setupMock(t)

	authToken := "TOKEN_MISMATCH"
	systemIdentifier := "SYSTEM_IDENTIFIER"
	status, _, err := activateUser(authToken, systemIdentifier)
	assert.Nil(t, err)
	assert.Equal(t, "mismatch_failure", status)
}

func TestActivateUserInvalidToken(t *testing.T) {
	c.ExternalSystemActivation = setupMock(t)
	authToken := "TOKEN_INVALID"
	systemIdentifier := "SYSTEM_IDENTIFIER"
	status, _, err := activateUser(authToken, systemIdentifier)
	assert.Equal(t, "failure", status)
	assert.Equal(t, fmt.Errorf("InvalidAuthenticationTokenType"), err)
}
