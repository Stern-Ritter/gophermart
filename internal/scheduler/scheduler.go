package scheduler

import (
	"time"

	"github.com/cenkalti/backoff/v4"
	"gopkg.in/h2non/gentleman.v2"

	"github.com/Stern-Ritter/gophermart/internal/logger"
	"github.com/Stern-Ritter/gophermart/internal/model"
	"github.com/Stern-Ritter/gophermart/internal/service"
)

type Scheduler interface {
	RunTasks()
	StopTasks()
}

type AccrualsScheduler struct {
	HTTPClient                    *gentleman.Client
	accrualService                service.AccrualService
	processingAccrualsCh          chan []model.Accrual
	doneCh                        chan struct{}
	processAccrualsBatchMaxSize   int
	processAccrualsWorkerPoolSize int
	getNewAccrualsInterval        int
	processAccrualsRetryInterval  *backoff.ExponentialBackOff
	logger                        *logger.ServerLogger
}

func NewAccrualsScheduler(accrualService service.AccrualService,
	processAccrualsSystemURL string, processAccrualsBufferSize int, processAccrualsBatchMaxSize int,
	processAccrualsWorkerPoolSize int, getNewAccrualsInterval int, logger *logger.ServerLogger) Scheduler {

	httpClient := gentleman.New()
	httpClient.URL(processAccrualsSystemURL)

	processingAccrualsCh := make(chan []model.Accrual, processAccrualsBufferSize)
	doneCh := make(chan struct{})

	processAccrualsRetryInterval := backoff.NewExponentialBackOff(
		backoff.WithInitialInterval(1*time.Second),
		backoff.WithRandomizationFactor(0),
		backoff.WithMultiplier(5),
		backoff.WithMaxInterval(60*time.Second),
		backoff.WithMaxElapsedTime(120*time.Second))

	return &AccrualsScheduler{
		HTTPClient:                    httpClient,
		accrualService:                accrualService,
		processingAccrualsCh:          processingAccrualsCh,
		doneCh:                        doneCh,
		processAccrualsBatchMaxSize:   processAccrualsBatchMaxSize,
		processAccrualsWorkerPoolSize: processAccrualsWorkerPoolSize,
		getNewAccrualsInterval:        getNewAccrualsInterval,
		processAccrualsRetryInterval:  processAccrualsRetryInterval,
		logger:                        logger,
	}
}
