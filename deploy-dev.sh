#!/bin/sh
set -e
PROJECT=odch-aircraft-logbook-dev
REGION=europe-west1
RUNTIME=go124
# Firestore trigger location must match the database location (dev DB is nam5).
TRIGGER_LOCATION=nam5
ENTRY_POINT1=SyncPilotLogBook
ENTRY_POINT2=ActivatePilotLogBookSync

gcloud config set project $PROJECT

gcloud functions deploy syncPilotLogBook \
  --gen2 \
  --region $REGION \
  --runtime $RUNTIME \
  --entry-point $ENTRY_POINT1 \
  --trigger-location $TRIGGER_LOCATION \
  --trigger-event-filters="type=google.cloud.firestore.document.v1.written" \
  --trigger-event-filters="database=(default)" \
  --trigger-event-filters-path-pattern="document=pilotLogbookSync/{sync}" \
  --env-vars-file .env.dev

gcloud functions deploy activatePilotLogBookSync \
  --gen2 \
  --region $REGION \
  --runtime $RUNTIME \
  --entry-point $ENTRY_POINT2 \
  --trigger-location $TRIGGER_LOCATION \
  --trigger-event-filters="type=google.cloud.firestore.document.v1.written" \
  --trigger-event-filters="database=(default)" \
  --trigger-event-filters-path-pattern="document=users/{user}" \
  --env-vars-file .env.dev

gcloud functions deploy SyncPilotLookbookWebhook \
  --gen2 \
  --region $REGION \
  --runtime $RUNTIME \
  --entry-point SyncPilotLookbookWebhook \
  --trigger-http \
  --env-vars-file .env.dev
