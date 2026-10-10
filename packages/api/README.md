# Craft API package 0.1.50

`api` is a first-party Craft library for routing, middleware, JSON/text responses, query helpers, and basic validation on the existing `std.http` server. It is written in `.craft`; the HTTP listener, limits, cancellation, and shutdown remain in the Runtime. It does not include an HTTP client or database layer.

## Add to a project

```toml
[dependencies]
api = { path = "packages/api" }
```

The bundled package uses the Craft toolchain version automatically. Run `craft package lock`, then import it in every source file that uses it:

```craft
import package api "api"

func health(request: api.Request): HttpResponse {
    return api.json(200, std.json.fromObject({
        "status": std.json.fromString("ok"),
    }))
}

func main() {
    var app: api.App = api.create()
    app = api.add(app, "GET", "/health", health)
    api.serve(std.http.config("127.0.0.1:8080"), app)
}
```

See [`examples/api-server-package`](../../examples/api-server-package) for a working application. The older [`examples/api-server`](../../examples/api-server) continues to show direct `std.http` usage.

## Public API

| Area | Functions and types |
| --- | --- |
| Application | `App`, `create`, `add`, `group`, `serve`, `dispatch` |
| Request and middleware | `Request`, `Gate`, `proceed`, `stop`, `useBefore`, `useAfter` |
| Responses | `json`, `text`, `errorResponse` |
| Request helpers | `parseJSON`, `queryValues` |
| Validation | `FieldError`, `requiredString`, `integerRange`, `validationResponse` |

`add`, `group`, `useBefore`, and `useAfter` return a new `App`; assign the result. Routes use case-sensitive exact segments and `:name` parameters. Trailing slashes are significant. Registration rejects duplicate/ambiguous routes, invalid parameter names, encoded path markers, query/fragment syntax, and empty interior segments. A matching path with the wrong method returns 405 with a sorted `Allow` header; no path match returns 404. HEAD and OPTIONS require explicit routes.

Before hooks run in registration order after route selection. They can continue with `proceed` or return an early response with `stop`. After hooks run in reverse order, including after an early response or 404/405. Handler/hook exceptions return generic JSON 500; internal details are printed to the server output. Transport errors before Craft dispatch are handled by `std.http` and do not enter package middleware.

`Request` contains the native `HttpRequest` in `raw`, route `params`, and request-local JSON `data`. `parseJSON` throws on invalid UTF-8/JSON; the handler decides the response status. Validation helpers return `FieldError[]`; `validationResponse` emits JSON status 422. The package does not apply validation automatically.

Run `craft check`, `craft test`, and `craft fmt --check` from this directory. See [Rev.44](../../roadmap/phase1/craft-phase1-rev44-api-package.md) and [HTTP Runtime contract](../../docs/REV4-LANGUAGE.md).
