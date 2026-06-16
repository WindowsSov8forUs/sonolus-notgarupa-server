# sonolus-notgarupa-server

Go rewrite of the local `sonolus-test-server`, built on `sonolus-core-go`, `sonolus-server-go`, and `sonolus-pack-go`.

The server exposes the Sonolus API and connects it to a local repository store. The repository store owns packing, generated blobs, and level source writes; the HTTP server only consumes repository snapshots and publishes uploads through a narrow interface.

## Run

```powershell
go run .
```

By default the server listens on `127.0.0.1:8000` to avoid Windows firewall prompts during local testing.

Useful environment variables:

- `PORT=8020` listens on `127.0.0.1:8020`
- `SONOLUS_LISTEN_ADDR=0.0.0.0:8000` allows external access
- `SONOLUS_REPOSITORY_PUBLIC_URL=http://localhost:8000` overrides repository SRL URL rewriting
- `SONOLUS_REPOSITORY_SOURCE_DIR=source` overrides the repository source directory
- `SONOLUS_REPOSITORY_DATA_DIR=data` overrides the generated repository data directory
- `SONOLUS_REPOSITORY_TMP_DIR=tmp` overrides temporary pack output
- `SONOLUS_REPOSITORY_WATCH_SOURCE=0` disables source watching
- `SONOLUS_REPOSITORY_POLL_INTERVAL=10s` overrides the manifest polling interval

At startup the repository store packs `source/` into `data/` and returns a snapshot. The server applies that snapshot to Sonolus routes and registers repository blobs under `/sonolus/repository/:hash`. Uploads publish through the repository interface, then the server refreshes from the latest snapshot.

## Test

```powershell
go test ./...
```

The integration tests use `httptest`, so they do not bind a network port.

