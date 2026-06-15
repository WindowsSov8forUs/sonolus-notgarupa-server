# sonolus-notgarupa-server

Go rewrite of the local `sonolus-test-server`, built on `sonolus-core-go` and `sonolus-server-go`.

The server exposes the Sonolus API, converts GarupaChart JSON into Sonolus level data, and publishes uploads through `sonolus-notgarupa-repository`.

## Run

```powershell
go run .
```

By default the server listens on `127.0.0.1:8000` to avoid Windows firewall prompts during local testing.

Useful environment variables:

- `PORT=8020` listens on `127.0.0.1:8020`
- `SONOLUS_LISTEN_ADDR=0.0.0.0:8000` allows external access
- `SONOLUS_REPOSITORY_ADMIN_URL=http://127.0.0.1:9000` overrides the repository admin URL
- `SONOLUS_REPOSITORY_MANIFEST_URL=http://localhost:9000/manifest.json` overrides the repository manifest URL
- `SONOLUS_REPOSITORY_POLL_INTERVAL=10s` overrides the manifest polling interval

## Test

```powershell
go test ./...
```

The integration tests use `httptest`, so they do not bind a network port.

