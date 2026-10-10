# API package example

This application exercises `packages/api` without modifying the HTTP Runtime. It has `/health`, `/items/:id`, and `POST /items`; middleware records request-local metadata, checks JSON content type for POST, and adds an `X-Framework` response header. The example is stateless; creating an item does not persist it.

From this directory:

```powershell
craft package lock
craft check
craft test
craft fmt --check
craft build
craft run
```

The server binds `127.0.0.1:8080` by default. Set `CRAFT_API_ADDRESS` to another address before `craft run` if needed. Try `GET /health`, `GET /items/1`, and `POST /items` with `Content-Type: application/json` and `{"name":"Craft"}`. The direct Runtime version remains at [`examples/api-server`](../api-server).
