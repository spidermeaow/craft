package stdlib

import (
	"context"
	rt "craft/internal/runtime"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

var errServerStop = errors.New("server stopped")
var errServerShutdown = errors.New("server shutdown")
var errRequestTimeout = errors.New("request timeout")

type HTTPDispatch func(context.Context, *rt.Struct) (*rt.Struct, error)

type httpSettings struct {
	address                                string
	concurrent, connections, body, headers int
	read, write, idle, request, shutdown   time.Duration
}

func settings(v *rt.Struct) (httpSettings, error) {
	s := httpSettings{address: v.Fields["address"].(string)}
	if s.address == "" {
		return s, fmt.Errorf("HTTP address is required")
	}
	get := func(n string, min, max int64) (int64, error) {
		x := v.Fields[n].(int64)
		if x < min || x > max {
			return 0, fmt.Errorf("%s must be %d..%d", n, min, max)
		}
		return x, nil
	}
	for _, item := range []struct {
		n        string
		dst      *int
		min, max int64
	}{{"maxConcurrent", &s.concurrent, 1, 256}, {"maxConnections", &s.connections, 1, 1024}, {"maxBodyBytes", &s.body, 1, MaxBodyBytes}, {"maxHeaderBytes", &s.headers, 1024, 1024 * 1024}} {
		n, e := get(item.n, item.min, item.max)
		if e != nil {
			return s, e
		}
		*item.dst = int(n)
	}
	for _, item := range []struct {
		n   string
		dst *time.Duration
	}{{"readTimeoutMs", &s.read}, {"writeTimeoutMs", &s.write}, {"idleTimeoutMs", &s.idle}, {"requestTimeoutMs", &s.request}, {"shutdownTimeoutMs", &s.shutdown}} {
		n, e := get(item.n, 1, 300000)
		if e != nil {
			return s, e
		}
		*item.dst = time.Duration(n) * time.Millisecond
	}
	if s.write < s.request {
		return s, fmt.Errorf("writeTimeoutMs must be >= requestTimeoutMs")
	}
	return s, nil
}

// Limit connections before net/http allocates a goroutine for each accepted
// connection. At most one extra connection is waiting inside Accept; there is
// no application request queue. OS listen backlog remains controlled by the OS.
type limitedListener struct {
	net.Listener
	slots chan struct{}
	done  chan struct{}
	once  sync.Once
}

func (l *limitedListener) Accept() (net.Conn, error) {
	select {
	case l.slots <- struct{}{}:
	case <-l.done:
		return nil, net.ErrClosed
	}
	c, e := l.Listener.Accept()
	if e != nil {
		<-l.slots
		return nil, e
	}
	return &limitedConn{Conn: c, release: func() { <-l.slots }}, nil
}
func (l *limitedListener) Close() error {
	l.once.Do(func() { close(l.done) })
	return l.Listener.Close()
}

type limitedConn struct {
	net.Conn
	once    sync.Once
	release func()
}

func (c *limitedConn) Close() error { e := c.Conn.Close(); c.once.Do(c.release); return e }

// ServeHTTPTransport knows no routes, controllers, validation or middleware.
// Every valid bounded HTTP request is handed to the same Craft dispatcher.
func ServeHTTPTransport(ctx context.Context, config *rt.Struct, dispatch HTTPDispatch, out io.Writer) error {
	s, e := settings(config)
	if e != nil {
		return e
	}
	listener, e := net.Listen("tcp", s.address)
	if e != nil {
		return e
	}
	return serveListener(ctx, s, listener, dispatch, out)
}
func serveListener(ctx context.Context, s httpSettings, listener net.Listener, dispatch HTTPDispatch, out io.Writer) error {
	ln := &limitedListener{Listener: listener, slots: make(chan struct{}, s.connections), done: make(chan struct{})}
	defer ln.Close()
	base, cancelBase := context.WithCancelCause(context.WithoutCancel(ctx))
	defer cancelBase(errServerShutdown)
	stop := make(chan struct{})
	var stopOnce sync.Once
	requestStop := func() { stopOnce.Do(func() { close(stop) }) }
	slots := make(chan struct{}, s.concurrent)
	var requests sync.WaitGroup
	var admission sync.Mutex
	closing := false
	closeAdmission := func() { admission.Lock(); closing = true; admission.Unlock() }
	server := &http.Server{ReadHeaderTimeout: s.read, ReadTimeout: s.read, WriteTimeout: s.write, IdleTimeout: s.idle, MaxHeaderBytes: s.headers, BaseContext: func(net.Listener) context.Context { return base }}
	server.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		admission.Lock()
		if closing {
			admission.Unlock()
			http.Error(w, "server draining", 503)
			return
		}
		requests.Add(1)
		admission.Unlock()
		defer requests.Done()
		select {
		case slots <- struct{}{}:
			defer func() { <-slots }()
		default:
			http.Error(w, "server busy", http.StatusServiceUnavailable)
			return
		}
		rctx, release := context.WithTimeoutCause(r.Context(), s.request, errRequestTimeout)
		defer release()
		handle := &Context{ctx: rctx, stop: requestStop}
		rctx = context.WithValue(rctx, requestContextKey{}, handle)
		defer handle.expired.Store(true)
		// Bound body reading by both the configured read and request durations.
		readLimit := min(s.request, s.read)
		_ = http.NewResponseController(w).SetReadDeadline(time.Now().Add(readLimit))
		body, e := io.ReadAll(http.MaxBytesReader(w, r.Body, int64(s.body)))
		_ = r.Body.Close()
		// The read deadline only protects request-body consumption. Leaving it on
		// the connection can make a completed request look like a later transport
		// failure while its handler is waiting on a database operation.
		_ = http.NewResponseController(w).SetReadDeadline(time.Time{})
		if e != nil {
			var tooLarge *http.MaxBytesError
			if errors.As(e, &tooLarge) {
				http.Error(w, "request body too large", 413)
			} else if timeout, ok := e.(net.Error); ok && timeout.Timeout() {
				http.Error(w, "request body timeout", 408)
			} else if rctx.Err() == nil {
				http.Error(w, "invalid request body", 400)
			}
			return
		}
		query, e := url.ParseQuery(r.URL.RawQuery)
		if e != nil {
			http.Error(w, "invalid query encoding", 400)
			return
		}
		// Decode path exactly once. Encoded separators cannot smuggle a new route
		// segment; query decoding is independent and preserves duplicate values.
		if !validRequestPath(r.URL) {
			http.Error(w, "invalid encoded path separator", 400)
			return
		}
		request := nativeStruct("HttpRequest", r.Method, r.RequestURI, r.URL.Path, multiMap(query, false), multiMap(r.Header, true), Bytes{string(body)}, handle)
		response, e := dispatch(rctx, request)
		if e != nil {
			if errors.Is(context.Cause(rctx), errRequestTimeout) {
				http.Error(w, "request timeout", 504)
				return
			}
			if rctx.Err() != nil {
				return
			}
			fmt.Fprintf(out, "HTTP handler failed: %v\n", e)
			http.Error(w, "internal server error", 500)
			return
		}
		if rctx.Err() != nil {
			if errors.Is(context.Cause(rctx), errRequestTimeout) {
				http.Error(w, "request timeout", 504)
			}
			return
		}
		status, headers, data, e := validateResponse(response)
		if e != nil {
			fmt.Fprintf(out, "HTTP response rejected: %v\n", e)
			http.Error(w, "internal server error", 500)
			return
		}
		for k, values := range headers {
			for _, v := range values {
				w.Header().Add(k, v)
			}
		}
		w.WriteHeader(status)
		if r.Method != "HEAD" && status != 204 && status != 304 {
			if _, e = io.WriteString(w, data); e != nil {
				fmt.Fprintf(out, "HTTP response write failed: %v\n", e)
			}
		}
	})
	served := make(chan error, 1)
	go func() { served <- server.Serve(ln) }()
	fmt.Fprintf(out, "HTTP listening on %s\n", listener.Addr())
	var cause error
	select {
	case e := <-served:
		closeAdmission()
		cancelBase(errServerShutdown)
		_ = server.Close()
		requests.Wait()
		if errors.Is(e, http.ErrServerClosed) {
			return nil
		}
		return e
	case <-ctx.Done():
		cause = ctx.Err()
	case <-stop:
		cause = errServerStop
	}
	closeAdmission()
	drain, cancelDrain := context.WithTimeout(context.Background(), s.shutdown)
	e := server.Shutdown(drain)
	cancelDrain()
	if e != nil {
		cancelBase(errServerShutdown)
		_ = server.Close()
		fmt.Fprintln(out, "HTTP shutdown drain deadline exceeded; cancelling active requests and waiting for cleanup")
	}
	<-served
	// Craft loops and native waits are cooperative. Blocking OS filesystem calls
	// must return before their task can unwind; this limitation is documented.
	requests.Wait()
	if e != nil {
		return errors.Join(cause, fmt.Errorf("HTTP shutdown drain deadline exceeded; active request cleanup completed: %w", e))
	}
	if cause == errServerStop {
		return nil
	}
	return cause
}

