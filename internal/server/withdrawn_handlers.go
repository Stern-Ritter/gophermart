package server

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	er "github.com/Stern-Ritter/gophermart/internal/errors"
	"github.com/Stern-Ritter/gophermart/internal/model"
	"github.com/Stern-Ritter/gophermart/internal/utils"
)

func (s *Server) WithdrawLoyaltyPointsHandler(res http.ResponseWriter, req *http.Request) {
	currentUser, err := s.UserService.GetCurrentUser(req.Context())
	if err != nil {
		http.Error(res, err.Error(), http.StatusUnauthorized)
		return
	}

	createWithdrawnDto, err := decodeCreateWithdrawnDto(req.Body)
	if err != nil {
		http.Error(res, "Error decode request JSON body", http.StatusBadRequest)
		return
	}
	if err := createWithdrawnDto.Validate(s.Validate); err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}
	parsedOrderNumber, err := utils.ParseOrderNumber(createWithdrawnDto.OrderNumber)
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	withdrawn := model.NewWithdrawn(currentUser.ID, parsedOrderNumber, createWithdrawnDto.PointsAmount)
	err = s.WithdrawnService.CreateWithdrawn(req.Context(), withdrawn)
	if err != nil {
		var paymentRequiredError er.PaymentRequiredError
		switch {
		case errors.As(err, &paymentRequiredError):
			http.Error(res, err.Error(), http.StatusPaymentRequired)
			return
		}
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	res.WriteHeader(http.StatusOK)
}

func (s *Server) FindAllWithdrawalsByUserHandler(res http.ResponseWriter, req *http.Request) {
	currentUser, err := s.UserService.GetCurrentUser(req.Context())
	if err != nil {
		http.Error(res, err.Error(), http.StatusUnauthorized)
		return
	}

	withdrawals, err := s.WithdrawnService.GetAllWithdrawalsByUserID(req.Context(), currentUser.ID)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
	if len(withdrawals) == 0 {
		res.WriteHeader(http.StatusNoContent)
		return
	}

	withdrawalsDto := model.ToWithdrawalsDto(withdrawals)

	body, err := json.Marshal(withdrawalsDto)
	if err != nil {
		http.Error(res, "Error encoding response", http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-Type", "application/json")
	_, err = res.Write(body)
	if err != nil {
		http.Error(res, "Error encoding response", http.StatusInternalServerError)
		return
	}
}

func decodeCreateWithdrawnDto(source io.ReadCloser) (model.CreateWithdrawnDto, error) {
	dto := model.CreateWithdrawnDto{}
	var buf bytes.Buffer
	_, err := buf.ReadFrom(source)
	if err != nil {
		return dto, err
	}

	err = json.Unmarshal(buf.Bytes(), &dto)
	return dto, err
}
