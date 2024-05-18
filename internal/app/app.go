package app

import (
	"context"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/Stern-Ritter/gophermart/internal/auth"
	"github.com/Stern-Ritter/gophermart/internal/compress"
	"github.com/Stern-Ritter/gophermart/internal/config"
	"github.com/Stern-Ritter/gophermart/internal/logger"
	"github.com/Stern-Ritter/gophermart/internal/scheduler"
	"github.com/Stern-Ritter/gophermart/internal/server"
	"github.com/Stern-Ritter/gophermart/internal/service"
	"github.com/Stern-Ritter/gophermart/internal/storage"
	"github.com/Stern-Ritter/gophermart/internal/validator"
	"github.com/Stern-Ritter/gophermart/migrations"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func Run(config *config.ServerConfig, logger *logger.ServerLogger) error {
	ctx := context.Background()
	db, err := pgxpool.New(ctx, config.DatabaseURL)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.String("event", "connect database"),
			zap.String("database url", config.DatabaseURL), zap.Error(err))
		return err
	}
	defer db.Close()

	err = migrations.Migrate(config.DatabaseURL, "postgres", "pgx")
	if err != nil {
		logger.Fatal("Failed to migrate database", zap.String("event", "migrate database"),
			zap.String("database url", config.DatabaseURL), zap.Error(err))
	}

	validate, err := validator.GetValidator()
	if err != nil {
		logger.Fatal("Failed to init validator", zap.String("event", "init validator"), zap.Error(err))
	}
	authToken := auth.GenerateAuthToken(config.JwtSecretKey)

	userStorage := storage.NewUserStorage(db, logger)
	accrualStorage := storage.NewAccrualStorage(db, logger)
	withdrawnStorage := storage.NewWithdrawnStorage(db, logger)
	balanceStorage := storage.NewBalanceStorage(db, logger)

	userService := service.NewUserService(userStorage, logger)
	authService := service.NewAuthService(userService, authToken, logger)
	accrualService := service.NewAccrualService(accrualStorage, logger)
	withdrawnService := service.NewWithdrawnService(withdrawnStorage, logger)
	balanceService := service.NewBalanceService(balanceStorage, logger)

	accrualsScheduler := scheduler.NewAccrualsScheduler(accrualService, config.AccrualSystemURL,
		config.ProcessAccrualsBatchMaxSize, config.ProcessAccrualsBufferSize, config.ProcessAccrualsWorkerPoolSize,
		config.GetNewAccrualsInterval, logger)
	accrualsScheduler.RunTasks()
	defer accrualsScheduler.StopTasks()

	server := server.NewServer(
		authService,
		userService,
		accrualService,
		withdrawnService,
		balanceService,
		validate,
		authToken,
		config,
		logger,
	)

	r := addRoutes(server)

	err = http.ListenAndServe(server.Config.URL, r)
	if err != nil {
		logger.Fatal("Failed to start server", zap.String("event", "start server"),
			zap.String("url", server.Config.URL), zap.Error(err), zap.Error(err))
	}

	return err
}

func addRoutes(s *server.Server) *chi.Mux {
	r := chi.NewRouter()
	r.Use(s.Logger.LoggerMiddleware)
	r.Use(compress.GzipMiddleware)

	r.Route("/api", func(r chi.Router) {
		r.Route("/user", func(r chi.Router) {
			r.Group(func(r chi.Router) {
				r.Post("/register", s.SignUpHandler)
				r.Post("/login", s.SignInHandler)
			})

			r.Group(func(r chi.Router) {
				r.Use(auth.Verifier(s.AuthToken))
				r.Use(auth.Authenticator(s.AuthToken))

				r.Route("/orders", func(r chi.Router) {
					r.Post("/", s.LoadOrderHandler)
					r.Get("/", s.FindAllOrdersLoadedByUserHandler)
				})

				r.Route("/balance", func(r chi.Router) {
					r.Get("/", s.GetLoyaltyPointsBalanceHandler)
					r.Post("/withdraw", s.WithdrawLoyaltyPointsHandler)
				})

				r.Route("/withdrawals", func(r chi.Router) {
					r.Get("/", s.FindAllWithdrawalsByUserHandler)
				})
			})
		})
	})

	return r
}
