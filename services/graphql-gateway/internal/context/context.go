package context

import (
	"context"
	"net/http"
)

type contextKey string

const (
	httpRequestKey  contextKey = "http_request"
	httpResponseKey contextKey = "http_response"
)

// WithHTTPRequest adds the HTTP request to the context
func WithHTTPRequest(ctx context.Context, r *http.Request) context.Context {
	return context.WithValue(ctx, httpRequestKey, r)
}

// WithHTTPResponse adds the HTTP response writer to the context
func WithHTTPResponse(ctx context.Context, w http.ResponseWriter) context.Context {
	return context.WithValue(ctx, httpResponseKey, w)
}

// GetHTTPRequest retrieves the HTTP request from the context
func GetHTTPRequest(ctx context.Context) (*http.Request, bool) {
	r, ok := ctx.Value(httpRequestKey).(*http.Request)
	return r, ok
}

// GetHTTPResponse retrieves the HTTP response writer from the context
func GetHTTPResponse(ctx context.Context) (http.ResponseWriter, bool) {
	w, ok := ctx.Value(httpResponseKey).(http.ResponseWriter)
	return w, ok
}
