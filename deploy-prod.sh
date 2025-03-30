#!/bin/sh
set -e
PROJECT=odch-aircraft-logbook-prod
DATABASE_INSTANCE=odch-aircraft-logbook-prod
ENTRY_POINT1=SyncPilotLogBook
ENTRY_POINT2=ActivatePilotLogBookSync

RUNTIME=go120

gcloud config set project $PROJECT

gcloud functions deploy syncPilotLogBook \
  --entry-point $ENTRY_POINT1 \
  --trigger-location europe-west1 \
  --trigger-event=providers/cloud.firestore/eventTypes/document.write \
  --trigger-resource "projects/$PROJECT/databases/(default)/documents/pilotLogbookSync/{sync}" \
  --runtime $RUNTIME \
  --env-vars-file .env.prod

gcloud functions deploy activatePilotLogBookSync \
  --entry-point $ENTRY_POINT2 \
  --trigger-event=providers/cloud.firestore/eventTypes/document.write \
  --trigger-resource "projects/$PROJECT/databases/(default)/documents/users/{user}" \
  --runtime $RUNTIME \
  --env-vars-file .env.prod

gcloud functions deploy SyncPilotLookbookWebhook --runtime go120 --trigger-http --env-vars-file .env.dev
