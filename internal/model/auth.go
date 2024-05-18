package model

import (
	"github.com/go-playground/validator/v10"

	v "github.com/Stern-Ritter/gophermart/internal/validator"
)

type SignInRequest struct {
	Login    string `json:"login" validate:"required,min=5,max=30" msg:"Login length should be between 5 and 30 characters."`
	Password string `json:"password" validate:"required,min=8,max=256" msg:"Password length should be between 8 and 256 characters."`
}

func (s *SignInRequest) Validate(validate *validator.Validate) error {
	return v.Validate[SignInRequest](*s, validate)
}

type SignUpRequest struct {
	Login    string `json:"login" validate:"required,min=5,max=30" msg:"Login length should be between 5 and 30 characters."`
	Password string `json:"password" validate:"required,min=8,max=256" msg:"Password length should be between 8 and 256 characters."`
}

func (s *SignUpRequest) Validate(validate *validator.Validate) error {
	return v.Validate[SignUpRequest](*s, validate)
}

func SignUpRequestToUser(request SignUpRequest) User {
	return User{
		Login:    request.Login,
		Password: request.Password,
	}
}
