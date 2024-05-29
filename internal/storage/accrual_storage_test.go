package storage

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/pashagolub/pgxmock/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Stern-Ritter/gophermart/internal/logger"
	"github.com/Stern-Ritter/gophermart/internal/model"
)

func TestAccrualStorageSave(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err, "Error init connection mock")
	defer mock.Close()

	l, err := logger.Initialize("error")
	require.NoError(t, err, "Error init logger")

	accrualStorage := NewAccrualStorage(mock, l)

	accrual := model.Accrual{
		UserID:       1,
		OrderNumber:  int64(12345678903),
		UploadedAt:   time.Now(),
		Status:       model.AccrualNew,
		PointsAmount: 0,
	}

	mock.ExpectExec(`
		INSERT INTO loyalty_points_accrual 
		    \(user_id, order_number, uploaded_at, status, amount\)
		VALUES \(@userId, @orderNumber, @uploadedAt, @status, @amount\)
	`).
		WithArgs(
			accrual.UserID,
			accrual.OrderNumber,
			accrual.UploadedAt,
			accrual.Status,
			accrual.PointsAmount).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	err = accrualStorage.Save(context.Background(), accrual)
	assert.NoError(t, err, "Error saving accrual")

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err, "The expected sql commands were not executed")
}

func TestAccrualStorageUpdate(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err, "Error init connection mock")
	defer mock.Close()

	l, err := logger.Initialize("error")
	require.NoError(t, err, "Error init logger")

	accrualStorage := NewAccrualStorage(mock, l)

	accrual := model.Accrual{
		UserID:       1,
		OrderNumber:  int64(12345678903),
		ProcessedAt:  time.Now(),
		Status:       model.AccrualProcessed,
		PointsAmount: 100,
	}

	mock.ExpectExec(`
		UPDATE loyalty_points_accrual
		SET processed_at = @processedAt, status = @status, amount = @amount
		WHERE user_id = @userId AND order_number = @orderNumber
	`).
		WithArgs(
			accrual.ProcessedAt,
			accrual.Status,
			accrual.PointsAmount,
			accrual.UserID,
			accrual.OrderNumber).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))

	err = accrualStorage.Update(context.Background(), accrual)
	assert.NoError(t, err, "Error updating accrual")

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err, "The expected sql commands were not executed")
}

func TestAccrualStorageUpdateInBatch(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err, "Error init connection mock")
	defer mock.Close()

	l, err := logger.Initialize("error")
	require.NoError(t, err, "Error init logger")

	accrualStorage := NewAccrualStorage(mock, l)

	accruals := []model.Accrual{
		{
			UserID:       1,
			OrderNumber:  int64(12345678903),
			ProcessedAt:  time.Now(),
			Status:       model.AccrualProcessed,
			PointsAmount: 300,
		},
		{
			UserID:       2,
			OrderNumber:  int64(12345678904),
			ProcessedAt:  time.Now(),
			Status:       model.AccrualInvalid,
			PointsAmount: 0,
		},
		{
			UserID:       3,
			OrderNumber:  int64(12345678905),
			ProcessedAt:  time.Now(),
			Status:       model.AccrualProcessing,
			PointsAmount: 0,
		},
	}

	mock.ExpectBegin()

	for _, accrual := range accruals {
		mock.ExpectExec(`
			UPDATE loyalty_points_accrual
			SET processed_at = @processedAt, status = @status, amount = @amount, processing_lock = FALSE
			WHERE user_id = @userId AND order_number = @orderNumber
		`).
			WithArgs(
				accrual.ProcessedAt,
				accrual.Status,
				accrual.PointsAmount,
				accrual.UserID,
				accrual.OrderNumber).
			WillReturnResult(pgxmock.NewResult("UPDATE", 1))
	}

	mock.ExpectCommit()

	err = accrualStorage.UpdateInBatch(context.Background(), accruals)
	assert.NoError(t, err, "Error updating accrual in batch")

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err, "The expected sql commands were not executed")
}

