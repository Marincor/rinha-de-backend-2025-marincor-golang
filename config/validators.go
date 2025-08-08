package config

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"

	validator "github.com/rezakhademix/govalidator/v2"
)

func validateStruct(validator *validator.Validator, s interface{}) {
	val := reflect.ValueOf(s)
	typ := reflect.TypeOf(s)

	// if pointer, get the value
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
		typ = typ.Elem()
	}

	//nolint:intrange // false positive
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := typ.Field(i)
		fieldName := fieldType.Tag.Get("json")

		if fieldName == "" {
			if fieldType.Tag.Get("optional") != "" {
				continue
			}

			fieldName = fieldType.Name
		}

		if field.Kind() == reflect.String {
			validator.RequiredString(field.String(), fieldName, "")
		}

		if field.Kind() == reflect.Int {
			validator.RequiredString(strconv.FormatInt(field.Int(), 10), fieldName, "")
		}

		if field.Kind() == reflect.Float64 {
			validator.RequiredString(strconv.FormatFloat(field.Float(), 'f', -1, 64), fieldName, "")
		}
	}
}

var ErrEnvironment = errors.New("environment validation failed")

func newValidateError(validationErrors map[string]string) error {
	return fmt.Errorf("%w: %v", ErrEnvironment, validationErrors)
}

func validate(config *Config) error {
	v := validator.New()

	validateStruct(&v, config)

	if v.IsFailed() {
		return newValidateError(v.Errors())
	}

	return nil
}
