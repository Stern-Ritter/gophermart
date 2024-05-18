package service

import (
	"context"

	"github.com/Stern-Ritter/gophermart/internal/logger"
	"github.com/Stern-Ritter/gophermart/internal/model"
	"github.com/Stern-Ritter/gophermart/internal/storage"
)

type WithdrawnService interface {
	CreateWithdrawn(ctx context.Context, withdrawn model.Withdrawn) error
	GetAllWithdrawalsByUserID(ctx context.Context, userID int64) ([]model.Withdrawn, error)
}

type WithdrawnServiceImpl struct {
	withdrawnStorage storage.WithdrawnStorage
	logger           *logger.ServerLogger
}

func NewWithdrawnService(withdrawnStorage storage.WithdrawnStorage, logger *logger.ServerLogger) WithdrawnService {
	return &WithdrawnServiceImpl{
		withdrawnStorage: withdrawnStorage,
		logger:           logger,
	}
}

func (s *WithdrawnServiceImpl) CreateWithdrawn(ctx context.Context, withdrawn model.Withdrawn) error {
	return s.withdrawnStorage.Save(ctx, withdrawn)
}

func (s *WithdrawnServiceImpl) GetAllWithdrawalsByUserID(ctx context.Context, userID int64) ([]model.Withdrawn, error) {
	return s.withdrawnStorage.GetAllByUserIDOrderByProcessedAtAsc(ctx, userID)
}
