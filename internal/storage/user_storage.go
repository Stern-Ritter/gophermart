package storage

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Stern-Ritter/gophermart/internal/logger"
	"github.com/Stern-Ritter/gophermart/internal/model"
)

type UserStorage interface {
	Save(ctx context.Context, user model.User) error
	GetOneByLogin(ctx context.Context, login string) (model.User, error)
}

type UserStorageImpl struct {
	db     *pgxpool.Pool
	logger *logger.ServerLogger
}

func NewUserStorage(db *pgxpool.Pool, logger *logger.ServerLogger) UserStorage {
	return &UserStorageImpl{
		db:     db,
		logger: logger,
	}
}

func (s *UserStorageImpl) Save(ctx context.Context, user model.User) error {
	_, err := s.db.Exec(ctx, `
		INSERT INTO users 
		    (login, password)
		VALUES 
		    (@login, @password)
	`, pgx.NamedArgs{
		"login":    user.Login,
		"password": user.Password,
	})

	return err
}

func (s *UserStorageImpl) GetOneByLogin(ctx context.Context, login string) (model.User, error) {
	row := s.db.QueryRow(ctx, `
		SELECT 
			id,
			login,
			password
		FROM users
		WHERE 
		    login = @login
`, pgx.NamedArgs{
		"login": login,
	})

	user := model.User{}
	err := row.Scan(&user.ID, &user.Login, &user.Password)

	return user, err
}
