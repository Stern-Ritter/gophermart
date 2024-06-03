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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/Stern-Ritter/gophermart/internal/auth"
	"github.com/Stern-Ritter/gophermart/internal/config"
	er "github.com/Stern-Ritter/gophermart/internal/errors"
	"github.com/Stern-Ritter/gophermart/internal/logger"
	"github.com/Stern-Ritter/gophermart/internal/model"
	"github.com/Stern-Ritter/gophermart/internal/service"
	"github.com/Stern-Ritter/gophermart/internal/validator"
)

func TestWithdrawLoyaltyPointsHandler(t *testing.T) {
	tests := []struct {
		name                string
		body                string
		isAuthorized        bool
		useUserStorage      bool
		useWithdrawnStorage bool
		userStorageErr      error
		withdrawnStorageErr error
		expectedStatusCode  int
	}{
		{
			name:                "should return status 401 when user is unauthorized",
			isAuthorized:        false,
			useUserStorage:      false,
			useWithdrawnStorage: false,
			expectedStatusCode:  http.StatusUnauthorized,
		},
		{
			name:                "should return status 400 when request body is empty",
			body:                "",
			isAuthorized:        true,
			useUserStorage:      true,
			useWithdrawnStorage: false,
			expectedStatusCode:  http.StatusBadRequest,
		},
		{
			name:                "should return status 400 when required order number field is missing",
			body:                `{"order1":"12345678903","sum":100}`,
			isAuthorized:        true,
			useUserStorage:      true,
			useWithdrawnStorage: false,
			expectedStatusCode:  http.StatusBadRequest,
		},
		{
			name:                "should return status 400 when order number is not numeric",
			body:                `{"order":"s12345678903","sum":100}`,
			isAuthorized:        true,
			useUserStorage:      true,
			useWithdrawnStorage: false,
			expectedStatusCode:  http.StatusBadRequest,
		},
		{
			name:                "should return status 400 when order number is invalid",
			body:                `{"order":"49927398717","sum":100}`,
			isAuthorized:        true,
			useUserStorage:      true,
			useWithdrawnStorage: false,
			expectedStatusCode:  http.StatusBadRequest,
		},
		{
			name:                "should return status 400 when required sum field is missing",
			body:                `{"order":"12345678903","sum1":100}`,
			isAuthorized:        true,
			useUserStorage:      true,
			useWithdrawnStorage: false,
			expectedStatusCode:  http.StatusBadRequest,
		},
		{
			name:                "should return status 400 when sum is less or equal to 0",
			body:                `{"order":"12345678903","sum":0}`,
			isAuthorized:        true,
			useUserStorage:      true,
			useWithdrawnStorage: false,
			expectedStatusCode:  http.StatusBadRequest,
		},
		{
			name:                "should return status 402 when user does not have enough points to be withdrawn",
			body:                `{"order":"12345678903","sum":10}`,
			isAuthorized:        true,
			useUserStorage:      true,
			useWithdrawnStorage: true,
			withdrawnStorageErr: er.PaymentRequiredError{},
			expectedStatusCode:  http.StatusPaymentRequired,
		},

		{
			name:                "should return status 500 when unexpected error occurred",
			body:                `{"order":"12345678903","sum":10}`,
			isAuthorized:        true,
			useUserStorage:      true,
			useWithdrawnStorage: true,
			withdrawnStorageErr: errors.New("unexpected error"),
			expectedStatusCode:  http.StatusInternalServerError,
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

			handler := http.HandlerFunc(server.WithdrawLoyaltyPointsHandler)

			if tt.useUserStorage {
				userStorage.EXPECT().GetOneByLogin(gomock.Any(), "user").
					Return(model.User{ID: 1, Login: "user", Password: "password"}, tt.userStorageErr)
			}
			if tt.useWithdrawnStorage {
				withdrawnStorage.EXPECT().Save(gomock.Any(), gomock.Any()).
					Return(tt.withdrawnStorageErr)
			}

			claims := map[string]interface{}{"login": "user"}
			jwtauth.SetExpiry(claims, time.Now().Add(time.Hour*24))
			_, tokenString, err := server.AuthToken.Encode(claims)
			require.NoError(t, err, "Error encoding token")
			decodedClaims, _ := server.AuthToken.Decode(tokenString)

			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/api/user/balance/withdraw", strings.NewReader(tt.body))
			req.Header.Set("Authorization", tokenString)
			req.Header.Set("Content-Type", "application/json")
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

func TestFindAllWithdrawalsByUserHandler(t *testing.T) {
	tests := []struct {
		name                          string
		isAuthorized                  bool
		useUserStorage                bool
		useWithdrawnStorage           bool
		userStorageErr                error
		withdrawnStorageReturnedValue []model.Withdrawn
		withdrawnStorageErr           error
		expectedBody                  string
		expectedStatusCode            int
	}{
		{
			name:                "should return status 401 when user is unauthorized",
			isAuthorized:        false,
			useUserStorage:      false,
			useWithdrawnStorage: false,
			expectedStatusCode:  http.StatusUnauthorized,
		},
		{
			name:                "should return status 200 when user had at least one withdrawn of points",
			isAuthorized:        true,
			useUserStorage:      true,
			useWithdrawnStorage: true,
			withdrawnStorageReturnedValue: []model.Withdrawn{
				{
					UserID:       1,
					OrderNumber:  12345678903,
					PointsAmount: 42,
					ProcessedAt:  time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				},
			},
			expectedBody:       `[{"order":"12345678903","sum":42,"processed_at":"2024-01-01T00:00:00Z"}]`,
			expectedStatusCode: http.StatusOK,
		},
		{
			name:                          "should return status 204 when user did not make any withdrawn of points",
			isAuthorized:                  true,
			useUserStorage:                true,
			useWithdrawnStorage:           true,
			withdrawnStorageReturnedValue: make([]model.Withdrawn, 0),
			expectedStatusCode:            http.StatusNoContent,
		},
		{
			name:                "should return status 500 when unexpected error occurred",
			isAuthorized:        true,
			useUserStorage:      true,
			useWithdrawnStorage: true,
			withdrawnStorageErr: errors.New("unexpected error"),
			expectedStatusCode:  http.StatusInternalServerError,
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

			handler := http.HandlerFunc(server.FindAllWithdrawalsByUserHandler)

			if tt.useUserStorage {
				userStorage.EXPECT().GetOneByLogin(gomock.Any(), "user").
					Return(model.User{ID: 1, Login: "user", Password: "password"}, tt.userStorageErr)
			}
			if tt.useWithdrawnStorage {
				withdrawnStorage.EXPECT().GetAllByUserIDOrderByProcessedAtAsc(gomock.Any(), int64(1)).
					Return(tt.withdrawnStorageReturnedValue, tt.withdrawnStorageErr)
			}

			claims := map[string]interface{}{"login": "user"}
			jwtauth.SetExpiry(claims, time.Now().Add(time.Hour*24))
			_, tokenString, err := server.AuthToken.Encode(claims)
			require.NoError(t, err, "Error encoding token")
			decodedClaims, _ := server.AuthToken.Decode(tokenString)

			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/api/user/withdrawals", nil)
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
				require.NoError(t, err)
				require.Equal(t, tt.expectedBody, string(body), "Response body does not match expected body")
			}
		})
	}
}
