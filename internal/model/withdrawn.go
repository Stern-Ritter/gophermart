package model

import (
	"time"

	"github.com/go-playground/validator/v10"

	"github.com/Stern-Ritter/gophermart/internal/utils"
	v "github.com/Stern-Ritter/gophermart/internal/validator"
)

type Withdrawn struct {
	UserID       int64
	OrderNumber  int64
	PointsAmount float64
	ProcessedAt  time.Time
}

type CreateWithdrawnDto struct {
	OrderNumber  string  `json:"order" validate:"required,numeric,order_number" msg:"Order should be correct numeric value"`
	PointsAmount float64 `json:"sum" validate:"required,gt=0" msg:"Sum should be greater than 0"`
}

func (s *CreateWithdrawnDto) Validate(validate *validator.Validate) error {
	return v.Validate[CreateWithdrawnDto](*s, validate)
}

type WithdrawnDto struct {
	OrderNumber  string  `json:"order"`
	PointsAmount float64 `json:"sum"`
	ProcessedAt  Time    `json:"processed_at"`
}

func NewWithdrawn(userID int64, orderNumber int64, pointsAmount float64) Withdrawn {
	return Withdrawn{
		UserID:       userID,
		OrderNumber:  orderNumber,
		PointsAmount: pointsAmount,
		ProcessedAt:  time.Now(),
	}
}

func ToWithdrawnDto(withdrawn Withdrawn) WithdrawnDto {
	return WithdrawnDto{
		OrderNumber:  utils.FormatOrderNumber(withdrawn.OrderNumber),
		PointsAmount: withdrawn.PointsAmount,
		ProcessedAt:  Time{withdrawn.ProcessedAt},
	}
}

func ToWithdrawalsDto(withdrawals []Withdrawn) []WithdrawnDto {
	withdrawalsResponse := make([]WithdrawnDto, len(withdrawals))
	for i, withdrawal := range withdrawals {
		withdrawalsResponse[i] = ToWithdrawnDto(withdrawal)
	}

	return withdrawalsResponse
}
