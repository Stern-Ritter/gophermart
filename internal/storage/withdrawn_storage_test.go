package storage

import (
	"context"
	"testing"
	"time"

	"github.com/pashagolub/pgxmock/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Stern-Ritter/gophermart/internal/logger"
	"github.com/Stern-Ritter/gophermart/internal/model"
)

func TestWithdrawnStorageSaveWhenLoyaltyPointEnough(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err, "Error init connection mock")
	defer mock.Close()

	l, err := logger.Initialize("error")
	require.NoError(t, err, "Error init logger")

	withdrawnStorage := NewWithdrawnStorage(mock, l)

	userID := int64(1)
	accrualPoints := 100.0
	withdrawnPoints := 50.0

	withdrawn := model.Withdrawn{
		UserID:       1,
		OrderNumber:  int64(12345678903),
		ProcessedAt:  time.Now(),
		PointsAmount: 50.0,
	}

	mock.ExpectBegin()

	mock.ExpectQuery(`
		SELECT COALESCE\(SUM\(amount\),0\) as accrual_points
		FROM loyalty_points_accrual
		WHERE 
		    user_id = \@userId AND
		    status = 'PROCESSED'
	`).
		WithArgs(userID).
		WillReturnRows(pgxmock.
			NewRows([]string{"accrual_points"}).
			AddRow(accrualPoints))

	mock.ExpectQuery(`
		SELECT COALESCE\(SUM\(amount\),0\) as withdrawn_points
		FROM loyalty_points_withdrawn
		WHERE 
		    user_id = \@userId
	`).
		WithArgs(userID).
		WillReturnRows(pgxmock.
			NewRows([]string{"withdrawn_points"}).
			AddRow(withdrawnPoints))

	mock.ExpectExec(`
		INSERT INTO loyalty_points_withdrawn
		\(user_id, order_number, processed_at, amount\)
		VALUES \(@userId, @orderNumber, @processedAt, @amount\)
	`).
		WithArgs(
			withdrawn.UserID,
			withdrawn.OrderNumber,
			withdrawn.ProcessedAt,
			withdrawn.PointsAmount).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))
	mock.ExpectCommit()

	err = withdrawnStorage.Save(context.Background(), withdrawn)

	assert.NoError(t, err, "Error saving withdrawn")

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err, "The expected sql commands were not executed")
}

func TestWithdrawnStorageSaveWhenLoyaltyPointNotEnough(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err, "Error init connection mock")
	defer mock.Close()

	l, err := logger.Initialize("error")
	require.NoError(t, err, "Error init logger")

	withdrawnStorage := NewWithdrawnStorage(mock, l)

	userID := int64(1)
	accrualPoints := 100.0
	withdrawnPoints := 50.0

	withdrawn := model.Withdrawn{
		UserID:       1,
		OrderNumber:  int64(12345678903),
		ProcessedAt:  time.Now(),
		PointsAmount: 50.1,
	}

	mock.ExpectBegin()

	mock.ExpectQuery(`
		SELECT COALESCE\(SUM\(amount\),0\) as accrual_points
		FROM loyalty_points_accrual
		WHERE 
		    user_id = \@userId AND
		    status = 'PROCESSED'
	`).
		WithArgs(userID).
		WillReturnRows(pgxmock.
			NewRows([]string{"accrual_points"}).
			AddRow(accrualPoints))

	mock.ExpectQuery(`
		SELECT COALESCE\(SUM\(amount\),0\) as withdrawn_points
		FROM loyalty_points_withdrawn
		WHERE 
		    user_id = \@userId
	`).
		WithArgs(userID).
		WillReturnRows(pgxmock.
			NewRows([]string{"withdrawn_points"}).
			AddRow(withdrawnPoints))

	mock.ExpectRollback()

	err = withdrawnStorage.Save(context.Background(), withdrawn)

	assert.Error(t, err, "Expected error does not returned")

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err, "The expected sql commands were not executed")
}

func TestWithdrawnStorage_GetAllByUserIDOrderByProcessedAtAsc(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err, "Error init connection mock")
	defer mock.Close()

	l, err := logger.Initialize("error")
	require.NoError(t, err, "Error init logger")

	withdrawnStorage := NewWithdrawnStorage(mock, l)

	userID := int64(1)
	withdrawals := []model.Withdrawn{
		{
			UserID:       userID,
			OrderNumber:  int64(12345678903),
			ProcessedAt:  time.Now(),
			PointsAmount: 100,
		},
		{
			UserID:       userID,
			OrderNumber:  int64(12345678904),
			ProcessedAt:  time.Now(),
			PointsAmount: 200,
		},
		{
			UserID:       userID,
			OrderNumber:  int64(12345678905),
			ProcessedAt:  time.Now(),
			PointsAmount: 300,
		},
	}

	rows := pgxmock.NewRows([]string{
		"user_id",
		"order_number",
		"processed_at",
		"amount",
	})

	for _, withdrawn := range withdrawals {
		rows.AddRow(
			withdrawn.UserID,
			withdrawn.OrderNumber,
			withdrawn.ProcessedAt,
			withdrawn.PointsAmount,
		)
	}

	mock.ExpectQuery(`
		SELECT
		    user_id,
		    order_number,
		    processed_at,
		    amount
		FROM loyalty_points_withdrawn
		WHERE
		    user_id = @userId
		ORDER BY processed_at
	`).
		WithArgs(userID).
		WillReturnRows(rows)

	savedWithdrawals, err := withdrawnStorage.GetAllByUserIDOrderByProcessedAtAsc(context.Background(), userID)

	assert.NoError(t, err, "Error getting all new processing accruals with limit")
	assert.Equal(t, withdrawals, savedWithdrawals, "Returned accruals does not match expected")

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err, "The expected sql commands were not executed")
}
