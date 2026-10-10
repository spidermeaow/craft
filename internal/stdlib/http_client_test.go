package stdlib

import (
	"context"
	rt "craft/internal/runtime"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func clientFixture(method, target string, body Bytes) *rt.Struct {
	return nativeStruct("HttpClientRequest", method, target, rt.NewMap(), body)
}

func clientErrorKind(t *testing.T, err error, kind string) {
	t.Helper()
	var exception *rt.Exception
	if !errors.As(err, &exception) || exception.Kind != kind {
		t.Fatalf("expected %s, got %v", kind, err)
	}
}

func TestRev46HTTPClientMethodsStatusAndHeaders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Set-Cookie", "a=1")
		w.Header().Add("Set-Cookie", "b=2")
		if r.Header.Get("X-Check") != "✓" {
			t.Errorf("missing request header")
		}
		body, _ := io.ReadAll(r.Body)
		w.WriteHeader(418)
		_, _ = w.Write(append([]byte(r.Method+":"), body...))
	}))
	defer server.Close()
	for _, method := range []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD"} {
		request := clientFixture(method, server.URL, Bytes{"data"})
		request.Fields["headers"].(*rt.Map).Set("X-Check", &rt.Array{Items: []any{"✓"}})
		value, err := sendHTTPClient(context.Background(), request, invokeClientConfig(t))
		if err != nil {
			t.Fatalf("%s: %v", method, err)
		}
		result := value.(*rt.Struct)
		if result.Fields["status"] != int64(418) {
			t.Fatalf("%s: status %v", method, result.Fields["status"])
		}
		cookies := result.Fields["headers"].(*rt.Map).Values["set-cookie"].(*rt.Array)
		if len(cookies.Items) != 2 {
			t.Fatalf("%s: repeated headers %v", method, cookies.Items)
		}
		if method == "HEAD" {
			if result.Fields["body"].(Bytes).data != "" {
				t.Fatal("HEAD response has body")
			}
		} else if !strings.HasPrefix(result.Fields["body"].(Bytes).data, method+":") {
			t.Fatalf("%s: unexpected body", method)
		}
	}
}

func invokeClientConfig(t *testing.T) *rt.Struct {
	t.Helper()
	value, err := invokeRev4(context.Background(), "std.http.clientConfig", nil)
	if err != nil {
		t.Fatal(err)
	}
	return value.(*rt.Struct)
}

func TestRev46HTTPClientRedirectAndTLS(t *testing.T) {
	var redirected bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/next" {
			redirected = true
			return
		}
		w.Header().Set("Location", "/next")
		w.WriteHeader(302)
	}))
	defer server.Close()
	result, err := sendHTTPClient(context.Background(), clientFixture("GET", server.URL, Bytes{}), invokeClientConfig(t))
	if err != nil || result.(*rt.Struct).Fields["status"] != int64(302) || redirected {
		t.Fatalf("redirect was followed or lost: %v, %v", result, err)
	}
	tlsServer := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(200) }))
	defer tlsServer.Close()
	_, err = sendHTTPClient(context.Background(), clientFixture("GET", tlsServer.URL, Bytes{}), invokeClientConfig(t))
	clientErrorKind(t, err, "HttpClientTLSError")
}

func TestRev46HTTPClientLimitsAndValidation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/headers" {
			w.Header().Set("X-Large", strings.Repeat("a", 1100))
		}
		_, _ = w.Write([]byte(strings.Repeat("b", 100)))
	}))
	defer server.Close()
	config := invokeClientConfig(t)
	config.Fields["maxResponseBytes"] = int64(10)
	_, err := sendHTTPClient(context.Background(), clientFixture("GET", server.URL, Bytes{}), config)
	clientErrorKind(t, err, "HttpClientLimitError")
	config.Fields["maxResponseBytes"] = int64(100)
	config.Fields["maxResponseHeaderBytes"] = int64(1024)
	_, err = sendHTTPClient(context.Background(), clientFixture("GET", server.URL+"/headers", Bytes{}), config)
	clientErrorKind(t, err, "HttpClientLimitError")
	for _, target := range []string{"file:///etc/passwd", "http://user:password@example.com/", "/relative", "https://example.com/#fragment"} {
		_, err := sendHTTPClient(context.Background(), clientFixture("GET", target, Bytes{}), config)
		clientErrorKind(t, err, "HttpClientConfigError")
	}
	request := clientFixture("GET", server.URL, Bytes{})
	request.Fields["headers"].(*rt.Map).Set("Authorization", &rt.Array{Items: []any{"secret\r\nX-Evil: yes"}})
	_, err = sendHTTPClient(context.Background(), request, config)
	clientErrorKind(t, err, "HttpClientConfigError")
	if strings.Contains(err.Error(), "secret") {
		t.Fatal("secret leaked in header error")
	}
}

func TestRev46HTTPClientTimeoutAndCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/slow" {
			<-r.Context().Done()
			return
		}
		w.WriteHeader(200)
	}))
	defer server.Close()
	config := invokeClientConfig(t)
	config.Fields["timeoutMs"] = int64(30)
	_, err := sendHTTPClient(context.Background(), clientFixture("GET", server.URL+"/slow", Bytes{}), config)
	clientErrorKind(t, err, "HttpClientTimeoutError")
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	config.Fields["timeoutMs"] = int64(1000)
	_, err = sendHTTPClient(ctx, clientFixture("GET", server.URL, Bytes{}), config)
	clientErrorKind(t, err, "HttpClientCancelledError")
	value, err := sendHTTPClient(context.Background(), clientFixture("GET", server.URL, Bytes{}), config)
	if err != nil || value.(*rt.Struct).Fields["status"] != int64(200) {
		t.Fatalf("connection did not recover: %v", err)
	}
}

func TestRev46HTTPClientConnectionReuse(t *testing.T) {
	var opened atomic.Int32
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(204) }))
	server.Config.ConnState = func(_ net.Conn, state http.ConnState) {
		if state == http.StateNew {
			opened.Add(1)
		}
	}
	server.Start()
	defer server.Close()
	config := invokeClientConfig(t)
	for i := 0; i < 20; i++ {
		_, err := sendHTTPClient(context.Background(), clientFixture("GET", server.URL, Bytes{}), config)
		if err != nil {
			t.Fatal(err)
		}
	}
	// A subsequent request must still succeed after all response bodies were closed.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if _, err := sendHTTPClient(ctx, clientFixture("GET", server.URL, Bytes{}), config); err != nil {
		t.Fatal(err)
	}
	if got := opened.Load(); got != 1 {
		t.Fatalf("expected one reused connection, got %d", got)
	}
}
