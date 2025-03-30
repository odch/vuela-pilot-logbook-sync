package capzlog

import (
	"encoding/json"
	"log"
	"strings"

	"github.com/odch/go-capzlog/client/external_system_activation"
	"github.com/odch/go-capzlog/models"
)

type externalSystemActivation struct {
	ActivationToken string `json:"ActivationToken"`
}

func ActivateUser(authToken, systemInstanceIdentifier string) (status, description string, err error) {

	return activateUser(authToken, systemInstanceIdentifier)
}

func activateUser(authToken, systemInstanceIdentifier string) (status, description string, err error) {
	bt := BearerBasicToken(authToken, systemInstanceIdentifier)

	esa := externalSystemActivation{
		ActivationToken: systemInstanceIdentifier,
	}
	act, _ := json.Marshal(&esa)
	_ = act
	i := models.ExternalSystemConnectionActivationInput{
		ExternalSystemConnectionActivationBase: &esa,
	}

	params := external_system_activation.NewExternalSystemActivationPostParams().WithInputEntity(&i)

	_, err = c.ExternalSystemActivation.ExternalSystemActivationPost(params, bt)

	status = "success"
	description = ""

	if err != nil {
		status = "failure"
		description = err.Error()
		status = "failure"
		log.Printf("ERROR: %v", err)
		if strings.Contains(description, "MissingRequiredSubscription") {
			status = "account_failure"
			err = nil
		}
		if strings.Contains(description, "MismatchBetweenPilotTokenAndExternalSystemIdentifier") {
			description = "MismatchBetweenPilotTokenAndExternalSystemIdentifier: token is not for this connection. Make sure you use the correct token"
			status = "mismatch_failure"
			err = nil
		}
	}
	return
}
