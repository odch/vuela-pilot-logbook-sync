# vuela-pilot-logbook-sync

Google Cloud Functions (Go, 2nd gen) that sync Vuela flights and connections to
external pilot-logbook systems.

## Functions

| Function | Trigger | Purpose |
|---|---|---|
| `activatePilotLogBookSync` | Firestore write on `users/{user}` | When a user's `pilotLogbookSync.status` becomes `new`, activate (connect) their account against the chosen system. |
| `syncPilotLogBook` | Firestore write on `pilotLogbookSync/{sync}` | When a sync job's `status` is `new`, import the referenced flight into the user's logbook system. |
| `SyncPilotLookbookWebhook` | HTTP | Re-queues stale sync jobs (`internal.RetryJobs`). |

## Providers

Each external system is a provider package under `internal/`, selected by a
`type` string (see `internal/const.go`). Dispatch happens via the `switch`
statements in `internal/activate.go` (activation) and `internal/import.go`
(import).

| `type` | Package | Notes |
|---|---|---|
| `capzlog` | `internal/capzlog` | capzlog.aero (host via `CAPZLOG_HOST`) |
| `mycontrol` | `internal/mycontrol` | myControl |
| `mock` | `internal/mock` | No external system — for testing in dev (see below) |

## Testing with the mock provider

`mock` talks to no external system. It logs what it *would* send and returns a
deterministic outcome **chosen by the apiKey value**, so you can exercise the
full Firestore-trigger → activation/import → status state-machine in dev by just
writing Firestore documents.

### apiKey → outcome

Matched case-insensitively; anything not listed succeeds, so the happy path
"just works".

| apiKey contains | Activation | Import | Resulting status |
|---|---|---|---|
| _(default, e.g. `mock`)_ | success | success | activation `success`, flight `success` |
| `fail-activate` | error | — | user `failure` |
| `fail-import` | success | generic error | flight `failure` |
| `fail-auth` | success | auth error | flight `auth_failure`, user `account_failure` |

> Import-failure keys still activate successfully, because a flight is only
> imported once the user connection is `success`.

### Steps (against `odch-aircraft-logbook-dev`)

1. **Activate a connection.** On a user document (`users/{userId}`), set:
   ```jsonc
   "pilotLogbookSync": { "type": "mock", "apiKey": "mock", "status": "new" }
   ```
   `activatePilotLogBookSync` runs and flips `pilotLogbookSync.status` to
   `success` (logs show `[mock] ActivateUser`).

2. **Sync a flight.** Easiest path: with the connection from step 1 active,
   create/sync a flight for that user in the Vuela dev app — the app writes the
   `pilotLogbookSync` job document for you, and `syncPilotLogBook` picks it up.

   To trigger it by hand instead, create a document in the `pilotLogbookSync`
   collection:
   ```jsonc
   {
     "type": "mock",
     "status": "new",
     "user":   "<reference to users/{userId}>",
     "flight": "<reference to an existing flight document>"
   }
   ```
   Notes that will otherwise silently stall the job:
   - `user` and `flight` must be Firestore **reference** fields (not string
     paths) — `loadSync` decodes them into `DocumentRef`s.
   - `flight` must point at a **real Vuela flight nested under an aircraft**
     (`…/{aircraft}/flights/{flight}`); `loadFlight` reads the aircraft
     registration via the flight's grandparent. A top-level flight ref fails
     before any status is written, leaving the job at `new`.

   `syncPilotLogBook` runs, logs `[mock] Import flight …`, and sets the job's
   `status` to `success`.

3. **Test failure branches.** Repeat step 1 with apiKey `mock-fail-activate`, or
   step 2 with `mock-fail-import` / `mock-fail-auth`, and confirm the resulting
   statuses match the table above.

### Reading logs

```sh
gcloud functions logs read syncPilotLogBook \
  --gen2 --region=europe-west1 --project=odch-aircraft-logbook-dev --limit=20
gcloud functions logs read activatePilotLogBookSync \
  --gen2 --region=europe-west1 --project=odch-aircraft-logbook-dev --limit=20
```

## Deployment

Deploys run via GitHub Actions: push to `main` deploys **dev**
(`.github/workflows/firebase-hosting-dev.yml`); pushing a `v*` tag deploys
**prod** (`firebase-hosting-prod.yml`). For manual deploys use `deploy-dev.sh` /
`deploy-prod.sh`.
