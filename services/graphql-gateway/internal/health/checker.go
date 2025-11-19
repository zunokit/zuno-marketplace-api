package health

import (
	"context"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health/grpc_health_v1"
)

// Checkable defines the interface for health checks
type Checkable interface {
	Check(ctx context.Context) (string, error)
}

// ServiceHealthChecker implements Checkable for gRPC services
type ServiceHealthChecker struct {
	conn *grpc.ClientConn
}

// NewServiceHealthChecker creates a new health checker for a gRPC service
func NewServiceHealthChecker(conn *grpc.ClientConn) *ServiceHealthChecker {
	return &ServiceHealthChecker{conn: conn}
}

// Check verifies the health of the gRPC service
func (c *ServiceHealthChecker) Check(ctx context.Context) (string, error) {
	client := grpc_health_v1.NewHealthClient(c.conn)
	resp, err := client.Check(ctx, &grpc_health_v1.HealthCheckRequest{})
	if err != nil {
		return "unhealthy", err
	}
	if resp.Status != grpc_health_v1.HealthCheckResponse_SERVING {
		return "unhealthy", nil
	}
	return "healthy", nil
}

// Registry manages multiple health checkers
type Registry struct {
	checkers map[string]Checkable
	mu       sync.RWMutex
}

// NewRegistry creates a new health registry
func NewRegistry() *Registry {
	return &Registry{
		checkers: make(map[string]Checkable),
	}
}

// Register adds a checker to the registry
func (r *Registry) Register(name string, checker Checkable) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.checkers[name] = checker
}

// CheckAll runs all registered health checks
func (r *Registry) CheckAll(ctx context.Context) map[string]string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	results := make(map[string]string)
	overallStatus := "healthy"

	var wg sync.WaitGroup
	var mu sync.Mutex

	for name, checker := range r.checkers {
		wg.Add(1)
		go func(n string, c Checkable) {
			defer wg.Done()
			status, _ := c.Check(ctx)

			mu.Lock()
			results[n] = status
			if status != "healthy" {
				overallStatus = "unhealthy"
			}
			mu.Unlock()
		}(name, checker)
	}

	wg.Wait()
	results["status"] = overallStatus
	return results
}
