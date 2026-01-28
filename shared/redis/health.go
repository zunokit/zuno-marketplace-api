package redis

import (
	"context"
	"fmt"
	"time"
)

// Health checks Redis connectivity
func Health(ctx context.Context) error {
	if client == nil {
		return fmt.Errorf("redis client not initialized")
	}

	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	return client.Ping(ctx).Err()
}
