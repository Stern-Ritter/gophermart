package server

import (
	"encoding/json"
	"net/http"

	"github.com/Stern-Ritter/gophermart/internal/model"
)

func (s *Server) GetLoyaltyPointsBalanceHandler(res http.ResponseWriter, req *http.Request) {
	currentUser, err := s.UserService.GetCurrentUser(req.Context())
	if err != nil {
		http.Error(res, err.Error(), http.StatusUnauthorized)
		return
	}

	balance, err := s.BalanceService.GetBalanceByUserID(req.Context(), currentUser.ID)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	balanceDto := model.ToBalanceDto(balance)

	body, err := json.Marshal(balanceDto)
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
