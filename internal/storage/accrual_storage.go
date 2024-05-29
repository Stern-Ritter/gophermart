package storage

import (
	"context"
	"database/sql"

	"github.com/jackc/pgx/v5"

	"github.com/Stern-Ritter/gophermart/internal/logger"
	"github.com/Stern-Ritter/gophermart/internal/model"
)

type AccrualStorage interface {
	Save(ctx context.Context, accrual model.Accrual) error
	Update(ctx context.Context, accrual model.Accrual) error
	UpdateInBatch(ctx context.Context, accruals []model.Accrual) error
	GetAllByUserIDOrderByUploadedAtAsc(ctx context.Context, userID int64) ([]model.Accrual, error)
	GetAllUnprocessedWithLimit(ctx context.Context, limit int64) ([]model.Accrual, error)
}

type AccrualStorageImpl struct {
	db     PgxIface
	logger *logger.ServerLogger
}

func NewAccrualStorage(db PgxIface, logger *logger.ServerLogger) AccrualStorage {
	return &AccrualStorageImpl{
		db:     db,
		logger: logger,
	}
}

func (s *AccrualStorageImpl) Save(ctx context.Context, accrual model.Accrual) error {
	_, err := s.db.Exec(ctx, `
		INSERT INTO loyalty_points_accrual 
		    (user_id, order_number, uploaded_at, status, amount)
		VALUES (@userId, @orderNumber, @uploadedAt, @status, @amount)
	`, pgx.NamedArgs{
		"userId":      accrual.UserID,
		"orderNumber": accrual.OrderNumber,
		"uploadedAt":  accrual.UploadedAt,
		"status":      accrual.Status,
		"amount":      accrual.PointsAmount,
	})

	return err
}

func (s *AccrualStorageImpl) Update(ctx context.Context, accrual model.Accrual) error {
	_, err := s.db.Exec(ctx, `
		UPDATE loyalty_points_accrual
		SET processed_at = @processedAt, status = @status, amount = @amount
		WHERE user_id = @userId AND order_number = @orderNumber
	`, pgx.NamedArgs{
		"userId":      accrual.UserID,
		"orderNumber": accrual.OrderNumber,
		"processedAt": accrual.ProcessedAt,
		"status":      accrual.Status,
		"amount":      accrual.PointsAmount,
	})

	return err
}

func (s *AccrualStorageImpl) UpdateInBatch(ctx context.Context, accruals []model.Accrual) error {
	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	for _, accrual := range accruals {
		_, err := tx.Exec(ctx, `
		UPDATE loyalty_points_accrual
		SET processed_at = @processedAt, status = @status, amount = @amount, processing_lock = FALSE
		WHERE user_id = @userId AND order_number = @orderNumber
		`, pgx.NamedArgs{
			"userId":      accrual.UserID,
			"orderNumber": accrual.OrderNumber,
			"processedAt": accrual.ProcessedAt,
			"status":      accrual.Status,
			"amount":      accrual.PointsAmount,
		})
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (s *AccrualStorageImpl) GetAllByUserIDOrderByUploadedAtAsc(ctx context.Context, userID int64) ([]model.Accrual, error) {
	rows, err := s.db.Query(ctx, `
		SELECT
		    user_id,
			order_number,
			uploaded_at,
			processed_at,
			status,
			amount
		FROM loyalty_points_accrual
		WHERE 
		    user_id = @userId
		ORDER BY uploaded_at 
	`, pgx.NamedArgs{
		"userId": userID,
	})

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	accruals := make([]model.Accrual, 0)

	for rows.Next() {
		accrual := model.Accrual{}
		var processedAt sql.NullTime
		if err := rows.Scan(&accrual.UserID, &accrual.OrderNumber, &accrual.UploadedAt, &processedAt, &accrual.Status,
			&accrual.PointsAmount); err != nil {
			return nil, err
		}
		if processedAt.Valid {
			accrual.ProcessedAt = processedAt.Time
		}

		accruals = append(accruals, accrual)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return accruals, nil
}

func (s *AccrualStorageImpl) GetAllUnprocessedWithLimit(ctx context.Context, limit int64) ([]model.Accrual, error) {
	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	rows, err := tx.Query(ctx, `
		SELECT
			user_id,
			order_number,
			uploaded_at,
			processed_at,
			status,
			amount
		FROM loyalty_points_accrual
		WHERE status IN ('NEW', 'PROCESSING')
		AND processing_lock = FALSE
		ORDER BY uploaded_at
		LIMIT @limit
	`, pgx.NamedArgs{
		"limit": limit,
	})

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	accruals := make([]model.Accrual, 0)

	for rows.Next() {
		accrual := model.Accrual{}
		var processedAt sql.NullTime
		if err := rows.Scan(&accrual.UserID, &accrual.OrderNumber, &accrual.UploadedAt, &processedAt, &accrual.Status,
			&accrual.PointsAmount); err != nil {
			return nil, err
		}
		if processedAt.Valid {
			accrual.ProcessedAt = processedAt.Time
		}

		accruals = append(accruals, accrual)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	err = updateStatusInBatch(ctx, tx, model.AccrualProcessing, true, accruals)
	if err != nil {
		return nil, err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil, err
	}

	for i := range accruals {
		accruals[i].Status = model.AccrualProcessing
	}

	return accruals, nil
}

func updateStatusInBatch(ctx context.Context, tx pgx.Tx, status model.AccrualStatus, processingLock bool,
	accruals []model.Accrual) error {
	for _, accrual := range accruals {
		_, err := tx.Exec(ctx, `
		UPDATE loyalty_points_accrual
		SET status = @status, processing_lock = @processingLock
		WHERE user_id = @userId AND order_number = @orderNumber
		`, pgx.NamedArgs{
			"userId":         accrual.UserID,
			"orderNumber":    accrual.OrderNumber,
			"status":         status,
			"processingLock": processingLock,
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func getAccrualPointsSumByUserID(ctx context.Context, tx pgx.Tx, userID int64) (float64, error) {
	var accrualPoints float64

	row := tx.QueryRow(ctx, `
		SELECT COALESCE(SUM(amount),0) as accrual_points
		FROM loyalty_points_accrual
		WHERE 
		    user_id = @userId AND
		    status = 'PROCESSED'  
	`, pgx.NamedArgs{
		"userId": userID,
	})

	err := row.Scan(&accrualPoints)
	if err != nil {
		return 0, err
	}

	return accrualPoints, nil
}
