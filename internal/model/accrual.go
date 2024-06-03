package model

import (
	"time"

	"github.com/Stern-Ritter/gophermart/internal/utils"
)

type AccrualStatus string

const (
	AccrualNew        AccrualStatus = "NEW"
	AccrualProcessing AccrualStatus = "PROCESSING"
	AccrualInvalid    AccrualStatus = "INVALID"
	AccrualProcessed  AccrualStatus = "PROCESSED"
)

type Accrual struct {
	UserID       int64
	OrderNumber  int64
	Status       AccrualStatus
	PointsAmount float64
	UploadedAt   time.Time
	ProcessedAt  time.Time
}

type AccrualDto struct {
	OrderNumber  string        `json:"number"`
	Status       AccrualStatus `json:"status"`
	PointsAmount float64       `json:"accrual,omitempty"`
	UploadedAt   Time          `json:"uploaded_at"`
}

type AccrualProcessStatus string

const (
	AccrualProcessRegistered AccrualProcessStatus = "REGISTERED"
	AccrualProcessProcessing AccrualProcessStatus = "PROCESSING"
	AccrualProcessInvalid    AccrualProcessStatus = "INVALID"
	AccrualProcessProcessed  AccrualProcessStatus = "PROCESSED"
)

type AccrualProcessDto struct {
	OrderNumber string               `json:"order"`
	Status      AccrualProcessStatus `json:"status"`
	Accrual     float64              `json:"accrual"`
}

func NewAccrual(userID int64, orderNumber int64) Accrual {
	return Accrual{
		UserID:       userID,
		OrderNumber:  orderNumber,
		Status:       AccrualNew,
		PointsAmount: 0,
		UploadedAt:   time.Now(),
	}
}

func ToAccrualDto(accrual Accrual) AccrualDto {
	return AccrualDto{
		OrderNumber:  utils.FormatOrderNumber(accrual.OrderNumber),
		Status:       accrual.Status,
		PointsAmount: accrual.PointsAmount,
		UploadedAt:   Time{accrual.UploadedAt},
	}
}

func ToAccrualsDto(accruals []Accrual) []AccrualDto {
	accrualsResponse := make([]AccrualDto, len(accruals))
	for i, accrual := range accruals {
		accrualsResponse[i] = ToAccrualDto(accrual)
	}
	return accrualsResponse
}

func UpdateAccrualFormAccrualProcessDto(accrual Accrual, dto AccrualProcessDto) Accrual {
	return Accrual{
		UserID:       accrual.UserID,
		OrderNumber:  accrual.OrderNumber,
		Status:       mapAccrualProcessStatusToAccrualStatus(dto.Status),
		PointsAmount: dto.Accrual,
		UploadedAt:   accrual.UploadedAt,
		ProcessedAt:  time.Now(),
	}
}

func mapAccrualProcessStatusToAccrualStatus(status AccrualProcessStatus) AccrualStatus {
	switch status {
	case AccrualProcessInvalid:
		return AccrualInvalid
	case AccrualProcessProcessed:
		return AccrualProcessed
	default:
		return AccrualProcessing
	}
}
