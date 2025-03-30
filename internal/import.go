package internal

import (
	"context"
	"fmt"
	"log"
	"strings"

	"cloud.google.com/go/firestore"
	"github.com/mitchellh/mapstructure"
	"github.com/odch/aircraft-logbook/functions-go/flightsync/internal/capzlog"
	"github.com/odch/aircraft-logbook/functions-go/flightsync/internal/mycontrol"
	"github.com/odch/aircraft-logbook/functions-go/flightsync/internal/vuela"
)

func refcmp(ref1, ref2 *firestore.DocumentRef) bool {
	if ref1 == nil || ref2 == nil {
		return false
	}
	return ref1.Path == ref2.Path
}

func loadSync(ctx context.Context, client *firestore.Client, ref *firestore.DocumentRef) (*vuela.PilotLogbookSync, error) {
	log.Printf("Loading sync %s", ref.Path)
	doc, err := ref.Get(ctx)
	if err != nil {
		return nil, err
	}
	sync := vuela.PilotLogbookSync{}
	err = doc.DataTo(&sync)

	if err != nil {
		return nil, err
	}
	return &sync, nil
}

func loadFlight(ctx context.Context, client *firestore.Client, ref *firestore.DocumentRef) (string, string, *vuela.FlightLogRecord, error) {
	log.Printf("Loading flight %s", ref.Path)

	doc, err := ref.Get(ctx)
	if err != nil {
		return "", "", nil, err
	}
	data := doc.Data()

	// hack ignore nature, if invalid
	n := data["nature"]
	if _, ok := n.(string); !ok {
		data["nature"] = ""
	}
	var frecord vuela.FlightLogRecord
	err = mapstructure.Decode(data, &frecord)
	if err != nil {
		log.Println(doc.Ref.ID, data["nature"])
		return "", "", nil, err
	}

	// fetch aircraft
	acref := doc.Ref.Parent.Parent
	acDoc, err := acref.Get(ctx)
	if err != nil {
		return "", "", nil, err
	}
	acData := acDoc.Data()
	var ac vuela.Aircraft
	err = mapstructure.Decode(acData, &ac)
	if err != nil {
		return "", "", nil, err
	}

	return doc.Ref.ID, ac.Registration, &frecord, nil
}

func syncUser(ctx context.Context, client *firestore.Client, ref *firestore.DocumentRef, id, registration string, f *vuela.FlightLogRecord, sync *vuela.PilotLogbookSync) error {
	log.Printf("Syncing flight %s for %s", ref.Path, sync.User.Path)

	// if sync.Status != "new" {
	// 	return nil
	// }
	status := "pending"
	description := ""
	if err := updateFlightSyncStatus(ctx, client, ref, status, description); err != nil {
		return err
	}

	user, err := loadUser(ctx, client, sync.User)
	if err != nil {
		// TODO
		status = "error"
		log.Printf("ERROR: unable to find user reference: %v", sync.User)
		if err = updateFlightSyncStatus(ctx, client, ref, status, description); err != nil {
			return err
		}
		// TODO
		//return err
	}
	status = "success"
	if user.PilotLogbookSync.Status != "success" {
		//unable to sync, because credentials are missing
		status = "auth_failure"
		if err = updateFlightSyncStatus(ctx, client, ref, status, description); err != nil {
			log.Println(err)
			return err
		}
		return nil
	}
	pf := vuela.PilotFunctionInvalid
	if refcmp(ref, f.PilotLogbookSync) {
		pf = vuela.PilotFunctionPilot
	} else if refcmp(ref, f.InstructorLogbookSync) {
		pf = vuela.PilotFunctionInstructor
	}

	switch sync.Type {
	case systemCapzlog:
		err = capzlog.Import(user.PilotLogbookSync.ApiKey, systemInstanceIdentifier, id, registration, pf, f)
	case systemMycontrol:
		err = mycontrol.Import(user.PilotLogbookSync.ApiKey, id, registration, pf, f)
	default:
		err = fmt.Errorf("invalid sync system: %v", sync.Type)
	}
	if err != nil {
		description = err.Error()
		status = "failure"
		log.Printf("ERROR: %v", err)
		if strings.Contains(description, "MissingRequiredSubscription") || strings.Contains(description, "InvalidAuthenticationTokenType") {
			status = "auth_failure"
			userStatus := "account_failure"
			log.Println(sync.User.Path)
			if err = updateUserSyncStatus(ctx, client, sync.User, userStatus, description); err != nil {
				log.Printf("updating account failed: %s / %s", ref.Path, sync.User.Path)
			}
		}
	}
	log.Println("updating")
	if err = updateFlightSyncStatus(ctx, client, ref, status, description); err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func SyncFlight(ctx context.Context, client *firestore.Client, refs string) error {
	ref := client.Doc(refs)
	log.Printf("Syncing %s", ref.Path)

	sync, err := loadSync(ctx, client, ref)
	if err != nil {
		log.Println("could not load sync entry", err)
		return err
	}
	log.Printf("sync: %v", sync)
	id, registration, flight, err := loadFlight(ctx, client, sync.Flight)
	if err != nil {
		log.Printf("could not load flight entry %s %v", sync.Flight.Path, err)
		return err
	}
	err = syncUser(ctx, client, ref, id, registration, flight, sync)
	if err != nil {
		log.Println("could not sync entry", err)
		return err
	}

	return nil
}
