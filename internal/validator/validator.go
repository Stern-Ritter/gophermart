package validator

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/go-playground/validator/v10"

	"github.com/Stern-Ritter/gophermart/internal/utils"
)

func GetValidator() (*validator.Validate, error) {
	validate := validator.New(validator.WithRequiredStructEnabled())
	err := validate.RegisterValidation("order_number", func(fl validator.FieldLevel) bool {
		value := fl.Field().String()
		return utils.ValidateOrderNumber(value) == nil
	})
	return validate, err
}

const errMsgTag = "msg"

func Validate[T interface{}](obj interface{}, validate *validator.Validate) (errs error) {
	o := obj.(T)

	defer func() {
		if r := recover(); r != nil {
			errs = fmt.Errorf("can't validate: %+v", r)
		}
	}()

	if err := validate.Struct(o); err != nil {
		errorValid := err.(validator.ValidationErrors)
		for _, e := range errorValid {
			errMsg := errorTagFunc(obj, e.Field(), errMsgTag)
			if errMsg != nil {
				errs = errors.Join(errs, fmt.Errorf("%w", errMsg))
			} else {
				errs = errors.Join(errs, fmt.Errorf("%w", e))
			}
		}
	}
	return
}

func errorTagFunc(obj any, fieldName string, tagName string) error {
	val := reflect.ValueOf(obj)
	for i := 0; i < val.NumField(); i++ {
		if val.Type().Field(i).Name == fieldName {
			tags := val.Type().Field(i).Tag
			customMessage := tags.Get(tagName)
			if customMessage != "" {
				return fmt.Errorf(customMessage)
			}
			return nil
		}
	}
	return nil
}
