package internal

import (
	"context"
	"fmt"
	"log"

	"cloud.google.com/go/firestore"
	"github.com/odch/aircraft-logbook/functions-go/flightsync/internal/capzlog"
	"github.com/odch/aircraft-logbook/functions-go/flightsync/internal/mycontrol"
)

func EnableSync(ctx context.Context, client *firestore.Client, userId, system, apiKey string) error {
	userRef := client.Doc(fmt.Sprintf("users/%s", userId))

	log.Printf("Enabling sync for user: %s, system: %s", userId, system)
	if _, err := loadUser(ctx, client, userRef); err != nil {
		return err
	}

	// mark entry as pending
	if err := updateUserSyncStatus(ctx, client, userRef, "pending", "trying to connect"); err != nil {
		return err
	}

	var err error
	var status, description string

	switch system {
	case systemCapzlog:
		status, description, err = capzlog.ActivateUser(apiKey, systemInstanceIdentifier)
	case systemMycontrol:
		status, description, err = mycontrol.ActivateUser(apiKey)

	default:
		err = fmt.Errorf("system %s not yet supported", system)
	}

	if err != nil {
		log.Printf("activation returned error: %v", err)

		status = "failure"
		description = err.Error()
	}

	return updateUserSyncStatus(ctx, client, userRef, status, description)
}
