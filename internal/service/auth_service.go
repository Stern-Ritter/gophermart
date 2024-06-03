package service

import (
	"context"
	"errors"
	"time"

	"github.com/go-chi/jwtauth/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/Stern-Ritter/gophermart/internal/auth"
	er "github.com/Stern-Ritter/gophermart/internal/errors"
	"github.com/Stern-Ritter/gophermart/internal/logger"
	"github.com/Stern-Ritter/gophermart/internal/model"
)

type AuthService interface {
	SignUp(ctx context.Context, request model.SignUpRequest) (string, error)
	SignIn(ctx context.Context, request model.SignInRequest) (string, error)
}

type AuthServiceImpl struct {
	userService UserService
	authToken   *jwtauth.JWTAuth
	logger      *logger.ServerLogger
}

func NewAuthService(userService UserService, authToken *jwtauth.JWTAuth, logger *logger.ServerLogger) AuthService {
	return &AuthServiceImpl{
		userService: userService,
		authToken:   authToken,
		logger:      logger,
	}
}

func (s *AuthServiceImpl) SignUp(ctx context.Context, request model.SignUpRequest) (string, error) {
	user := model.SignUpRequestToUser(request)

	passwordHash, err := auth.GetPasswordHash(user.Password)
	if err != nil {
		return "", err
	}
	user.Password = passwordHash

	err = s.userService.CreateUser(ctx, user)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.ConstraintName == "users_login_unique" {
			return "", er.NewConflictError("User with this login already exists", err)
		}

		return "", err
	}

	claims := map[string]interface{}{"login": user.Login}
	jwtauth.SetExpiry(claims, time.Now().Add(time.Hour*24))
	_, tokenString, err := s.authToken.Encode(claims)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (s *AuthServiceImpl) SignIn(ctx context.Context, request model.SignInRequest) (string, error) {
	user, err := s.userService.GetUserByLogin(ctx, request.Login)
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return "", er.NewUnauthorizedError("Invalid login or password", err)
	case err != nil:
		return "", err
	}

	if !auth.CheckPasswordHash(request.Password, user.Password) {
		return "", er.NewUnauthorizedError("Invalid login or password", err)
	}

	claims := map[string]interface{}{"login": user.Login}
	jwtauth.SetExpiry(claims, time.Now().Add(time.Hour*24))
	_, tokenString, err := s.authToken.Encode(claims)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
