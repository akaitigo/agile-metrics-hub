package handler

import (
	"encoding/json"
	"net/http"

	"github.com/akaitigo/agile-metrics-hub/internal/adapter"
)

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

	a, err := h.Registry.Create(req.Source, req.APIKey, req.Config)
	if err != nil {
		JSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := a.TestConnection(r.Context()); err != nil {
		JSONError(w, http.StatusUnauthorized, "connection test failed: "+err.Error())
		return
	}

	JSONResponse(w, http.StatusOK, map[string]string{"status": "ok"})
}
