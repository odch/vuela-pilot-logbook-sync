package functions

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"cloud.google.com/go/firestore"
	"cloud.google.com/go/functions/metadata"
	firebase "firebase.google.com/go"
	"github.com/odch/aircraft-logbook/functions-go/flightsync/internal"
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
}

type FirestoreSyncEvent struct {
	OldValue   FirestoreSyncValue `json:"oldValue"`
	Value      FirestoreSyncValue `json:"value"`
	UpdateMask struct {
		FieldPaths []string `json:"fieldPaths"`
	} `json:"updateMask"`
}

type FirestoreSyncValue struct {
	CreateTime time.Time `json:"createTime"`
	//Fields     interface{} `json:"fields"`
	//Fields internal.FlightLogRecord `json:"fields"`
	Fields struct {
		Type   FirestoreStringValue `json:"type"`
		Status FirestoreStringValue `json:"status"`
		User   FirestoreStringValue `json:"user"`
		Flight FirestoreStringValue `json:"flight"`
	} `json:"fields"`
	Name       string    `json:"name"`
	UpdateTime time.Time `json:"updateTime"`
}

func SyncPilotLogBook(ctx context.Context, e FirestoreSyncEvent) error {
	meta, err := metadata.FromContext(ctx)
	if err != nil {
		return fmt.Errorf("metadata.FromContext: %w", err)
	}
	log.Printf("Function triggered by change to: %v", meta.Resource)
	log.Printf("%+v", e)
	if err != nil {
		log.Println(err)
	}

	if e.Value.Fields.Status.StringValue == "new" {
		x := strings.Split(e.Value.Name, "/databases/(default)/documents/")
		ref := x[len(x)-1]
		return internal.SyncFlight(ctx, firestoreClient, ref)
	}
	return nil
}

type FirestoreUserEvent struct {
	OldValue   FirestoreUserValue `json:"oldValue"`
	Value      FirestoreUserValue `json:"value"`
	UpdateMask struct {
		FieldPaths []string `json:"fieldPaths"`
	} `json:"updateMask"`
}

type FirestoreStringValue struct {
	StringValue string `json:"stringValue"`
}
type FirestoreUserValue struct {
	CreateTime time.Time `json:"createTime"`
	//Fields     interface{} `json:"fields"`
	Fields struct {
		PilotLogbookSync struct {
			MapValue struct {
				Fields struct {
					ApiKey FirestoreStringValue `json:"apiKey"`
					Status FirestoreStringValue `json:"status"`
					Type   FirestoreStringValue `json:"type"`
				} `json:"fields"`
			} `json:"mapValue"`
		} `json:"pilotLogbookSync"`
	} `json:"fields"`
	Name       string    `json:"name"`
	UpdateTime time.Time `json:"updateTime"`
}

/*
Value:{
	CreateTime:2021-02-20 15:45:46.016415 +0000 UTC
	Fields:map[firstname:map[stringValue:Philipp] lastLogin:map[timestampValue:2023-07-30T06:56:53.656Z] lastname:map[stringValue:Hug] organizations:map[arrayValue:map[values:[map[referenceValue:projects/odch-aircraft-logbook-dev/databases/(default)/documents/organizations/phil-test] map[referenceValue:projects/odch-aircraft-logbook-dev/databases/(default)/documents/organizations/mfgt]]]] orgs:map[mapValue:map[fields:map[mfgt:map[mapValue:map[fields:map[ref:map[referenceValue:projects/odch-aircraft-logbook-dev/databases/(default)/documents/organizations/mfgt] roles:map[arrayValue:map[values:[map[stringValue:manager] map[stringValue:techlogmanager]]]]]]] phil-test:map[mapValue:map[fields:map[ref:map[referenceValue:projects/odch-aircraft-logbook-dev/databases/(default)/documents/organizations/phil-test] roles:map[arrayValue:map[values:[map[stringValue:manager]]]]]]]]]]
	  pilotLogbookSync:map[mapValue:map[fields:map[apiKey:map[stringValue:1sh0n8YCYqpGhtOCy3i0pQT1nO0gNSGooXcW1ySofsLuhgE4ZKELkWrPsQwKZXbCX1W3kU2bNUjW8xC8IwUdH4uaYahupGl2rdRBkcgBLvuhtAzP4JWIsUiZt+zEARdLPcsDPp+1LJWzxIXnBWVd4wbEeVrr+jQLzIRLiEgWQf4zLom9eyuBEVXwOGqDfi0f//Q1gT883uK+7lZjho+1SYDB] description:map[stringValue:invalid token] status:map[stringValue:failure] type:map[stringValue:capzlog]]]] selectedOrganization:map[stringValue:mfgt]] Name:projects/odch-aircraft-logbook-dev/databases/(default)/documents/users/dcZzh34cCgWFoVqwSaw6eJAXoQs1 UpdateTime:2023-07-30 07:07:56.350319 +0000 UTC}
*/

func ActivatePilotLogBookSync(ctx context.Context, e FirestoreUserEvent) error {
	meta, err := metadata.FromContext(ctx)
	if err != nil {
		return fmt.Errorf("metadata.FromContext: %w", err)
	}
	log.Printf("Function triggered by change to: %v", meta.Resource)
	log.Printf("%+v", e)
	if err != nil {
		log.Println(err)
	}

	if e.Value.Fields.PilotLogbookSync.MapValue.Fields.Status.StringValue == "new" {
		x := strings.Split(e.Value.Name, "/")
		userId := x[len(x)-1]
		return internal.EnableSync(ctx, firestoreClient, userId, e.Value.Fields.PilotLogbookSync.MapValue.Fields.Type.StringValue, e.Value.Fields.PilotLogbookSync.MapValue.Fields.ApiKey.StringValue)
	}
	return nil
}

func SyncPilotLookbookWebhook(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	if err := internal.RetryJobs(ctx, firestoreClient); err != nil {
		panic(err)
	}
}
