package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/zunokit/zuno-marketplace-api/services/collection-service/internal/models"
	"github.com/zunokit/zuno-marketplace-api/shared/utils"
	"gorm.io/gorm"
)

var (
	ErrEventAlreadyProcessed = errors.New("event already processed")
	ErrLockNotAcquired       = errors.New("advisory lock not acquired")
)

// ProcessedEventRepository handles processed event data operations
type ProcessedEventRepository interface {
	// WithTransaction executes fn within a single database transaction.
	// Useful for holding transaction-scoped advisory locks for the duration of webhook processing.
	WithTransaction(ctx context.Context, fn func(repo ProcessedEventRepository) error) error

	// AcquireEventLock tries to acquire a PostgreSQL transaction-scoped advisory lock for event serialization.
	// Returns ErrLockNotAcquired if another transaction is currently processing the same event.
	AcquireEventLock(ctx context.Context, eventID string) error

	// IsEventProcessed checks if an event has been processed
	IsEventProcessed(ctx context.Context, eventID string) (bool, error)

	// CreateProcessedEvent marks an event as processed
	CreateProcessedEvent(ctx context.Context, event *models.ProcessedEvent) error

	// GetProcessedEvent retrieves a processed event by event ID
	GetProcessedEvent(ctx context.Context, eventID string) (*models.ProcessedEvent, error)

	// DeleteProcessedEventsBefore cleanup old events (optional retention policy)
	DeleteProcessedEventsBefore(ctx context.Context, beforeDate interface{}) (int64, error)
}

// processedEventRepository implements ProcessedEventRepository
type processedEventRepository struct {
	db *gorm.DB
}

// NewProcessedEventRepository creates a new ProcessedEventRepository
func NewProcessedEventRepository(db *gorm.DB) ProcessedEventRepository {
	return &processedEventRepository{db: db}
}

func (r *processedEventRepository) WithTransaction(ctx context.Context, fn func(repo ProcessedEventRepository) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txRepo := &processedEventRepository{db: tx}
		return fn(txRepo)
	})
}

func (r *processedEventRepository) AcquireEventLock(ctx context.Context, eventID string) error {
	lockKey := utils.GenerateAdvisoryLockKey(eventID)

	var acquired bool
	err := r.db.WithContext(ctx).
		Raw("SELECT pg_try_advisory_xact_lock(?) AS acquired", lockKey).
		Scan(&acquired).
		Error
	if err != nil {
		return fmt.Errorf("failed to acquire advisory lock: %w", err)
	}

	if !acquired {
		return ErrLockNotAcquired
	}

	return nil
}

// IsEventProcessed checks if an event has been processed
func (r *processedEventRepository) IsEventProcessed(ctx context.Context, eventID string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.ProcessedEvent{}).
		Where("event_id = ?", eventID).
		Count(&count).
		Error

	if err != nil {
		return false, fmt.Errorf("failed to check if event is processed: %w", err)
	}

	return count > 0, nil
}

// CreateProcessedEvent marks an event as processed
func (r *processedEventRepository) CreateProcessedEvent(ctx context.Context, event *models.ProcessedEvent) error {
	// Check if already exists
	exists, err := r.IsEventProcessed(ctx, event.EventID)
	if err != nil {
		return err
	}
	if exists {
		return ErrEventAlreadyProcessed
	}

	// Create the record
	if err := r.db.WithContext(ctx).Create(event).Error; err != nil {
		return fmt.Errorf("failed to create processed event: %w", err)
	}

	return nil
}

// GetProcessedEvent retrieves a processed event by event ID
func (r *processedEventRepository) GetProcessedEvent(ctx context.Context, eventID string) (*models.ProcessedEvent, error) {
	var event models.ProcessedEvent
	err := r.db.WithContext(ctx).
		Where("event_id = ?", eventID).
		First(&event).
		Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get processed event: %w", err)
	}

	return &event, nil
}

// DeleteProcessedEventsBefore cleanup old events (optional retention policy)
func (r *processedEventRepository) DeleteProcessedEventsBefore(ctx context.Context, beforeDate interface{}) (int64, error) {
	result := r.db.WithContext(ctx).
		Where("processed_at < ?", beforeDate).
		Delete(&models.ProcessedEvent{})

	if result.Error != nil {
		return 0, fmt.Errorf("failed to delete old processed events: %w", result.Error)
	}

	return result.RowsAffected, nil
}
