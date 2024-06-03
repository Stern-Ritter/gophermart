package storage

import (
	"context"
	"testing"

	"github.com/pashagolub/pgxmock/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Stern-Ritter/gophermart/internal/logger"
	"github.com/Stern-Ritter/gophermart/internal/model"
)

func TestUserStorageSave(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err, "Error init connection mock")
	defer mock.Close()

	l, err := logger.Initialize("error")
	require.NoError(t, err, "Error init logger")

	userStorage := NewUserStorage(mock, l)

	user := model.User{
		Login:    "testUser",
		Password: "secretPassword",
	}

	mock.ExpectExec("INSERT INTO users").
		WithArgs(user.Login, user.Password).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	err = userStorage.Save(context.Background(), user)
	assert.NoError(t, err, "Error saving user")

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err, "The expected sql commands were not executed")
}

func TestUserStorageGetOneByLogin(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err, "Error init connection mock")
	defer mock.Close()

	l, err := logger.Initialize("error")
	require.NoError(t, err, "Error init logger")

	userStorage := NewUserStorage(mock, l)

	expectedUser := model.User{
		ID:       1,
		Login:    "testUser",
		Password: "secretPassword",
	}

	rows := mock.NewRows([]string{"id", "login", "password"}).
		AddRow(expectedUser.ID, expectedUser.Login, expectedUser.Password)

	mock.ExpectQuery("SELECT id, login, password FROM users WHERE login =").
		WithArgs("testUser").
		WillReturnRows(rows)

	user, err := userStorage.GetOneByLogin(context.Background(), "testUser")

	assert.NoError(t, err, "Error getting user by login")
	assert.Equal(t, expectedUser, user, "Returned user does not match expected")

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err, "The expected sql commands were not executed")
}
