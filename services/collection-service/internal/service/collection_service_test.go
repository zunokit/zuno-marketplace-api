package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/zunokit/zuno-marketplace-api/services/collection-service/internal/models"
	"github.com/zunokit/zuno-marketplace-api/services/collection-service/internal/repository"
)

// Mock repositories
type mockCollectionRepo struct {
	createFunc               func(ctx context.Context, collection *models.Collection) error
	getByIDFunc              func(ctx context.Context, id uuid.UUID) (*models.Collection, error)
	getByContractAddressFunc func(ctx context.Context, address, chainID string) (*models.Collection, error)
	updateFunc               func(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error
	listByUserFunc           func(ctx context.Context, userID uuid.UUID, page, limit int) ([]*models.Collection, int64, error)
	listFunc                 func(ctx context.Context, filters *repository.ListFilters, page, limit int) ([]*models.Collection, int64, error)
	deleteFunc               func(ctx context.Context, id uuid.UUID) error
	incrementTotalMintedFunc func(ctx context.Context, id uuid.UUID, increment int64) error
}

func (m *mockCollectionRepo) Create(ctx context.Context, collection *models.Collection) error {
	if m.createFunc != nil {
		return m.createFunc(ctx, collection)
	}
	return nil
}

func (m *mockCollectionRepo) GetByID(ctx context.Context, id uuid.UUID) (*models.Collection, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, id)
	}
	return nil, errors.New("not found")
}

func (m *mockCollectionRepo) GetByContractAddress(ctx context.Context, address, chainID string) (*models.Collection, error) {
	if m.getByContractAddressFunc != nil {
		return m.getByContractAddressFunc(ctx, address, chainID)
	}
	return nil, errors.New("not found")
}

func (m *mockCollectionRepo) Update(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error {
	if m.updateFunc != nil {
		return m.updateFunc(ctx, id, updates)
	}
	return nil
}

func (m *mockCollectionRepo) ListByUser(ctx context.Context, userID uuid.UUID, page, limit int) ([]*models.Collection, int64, error) {
	if m.listByUserFunc != nil {
		return m.listByUserFunc(ctx, userID, page, limit)
	}
	return nil, 0, nil
}

func (m *mockCollectionRepo) List(ctx context.Context, filters *repository.ListFilters, page, limit int) ([]*models.Collection, int64, error) {
	if m.listFunc != nil {
		return m.listFunc(ctx, filters, page, limit)
	}
	return nil, 0, nil
}

func (m *mockCollectionRepo) Delete(ctx context.Context, id uuid.UUID) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, id)
	}
	return nil
}

func (m *mockCollectionRepo) IncrementTotalMinted(ctx context.Context, id uuid.UUID, increment int64) error {
	if m.incrementTotalMintedFunc != nil {
		return m.incrementTotalMintedFunc(ctx, id, increment)
	}
	return nil
}

type mockAllowlistRepo struct {
	addWalletsFunc      func(ctx context.Context, entries []*models.CollectionAllowlist) error
	getByCollectionFunc func(ctx context.Context, collectionID uuid.UUID) ([]*models.CollectionAllowlist, error)
	removeWalletFunc    func(ctx context.Context, collectionID uuid.UUID, walletAddress string) error
}

func (m *mockAllowlistRepo) AddWallets(ctx context.Context, entries []*models.CollectionAllowlist) error {
	if m.addWalletsFunc != nil {
		return m.addWalletsFunc(ctx, entries)
	}
	return nil
}

func (m *mockAllowlistRepo) GetByCollection(ctx context.Context, collectionID uuid.UUID) ([]*models.CollectionAllowlist, error) {
	if m.getByCollectionFunc != nil {
		return m.getByCollectionFunc(ctx, collectionID)
	}
	return nil, nil
}

func (m *mockAllowlistRepo) RemoveWallet(ctx context.Context, collectionID uuid.UUID, walletAddress string) error {
	if m.removeWalletFunc != nil {
		return m.removeWalletFunc(ctx, collectionID, walletAddress)
	}
	return nil
}

