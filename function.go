package functions

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/googleapis/google-cloudevents-go/cloud/firestoredata"
	"github.com/odch/aircraft-logbook/functions-go/flightsync/internal"
	"google.golang.org/protobuf/proto"
)

var firestoreClient *firestore.Client

func init() {
	// err is pre-declared to avoid shadowing client.
	var err error

	// client is initialized with context.Background() because it should
	// persist between function invocations.
	ctx := context.Background()

	firebaseApp, err := firebase.NewApp(ctx, nil)
	if err != nil {
		log.Fatalf("firebase.NewApp: %v", err)
	}

	firestoreClient, err = firebaseApp.Firestore(ctx)
	if err != nil {
		log.Fatalf("firebase.Firestore: %v", err)
	}

	// Register the entry points with the functions framework (Cloud Functions 2nd gen).
	functions.CloudEvent("SyncPilotLogBook", SyncPilotLogBook)
	functions.CloudEvent("ActivatePilotLogBookSync", ActivatePilotLogBookSync)
	functions.HTTP("SyncPilotLookbookWebhook", SyncPilotLookbookWebhook)
}

func SyncPilotLogBook(ctx context.Context, e event.Event) error {
	var data firestoredata.DocumentEventData
	if err := proto.Unmarshal(e.Data(), &data); err != nil {
		return fmt.Errorf("proto.Unmarshal: %w", err)
	}

	// Getters are nil-safe: on delete events the value is nil, so the status
	// reads as "" and the sync is skipped.
	doc := data.GetValue()
	log.Printf("Function triggered by change to: %v", doc.GetName())

	if doc.GetFields()["status"].GetStringValue() == "new" {
		// Name is projects/.../databases/(default)/documents/pilotLogbookSync/{id};
		// SyncFlight expects the path relative to .../documents/.
		x := strings.SplitN(doc.GetName(), "/documents/", 2)
		ref := x[len(x)-1]
		return internal.SyncFlight(ctx, firestoreClient, ref)
	}
	return nil
}

func ActivatePilotLogBookSync(ctx context.Context, e event.Event) error {
	var data firestoredata.DocumentEventData
	if err := proto.Unmarshal(e.Data(), &data); err != nil {
		return fmt.Errorf("proto.Unmarshal: %w", err)
	}

	doc := data.GetValue()
	log.Printf("Function triggered by change to: %v", doc.GetName())

	pls := doc.GetFields()["pilotLogbookSync"].GetMapValue().GetFields()
	if pls["status"].GetStringValue() == "new" {
		x := strings.Split(doc.GetName(), "/")
		userId := x[len(x)-1]
		return internal.EnableSync(ctx, firestoreClient, userId,
			pls["type"].GetStringValue(), pls["apiKey"].GetStringValue())
	}
	return nil
}

func SyncPilotLookbookWebhook(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	if err := internal.RetryJobs(ctx, firestoreClient); err != nil {
		panic(err)
	}
}
