package health

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockCheckable is a mock implementation of Checkable
type MockCheckable struct {
	mock.Mock
}

func (m *MockCheckable) Check(ctx context.Context) (string, error) {
	args := m.Called(ctx)
	return args.String(0), args.Error(1)
}

func TestRegistry_CheckAll(t *testing.T) {
	t.Run("all healthy", func(t *testing.T) {
		registry := NewRegistry()
		mockChecker1 := new(MockCheckable)
		mockChecker2 := new(MockCheckable)

		mockChecker1.On("Check", mock.Anything).Return("healthy", nil)
		mockChecker2.On("Check", mock.Anything).Return("healthy", nil)

		registry.Register("service1", mockChecker1)
		registry.Register("service2", mockChecker2)

		results := registry.CheckAll(context.Background())

		assert.Equal(t, "healthy", results["status"])
		assert.Equal(t, "healthy", results["service1"])
		assert.Equal(t, "healthy", results["service2"])
	})

	t.Run("one unhealthy", func(t *testing.T) {
		registry := NewRegistry()
		mockChecker1 := new(MockCheckable)
		mockChecker2 := new(MockCheckable)

		mockChecker1.On("Check", mock.Anything).Return("healthy", nil)
		mockChecker2.On("Check", mock.Anything).Return("unhealthy", nil)

		registry.Register("service1", mockChecker1)
		registry.Register("service2", mockChecker2)

		results := registry.CheckAll(context.Background())

		assert.Equal(t, "unhealthy", results["status"])
		assert.Equal(t, "healthy", results["service1"])
		assert.Equal(t, "unhealthy", results["service2"])
	})

	t.Run("checker error", func(t *testing.T) {
		registry := NewRegistry()
		mockChecker1 := new(MockCheckable)

		mockChecker1.On("Check", mock.Anything).Return("unhealthy", errors.New("connection failed"))

		registry.Register("service1", mockChecker1)

		results := registry.CheckAll(context.Background())

		assert.Equal(t, "unhealthy", results["status"])
		assert.Equal(t, "unhealthy", results["service1"])
	})
}
