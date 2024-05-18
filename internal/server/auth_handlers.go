package server

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	er "github.com/Stern-Ritter/gophermart/internal/errors"
	"github.com/Stern-Ritter/gophermart/internal/model"
)

func (s *Server) SignUpHandler(res http.ResponseWriter, req *http.Request) {
	signUpRequest, err := decodeSignUpRequest(req.Body)
	if err != nil {
		http.Error(res, "Error decode request JSON body", http.StatusBadRequest)
		return
	}

	if err := signUpRequest.Validate(s.Validate); err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	tokenString, err := s.AuthService.SignUp(req.Context(), signUpRequest)
	if err != nil {
		var conflictError er.ConflictError
		if errors.As(err, &conflictError) {
			http.Error(res, conflictError.Error(), http.StatusConflict)
			return
		}
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	res.Header().Set("Authorization", tokenString)
	res.WriteHeader(http.StatusOK)
}

func (s *Server) SignInHandler(res http.ResponseWriter, req *http.Request) {
	signInRequest, err := decodeSignInRequest(req.Body)
	if err != nil {
		http.Error(res, "Error decode request JSON body", http.StatusBadRequest)
		return
	}

	if err := signInRequest.Validate(s.Validate); err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	tokenString, err := s.AuthService.SignIn(req.Context(), signInRequest)
	if err != nil {
		var unauthorizedError er.UnauthorizedError
		if errors.As(err, &unauthorizedError) {
			http.Error(res, unauthorizedError.Error(), http.StatusUnauthorized)
			return
		}
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	res.Header().Set("Authorization", tokenString)
	res.WriteHeader(http.StatusOK)
}

func decodeSignUpRequest(source io.ReadCloser) (model.SignUpRequest, error) {
	request := model.SignUpRequest{}
	var buf bytes.Buffer
	_, err := buf.ReadFrom(source)
	if err != nil {
		return request, err
	}

	err = json.Unmarshal(buf.Bytes(), &request)
	return request, err
}

func decodeSignInRequest(source io.ReadCloser) (model.SignInRequest, error) {
	request := model.SignInRequest{}
	var buf bytes.Buffer
	_, err := buf.ReadFrom(source)
	if err != nil {
		return request, err
	}

	err = json.Unmarshal(buf.Bytes(), &request)
	return request, err
}