type mockMetadataRepo struct {
	updateFunc            func(ctx context.Context, collectionID uuid.UUID, updates map[string]interface{}) error
	getByCollectionIDFunc func(ctx context.Context, collectionID uuid.UUID) (*models.CollectionMetadata, error)
}

func (m *mockMetadataRepo) Update(ctx context.Context, collectionID uuid.UUID, updates map[string]interface{}) error {
	if m.updateFunc != nil {
		return m.updateFunc(ctx, collectionID, updates)
	}
	return nil
}

func (m *mockMetadataRepo) GetByCollectionID(ctx context.Context, collectionID uuid.UUID) (*models.CollectionMetadata, error) {
	if m.getByCollectionIDFunc != nil {
		return m.getByCollectionIDFunc(ctx, collectionID)
	}
	return nil, nil
}

func TestCreateCollection(t *testing.T) {
	ctx := context.Background()
	collectionID := uuid.New()
	userID := uuid.New()

	tests := []struct {
		name       string
		collection *models.Collection
		setupMock  func(*mockCollectionRepo)
		wantErr    bool
	}{
		{
			name: "successful creation",
			collection: &models.Collection{
				ID:              collectionID,
				UserID:          userID,
				Name:            "Test Collection",
				Symbol:          "TEST",
				DeployerAddress: "0xdeployer",
				ImageURL:        "https://example.com/image.png",
			},
			setupMock: func(m *mockCollectionRepo) {
				m.createFunc = func(ctx context.Context, collection *models.Collection) error {
					// Verify defaults are set
					if collection.Status != models.CollectionStatusPending {
						t.Errorf("Status should be set to PENDING, got %v", collection.Status)
					}
					if collection.IndexStatus != models.IndexStatusNotStarted {
						t.Errorf("IndexStatus should be set to NOT_STARTED, got %v", collection.IndexStatus)
					}
					return nil
				}
				m.getByIDFunc = func(ctx context.Context, id uuid.UUID) (*models.Collection, error) {
					return &models.Collection{
						ID:              collectionID,
						UserID:          userID,
						Name:            "Test Collection",
						Symbol:          "TEST",
						DeployerAddress: "0xdeployer",
						Status:          models.CollectionStatusPending,
						IndexStatus:     models.IndexStatusNotStarted,
						CreatedAt:       time.Now(),
						UpdatedAt:       time.Now(),
					}, nil
				}
			},
			wantErr: false,
		},
		{
			name: "create error",
			collection: &models.Collection{
				UserID: userID,
				Name:   "Test Collection",
			},
			setupMock: func(m *mockCollectionRepo) {
				m.createFunc = func(ctx context.Context, collection *models.Collection) error {
					return errors.New("database error")
				}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockCollectionRepo{}
			tt.setupMock(mockRepo)

			service := NewCollectionService(mockRepo, &mockAllowlistRepo{}, &mockMetadataRepo{})
			got, err := service.CreateCollection(ctx, tt.collection)

			if (err != nil) != tt.wantErr {
				t.Errorf("CreateCollection() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && got == nil {
				t.Error("CreateCollection() should return collection")
			}
		})
	}
}

func TestGetCollection(t *testing.T) {
	ctx := context.Background()
	collectionID := uuid.New()

	tests := []struct {
		name      string
		id        uuid.UUID
		setupMock func(*mockCollectionRepo)
		wantErr   bool
	}{
		{
			name: "successful retrieval",
			id:   collectionID,
			setupMock: func(m *mockCollectionRepo) {
				m.getByIDFunc = func(ctx context.Context, id uuid.UUID) (*models.Collection, error) {
					return &models.Collection{
						ID:   collectionID,
						Name: "Test Collection",
					}, nil
				}
			},
			wantErr: false,
		},
		{
			name: "not found",
			id:   uuid.New(),
			setupMock: func(m *mockCollectionRepo) {
				m.getByIDFunc = func(ctx context.Context, id uuid.UUID) (*models.Collection, error) {
					return nil, errors.New("not found")
				}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockCollectionRepo{}
			tt.setupMock(mockRepo)

			service := NewCollectionService(mockRepo, &mockAllowlistRepo{}, &mockMetadataRepo{})
			got, err := service.GetCollection(ctx, tt.id)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetCollection() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && got == nil {
				t.Error("GetCollection() should return collection")
			}
		})
	}
}

func TestUpdateCollection(t *testing.T) {
	ctx := context.Background()
	collectionID := uuid.New()
	userID := uuid.New()
	otherUserID := uuid.New()
	contractAddress := "0x1234567890123456789012345678901234567890"

	tests := []struct {
		name      string
		id        uuid.UUID
		userID    uuid.UUID
		updates   map[string]interface{}
		setupMock func(*mockCollectionRepo)
		wantErr   bool
		errType   error
	}{
		{
			name:   "successful update",
			id:     collectionID,
			userID: userID,
			updates: map[string]interface{}{
				"description": "Updated description",
			},
			setupMock: func(m *mockCollectionRepo) {
				m.getByIDFunc = func(ctx context.Context, id uuid.UUID) (*models.Collection, error) {
					return &models.Collection{
						ID:     collectionID,
						UserID: userID,
						Status: models.CollectionStatusPending,
					}, nil
				}
				m.updateFunc = func(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error {
					return nil
				}
			},
			wantErr: false,
		},
		{
			name:   "unauthorized update",
			id:     collectionID,
			userID: otherUserID,
			updates: map[string]interface{}{
				"description": "Updated description",
			},
			setupMock: func(m *mockCollectionRepo) {
				m.getByIDFunc = func(ctx context.Context, id uuid.UUID) (*models.Collection, error) {
					return &models.Collection{
						ID:     collectionID,
						UserID: userID,
						Status: models.CollectionStatusPending,
					}, nil
				}
			},
			wantErr: true,
			errType: ErrUnauthorized,
		},
		{
			name:   "invalid status change - deployed to pending",
			id:     collectionID,
			userID: userID,
			updates: map[string]interface{}{
				"status": models.CollectionStatusPending,
			},
			setupMock: func(m *mockCollectionRepo) {
				m.getByIDFunc = func(ctx context.Context, id uuid.UUID) (*models.Collection, error) {
					return &models.Collection{
						ID:     collectionID,
						UserID: userID,
						Status: models.CollectionStatusDeployed,
					}, nil
				}
			},
			wantErr: true,
			errType: ErrInvalidStatusChange,
		},
		{
			name:   "deploy without contract address",
			id:     collectionID,
			userID: userID,
			updates: map[string]interface{}{
				"status": models.CollectionStatusDeployed,
			},
			setupMock: func(m *mockCollectionRepo) {
				m.getByIDFunc = func(ctx context.Context, id uuid.UUID) (*models.Collection, error) {
					return &models.Collection{
						ID:              collectionID,
						UserID:          userID,
						Status:          models.CollectionStatusPending,
						ContractAddress: nil,
					}, nil
				}
			},
			wantErr: true,
			errType: ErrContractRequired,
		},
		{
			name:   "deploy with contract address in updates",
			id:     collectionID,
			userID: userID,
			updates: map[string]interface{}{
				"status":           models.CollectionStatusDeployed,
				"contract_address": contractAddress,
			},
			setupMock: func(m *mockCollectionRepo) {
				m.getByIDFunc = func(ctx context.Context, id uuid.UUID) (*models.Collection, error) {
					return &models.Collection{
						ID:              collectionID,
						UserID:          userID,
						Status:          models.CollectionStatusPending,
						ContractAddress: nil,
					}, nil
				}
				m.updateFunc = func(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error {
					return nil
				}
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockCollectionRepo{}
			tt.setupMock(mockRepo)

			service := NewCollectionService(mockRepo, &mockAllowlistRepo{}, &mockMetadataRepo{})
			_, err := service.UpdateCollection(ctx, tt.id, tt.userID, tt.updates)

			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateCollection() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errType != nil && !errors.Is(err, tt.errType) {
				t.Errorf("UpdateCollection() error = %v, want error type %v", err, tt.errType)
			}
		})
	}
}

func TestUpdateMetadata(t *testing.T) {
	ctx := context.Background()
	collectionID := uuid.New()
	userID := uuid.New()
	otherUserID := uuid.New()

	tests := []struct {
		name                string
		collectionID        uuid.UUID
		userID              uuid.UUID
		metadataUpdates     map[string]interface{}
		setupCollectionMock func(*mockCollectionRepo)
		setupMetadataMock   func(*mockMetadataRepo)
		wantErr             bool
		errType             error
	}{
		{
			name:         "successful metadata update",
			collectionID: collectionID,
			userID:       userID,
			metadataUpdates: map[string]interface{}{
				"discord_url": "https://discord.gg/test",
			},
			setupCollectionMock: func(m *mockCollectionRepo) {
				m.getByIDFunc = func(ctx context.Context, id uuid.UUID) (*models.Collection, error) {
					return &models.Collection{
						ID:     collectionID,
						UserID: userID,
					}, nil
				}
			},
			setupMetadataMock: func(m *mockMetadataRepo) {
				m.updateFunc = func(ctx context.Context, collectionID uuid.UUID, updates map[string]interface{}) error {
					return nil
				}
			},
			wantErr: false,
		},
		{
			name:         "unauthorized metadata update",
			collectionID: collectionID,
			userID:       otherUserID,
			metadataUpdates: map[string]interface{}{
				"discord_url": "https://discord.gg/test",
			},
			setupCollectionMock: func(m *mockCollectionRepo) {
				m.getByIDFunc = func(ctx context.Context, id uuid.UUID) (*models.Collection, error) {
					return &models.Collection{
						ID:     collectionID,
						UserID: userID,
					}, nil
				}
			},
			setupMetadataMock: func(m *mockMetadataRepo) {},
			wantErr:           true,
			errType:           ErrUnauthorized,
		},
		{
			name:            "no updates",
			collectionID:    collectionID,
			userID:          userID,
			metadataUpdates: map[string]interface{}{},
			setupCollectionMock: func(m *mockCollectionRepo) {
				m.getByIDFunc = func(ctx context.Context, id uuid.UUID) (*models.Collection, error) {
					return &models.Collection{
						ID:     collectionID,
						UserID: userID,
					}, nil
				}
			},
			setupMetadataMock: func(m *mockMetadataRepo) {},
			wantErr:           false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCollectionRepo := &mockCollectionRepo{}
			mockMetadataRepo := &mockMetadataRepo{}
			tt.setupCollectionMock(mockCollectionRepo)
			tt.setupMetadataMock(mockMetadataRepo)

			service := NewCollectionService(mockCollectionRepo, &mockAllowlistRepo{}, mockMetadataRepo)
			err := service.UpdateMetadata(ctx, tt.collectionID, tt.userID, tt.metadataUpdates)

			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateMetadata() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errType != nil && !errors.Is(err, tt.errType) {
				t.Errorf("UpdateMetadata() error = %v, want error type %v", err, tt.errType)
			}
		})
	}
}

func TestAddToAllowlist(t *testing.T) {
	ctx := context.Background()
	collectionID := uuid.New()
	userID := uuid.New()
	otherUserID := uuid.New()

	tests := []struct {
		name                string
		collectionID        uuid.UUID
		userID              uuid.UUID
		wallets             []string
		maxMintAmount       int
		setupCollectionMock func(*mockCollectionRepo)
		setupAllowlistMock  func(*mockAllowlistRepo)
		wantCount           int
		wantErr             bool
		errType             error
	}{
		{
			name:          "successful allowlist addition",
			collectionID:  collectionID,
			userID:        userID,
			wallets:       []string{"0xwallet1", "0xwallet2", "0xwallet3"},
			maxMintAmount: 5,
			setupCollectionMock: func(m *mockCollectionRepo) {
				m.getByIDFunc = func(ctx context.Context, id uuid.UUID) (*models.Collection, error) {
					return &models.Collection{
						ID:     collectionID,
						UserID: userID,
					}, nil
				}
			},
			setupAllowlistMock: func(m *mockAllowlistRepo) {
				m.addWalletsFunc = func(ctx context.Context, entries []*models.CollectionAllowlist) error {
					if len(entries) != 3 {
						t.Errorf("Expected 3 entries, got %d", len(entries))
					}
					for _, entry := range entries {
						if entry.MaxMintAmount != 5 {
							t.Errorf("Expected MaxMintAmount 5, got %d", entry.MaxMintAmount)
						}
					}
					return nil
				}
			},
			wantCount: 3,
			wantErr:   false,
		},
		{
			name:          "unauthorized allowlist addition",
			collectionID:  collectionID,
			userID:        otherUserID,
			wallets:       []string{"0xwallet1"},
			maxMintAmount: 5,
			setupCollectionMock: func(m *mockCollectionRepo) {
				m.getByIDFunc = func(ctx context.Context, id uuid.UUID) (*models.Collection, error) {
					return &models.Collection{
						ID:     collectionID,
						UserID: userID,
					}, nil
				}
			},
			setupAllowlistMock: func(m *mockAllowlistRepo) {},
			wantCount:          0,
			wantErr:            true,
			errType:            ErrUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCollectionRepo := &mockCollectionRepo{}
			mockAllowlistRepo := &mockAllowlistRepo{}
			tt.setupCollectionMock(mockCollectionRepo)
			tt.setupAllowlistMock(mockAllowlistRepo)

			service := NewCollectionService(mockCollectionRepo, mockAllowlistRepo, &mockMetadataRepo{})
			count, err := service.AddToAllowlist(ctx, tt.collectionID, tt.userID, tt.wallets, tt.maxMintAmount)

			if (err != nil) != tt.wantErr {
				t.Errorf("AddToAllowlist() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && count != tt.wantCount {
				t.Errorf("AddToAllowlist() count = %v, want %v", count, tt.wantCount)
			}

			if tt.wantErr && tt.errType != nil && !errors.Is(err, tt.errType) {
				t.Errorf("AddToAllowlist() error = %v, want error type %v", err, tt.errType)
			}
		})
	}
}

func TestDeleteCollection(t *testing.T) {
	ctx := context.Background()
	collectionID := uuid.New()
	userID := uuid.New()
	otherUserID := uuid.New()

	tests := []struct {
		name      string
		id        uuid.UUID
		userID    uuid.UUID
		setupMock func(*mockCollectionRepo)
		wantErr   bool
		errType   error
	}{
		{
			name:   "successful deletion",
			id:     collectionID,
			userID: userID,
			setupMock: func(m *mockCollectionRepo) {
				m.getByIDFunc = func(ctx context.Context, id uuid.UUID) (*models.Collection, error) {
					return &models.Collection{
						ID:     collectionID,
						UserID: userID,
					}, nil
				}
				m.deleteFunc = func(ctx context.Context, id uuid.UUID) error {
					return nil
				}
			},
			wantErr: false,
		},
		{
			name:   "unauthorized deletion",
			id:     collectionID,
			userID: otherUserID,
			setupMock: func(m *mockCollectionRepo) {
				m.getByIDFunc = func(ctx context.Context, id uuid.UUID) (*models.Collection, error) {
					return &models.Collection{
						ID:     collectionID,
						UserID: userID,
					}, nil
				}
			},
			wantErr: true,
			errType: ErrUnauthorized,
		},
		{
			name:   "collection not found",
			id:     uuid.New(),
			userID: userID,
			setupMock: func(m *mockCollectionRepo) {
				m.getByIDFunc = func(ctx context.Context, id uuid.UUID) (*models.Collection, error) {
					return nil, errors.New("not found")
				}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockCollectionRepo{}
			tt.setupMock(mockRepo)

			service := NewCollectionService(mockRepo, &mockAllowlistRepo{}, &mockMetadataRepo{})
			err := service.DeleteCollection(ctx, tt.id, tt.userID)

			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteCollection() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errType != nil && !errors.Is(err, tt.errType) {
				t.Errorf("DeleteCollection() error = %v, want error type %v", err, tt.errType)
			}
		})
	}
}
