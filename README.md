# sonolus-notgarupa-server

Go rewrite of the local `sonolus-test-server`, built on `sonolus-core-go`, `sonolus-server-go`, and `sonolus-pack-go`.

The server exposes the Sonolus API and connects it to a local repository store. The repository store owns packing, generated blobs, and level source writes; the HTTP server only consumes repository snapshots and publishes uploads through a narrow interface.

## Run

```powershell
go run .
```

Use a specific config file at startup:

```powershell
go run . -c .\config.local.ini
# or
go run . --config .\config.local.ini
```

By default the server listens on `127.0.0.1:8000` to avoid Windows firewall prompts during local testing.

Useful environment variables:

- `SONOLUS_CONFIG=config.local.ini` overrides the default config file path; `-c` and `--config` take precedence
- `PORT=8020` listens on `127.0.0.1:8020`
- `SONOLUS_LISTEN_ADDR=0.0.0.0:8000` allows external access
- `SONOLUS_REPOSITORY_SOURCE_DIR=source` overrides the repository source directory
- `SONOLUS_REPOSITORY_PACK_DIR=pack` overrides the generated repository pack directory
- `SONOLUS_REPOSITORY_TMP_DIR=tmp` overrides temporary pack output

Useful `config.ini` sections:

```ini
[server]
port = 8000

[repository]
source-dir = ./source
pack-dir = ./pack
tmp-dir = ./tmp
```

At startup the repository store packs `source/` through a temporary `tmp/pack-{timestamp}` directory into `pack/` and returns a snapshot. The server applies that snapshot to Sonolus routes and registers repository blobs under `/sonolus/repository/:hash`. Uploads write level source files, partially pack the uploaded level into `pack/repository`, append the level to `pack/db.json`, then refresh the server from that snapshot.

## Test

```powershell
go test ./...
```

The integration tests use `httptest`, so they do not bind a network port.

