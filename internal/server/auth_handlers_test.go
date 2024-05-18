package server

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"

	"github.com/Stern-Ritter/gophermart/internal/auth"
	"github.com/Stern-Ritter/gophermart/internal/config"
	"github.com/Stern-Ritter/gophermart/internal/logger"
	"github.com/Stern-Ritter/gophermart/internal/model"
	"github.com/Stern-Ritter/gophermart/internal/service"
	"github.com/Stern-Ritter/gophermart/internal/validator"
)

func TestSignUpHandler(t *testing.T) {
	tests := []struct {
		name                      string
		body                      string
		useUserStorage            bool
		userStorageErr            error
		expectedStatusCode        int
		expectAuthorizationHeader bool
	}{
		{
			name:                      "should return status 200 when user with this login not exists",
			body:                      `{"login":"user42","password":"password"}`,
			useUserStorage:            true,
			expectedStatusCode:        http.StatusOK,
			expectAuthorizationHeader: true,
		},
		{
			name:               "should return status 400 when request body is empty",
			body:               "",
			useUserStorage:     false,
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "should return status 400 when required login field is missing",
			body:               `{"login1":"user42","password":"password"}`,
			useUserStorage:     false,
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "should return status 400 when login field length is less than 4 characters",
			body:               `{"login":"usr","password":"password"}`,
			useUserStorage:     false,
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "should return status 400 when login field length is more than 30 characters",
			body:               `{"login":"user42user42user42user42user42user42","password":"password"}`,
			useUserStorage:     false,
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "should return status 400 when required password field is missing",
			body:               `{"login":"user42","password1":"password"}`,
			useUserStorage:     false,
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "should return status 400 when password field length is less than 8 characters",
			body:               `{"login":"user42","password":"word"}`,
			useUserStorage:     false,
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "should return status 400 when password field length is more than 256 characters",
			body:               `{"login":"user","password":"passwordpasswordpasswordpasswordpasswordpasswordpasswordpasswordpasswordpasswordpasswordpasswordpasswordpasswordpasswordpasswordpasswordpasswordpasswordpasswordpasswordpasswordpasswordpasswordpasswordpasswordpasswordpasswordpasswordpasswordpasswordpasswordpassword"}`,
			useUserStorage:     false,
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "should return status 409 when user with this login already exists",
			body:               `{"login":"user42","password":"password"}`,
			useUserStorage:     true,
			userStorageErr:     &pgconn.PgError{ConstraintName: "users_login_unique"},
			expectedStatusCode: http.StatusConflict,
		},
		{
			name:               "should return status 500 when unexpected error occurred",
			body:               `{"login":"user42","password":"password"}`,
			useUserStorage:     true,
			userStorageErr:     errors.New("unexpected error"),
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

			handler := http.HandlerFunc(server.SignUpHandler)
			if tt.useUserStorage {
				userStorage.EXPECT().Save(gomock.Any(), gomock.Any()).Return(tt.userStorageErr)
			}

			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/api/user/register", strings.NewReader(tt.body))

			handler.ServeHTTP(w, req)

			resp := w.Result()
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedStatusCode, resp.StatusCode, "Response status code does not match expected status")
			if tt.expectAuthorizationHeader {
				assert.NotEmpty(t, resp.Header.Get("Authorization"), "Response header should contain Authorization header")
			} else {
				assert.Empty(t, resp.Header.Get("Authorization"), "Response header should not contain Authorization header")
			}
		})
	}
}

func TestSignInHandler(t *testing.T) {
	tests := []struct {
		name                      string
		body                      string
		useUserStorage            bool
		userStorageErr            error
		expectedStatusCode        int
		expectAuthorizationHeader bool
	}{
		{
			name:                      "should return status 200 when user with this login exist and password is valid",
			body:                      `{"login":"user42","password":"password"}`,
			useUserStorage:            true,
			expectedStatusCode:        http.StatusOK,
			expectAuthorizationHeader: true,
		},
		{
			name:               "should return status 401 when user with this login not exists",
			body:               `{"login":"user42","password":"password"}`,
			useUserStorage:     true,
			userStorageErr:     pgx.ErrNoRows,
			expectedStatusCode: http.StatusUnauthorized,
		},
		{
			name:               "should return status 401 when password is invalid",
			body:               `{"login":"user42","password":"invalidPassword"}`,
			useUserStorage:     true,
			expectedStatusCode: http.StatusUnauthorized,
		},
		{
			name:               "should return status 400 when request body is empty",
			body:               "",
			useUserStorage:     false,
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "should return status 400 when required login field is missing",
			body:               `{"login1":"user42","password":"password"}`,
			useUserStorage:     false,
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "should return status 400 when login field length is less than 4 characters",
			body:               `{"login":"usr","password":"password"}`,
			useUserStorage:     false,
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "should return status 400 when login field length is more than 30 characters",
			body:               `{"login":"user42user42user42user42user42user42","password":"password"}`,
			useUserStorage:     false,
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "should return status 400 when required password field is missing",
			body:               `{"login":"user42","password1":"password"}`,
			useUserStorage:     false,
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "should return status 400 when password field length is less than 8 characters",
			body:               `{"login":"user42","password":"word"}`,
			useUserStorage:     false,
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "should return status 400 when password field length is more than 256 characters",
			body:               `{"login":"user","password":"passwordpasswordpasswordpasswordpasswordpasswordpasswordpasswordpasswordpasswordpasswordpasswordpasswordpasswordpasswordpasswordpasswordpasswordpasswordpasswordpasswordpasswordpasswordpasswordpasswordpasswordpasswordpasswordpasswordpasswordpasswordpasswordpassword"}`,
			useUserStorage:     false,
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "should return status 500 when unexpected error occurred",
			body:               `{"login":"user42","password":"password"}`,
			useUserStorage:     true,
			userStorageErr:     errors.New("unexpected error"),
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

			handler := http.HandlerFunc(server.SignInHandler)

			hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
			require.NoError(t, err, "Error hashing password")
			if tt.useUserStorage {
				userStorage.EXPECT().GetOneByLogin(gomock.Any(), gomock.Any()).
					Return(model.User{ID: 1, Login: "user42", Password: string(hashedPassword)}, tt.userStorageErr)
			}

			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/api/user/login", strings.NewReader(tt.body))

			handler.ServeHTTP(w, req)

			resp := w.Result()
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedStatusCode, resp.StatusCode, "Response status code does not match expected status")
			if tt.expectAuthorizationHeader {
				assert.NotEmpty(t, resp.Header.Get("Authorization"), "Response header should contain Authorization header")
			} else {
				assert.Empty(t, resp.Header.Get("Authorization"), "Response header should not contain Authorization header")
			}
		})
	}
}
