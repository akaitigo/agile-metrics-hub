package handler

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/akaitigo/agile-metrics-hub/internal/adapter"
)

// sanitizeLogValue はログインジェクション防止のため改行・制御文字を除去する。
func sanitizeLogValue(s string) string {
	r := strings.NewReplacer("\n", "", "\r", "", "\t", "")
	result := r.Replace(s)
	if len(result) > 64 {
		result = result[:64]
	}
	return result
}

const maxRequestBodySize = 1 << 16 // 64KB

// ConnectionsHandler は接続管理のHTTPハンドラー。
type ConnectionsHandler struct {
	Registry *adapter.Registry
}

// CreateConnectionRequest は接続作成リクエスト。
type CreateConnectionRequest struct {
	Source      string            `json:"source"`
	DisplayName string            `json:"display_name"`
	APIKey      string            `json:"api_key"`
	Config      map[string]string `json:"config"`
}

// TestConnection はAPIキーの有効性をテストする。
func (h *ConnectionsHandler) TestConnection(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBodySize)

	var req CreateConnectionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		JSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Source == "" || req.APIKey == "" {
		JSONError(w, http.StatusBadRequest, "source and api_key are required")
		return
	}

	if req.DisplayName == "" || len(req.DisplayName) > 128 {
		JSONError(w, http.StatusBadRequest, "display_name is required and must be 128 chars or less")
		return
	}

	if len(req.APIKey) > 256 {
		JSONError(w, http.StatusBadRequest, "api_key exceeds maximum length")
		return
	}

	// ソース名をサニタイズ（ログインジェクション防止）
	sanitizedSource := sanitizeLogValue(req.Source)

	a, err := h.Registry.Create(req.Source, req.APIKey, req.Config)
	if err != nil {
		JSONError(w, http.StatusBadRequest, "unsupported source")
		return
	}

	if err := a.TestConnection(r.Context()); err != nil {
		log.Printf("connection test failed for source=%s: %v", sanitizedSource, err)
		switch {
		case errors.Is(err, adapter.ErrUnauthorized):
			JSONError(w, http.StatusUnauthorized, "invalid credentials")
		case errors.Is(err, adapter.ErrRateLimited):
			JSONError(w, http.StatusTooManyRequests, "rate limit exceeded, try again later")
		default:
			JSONError(w, http.StatusBadGateway, "connection test failed")
		}
		return
	}

	JSONResponse(w, http.StatusOK, map[string]string{"status": "ok"})
}
