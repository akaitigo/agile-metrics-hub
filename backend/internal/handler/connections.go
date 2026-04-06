package handler

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/akaitigo/agile-metrics-hub/internal/adapter"
)

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

	if req.DisplayName == "" {
		JSONError(w, http.StatusBadRequest, "display_name is required")
		return
	}

	if len(req.APIKey) > 256 {
		JSONError(w, http.StatusBadRequest, "api_key exceeds maximum length")
		return
	}

	a, err := h.Registry.Create(req.Source, req.APIKey, req.Config)
	if err != nil {
		JSONError(w, http.StatusBadRequest, "unsupported source: "+req.Source)
		return
	}

	if err := a.TestConnection(r.Context()); err != nil {
		// エラーメッセージをサニタイズ: 内部情報を漏洩しない
		log.Printf("connection test failed for source=%s: %v", req.Source, err)
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
