package internal

import (
	"context"
	"log"
	"time"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
)

func doRetry(ctx context.Context, client *firestore.Client, ref *firestore.DocumentRef) error {
	return updateFlightSyncStatus(ctx, client, ref, "new", "retrying")
}

func RetryJobs(ctx context.Context, client *firestore.Client) error {
	ts := time.Now().Add(-1 * time.Hour)
	log.Println("retrying new")
	iter := client.Collection("pilotLogSync").Where("status", "==", "new").Where("timestamp", "<=", ts).Documents(ctx)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return err
		}
		log.Println(doc.Ref.ID)
		log.Println(doc.Data())
		err = doRetry(ctx, client, doc.Ref)
		if err != nil {
			return err
		}
	}
	log.Println("retrying pending")
	iter = client.Collection("pilotLogSync").Where("status", "==", "pending").Where("timestamp", "<=", ts).Documents(ctx)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return err
		}
		log.Println(doc.Ref.ID)
		log.Println(doc.Data())
		err = doRetry(ctx, client, doc.Ref)
		if err != nil {
			return err
		}
	}
	log.Println("retrying failed")
	iter = client.Collection("pilotLogSync").Where("status", "==", "failure").Documents(ctx)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return err
		}
		log.Println(doc.Ref.ID)
		log.Println(doc.Data())
		continue
		err = doRetry(ctx, client, doc.Ref)
		if err != nil {
			return err
		}
	}
	return nil
}
