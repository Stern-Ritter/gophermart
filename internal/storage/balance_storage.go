package storage

import (
	"context"

	"github.com/jackc/pgx/v5"

	"github.com/Stern-Ritter/gophermart/internal/logger"
	"github.com/Stern-Ritter/gophermart/internal/model"
)

type BalanceStorage interface {
	GetByUserID(ctx context.Context, userID int64) (model.Balance, error)
}

type BalanceStorageImpl struct {
	db     PgxIface
	logger *logger.ServerLogger
}

func NewBalanceStorage(db PgxIface, logger *logger.ServerLogger) BalanceStorage {
	return &BalanceStorageImpl{
		db:     db,
		logger: logger,
	}
}

func (s *BalanceStorageImpl) GetByUserID(ctx context.Context, userID int64) (model.Balance, error) {
	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return model.Balance{}, err
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	accrualPoints, err := getAccrualPointsSumByUserID(ctx, tx, userID)
	if err != nil {
		return model.Balance{}, err
	}
	withdrawnPoints, err := getWithdrawnPointsSumByUserID(ctx, tx, userID)
	if err != nil {
		return model.Balance{}, err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return model.Balance{}, err
	}

	balance := model.Balance{
		UserID:                userID,
		CurrentPointsAmount:   accrualPoints - withdrawnPoints,
		WithdrawnPointsAmount: withdrawnPoints,
	}

	return balance, nil
}
