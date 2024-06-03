package server

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	er "github.com/Stern-Ritter/gophermart/internal/errors"
	"github.com/Stern-Ritter/gophermart/internal/model"
)

func (s *Server) SignUpHandler(res http.ResponseWriter, req *http.Request) {
	signUpRequest := model.SignUpRequest{}
	err := decodeWithUnknownAndDuplicateFieldsCheck(req.Body, &signUpRequest)
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
	signInRequest := model.SignInRequest{}
	err := decodeWithUnknownAndDuplicateFieldsCheck(req.Body, &signInRequest)
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

func decodeWithUnknownAndDuplicateFieldsCheck(source io.ReadCloser, target any) error {
	var buf bytes.Buffer
	reader := io.TeeReader(source, &buf)

	dec := json.NewDecoder(reader)
	err := checkDuplicateKeys(dec)
	if err != nil {
		return err
	}

	dec = json.NewDecoder(&buf)
	dec.DisallowUnknownFields()
	return dec.Decode(target)
}

func checkDuplicateKeys(dec *json.Decoder) error {
	t, err := dec.Token()
	if err != nil {
		return err
	}

	delim, ok := t.(json.Delim)
	if !ok {
		return nil
	}

	switch delim {
	case '{':
		keys := make(map[string]bool)
		for dec.More() {
			t, err := dec.Token()
			if err != nil {
				return err
			}
			key := t.(string)

			if keys[key] {
				return fmt.Errorf("duplicate key found: %s", key)
			}
			keys[key] = true

			if err := checkDuplicateKeys(dec); err != nil {
				return err
			}
		}

		if _, err := dec.Token(); err != nil {
			return err
		}

	case '[':
		for dec.More() {
			if err := checkDuplicateKeys(dec); err != nil {
				return err
			}
		}

		if _, err := dec.Token(); err != nil {
			return err
		}

	}
	return nil
}
