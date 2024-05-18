package server

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/jwtauth/v5"
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

func TestGetLoyaltyPointsBalanceHandler(t *testing.T) {
	tests := []struct {
		name                        string
		isAuthorized                bool
		userUserStorage             bool
		useBalanceStorage           bool
		userStorageErr              error
		balanceStorageReturnedValue model.Balance
		balanceStorageErr           error
		expectedBody                string
		expectedStatusCode          int
	}{
		{
			name:               "should return status 401 when user is unauthorized",
			useBalanceStorage:  false,
			expectedStatusCode: http.StatusUnauthorized,
		},
		{
			name:                        "should return status 200 when user user exists",
			isAuthorized:                true,
			userUserStorage:             true,
			useBalanceStorage:           true,
			balanceStorageReturnedValue: model.Balance{CurrentPointsAmount: 400, WithdrawnPointsAmount: 300},
			expectedBody:                `{"current":400,"withdrawn":300}`,
			expectedStatusCode:          http.StatusOK,
		},
		{
			name:               "should return status 500 when unexpected error occurred",
			isAuthorized:       true,
			userUserStorage:    true,
			useBalanceStorage:  true,
			balanceStorageErr:  errors.New("unexpected error"),
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

			handler := http.HandlerFunc(server.GetLoyaltyPointsBalanceHandler)

			if tt.userUserStorage {
				userStorage.EXPECT().GetOneByLogin(gomock.Any(), "user").
					Return(model.User{ID: 1, Login: "user", Password: "password"}, tt.userStorageErr)
			}
			if tt.useBalanceStorage {
				balanceStorage.EXPECT().GetByUserID(gomock.Any(), int64(1)).
					Return(tt.balanceStorageReturnedValue, tt.balanceStorageErr)
			}

			claims := map[string]interface{}{"login": "user"}
			jwtauth.SetExpiry(claims, time.Now().Add(time.Hour*24))
			_, tokenString, err := server.AuthToken.Encode(claims)
			require.NoError(t, err, "Error encoding token")
			decodedClaims, err := server.AuthToken.Decode(tokenString)
			require.NoError(t, err, "Error decoding token")

			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/api/user/balance", nil)
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
