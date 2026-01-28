package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/zunokit/zuno-marketplace-api/services/collection-service/internal/models"
	"gorm.io/gorm"
)

var (
	ErrEventAlreadyProcessed = errors.New("event already processed")
)

// ProcessedEventRepository handles processed event data operations
type ProcessedEventRepository interface {
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
