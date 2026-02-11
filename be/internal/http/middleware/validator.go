package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

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

func ValidateBodyWithScope[T any](
	v *validator.Validate,
	scope string,
) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var body T

			decoder := json.NewDecoder(r.Body)
			decoder.DisallowUnknownFields()

			if err := decoder.Decode(&body); err != nil {
				response.SendErrorWithCode(
					w,
					http.StatusBadRequest,
					fmt.Sprintf("validation.%s.body.invalid", strings.ToLower(scope)),
					"Invalid request body.",
				)
				return
			}

			if err := v.Struct(body); err != nil {
				if validatorErrors, ok := err.(validator.ValidationErrors); ok {
					firstValidationError := validatorErrors[0]
					field := strings.ToLower(firstValidationError.Field())
					rule := strings.ToLower(firstValidationError.Tag())
					code := fmt.Sprintf("validation.%s.%s.%s", strings.ToLower(scope), field, rule)

					response.SendErrorWithCodeAndDetails(
						w,
						http.StatusUnprocessableEntity,
						code,
						getValidationErrorMessage(firstValidationError),
						map[string]string{
							"field": field,
							"rule":  rule,
						},
					)
				} else {
					response.SendErrorWithCode(
						w,
						http.StatusUnprocessableEntity,
						fmt.Sprintf("validation.%s.invalid", strings.ToLower(scope)),
						"Invalid request payload.",
					)
				}
				return
			}

			ctx := context.WithValue(r.Context(), ctxKey[T]{}, body)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func getValidationErrorMessage(fieldError validator.FieldError) string {
	fieldName := fieldError.Field()
	switch fieldError.Tag() {
	case "required":
		return fmt.Sprintf("%s is required.", fieldName)
	case "email":
		return fmt.Sprintf("%s must be a valid email address.", fieldName)
	case "min":
		return fmt.Sprintf("%s must be at least %s characters.", fieldName, fieldError.Param())
	case "max":
		return fmt.Sprintf("%s must be at most %s characters.", fieldName, fieldError.Param())
	default:
		return fmt.Sprintf("%s failed validation for '%s'.", fieldName, fieldError.Tag())
	}
}
