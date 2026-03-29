package httpresponse

import (
	"encoding/json"
	"net/http"
)

type ErrorResponse struct {
	Error   string `json:"error" example:"invalid request"`
	Details string `json:"details,omitempty" example:"end_date must be greater than or equal to start_date"`
}

func JSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func Error(w http.ResponseWriter, status int, errMsg string, details string) {
	JSON(w, status, ErrorResponse{
		Error:   errMsg,
		Details: details,
	})
}
