package service

import (
	"context"

	"github.com/go-chi/jwtauth/v5"

	er "github.com/Stern-Ritter/gophermart/internal/errors"
	"github.com/Stern-Ritter/gophermart/internal/logger"
	"github.com/Stern-Ritter/gophermart/internal/model"
	"github.com/Stern-Ritter/gophermart/internal/storage"
)

type UserService interface {
	CreateUser(ctx context.Context, user model.User) error
	GetUserByLogin(ctx context.Context, login string) (model.User, error)
	GetCurrentUser(ctx context.Context) (model.User, error)
}

type UserServiceImpl struct {
	userStorage storage.UserStorage
	logger      *logger.ServerLogger
}

func NewUserService(userStorage storage.UserStorage, logger *logger.ServerLogger) UserService {
	return &UserServiceImpl{
		userStorage: userStorage,
		logger:      logger,
	}
}

func (s *UserServiceImpl) CreateUser(ctx context.Context, user model.User) error {
	return s.userStorage.Save(ctx, user)
}

func (s *UserServiceImpl) GetUserByLogin(ctx context.Context, login string) (model.User, error) {
	return s.userStorage.GetOneByLogin(ctx, login)
}

func (s *UserServiceImpl) GetCurrentUser(ctx context.Context) (model.User, error) {
	_, claims, err := jwtauth.FromContext(ctx)
	if err != nil {
		return model.User{}, er.NewUnauthorizedError("User is not authorized to access this resource", err)
	}

	if _, ok := claims["login"]; !ok {
		return model.User{}, er.NewUnauthorizedError("User is not authorized to access this resource", err)
	}

	login := claims["login"].(string)
	currentUser, err := s.userStorage.GetOneByLogin(ctx, login)
	return currentUser, err
}
