// Package mock is a logbook provider that talks to no external system. It logs
// what it "would" send and returns a deterministic outcome chosen by the apiKey
// value, so the sync/activation functions can be exercised end-to-end in dev.
//
// The apiKey selects the outcome (matched case-insensitively):
//
//	(default, e.g. "mock")  activate -> success, import -> success
//	contains "fail-activate" activate -> error          (user status "failure")
//	contains "fail-import"   import   -> generic error  (flight status "failure")
//	contains "fail-auth"     import   -> auth error      (flight "auth_failure", user "account_failure")
//
// Import-failure outcomes still activate successfully, because a flight is only
// imported once the user connection is "success".
package mock

import (
	"fmt"
	"log"
	"strings"

	"github.com/odch/aircraft-logbook/functions-go/flightsync/internal/vuela"
)

func ActivateUser(apiKey string) (status, description string, err error) {
	log.Printf("[mock] ActivateUser apiKey=%q", apiKey)
	if strings.Contains(strings.ToLower(apiKey), "fail-activate") {
		return "", "", fmt.Errorf("mock: simulated activation failure")
	}
	return "success", "", nil
}

func Import(authToken, id, registration string, pf vuela.PilotFunction, f *vuela.FlightLogRecord) error {
	log.Printf("[mock] Import flight id=%s registration=%s pilotFunction=%d apiKey=%q", id, registration, pf, authToken)
	log.Printf("[mock] flight record: %+v", f)
	key := strings.ToLower(authToken)
	switch {
	case strings.Contains(key, "fail-auth"):
		// Message must contain a token the dispatcher recognises so it maps to
		// flight "auth_failure" + user "account_failure" (see internal/import.go).
		return fmt.Errorf("mock: simulated MissingRequiredSubscription")
	case strings.Contains(key, "fail-import"):
		return fmt.Errorf("mock: simulated import failure")
	default:
		return nil
	}
}