func validHeaderName(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r > 127 || !(r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || strings.ContainsRune("!#$%&'*+-.^_`|~", r)) {
			return false
		}
	}
	return true
}
func validateResponse(v *rt.Struct) (int, map[string][]string, string, error) {
	if v == nil || v.Name != "HttpResponse" {
		return 0, nil, "", fmt.Errorf("dispatch did not return HttpResponse")
	}
	status := v.Fields["status"].(int64)
	if status < 200 || status > 599 {
		return 0, nil, "", fmt.Errorf("response status must be 200..599")
	}
	body := v.Fields["body"].(Bytes).data
	if len(body) > MaxBodyBytes {
		return 0, nil, "", fmt.Errorf("response exceeds 16 MiB")
	}
	headers := map[string][]string{}
	size := 0
	for k, a := range v.Fields["headers"].(*rt.Map).Values {
		if !validHeaderName(k) {
			return 0, nil, "", fmt.Errorf("invalid response header")
		}
		switch strings.ToLower(k) {
		case "content-length", "transfer-encoding", "connection", "trailer", "upgrade":
			return 0, nil, "", fmt.Errorf("transport-owned response header %s", k)
		}
		for _, item := range a.(*rt.Array).Items {
			value := item.(string)
			for _, r := range value {
				if r == '\r' || r == '\n' || r == 127 || (r < 32 && r != '\t') {
					return 0, nil, "", fmt.Errorf("invalid response header value")
				}
			}
			size += len(k) + len(value)
			if size > 65536 {
				return 0, nil, "", fmt.Errorf("response headers exceed 64 KiB")
			}
			key := http.CanonicalHeaderKey(k)
			headers[key] = append(headers[key], value)
		}
	}
	return int(status), headers, body, nil
}
