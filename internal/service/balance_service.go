package service

import (
	"context"

	"github.com/Stern-Ritter/gophermart/internal/logger"
	"github.com/Stern-Ritter/gophermart/internal/model"
	"github.com/Stern-Ritter/gophermart/internal/storage"
)

type BalanceService interface {
	GetBalanceByUserID(ctx context.Context, userID int64) (model.Balance, error)
}

type BalanceServiceImpl struct {
	balanceStorage storage.BalanceStorage
	logger         *logger.ServerLogger
}

func NewBalanceService(balanceStorage storage.BalanceStorage, logger *logger.ServerLogger) BalanceService {
	return &BalanceServiceImpl{
		balanceStorage: balanceStorage,
		logger:         logger,
	}
}

func (s *BalanceServiceImpl) GetBalanceByUserID(ctx context.Context, userID int64) (model.Balance, error) {
	return s.balanceStorage.GetByUserID(ctx, userID)
}
