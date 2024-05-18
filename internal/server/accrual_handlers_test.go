package server

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/jwtauth/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/Stern-Ritter/gophermart/internal/auth"
	"github.com/Stern-Ritter/gophermart/internal/config"
	"github.com/Stern-Ritter/gophermart/internal/logger"
	"github.com/Stern-Ritter/gophermart/internal/model"
	"github.com/Stern-Ritter/gophermart/internal/service"
	"github.com/Stern-Ritter/gophermart/internal/validator"
)

func TestLoadOrderHandler(t *testing.T) {
	tests := []struct {
		name               string
		body               string
		isAuthorized       bool
		useUserStorage     bool
		useAccrualStorage  bool
		userStorageErr     error
		accrualStorageErr  error
		expectedStatusCode int
	}{
		{
			name:               "should return status 401 when user is unauthorized",
			body:               "12345678903",
			isAuthorized:       false,
			useUserStorage:     false,
			useAccrualStorage:  false,
			expectedStatusCode: http.StatusUnauthorized,
		},
		{
			name:               "should return status 202 when order is loaded",
			body:               "12345678903",
			isAuthorized:       true,
			useUserStorage:     true,
			useAccrualStorage:  true,
			expectedStatusCode: http.StatusAccepted,
		},
		{
			name:               "should return status 400 when request body is empty",
			body:               "",
			isAuthorized:       true,
			useUserStorage:     true,
			useAccrualStorage:  false,
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "should return status 422 when order number in request body is invalid",
			body:               "49927398717",
			isAuthorized:       true,
			useUserStorage:     true,
			useAccrualStorage:  false,
			expectedStatusCode: http.StatusUnprocessableEntity,
		},
		{
			name:               "should return status 200 when order number in already loaded by this user",
			body:               "12345678903",
			isAuthorized:       true,
			useUserStorage:     true,
			useAccrualStorage:  true,
			accrualStorageErr:  &pgconn.PgError{ConstraintName: "pk_loyalty_points_accrual"},
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "should return status 409 when order number in already loaded by other user",
			body:               "12345678903",
			isAuthorized:       true,
			useUserStorage:     true,
			useAccrualStorage:  true,
			accrualStorageErr:  &pgconn.PgError{ConstraintName: "loyalty_points_accrual_order_number_unique"},
			expectedStatusCode: http.StatusConflict,
		},
		{
			name:               "should return status 500 when unexpected error occurred",
			body:               "12345678903",
			isAuthorized:       true,
			useUserStorage:     true,
			useAccrualStorage:  true,
			accrualStorageErr:  errors.New("unexpected error"),
			expectedStatusCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			validate, err := validator.GetValidator()
			require.NoError(t, err, "Error init validator")
			authToken := auth.GenerateAuthToken("secret")
			cfg := &config.ServerConfig{}
			logger, err := logger.Initialize("error")
			require.NoError(t, err, "Error init logger")

			userStorage := NewMockUserStorage(ctrl)
			accrualStorage := NewMockAccrualStorage(ctrl)
			withdrawnStorage := NewMockWithdrawnStorage(ctrl)
			balanceStorage := NewMockBalanceStorage(ctrl)

			userService := service.NewUserService(userStorage, logger)
			authService := service.NewAuthService(userService, authToken, logger)
			accrualService := service.NewAccrualService(accrualStorage, logger)
			withdrawnService := service.NewWithdrawnService(withdrawnStorage, logger)
			balanceService := service.NewBalanceService(balanceStorage, logger)

			server := NewServer(authService, userService, accrualService, withdrawnService, balanceService, validate,
				authToken, cfg, logger)

			handler := http.HandlerFunc(server.LoadOrderHandler)

			if tt.useUserStorage {
				userStorage.EXPECT().GetOneByLogin(gomock.Any(), "user").
					Return(model.User{ID: 1, Login: "user", Password: "password"}, tt.userStorageErr)
			}
			if tt.useAccrualStorage {
				accrualStorage.EXPECT().Save(gomock.Any(), gomock.Any()).
					Return(tt.accrualStorageErr)
			}

			claims := map[string]interface{}{"login": "user"}
			jwtauth.SetExpiry(claims, time.Now().Add(time.Hour*24))
			_, tokenString, err := server.AuthToken.Encode(claims)
			require.NoError(t, err, "Error encoding token")
			decodedClaims, _ := server.AuthToken.Decode(tokenString)

			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/api/user/orders", strings.NewReader(tt.body))
			req.Header.Set("Authorization", tokenString)
			req.Header.Set("Content-Type", "text/plain")
			if tt.isAuthorized {
				ctx := jwtauth.NewContext(req.Context(), decodedClaims, nil)
				req = req.WithContext(ctx)
			}

			handler.ServeHTTP(w, req)

			resp := w.Result()
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedStatusCode, resp.StatusCode, "Response status code does not match expected status")
		})
	}
}

