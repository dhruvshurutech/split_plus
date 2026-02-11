package response

import (
	"encoding/json"
	"net/http"
)

type StandardResponse[TData any, TError any] struct {
	Status bool    `json:"status"`
	Data   *TData  `json:"data,omitempty"`
	Error  *TError `json:"error,omitempty"`
}

type ErrorDetail struct {
	Code    string      `json:"code,omitempty"`
	Message interface{} `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

func SendSuccess[TData any](w http.ResponseWriter, statusCode int, data TData) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(StandardResponse[TData, ErrorDetail]{
		Status: true,
		Data:   &data,
		Error:  nil,
	})
}

func SendError(w http.ResponseWriter, statusCode int, errorMessage string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	errorDetail := ErrorDetail{Message: errorMessage}
	_ = json.NewEncoder(w).Encode(StandardResponse[interface{}, ErrorDetail]{
		Status: false,
		Data:   nil,
		Error:  &errorDetail,
	})
}

func SendErrorWithCode(w http.ResponseWriter, statusCode int, code string, errorMessage string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	errorDetail := ErrorDetail{
		Code:    code,
		Message: errorMessage,
	}
	_ = json.NewEncoder(w).Encode(StandardResponse[interface{}, ErrorDetail]{
		Status: false,
		Data:   nil,
		Error:  &errorDetail,
	})
}

func SendErrorWithCodeAndDetails(w http.ResponseWriter, statusCode int, code string, errorMessage string, details interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	errorDetail := ErrorDetail{
		Code:    code,
		Message: errorMessage,
		Details: details,
	}
	_ = json.NewEncoder(w).Encode(StandardResponse[interface{}, ErrorDetail]{
		Status: false,
		Data:   nil,
		Error:  &errorDetail,
	})
}

func SendValidationErrors(w http.ResponseWriter, statusCode int, errorMessages []string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	errorDetail := ErrorDetail{Message: errorMessages}
	_ = json.NewEncoder(w).Encode(StandardResponse[interface{}, ErrorDetail]{
		Status: false,
		Data:   nil,
		Error:  &errorDetail,
	})
}
