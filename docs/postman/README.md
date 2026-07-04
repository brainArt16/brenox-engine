# Postman — Brenox API

## Import

1. Open Postman → **Import**
2. Select:
   - `brenox.postman_collection.json`
   - `brenox.postman_environment.json` (local) or `brenox.postman_environment.prod.json` (production: `https://api.breno-x.com`)
3. Activate the matching environment

## Usage

1. Start the API: `make db-start && make migrate && make run`
2. Run **Auth → Register** (once)
3. Run **Auth → Login** — saves `token` automatically
4. Run **Channels** requests
5. WebSocket: use Postman's WebSocket client or another tool with the URL from **WebSocket → Connect**

## Maintenance

The **Postman Agent** (see `AGENTS.md`) updates this collection whenever routes change in `cmd/api/main.go`.

Do not hand-edit paths here without updating the server and README.
