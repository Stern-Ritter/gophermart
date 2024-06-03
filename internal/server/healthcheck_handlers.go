package server

import "net/http"

func (s *Server) HealthcheckHandler(res http.ResponseWriter, req *http.Request) {
	res.WriteHeader(http.StatusOK)
	_, err := res.Write([]byte("OK"))
	if err != nil {
		http.Error(res, "Error writing response", http.StatusInternalServerError)
	}
}