func TestAccrualStorageGetAllByUserIDOrderByUploadedAtAsc(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err, "Error init connection mock")
	defer mock.Close()

	l, err := logger.Initialize("error")
	require.NoError(t, err, "Error init logger")

	accrualStorage := NewAccrualStorage(mock, l)

	userID := int64(1)
	accruals := []model.Accrual{
		{
			UserID:       1,
			OrderNumber:  int64(12345678903),
			ProcessedAt:  time.Now(),
			Status:       model.AccrualProcessed,
			PointsAmount: 300,
		},
		{
			UserID:       2,
			OrderNumber:  int64(12345678904),
			ProcessedAt:  time.Now(),
			Status:       model.AccrualInvalid,
			PointsAmount: 0,
		},
	}

	rows := pgxmock.NewRows([]string{
		"user_id",
		"order_number",
		"uploaded_at",
		"processed_at",
		"status",
		"amount",
	})

	for _, accrual := range accruals {
		rows.AddRow(
			accrual.UserID,
			accrual.OrderNumber,
			accrual.UploadedAt,
			accrual.ProcessedAt,
			accrual.Status,
			accrual.PointsAmount,
		)
	}

	mock.ExpectQuery(`
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
	`).
		WithArgs(userID).
		WillReturnRows(rows)

	savedAccruals, err := accrualStorage.GetAllByUserIDOrderByUploadedAtAsc(context.Background(), userID)

	assert.NoError(t, err, "Error getting all accruals by user id")
	assert.Equal(t, accruals, savedAccruals, "Returned accruals does not match expected")

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err, "The expected sql commands were not executed")
}

func TestAccrualStorageGetAllUnprocessedWithLimit(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err, "Error init connection mock")
	defer mock.Close()

	l, err := logger.Initialize("error")
	require.NoError(t, err, "Error init logger")

	accrualStorage := NewAccrualStorage(mock, l)

	limit := int64(3)
	accruals := []model.Accrual{
		{
			UserID:       1,
			OrderNumber:  int64(12345678903),
			ProcessedAt:  time.Now(),
			Status:       model.AccrualProcessing,
			PointsAmount: 0,
		},
		{
			UserID:       2,
			OrderNumber:  int64(12345678904),
			ProcessedAt:  time.Now(),
			Status:       model.AccrualProcessing,
			PointsAmount: 0,
		},
		{
			UserID:       3,
			OrderNumber:  int64(12345678905),
			ProcessedAt:  time.Now(),
			Status:       model.AccrualProcessing,
			PointsAmount: 0,
		},
	}

	rows := pgxmock.NewRows([]string{
		"user_id",
		"order_number",
		"uploaded_at",
		"processed_at",
		"status",
		"amount",
	})

	for _, accrual := range accruals {
		rows.AddRow(
			accrual.UserID,
			accrual.OrderNumber,
			accrual.UploadedAt,
			accrual.ProcessedAt,
			accrual.Status,
			accrual.PointsAmount,
		)
	}

	mock.ExpectBegin()

	mock.ExpectQuery(regexp.QuoteMeta(`
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
	`)).WithArgs(limit).WillReturnRows(rows)

	for _, accrual := range accruals {
		mock.ExpectExec(`
		UPDATE loyalty_points_accrual
		SET status = @status, processing_lock = @processingLock
		WHERE user_id = @userId AND order_number = @orderNumber
		`).
			WithArgs(model.AccrualProcessing, true, accrual.UserID, accrual.OrderNumber).
			WillReturnResult(pgxmock.NewResult("UPDATE", 1))
	}

	mock.ExpectCommit()

	savedAccruals, err := accrualStorage.GetAllUnprocessedWithLimit(context.Background(), limit)

	assert.NoError(t, err, "Error getting all unprocessed accruals with limit")
	assert.Equal(t, accruals, savedAccruals, "Returned accruals does not match expected")

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err, "The expected sql commands were not executed")
}
