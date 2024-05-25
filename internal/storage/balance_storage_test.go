package storage

import (
	"context"
	"errors"
	"testing"

	"github.com/pashagolub/pgxmock/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Stern-Ritter/gophermart/internal/logger"
	"github.com/Stern-Ritter/gophermart/internal/model"
)

func TestBalanceStorageGetByUserID(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err, "Error init connection mock")
	defer mock.Close()

	l, err := logger.Initialize("error")
	require.NoError(t, err, "Error init logger")

	balanceStorage := NewBalanceStorage(mock, l)

	userID := int64(1)
	accrualPoints := 100.0
	withdrawnPoints := 50.0

	expectedBalance := model.Balance{
		UserID:                userID,
		CurrentPointsAmount:   accrualPoints - withdrawnPoints,
		WithdrawnPointsAmount: withdrawnPoints,
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

	mock.ExpectCommit()

	balance, err := balanceStorage.GetByUserID(context.Background(), userID)

	assert.NoError(t, err, "Error getting balance")
	assert.Equal(t, expectedBalance, balance, "Returned balance does not match expected")

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err, "The expected sql commands were not executed")
}

func TestBalanceStorageGetByUserIDWhenGetAccrualPointsSumByUserIDError(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err, "Error init connection mock")
	defer mock.Close()

	l, err := logger.Initialize("error")
	require.NoError(t, err, "Error init logger")

	balanceStorage := NewBalanceStorage(mock, l)

	userID := int64(1)

	expectedBalance := model.Balance{}

	mock.ExpectBegin()
	mock.ExpectQuery(`
		SELECT COALESCE\(SUM\(amount\),0\) as accrual_points
		FROM loyalty_points_accrual
		WHERE 
		    user_id = \@userId AND
		    status = 'PROCESSED'
	`).
		WithArgs(userID).
		WillReturnError(errors.New("error getting accrual points"))

	mock.ExpectRollback()

	balance, err := balanceStorage.GetByUserID(context.Background(), userID)

	assert.Error(t, err, "Expected error does not returned")
	assert.Equal(t, expectedBalance, balance, "Returned balance does not match expected")

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err, "The expected sql commands were not executed")
}

func TestBalanceStorageGetByUserIDWhenGetWithdrawnPointsSumByUserIDError(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err, "Error init connection mock")
	defer mock.Close()

	l, err := logger.Initialize("error")
	require.NoError(t, err, "Error init logger")

	balanceStorage := NewBalanceStorage(mock, l)

	userID := int64(1)
	accrualPoints := 100.0

	expectedBalance := model.Balance{}

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
		WillReturnError(errors.New("error getting accrual points"))

	mock.ExpectRollback()

	balance, err := balanceStorage.GetByUserID(context.Background(), userID)

	assert.Error(t, err, "Expected error does not returned")
	assert.Equal(t, expectedBalance, balance, "Returned balance does not match expected")

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err, "The expected sql commands were not executed")
}
