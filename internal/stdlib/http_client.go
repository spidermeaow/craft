package stdlib

import (
	"context"
	rt "craft/internal/runtime"
	"crypto/x509"
	"errors"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// No environment proxy or redirects: callers choose every outbound destination.
// The transport is shared so sequential calls can reuse idle connections.
var craftHTTPClientTransport = &http.Transport{
	Proxy:                  nil,
	MaxIdleConns:           64,
	MaxIdleConnsPerHost:    8,
	MaxConnsPerHost:        64,
	IdleConnTimeout:        30 * time.Second,
	TLSHandshakeTimeout:    0,
	ResponseHeaderTimeout:  0,
	MaxResponseHeaderBytes: 1024 * 1024,
	DisableCompression:     true,
}

var craftHTTPClient = &http.Client{
	Transport: craftHTTPClientTransport,
	CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
		return http.ErrUseLastResponse
	},
}

func httpClientError(kind, message string) error { return toolingError(kind, message) }

func httpClientFailure(ctx context.Context, err error) error {
	if errors.Is(ctx.Err(), context.Canceled) || errors.Is(err, context.Canceled) {
		return httpClientError("HttpClientCancelledError", "HTTP request was cancelled")
	}
	if errors.Is(ctx.Err(), context.DeadlineExceeded) || errors.Is(err, context.DeadlineExceeded) {
		return httpClientError("HttpClientTimeoutError", "HTTP request timed out")
	}
	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return httpClientError("HttpClientTimeoutError", "HTTP request timed out")
	}
	var unknownAuthority x509.UnknownAuthorityError
	var hostError x509.HostnameError
	var certificateError x509.CertificateInvalidError
	if errors.As(err, &unknownAuthority) || errors.As(err, &hostError) || errors.As(err, &certificateError) {
		return httpClientError("HttpClientTLSError", "TLS certificate validation failed")
	}
	return httpClientError("HttpClientNetworkError", "HTTP request failed")
}

func validHTTPToken(value string) bool {
	if value == "" {
		return false
	}
	for i := 0; i < len(value); i++ {
		c := value[i]
		if c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'z' {
			continue
		}
		if !strings.ContainsRune("!#$%&'*+-.^_`|~", rune(c)) {
			return false
		}
	}
	return true
}

func sendHTTPClient(ctx context.Context, request, config *rt.Struct) (any, error) {
	method, ok := request.Fields["method"].(string)
	if !ok || !validHTTPToken(method) {
		return nil, httpClientError("HttpClientConfigError", "invalid HTTP method")
	}
	switch method {
	case "GET", "POST", "PUT", "PATCH", "DELETE", "HEAD":
	default:
		return nil, httpClientError("HttpClientConfigError", "unsupported HTTP method")
	}
	rawURL, ok := request.Fields["url"].(string)
	if !ok || len(rawURL) > 1024*1024 {
		return nil, httpClientError("HttpClientConfigError", "invalid HTTP URL")
	}
	u, err := url.Parse(rawURL)
	if err != nil || u == nil || (u.Scheme != "http" && u.Scheme != "https") || u.Hostname() == "" || u.User != nil || u.Fragment != "" || u.Opaque != "" {
		return nil, httpClientError("HttpClientConfigError", "HTTP URL must be absolute http or https without userinfo or fragment")
	}
	body, ok := request.Fields["body"].(Bytes)
	if !ok || len(body.data) > MaxBodyBytes {
		return nil, httpClientError("HttpClientConfigError", "HTTP request body exceeds 16 MiB")
	}
	timeout, okTimeout := config.Fields["timeoutMs"].(int64)
	maxBody, okBody := config.Fields["maxResponseBytes"].(int64)
	maxHeaders, okHeaders := config.Fields["maxResponseHeaderBytes"].(int64)
	if !okTimeout || !okBody || !okHeaders || timeout < 1 || timeout > 300000 || maxBody < 1 || maxBody > MaxBodyBytes || maxHeaders < 1024 || maxHeaders > 1024*1024 {
		return nil, httpClientError("HttpClientConfigError", "HTTP client config is outside allowed ranges")
	}
	headers, ok := request.Fields["headers"].(*rt.Map)
	if !ok {
		return nil, httpClientError("HttpClientConfigError", "HTTP request headers must be a map")
	}

	requestCtx, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Millisecond)
	defer cancel()
	goRequest, err := http.NewRequestWithContext(requestCtx, method, rawURL, strings.NewReader(body.data))
	if err != nil {
		return nil, httpClientError("HttpClientConfigError", "invalid HTTP request")
	}
	headerBytes := 0
	for key, rawValues := range headers.Values {
		if !validHTTPToken(key) {
			return nil, httpClientError("HttpClientConfigError", "invalid HTTP header name")
		}
		switch strings.ToLower(key) {
		case "host", "content-length", "transfer-encoding", "connection", "trailer", "upgrade", "proxy-authorization", "proxy-connection":
			return nil, httpClientError("HttpClientConfigError", "transport-owned HTTP header is not configurable")
		}
		values, ok := rawValues.(*rt.Array)
		if !ok {
			return nil, httpClientError("HttpClientConfigError", "HTTP header values must be strings")
		}
		for _, rawValue := range values.Items {
			value, ok := rawValue.(string)
			if !ok || !validHTTPHeaderValue(value) {
				return nil, httpClientError("HttpClientConfigError", "invalid HTTP header value")
			}
			headerBytes += len(key) + len(value)
			if headerBytes > 65536 {
				return nil, httpClientError("HttpClientConfigError", "HTTP request headers exceed 64 KiB")
			}
			goRequest.Header.Add(key, value)
		}
	}

	response, err := craftHTTPClient.Do(goRequest)
	if err != nil {
		return nil, httpClientFailure(requestCtx, err)
	}
	defer response.Body.Close()
	responseHeaderBytes := 0
	for key, values := range response.Header {
		for _, value := range values {
			responseHeaderBytes += len(key) + len(value)
			if int64(responseHeaderBytes) > maxHeaders {
				return nil, httpClientError("HttpClientLimitError", "HTTP response headers exceed configured limit")
			}
		}
	}
	if response.ContentLength > maxBody {
		return nil, httpClientError("HttpClientLimitError", "HTTP response body exceeds configured limit")
	}
	data, err := io.ReadAll(io.LimitReader(response.Body, maxBody+1))
	if err != nil {
		return nil, httpClientFailure(requestCtx, err)
	}
	if int64(len(data)) > maxBody {
		return nil, httpClientError("HttpClientLimitError", "HTTP response body exceeds configured limit")
	}
	return nativeStruct("HttpClientResponse", int64(response.StatusCode), multiMap(response.Header, true), Bytes{string(data)}), nil
}

func validHTTPHeaderValue(value string) bool {
	for i := 0; i < len(value); i++ {
		if value[i] < 32 && value[i] != '\t' || value[i] == 127 {
			return false
		}
	}
	return true
}