func TestFindAllOrdersLoadedByUserHandler(t *testing.T) {
	tests := []struct {
		name                        string
		isAuthorized                bool
		useUserStorage              bool
		useAccrualStorage           bool
		userStorageErr              error
		accrualStorageReturnedValue []model.Accrual
		accrualStorageErr           error
		expectedBody                string
		expectedStatusCode          int
	}{
		{
			name:               "should return status 401 when user is unauthorized",
			isAuthorized:       false,
			useUserStorage:     false,
			useAccrualStorage:  false,
			expectedStatusCode: http.StatusUnauthorized,
		},
		{
			name:              "should return status 200 when user has uploaded at least one order",
			isAuthorized:      true,
			useUserStorage:    true,
			useAccrualStorage: true,
			accrualStorageReturnedValue: []model.Accrual{
				{
					OrderNumber:  1,
					Status:       model.AccrualNew,
					PointsAmount: 42,
					UploadedAt:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				},
			},
			expectedBody:       `[{"number":"1","status":"NEW","accrual":42,"uploaded_at":"2024-01-01T00:00:00Z"}]`,
			expectedStatusCode: http.StatusOK,
		},
		{
			name:                        "should return status 204 when user did not upload orders",
			isAuthorized:                true,
			useUserStorage:              true,
			useAccrualStorage:           true,
			accrualStorageReturnedValue: make([]model.Accrual, 0),
			expectedStatusCode:          http.StatusNoContent,
		},
		{
			name:               "should return status 500 when unexpected error occurred",
			isAuthorized:       true,
			useUserStorage:     true,
			useAccrualStorage:  true,
			accrualStorageErr:  errors.New("unexpected error"),
			expectedStatusCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			validate, err := validator.GetValidator()
			require.NoError(t, err, "Error init validator")
			authToken := auth.GenerateAuthToken("secret")
			cfg := &config.ServerConfig{}
			logger, err := logger.Initialize("error")
			require.NoError(t, err, "Error init logger")

			userStorage := NewMockUserStorage(ctrl)
			accrualStorage := NewMockAccrualStorage(ctrl)
			withdrawnStorage := NewMockWithdrawnStorage(ctrl)
			balanceStorage := NewMockBalanceStorage(ctrl)

			userService := service.NewUserService(userStorage, logger)
			authService := service.NewAuthService(userService, authToken, logger)
			accrualService := service.NewAccrualService(accrualStorage, logger)
			withdrawnService := service.NewWithdrawnService(withdrawnStorage, logger)
			balanceService := service.NewBalanceService(balanceStorage, logger)

			server := NewServer(authService, userService, accrualService, withdrawnService, balanceService, validate,
				authToken, cfg, logger)

			handler := http.HandlerFunc(server.FindAllOrdersLoadedByUserHandler)

			if tt.useUserStorage {
				userStorage.EXPECT().GetOneByLogin(gomock.Any(), "user").
					Return(model.User{ID: 1, Login: "user", Password: "password"}, tt.userStorageErr)
			}
			if tt.useAccrualStorage {
				accrualStorage.EXPECT().GetAllByUserIDOrderByUploadedAtAsc(gomock.Any(), gomock.Any()).
					Return(tt.accrualStorageReturnedValue, tt.accrualStorageErr)
			}

			claims := map[string]interface{}{"login": "user"}
			jwtauth.SetExpiry(claims, time.Now().Add(time.Hour*24))
			_, tokenString, err := server.AuthToken.Encode(claims)
			require.NoError(t, err, "Error encoding token")
			decodedClaims, _ := server.AuthToken.Decode(tokenString)

			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/api/user/orders", nil)
			req.Header.Set("Authorization", tokenString)
			if tt.isAuthorized {
				ctx := jwtauth.NewContext(req.Context(), decodedClaims, nil)
				req = req.WithContext(ctx)
			}

			handler.ServeHTTP(w, req)

			resp := w.Result()
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedStatusCode, resp.StatusCode, "Response status code does not match expected status")
			if tt.expectedBody != "" {
				body, err := io.ReadAll(resp.Body)
				require.NoError(t, err, "Error reading response body")
				require.Equal(t, tt.expectedBody, string(body), "Response body does not match expected body")
			}
		})
	}
}
