package server

import (
	"github.com/go-chi/jwtauth/v5"
	"github.com/go-playground/validator/v10"

	"github.com/Stern-Ritter/gophermart/internal/config"
	"github.com/Stern-Ritter/gophermart/internal/logger"
	"github.com/Stern-Ritter/gophermart/internal/service"
)

type Server struct {
	AuthService      service.AuthService
	UserService      service.UserService
	AccrualService   service.AccrualService
	WithdrawnService service.WithdrawnService
	BalanceService   service.BalanceService
	Validate         *validator.Validate
	AuthToken        *jwtauth.JWTAuth
	Logger           *logger.ServerLogger
	Config           *config.ServerConfig
}

func NewServer(authService service.AuthService, userService service.UserService, accrualService service.AccrualService,
	withdrawnService service.WithdrawnService, balanceService service.BalanceService, validate *validator.Validate,
	authToken *jwtauth.JWTAuth, config *config.ServerConfig, logger *logger.ServerLogger) *Server {
	return &Server{
		AuthService:      authService,
		UserService:      userService,
		AccrualService:   accrualService,
		WithdrawnService: withdrawnService,
		BalanceService:   balanceService,
		Validate:         validate,
		AuthToken:        authToken,
		Logger:           logger,
		Config:           config,
	}
}
