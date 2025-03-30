package internal

import (
	"context"
	"log"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/mitchellh/mapstructure"
	"github.com/odch/aircraft-logbook/functions-go/flightsync/internal/vuela"
)

const systemInstanceIdentifier = "jyDLhqCBa8Ombm6fLGqhSg=="

func loadUser(ctx context.Context, client *firestore.Client, ref *firestore.DocumentRef) (*vuela.User, error) {
	log.Printf("Loading User %s", ref.Path)

	doc, err := ref.Get(ctx)
	if err != nil {
		return nil, err
	}
	data := doc.Data()
	var user vuela.User
	err = mapstructure.Decode(data, &user)
	if err != nil {
		log.Println(doc.Ref.ID, data["nature"])
		return nil, err
	}
	return &user, nil
}

func updateUserSyncStatus(ctx context.Context, client *firestore.Client, ref *firestore.DocumentRef, status, description string) error {
	ts := time.Now().UTC()

	_, err := ref.Update(ctx, []firestore.Update{
		{
			Path:  "pilotLogbookSync.status",
			Value: status,
		},
		{
			Path:  "pilotLogbookSync.description",
			Value: description,
		},
		{
			Path:  "pilotLogbookSync.timestamp",
			Value: ts,
		},
	})

	return err
}

func updateFlightSyncStatus(ctx context.Context, client *firestore.Client, ref *firestore.DocumentRef, status, description string) error {
	ts := time.Now().UTC()
	log.Println(ref.Path, status)
	_, err := ref.Update(ctx, []firestore.Update{
		{
			Path:  "status",
			Value: status,
		},
		{
			Path:  "description",
			Value: description,
		},
		{
			Path:  "timestamp",
			Value: ts,
		},
	})

	return err
}
