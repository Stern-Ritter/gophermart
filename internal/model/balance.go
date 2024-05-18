package model

type Balance struct {
	UserID                int64
	CurrentPointsAmount   float64
	WithdrawnPointsAmount float64
}

type BalanceDto struct {
	CurrentPointsAmount   float64 `json:"current"`
	WithdrawnPointsAmount float64 `json:"withdrawn"`
}

func ToBalanceDto(balance Balance) BalanceDto {
	return BalanceDto{
		CurrentPointsAmount:   balance.CurrentPointsAmount,
		WithdrawnPointsAmount: balance.WithdrawnPointsAmount,
	}
}
