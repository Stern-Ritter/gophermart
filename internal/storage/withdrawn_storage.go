package storage

import (
	"context"

	"github.com/jackc/pgx/v5"

	er "github.com/Stern-Ritter/gophermart/internal/errors"
	"github.com/Stern-Ritter/gophermart/internal/logger"
	"github.com/Stern-Ritter/gophermart/internal/model"
	"github.com/Stern-Ritter/gophermart/internal/utils"
)

type WithdrawnStorage interface {
	Save(ctx context.Context, withdrawn model.Withdrawn) error
	GetAllByUserIDOrderByProcessedAtAsc(ctx context.Context, userID int64) ([]model.Withdrawn, error)
}

type WithdrawnStorageImpl struct {
	db     PgxIface
	logger *logger.ServerLogger
}

func NewWithdrawnStorage(db PgxIface, logger *logger.ServerLogger) WithdrawnStorage {
	return &WithdrawnStorageImpl{
		db:     db,
		logger: logger,
	}
}

func (s *WithdrawnStorageImpl) Save(ctx context.Context, withdrawn model.Withdrawn) error {
	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	accrualPoints, err := getAccrualPointsSumByUserID(ctx, tx, withdrawn.UserID)
	if err != nil {
		return err
	}
	withdrawnPoints, err := getWithdrawnPointsSumByUserID(ctx, tx, withdrawn.UserID)
	if err != nil {
		return err
	}

	currentPoints := accrualPoints - withdrawnPoints
	if utils.Float64Compare(currentPoints, withdrawn.PointsAmount) < 0 {
		return er.NewPaymentRequiredError("Not enough loyalty points to withdrawn", nil)
	}

	_, err = s.db.Exec(ctx, `
		INSERT INTO loyalty_points_withdrawn
		(user_id, order_number, processed_at, amount) 
		VALUES (@userId, @orderNumber, @processedAt, @amount)		
	`, pgx.NamedArgs{
		"userId":      withdrawn.UserID,
		"orderNumber": withdrawn.OrderNumber,
		"processedAt": withdrawn.ProcessedAt,
		"amount":      withdrawn.PointsAmount,
	})

	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (s *WithdrawnStorageImpl) GetAllByUserIDOrderByProcessedAtAsc(ctx context.Context, userID int64) ([]model.Withdrawn, error) {
	rows, err := s.db.Query(ctx, `
		SELECT
		    user_id, 
		    order_number, 
		    processed_at, 
		    amount
		FROM loyalty_points_withdrawn
		WHERE 
		    user_id = @userId
		ORDER BY processed_at
	`, pgx.NamedArgs{
		"userId": userID,
	})

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	withdrawals := make([]model.Withdrawn, 0)

	for rows.Next() {
		withdrawn := model.Withdrawn{}
		if err := rows.Scan(&withdrawn.UserID, &withdrawn.OrderNumber, &withdrawn.ProcessedAt,
			&withdrawn.PointsAmount); err != nil {
			return nil, err
		}
		withdrawals = append(withdrawals, withdrawn)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return withdrawals, nil
}

func getWithdrawnPointsSumByUserID(ctx context.Context, tx pgx.Tx, userID int64) (float64, error) {
	var withdrawnPoints float64

	row := tx.QueryRow(ctx, `
		SELECT COALESCE(SUM(amount),0) as withdrawn_points
		FROM loyalty_points_withdrawn
		WHERE 
		    user_id = @userId
	`, pgx.NamedArgs{
		"userId": userID,
	})

	err := row.Scan(&withdrawnPoints)
	if err != nil {
		return 0, err
	}

	return withdrawnPoints, nil
}
