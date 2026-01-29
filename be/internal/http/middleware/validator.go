package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/dhruvsaxena1998/splitplus/internal/http/response"
	"github.com/go-playground/validator/v10"
)

type ctxKey[T any] struct{}

func GetBody[T any](r *http.Request) (T, bool) {
	body, ok := r.Context().Value(ctxKey[T]{}).(T)
	return body, ok
}

func DecodeBody[T any](r *http.Request, v *T) error {
	decoder := json.NewDecoder(r.Body)
	return decoder.Decode(v)
}

func ValidateBody[T any](
	v *validator.Validate,
) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var body T

			decoder := json.NewDecoder(r.Body)
			decoder.DisallowUnknownFields()

			if err := decoder.Decode(&body); err != nil {
				response.SendError(w, http.StatusBadRequest, fmt.Sprintf("invalid request body: %s", err.Error()))
				return
			}

			if err := v.Struct(body); err != nil {
				var validationErrors []string
				if validatorErrors, ok := err.(validator.ValidationErrors); ok {
					for _, fieldError := range validatorErrors {
						validationErrors = append(validationErrors, fmt.Sprintf("%s: %s", fieldError.Field(), getValidationErrorMessage(fieldError)))
					}
				} else {
					validationErrors = append(validationErrors, err.Error())
				}
				response.SendValidationErrors(w, http.StatusUnprocessableEntity, validationErrors)
				return
			}

			ctx := context.WithValue(r.Context(), ctxKey[T]{}, body)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func getValidationErrorMessage(fieldError validator.FieldError) string {
	switch fieldError.Tag() {
	case "required":
		return "is required"
	case "email":
		return "must be a valid email address"
	case "min":
		return fmt.Sprintf("must be at least %s characters", fieldError.Param())
	case "max":
		return fmt.Sprintf("must be at most %s characters", fieldError.Param())
	default:
		return fmt.Sprintf("failed validation for '%s'", fieldError.Tag())
	}
}
