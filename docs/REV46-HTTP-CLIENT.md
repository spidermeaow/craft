# Craft 0.1.46 HTTP Client

`std.http` now has a client for bounded HTTP/HTTPS requests. It is independent of the existing Server API and `packages/api`. The client can be used from CLI, Desktop, task, or HTTP handler code.

```craft
var request: HttpClientRequest = std.http.clientRequest(
    "GET", "https://example.com/health", std.bytes.fromString(""),
)
request.headers.set("accept", ["application/json"])
var config: HttpClientConfig = std.http.clientConfig()
config.timeoutMs = 5000
let response: HttpClientResponse = std.http.send(request, config)
if response.status == 200 { print(response.body.text()) }
```

| API | Contract |
| --- | --- |
| `clientRequest(method, url, body)` | Returns `HttpClientRequest` with mutable `method`, `url`, `headers: Map<String,String[]>`, and `body: Bytes`. Supports GET, POST, PUT, PATCH, DELETE, and HEAD. |
| `clientConfig()` | Returns `HttpClientConfig` with `timeoutMs = 10000`, `maxResponseBytes = 1048576`, and `maxResponseHeaderBytes = 65536`. |
| `send(request, config)` | Returns `HttpClientResponse` with `status: Int`, lowercase `headers: Map<String,String[]>`, and materialized `body: Bytes`. 3xx, 4xx, and 5xx statuses are ordinary responses. |

`timeoutMs` must be 1–300000 and bounds the whole request, including DNS, connection, TLS, response headers, and body. `maxResponseBytes` must be 1–16777216; `maxResponseHeaderBytes` must be 1024–1048576. Request bodies cannot exceed 16 MiB and request headers cannot exceed 64 KiB. The response body is read completely with a bound; there is no streaming interface. Binary responses remain `Bytes`; `.text()` requires UTF-8.

Only absolute `http://` and `https://` URLs are accepted. URL userinfo and fragments are rejected. TLS certificate and hostname validation use Go's standard trust configuration; Craft has no insecure TLS option. Redirects are **not followed**: the application receives a 3xx response and may inspect `Location`. There is no automatic retry, cookie jar, or environment-proxy use. The Runtime owns `Host`, `Content-Length`, and other transport headers. Header names and values are checked before network activity.

The current Craft execution context is used for cancellation, so an HTTP handler or task cannot leave an outbound request running after its owner is cancelled. Errors have types `HttpClientConfigError`, `HttpClientNetworkError`, `HttpClientTLSError`, `HttpClientTimeoutError`, `HttpClientCancelledError`, and `HttpClientLimitError`. Their messages do not include the URL, query, credentials, headers, or body. Applications that accept user-provided URLs must apply their own destination/SSRF policy; this client is not a sandbox.

The client uses a shared Go transport for connection reuse, with up to 64 total idle connections, 8 idle connections per host, and 64 active connections per host. Idle connections expire after 30 seconds or close when the Craft process ends. Proxy discovery from environment variables is disabled in this revision; proxy configuration remains future work.

See [`examples/rev46-http-client`](../examples/rev46-http-client) for a small GET example and [Rev.46 plan](../roadmap/phase1/craft-phase1-rev46-http-client.md) for acceptance criteria.
