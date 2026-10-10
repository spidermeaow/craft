package tests

import (
	"bytes"
	"context"
	"craft/internal/cli"
	"craft/internal/interpreter"
	"craft/internal/project"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestRev46CraftHTTPClientPublicAPI(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.Header.Get("X-Check") != "craft" {
			t.Errorf("unexpected request method or header")
		}
		w.Header().Add("Set-Cookie", "a=1")
		w.Header().Add("Set-Cookie", "b=2")
		w.WriteHeader(201)
		_, _ = w.Write([]byte("created"))
	}))
	defer server.Close()
	t.Setenv("CRAFT_REV46_URL", server.URL)
	source := `func main() {
    var request: HttpClientRequest = std.http.clientRequest("POST", std.env.get("CRAFT_REV46_URL"), std.bytes.fromString("hello"))
    request.headers.set("x-check", ["craft"])
    var config: HttpClientConfig = std.http.clientConfig()
    config.timeoutMs = 3000
    let response: HttpClientResponse = std.http.send(request, config)
    assert response.status == 201
    assert response.body.text() == "created"
    assert response.headers.require("set-cookie").length == 2
    print("HTTP client public API passed")
}`
	p := checked(t, source)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var out bytes.Buffer
	if err := interpreter.Run(ctx, p, &out); err != nil {
		t.Fatalf("Craft HTTP client failed: %v", err)
	}
	if !strings.Contains(out.String(), "HTTP client public API passed") {
		t.Fatalf("unexpected output: %q", out.String())
	}
}

func TestRev46HTTPClientExampleAndBundle(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()
	t.Setenv("CRAFT_HTTP_CLIENT_URL", server.URL)
	const dir = "../examples/rev46-http-client"
	var out, stderr bytes.Buffer
	if code := cli.Run(context.Background(), []string{"run"}, dir, &out, &stderr); code != 0 || !strings.Contains(out.String(), "HTTP status: 200") {
		t.Fatalf("example run failed: code %d, output %q, error %q", code, out.String(), stderr.String())
	}
	p, err := project.Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	bundle, err := p.Build()
	if err != nil {
		t.Fatal(err)
	}
	loaded, err := project.LoadBundle(bundle)
	if err != nil {
		t.Fatal(err)
	}
	program, _, err := loaded.Compile()
	if err != nil {
		t.Fatal(err)
	}
	out.Reset()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := interpreter.Run(ctx, program, &out); err != nil || !strings.Contains(out.String(), "Response bytes: 11") {
		t.Fatalf("bundle run failed: %v, output %q", err, out.String())
	}
}
