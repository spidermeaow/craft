# Reference HTTP package written in Craft

แพ็กเกจนี้เป็นตัวอย่างการสร้างเครื่องมือระดับสูงด้วยภาษา Craft ไม่ใช่ส่วนบังคับของ Craft runtime และไม่ใช่ release gate ของภาษา ผู้พัฒนาสามารถนำแนวทางไปสร้าง framework หรือ package ของตนเองได้

Requires Craft 0.1.4. This local package has `entry = "library"`; its router,
middleware pipeline, validation and response policy are entirely `.craft`.
Go implements only the language, lifecycle and native HTTP transport.

Copy `examples/api-server` and this package while preserving their relative
layout, or edit `dependency.api` to point to the copied package root. Keep
`dependency-version.api = "0.1.0"`. Import `api` separately in every source
file that references exported declarations.

```craft
import api "api"

func hello(request: api.Request): HttpResponse {
    return api.text(200, "Hello from Craft")
}

func stamp(request: api.Request, response: HttpResponse): HttpResponse {
    var next: HttpResponse = response
    next.headers.set("x-example", ["Craft"])
    return next
}

func main() {
    var app: api.App = api.create()
    app = api.add(app, "GET", "/hello", hello)
    app = api.useAfter(app, stamp)
    api.serve(std.http.config("127.0.0.1:8080"), app)
}
```

Run `craft check`, `craft test`, then `craft run` in the application directory.
No Go change or CLI rebuild is required when adding this handler or hook.
Run this package's pure unit tests with `craft test` from this directory.

## Contracts

- `create/add/group/useBefore/useAfter` return a new App; always assign it.
- `add` accepts GET/POST/PUT/PATCH/DELETE/HEAD/OPTIONS; routes are case-sensitive,
  exact segments with `:name` parameters. Trailing slash is significant.
- Highest count of static segments wins among routes of the requested method.
  Equal-specificity overlapping routes for the same method are rejected.
  Empty/repeated parameter names and duplicate registrations are rejected.
- `group(app, prefix, routes)` concatenates prefix and each Route.path, then
  validates through add; use `/v1` + `/health`, not a double slash.
- Before hooks run in registration order, returning Gate via `proceed(request)`
  or `stop(request,response)`. After hooks run in reverse registration order,
  including after a short circuit and 404/405. Before hooks run after route
  selection; changing raw.path/method does not select a different route.
- A thrown handler/hook error stops the pipeline and yields generic500 JSON;
  failing hooks are not retried. Internal details go to server logs.
- No matching path gives404; matching path with another method gives405 with
  sorted Allow. HEAD and OPTIONS require explicit routes; no automatic GET
  fallback or OPTIONS synthesis. Native transport suppresses HEAD body.
- `Request.params` contains decoded path parameters; `Request.data` carries
  request-local JSON values. `queryValues` reads repeated query values.
- `json/text/errorResponse/parseJSON` provide explicit response/parse helpers.
  `requiredString` and `integerRange` return FieldError arrays;
  `validationResponse` returns422. The application decides when to validate.
- Error JSON uses `error.code` and `error.message`, with `error.fields` for
  validation errors. Transport errors before/after Craft dispatch use generic
  plain text and do not run Craft middleware.

This is a stateless reference framework. No database, shared mutable cache,
authentication, automatic JSON-to-Struct binding, dynamic routes or closures
are implied. Extend controllers, middleware and validation with Craft named
functions. See [language/native contracts](../../docs/REV4-LANGUAGE.md) and
[acceptance steps](../../docs/TRY-REV4.md).
