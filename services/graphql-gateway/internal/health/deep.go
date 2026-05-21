package health

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"
)

// CheckResult is the per-dependency result returned by a deep health check.
type CheckResult struct {
	Status    string `json:"status"`
	LatencyMs int64  `json:"latency_ms"`
	Error     string `json:"error,omitempty"`
}

// DeepResult is the response returned by Registry.CheckAllDeep and rendered by
// DeepHandler. Status is "healthy" only when every registered check succeeded.
type DeepResult struct {
	Status    string                 `json:"status"`
	LatencyMs int64                  `json:"latency_ms"`
	Checks    map[string]CheckResult `json:"checks"`
}

// CheckAllDeep runs every registered checker in parallel and returns a
// structured result that includes per-check latency and any error message.
//
// A per-check timeout can be supplied with `perCheckTimeout`; pass 0 to honour
// only the caller's context deadline. Checks that exceed the timeout are
// reported as `status: "timeout"` and the overall status is set to
// "unhealthy".
//
// The method never panics on a checker that returns an empty status string —
// it normalises to "unknown" and treats it as unhealthy.
func (r *Registry) CheckAllDeep(ctx context.Context, perCheckTimeout time.Duration) DeepResult {
	r.mu.RLock()
	checkers := make(map[string]Checkable, len(r.checkers))
	for name, c := range r.checkers {
		checkers[name] = c
	}
	r.mu.RUnlock()

	start := time.Now()
	results := make(map[string]CheckResult, len(checkers))
	var mu sync.Mutex
	var wg sync.WaitGroup

	for name, checker := range checkers {
		wg.Add(1)
		go func(n string, c Checkable) {
			defer wg.Done()

			callCtx := ctx
			var cancel context.CancelFunc
			if perCheckTimeout > 0 {
				callCtx, cancel = context.WithTimeout(ctx, perCheckTimeout)
				defer cancel()
			}

			checkStart := time.Now()
			status, err := c.Check(callCtx)
			latency := time.Since(checkStart)

			res := CheckResult{LatencyMs: latency.Milliseconds()}
			switch {
			case err != nil && (callCtx.Err() == context.DeadlineExceeded || ctx.Err() == context.DeadlineExceeded):
				res.Status = "timeout"
				res.Error = err.Error()
			case err != nil:
				if status == "" {
					status = "unhealthy"
				}
				res.Status = status
				res.Error = err.Error()
			case status == "":
				res.Status = "unknown"
			default:
				res.Status = status
			}

			mu.Lock()
			results[n] = res
			mu.Unlock()
		}(name, checker)
	}

	wg.Wait()

	overall := "healthy"
	for _, res := range results {
		if res.Status != "healthy" {
			overall = "unhealthy"
			break
		}
	}

	return DeepResult{
		Status:    overall,
		LatencyMs: time.Since(start).Milliseconds(),
		Checks:    results,
	}
}

// DeepHandler returns an http.Handler that serves the result of CheckAllDeep
// as JSON. It responds with 200 OK when every check is healthy and 503 Service
// Unavailable otherwise, which is the contract expected by Kubernetes-style
// readiness probes.
func (r *Registry) DeepHandler(perCheckTimeout time.Duration) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		result := r.CheckAllDeep(req.Context(), perCheckTimeout)
		w.Header().Set("Content-Type", "application/json")
		if result.Status == "healthy" {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusServiceUnavailable)
		}
		_ = json.NewEncoder(w).Encode(result)
	})
}
