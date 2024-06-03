package service

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"

	er "github.com/Stern-Ritter/gophermart/internal/errors"
	"github.com/Stern-Ritter/gophermart/internal/logger"
	"github.com/Stern-Ritter/gophermart/internal/model"
	"github.com/Stern-Ritter/gophermart/internal/storage"
)

type AccrualService interface {
	CreateAccrual(ctx context.Context, accrual model.Accrual) error
	UpdateAccrual(ctx context.Context, accrual model.Accrual) error
	UpdateAccruals(ctx context.Context, accruals []model.Accrual) error
	GetAllAccrualsByUserID(ctx context.Context, userID int64) ([]model.Accrual, error)
	GetAllNewAccrualsInProcessingWithLimit(ctx context.Context, limit int64) ([]model.Accrual, error)
}

type AccrualServiceImpl struct {
	accrualStorage storage.AccrualStorage
	logger         *logger.ServerLogger
}

func NewAccrualService(accrualStorage storage.AccrualStorage, logger *logger.ServerLogger) AccrualService {
	return &AccrualServiceImpl{
		accrualStorage: accrualStorage,
		logger:         logger,
	}
}

func (s *AccrualServiceImpl) CreateAccrual(ctx context.Context, accrual model.Accrual) error {
	err := s.accrualStorage.Save(ctx, accrual)

	var pgErr *pgconn.PgError
	if err != nil && errors.As(err, &pgErr) {
		switch pgErr.ConstraintName {
		case "pk_loyalty_points_accrual":
			return er.NewAlreadyExistsError("User already uploaded this order number", err)
		case "loyalty_points_accrual_order_number_unique":
			return er.NewConflictError("Other user already uploaded this order number", err)
		}
	}

	return err
}

func (s *AccrualServiceImpl) UpdateAccrual(ctx context.Context, accrual model.Accrual) error {
	return s.accrualStorage.Update(ctx, accrual)
}

func (s *AccrualServiceImpl) UpdateAccruals(ctx context.Context, accruals []model.Accrual) error {
	return s.accrualStorage.UpdateInBatch(ctx, accruals)
}

func (s *AccrualServiceImpl) GetAllAccrualsByUserID(ctx context.Context, userID int64) ([]model.Accrual, error) {
	return s.accrualStorage.GetAllByUserIDOrderByUploadedAtAsc(ctx, userID)
}

func (s *AccrualServiceImpl) GetAllNewAccrualsInProcessingWithLimit(ctx context.Context, limit int64) ([]model.Accrual, error) {
	return s.accrualStorage.GetAllUnprocessedWithLimit(ctx, limit)
}
