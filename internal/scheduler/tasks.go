package scheduler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/cenkalti/backoff/v4"
	"go.uber.org/zap"

	er "github.com/Stern-Ritter/gophermart/internal/errors"
	"github.com/Stern-Ritter/gophermart/internal/model"
	"github.com/Stern-Ritter/gophermart/internal/utils"
)

const (
	taskCount = 1
)

func (s *AccrualsScheduler) RunTasks() {
	go func() {
		wg := sync.WaitGroup{}
		wg.Add(taskCount)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		s.startProcessingAccrualsWorkerPool()
		setInterval(ctx, &wg, s.getNewAccrualsInProcessing, time.Duration(s.getNewAccrualsInterval)*time.Second)

		wg.Wait()
	}()
}

func (s *AccrualsScheduler) StopTasks() {
	close(s.doneCh)
}

func (s *AccrualsScheduler) getNewAccrualsInProcessing() {
	accruals, err := s.accrualService.GetAllNewAccrualsInProcessingWithLimit(context.Background(),
		int64(s.processAccrualsBatchMaxSize))
	if err != nil {
		s.logger.Error("Error getting new accruals from database",
			zap.String("event", "getting new accruals"), zap.Error(err))
		return
	}
	if len(accruals) == 0 {
		s.logger.Info("No new accruals found",
			zap.String("event", "getting new accruals"))
		return
	}
	s.logger.Info("Success getting new accruals from database",
		zap.String("event", "getting new accruals"))

	select {
	case <-s.doneCh:
		close(s.processingAccrualsCh)
		s.logger.Info("Getting new accruals stopped",
			zap.String("event", "getting new accruals stopped"))
	case s.processingAccrualsCh <- accruals:
	}
}

func (s *AccrualsScheduler) startProcessingAccrualsWorkerPool() {
	processAccrualsWorkerPoolSize := s.processAccrualsWorkerPoolSize
	if processAccrualsWorkerPoolSize <= 0 {
		s.logger.Error("Process accruals worker pool size can't be less than or equal to zero",
			zap.String("event", "start send accruals worker pool"))
		processAccrualsWorkerPoolSize = 1
	}

	for w := 1; w <= processAccrualsWorkerPoolSize; w++ {
		go s.processingAccrualsWorker(w, s.processingAccrualsCh)
	}

	s.logger.Debug("Worker pool started",
		zap.String("event", "start send accruals worker pool"))
}

func (s *AccrualsScheduler) processingAccrualsWorker(id int, processingAccrualsCh <-chan []model.Accrual) {
	s.logger.Debug("Worker started", zap.Int("worker id", id),
		zap.String("event", "start processing accruals worker"))

	for accrualsBatch := range processingAccrualsCh {
		processedAccruals := make([]model.Accrual, 0)

		for _, accrual := range accrualsBatch {
			processingAccrual := func() error {
				endpoint := strings.Join([]string{"/api/orders", utils.FormatOrderNumber(accrual.OrderNumber)}, "/")
				resp, err := sendGetRequest(s.HTTPClient, endpoint)
				s.logger.Debug("Received response", zap.String("event", "received response"),
					zap.String("body", string(resp.Bytes())), zap.Error(err))

				if err != nil {
					return backoff.Permanent(err)
				}

				switch resp.StatusCode {
				case http.StatusOK:
					accrualProcessDto, err := decodeAccrualProcessDto(resp.Bytes())
					if err != nil {
						return backoff.Permanent(err)
					}
					switch accrualProcessDto.Status {
					case model.AccrualProcessInvalid, model.AccrualProcessProcessed:
						processedAccrual := model.UpdateAccrualFormAccrualProcessDto(accrual, accrualProcessDto)
						processedAccruals = append(processedAccruals, processedAccrual)
					case model.AccrualProcessRegistered, model.AccrualProcessProcessing:
						processedAccruals = append(processedAccruals, accrual)
					}
				case http.StatusNoContent:
					processedAccruals = append(processedAccruals, accrual)
				case http.StatusTooManyRequests, http.StatusInternalServerError:
					return er.NewRequestProcessingError(
						fmt.Sprintf("Unsuccess request sent on url: %s, status code: %d", endpoint, resp.StatusCode), nil)
				default:
					return backoff.Permanent(fmt.Errorf("unexpected response status code: %d", resp.StatusCode))
				}
				return nil
			}

			if sendErr := backoff.Retry(processingAccrual, s.processAccrualsRetryInterval); sendErr != nil {
				s.logger.Error("Error processing accrual", zap.Int("worker id", id),
					zap.Error(sendErr), zap.String("event", "processing accrual"))
				processedAccruals = append(processedAccruals, accrual)
				continue
			}

			s.logger.Debug("Processing accrual done", zap.Int("worker id", id),
				zap.String("event", "processing accrual"))
		}

		err := s.accrualService.UpdateAccruals(context.Background(), processedAccruals)
		if err != nil {
			s.logger.Error("Error saving processed accruals in database",
				zap.String("event", "saving processed accruals"), zap.Error(err))
			return
		}
		s.logger.Info("Success saving processed accruals in database",
			zap.String("event", "saving processed accruals"))
	}

	s.logger.Debug("Worker stopped", zap.Int("worker id", id),
		zap.String("event", "stop processing accruals worker"))
}

func decodeAccrualProcessDto(buf []byte) (model.AccrualProcessDto, error) {
	dto := model.AccrualProcessDto{}
	err := json.Unmarshal(buf, &dto)
	return dto, err
}
