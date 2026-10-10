# Rev.46 HTTP client example

Set `CRAFT_HTTP_CLIENT_URL` to an endpoint that you control, then run `craft check` and `craft run`. The example performs one GET with a 5-second total timeout and prints the response status and byte count. It does not print the URL, credentials, or body. To run from a source checkout, build the current CLI first; the installed CLI must be version 0.1.46 or later.
