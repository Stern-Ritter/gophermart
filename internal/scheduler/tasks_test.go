package scheduler

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	gmock "gopkg.in/h2non/gentleman-mock.v2"
	"gopkg.in/h2non/gentleman.v2"

	"github.com/Stern-Ritter/gophermart/internal/logger"
	"github.com/Stern-Ritter/gophermart/internal/model"
	"github.com/Stern-Ritter/gophermart/internal/service"
)

func TestProcessingAccrualsWorker(t *testing.T) {
	processingAccruals := []model.Accrual{{
		UserID:       1,
		OrderNumber:  12345678903,
		Status:       model.AccrualNew,
		PointsAmount: 0,
		UploadedAt:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	}}

	tests := []struct {
		name               string
		processingAccruals []model.Accrual
		expectedAccruals   []model.Accrual
		responseStatus     int
		responseBody       string
	}{
		{
			name:               "should update accrual when response status code is 200 and status is 'PROCESSED'",
			processingAccruals: processingAccruals,
			responseStatus:     http.StatusOK,
			responseBody:       `{"order":"12345678903","status":"PROCESSED","accrual":500}`,

			expectedAccruals: []model.Accrual{{
				UserID:       1,
				OrderNumber:  12345678903,
				Status:       model.AccrualProcessed,
				PointsAmount: 500,
				UploadedAt:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			}},
		},
		{
			name:               "should not update accrual when response status code is 200 and status is 'REGISTERED'",
			processingAccruals: processingAccruals,
			responseStatus:     http.StatusOK,
			responseBody:       `{"order":"12345678903","status":"REGISTERED"}`,

			expectedAccruals: processingAccruals,
		},
		{
			name:               "should not update accrual when response status code is 200 and status is 'PROCESSING'",
			processingAccruals: processingAccruals,
			responseStatus:     http.StatusOK,
			responseBody:       `{"order":"12345678903","status":"PROCESSING"}`,
			expectedAccruals:   processingAccruals,
		},
		{
			name:               "should not update accrual when response status code is 204",
			processingAccruals: processingAccruals,
			responseStatus:     http.StatusNoContent,
			expectedAccruals:   processingAccruals,
		},
		{
			name:               "should not update accrual when response status code is 429",
			processingAccruals: processingAccruals,
			responseStatus:     http.StatusTooManyRequests,
			expectedAccruals:   processingAccruals,
		},
		{
			name:               "should not update accrual when response status code is 500",
			processingAccruals: processingAccruals,
			responseStatus:     http.StatusInternalServerError,
			expectedAccruals:   processingAccruals,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			defer gmock.Disable()
			gmock.New("").
				Get("/api/orders/12345678903").
				Reply(tt.responseStatus).
				Body(strings.NewReader(tt.responseBody))
			httpClient := gentleman.New()
			httpClient.Use(gmock.Plugin)

			logger, err := logger.Initialize("error")
			require.NoError(t, err, "Error init logger")

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			accrualStorage := NewMockAccrualStorage(ctrl)
			accrualService := service.NewAccrualService(accrualStorage, logger)

			s := &AccrualsScheduler{
				HTTPClient:     httpClient,
				accrualService: accrualService,
				processAccrualsRetryInterval: backoff.NewExponentialBackOff(
					backoff.WithInitialInterval(1*time.Millisecond),
					backoff.WithRandomizationFactor(0),
					backoff.WithMultiplier(1),
					backoff.WithMaxInterval(1*time.Millisecond),
					backoff.WithMaxElapsedTime(1*time.Millisecond)),
				logger: logger,
			}

			processingAccrualsCh := make(chan []model.Accrual, 1)

			processingAccrualsCh <- tt.processingAccruals
			close(processingAccrualsCh)

			var processedAccruals []model.Accrual
			accrualStorage.EXPECT().
				UpdateInBatch(gomock.Any(), gomock.Any()).
				Do(func(ctx context.Context, accruals []model.Accrual) {
					processedAccruals = accruals
				}).
				Return(nil)

			s.processingAccrualsWorker(1, processingAccrualsCh)
			equalAccruals(t, tt.expectedAccruals, processedAccruals)
		})
	}
}

func equalAccruals(t *testing.T, expectedAccruals []model.Accrual, gotAccruals []model.Accrual) bool {
	require.NotNil(t, expectedAccruals, "Expected accruals slice should not be nil")
	require.NotNil(t, gotAccruals, "Got accruals slice should not be nil")
	require.Equal(t, len(expectedAccruals), len(gotAccruals), "Got accruals length should be equal to expected accruals length")

	for i := range expectedAccruals {
		expectedAccrual := expectedAccruals[i]
		gotAccrual := gotAccruals[i]
		if gotAccrual.Status == model.AccrualProcessed {
			assert.Equal(t, expectedAccrual.UserID, gotAccrual.UserID, "Accruals user id should be equal")
			assert.Equal(t, expectedAccrual.OrderNumber, gotAccrual.OrderNumber, "Accruals order number should be equal")
			assert.Equal(t, expectedAccrual.Status, gotAccrual.Status, "Accruals status should be equal")
			assert.Equal(t, expectedAccrual.PointsAmount, gotAccrual.PointsAmount, "Accruals points amount should be equal")
			assert.Equal(t, expectedAccrual.UploadedAt, gotAccrual.UploadedAt, "Accruals uploaded should be equal")
			assert.True(t, isWithinLastTenMinutes(gotAccrual.ProcessedAt), "Accruals should have processed at last 10 minutes")
		} else {
			assert.Equal(t, expectedAccrual, gotAccrual, "Accruals should be equal")
		}
	}
	return true
}

func isWithinLastTenMinutes(t time.Time) bool {
	now := time.Now()
	tenMinutesAgo := now.Add(-10 * time.Minute)
	return t.After(tenMinutesAgo) && t.Before(now)
}
