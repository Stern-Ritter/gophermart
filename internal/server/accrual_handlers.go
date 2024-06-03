package server

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	er "github.com/Stern-Ritter/gophermart/internal/errors"
	"github.com/Stern-Ritter/gophermart/internal/model"
	"github.com/Stern-Ritter/gophermart/internal/utils"
)

func (s *Server) LoadOrderHandler(res http.ResponseWriter, req *http.Request) {
	currentUser, err := s.UserService.GetCurrentUser(req.Context())
	if err != nil {
		http.Error(res, err.Error(), http.StatusUnauthorized)
		return
	}

	body, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(res, "Error decode request plain/text body", http.StatusBadRequest)
	}
	orderNumber := string(body)
	parsedOrderNumber, err := utils.ParseOrderNumber(orderNumber)
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}
	if err := utils.ValidateOrderNumber(orderNumber); err != nil {
		http.Error(res, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	accrual := model.NewAccrual(currentUser.ID, parsedOrderNumber)
	err = s.AccrualService.CreateAccrual(req.Context(), accrual)
	if err != nil {
		var alreadyExistsError er.AlreadyExistsError
		var conflictError er.ConflictError
		switch {
		case errors.As(err, &alreadyExistsError):
			http.Error(res, err.Error(), http.StatusOK)
			return
		case errors.As(err, &conflictError):
			http.Error(res, err.Error(), http.StatusConflict)
			return
		}
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	res.WriteHeader(http.StatusAccepted)
}

func (s *Server) FindAllOrdersLoadedByUserHandler(res http.ResponseWriter, req *http.Request) {
	currentUser, err := s.UserService.GetCurrentUser(req.Context())
	if err != nil {
		http.Error(res, err.Error(), http.StatusUnauthorized)
		return
	}

	accruals, err := s.AccrualService.GetAllAccrualsByUserID(req.Context(), currentUser.ID)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
	if len(accruals) == 0 {
		res.WriteHeader(http.StatusNoContent)
		return
	}

	accrualsDto := model.ToAccrualsDto(accruals)

	body, err := json.Marshal(accrualsDto)
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
