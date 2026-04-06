package handler

import (
	"encoding/json"
	"log"
	"net/http"
)

// JSONResponse はJSON形式でレスポンスを返す。
func JSONResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("json encode error: %v", err)
	}
}

// ErrorResponse はエラーレスポンスを返す。
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

// JSONError はエラーをJSON形式で返す。
func JSONError(w http.ResponseWriter, status int, message string) {
	JSONResponse(w, status, ErrorResponse{
		Error:   http.StatusText(status),
		Message: message,
	})
}
