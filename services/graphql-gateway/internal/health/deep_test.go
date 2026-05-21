package health

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type funcChecker struct {
	fn func(ctx context.Context) (string, error)
}

func (f funcChecker) Check(ctx context.Context) (string, error) {
	return f.fn(ctx)
}

func TestCheckAllDeep_AllHealthy(t *testing.T) {
	r := NewRegistry()
	r.Register("svc-a", funcChecker{fn: func(ctx context.Context) (string, error) {
		return "healthy", nil
	}})
	r.Register("svc-b", funcChecker{fn: func(ctx context.Context) (string, error) {
		return "healthy", nil
	}})

	res := r.CheckAllDeep(context.Background(), 0)

	assert.Equal(t, "healthy", res.Status)
	assert.Len(t, res.Checks, 2)
	assert.Equal(t, "healthy", res.Checks["svc-a"].Status)
	assert.Equal(t, "healthy", res.Checks["svc-b"].Status)
	assert.Empty(t, res.Checks["svc-a"].Error)
}

func TestCheckAllDeep_OneFailedSurfacesUnhealthy(t *testing.T) {
	r := NewRegistry()
	r.Register("svc-a", funcChecker{fn: func(ctx context.Context) (string, error) {
		return "healthy", nil
	}})
	r.Register("svc-b", funcChecker{fn: func(ctx context.Context) (string, error) {
		return "unhealthy", errors.New("connection refused")
	}})

	res := r.CheckAllDeep(context.Background(), 0)

	assert.Equal(t, "unhealthy", res.Status)
	assert.Equal(t, "healthy", res.Checks["svc-a"].Status)
	assert.Equal(t, "unhealthy", res.Checks["svc-b"].Status)
	assert.Equal(t, "connection refused", res.Checks["svc-b"].Error)
}

func TestCheckAllDeep_PerCheckTimeout(t *testing.T) {
	r := NewRegistry()
	r.Register("fast", funcChecker{fn: func(ctx context.Context) (string, error) {
		return "healthy", nil
	}})
	r.Register("slow", funcChecker{fn: func(ctx context.Context) (string, error) {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-time.After(500 * time.Millisecond):
			return "healthy", nil
		}
	}})

	res := r.CheckAllDeep(context.Background(), 20*time.Millisecond)

	assert.Equal(t, "unhealthy", res.Status)
	assert.Equal(t, "healthy", res.Checks["fast"].Status)
	assert.Equal(t, "timeout", res.Checks["slow"].Status)
}

func TestCheckAllDeep_NormalisesEmptyStatus(t *testing.T) {
	r := NewRegistry()
	r.Register("no-status", funcChecker{fn: func(ctx context.Context) (string, error) {
		return "", nil
	}})

	res := r.CheckAllDeep(context.Background(), 0)

	assert.Equal(t, "unhealthy", res.Status)
	assert.Equal(t, "unknown", res.Checks["no-status"].Status)
}

func TestCheckAllDeep_RecordsLatency(t *testing.T) {
	r := NewRegistry()
	r.Register("slow-ish", funcChecker{fn: func(ctx context.Context) (string, error) {
		time.Sleep(10 * time.Millisecond)
		return "healthy", nil
	}})

	res := r.CheckAllDeep(context.Background(), 0)

	assert.GreaterOrEqual(t, res.Checks["slow-ish"].LatencyMs, int64(10))
	assert.GreaterOrEqual(t, res.LatencyMs, int64(10))
}

func TestCheckAllDeep_EmptyRegistry(t *testing.T) {
	r := NewRegistry()
	res := r.CheckAllDeep(context.Background(), 0)

	assert.Equal(t, "healthy", res.Status)
	assert.Empty(t, res.Checks)
}

func TestDeepHandler_Healthy200(t *testing.T) {
	r := NewRegistry()
	r.Register("svc-a", funcChecker{fn: func(ctx context.Context) (string, error) {
		return "healthy", nil
	}})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz/deep", nil)
	r.DeepHandler(0).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

	var got DeepResult
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&got))
	assert.Equal(t, "healthy", got.Status)
	assert.Equal(t, "healthy", got.Checks["svc-a"].Status)
}

func TestDeepHandler_Unhealthy503(t *testing.T) {
	r := NewRegistry()
	r.Register("svc-a", funcChecker{fn: func(ctx context.Context) (string, error) {
		return "unhealthy", errors.New("db down")
	}})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz/deep", nil)
	r.DeepHandler(0).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusServiceUnavailable, rec.Code)

	var got DeepResult
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&got))
	assert.Equal(t, "unhealthy", got.Status)
	assert.Equal(t, "db down", got.Checks["svc-a"].Error)
}
