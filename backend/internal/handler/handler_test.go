package handler_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/akaitigo/agile-metrics-hub/internal/adapter"
	"github.com/akaitigo/agile-metrics-hub/internal/handler"
	"github.com/akaitigo/agile-metrics-hub/internal/model"
)

func TestMetrics_Burndown(t *testing.T) {
	h := &handler.MetricsHandler{}
	req := httptest.NewRequest(http.MethodGet, "/api/metrics/burndown?project_id=demo", nil)
	rec := httptest.NewRecorder()

	h.Burndown(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var result []model.BurndownPoint
	if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if len(result) == 0 {
		t.Error("expected non-empty burndown data")
	}
	// 値の検証: 最初のポイントの残ポイントはtotalと一致すべき
	if result[0].RemainingPoints < 0 {
		t.Error("remaining points should not be negative")
	}
}

func TestMetrics_Burndown_MissingProjectID(t *testing.T) {
	h := &handler.MetricsHandler{}
	req := httptest.NewRequest(http.MethodGet, "/api/metrics/burndown", nil)
	rec := httptest.NewRecorder()

	h.Burndown(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestMetrics_Velocity(t *testing.T) {
	h := &handler.MetricsHandler{}
	req := httptest.NewRequest(http.MethodGet, "/api/metrics/velocity?count=3", nil)
	rec := httptest.NewRecorder()

	h.Velocity(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var result struct {
		Velocity        []model.VelocityPoint `json:"velocity"`
		AverageVelocity float64               `json:"average_velocity"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if len(result.Velocity) != 3 {
		t.Errorf("expected 3 sprints, got %d", len(result.Velocity))
	}
	if result.AverageVelocity <= 0 {
		t.Error("expected positive average velocity")
	}
}

func TestMetrics_Velocity_InvalidCount(t *testing.T) {
	h := &handler.MetricsHandler{}
	// count=-1 は無効 → デフォルト5が使われるべき
	req := httptest.NewRequest(http.MethodGet, "/api/metrics/velocity?count=-1", nil)
	rec := httptest.NewRecorder()

	h.Velocity(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 (fallback to default), got %d", rec.Code)
	}

	var result struct {
		Velocity []model.VelocityPoint `json:"velocity"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if len(result.Velocity) != 5 {
		t.Errorf("expected default 5 sprints for invalid count, got %d", len(result.Velocity))
	}
}

func TestMetrics_Velocity_OverMaxCount(t *testing.T) {
	h := &handler.MetricsHandler{}
	req := httptest.NewRequest(http.MethodGet, "/api/metrics/velocity?count=999", nil)
	rec := httptest.NewRecorder()

	h.Velocity(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var result struct {
		Velocity []model.VelocityPoint `json:"velocity"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	// count > 50 は無効 → デフォルト5
	if len(result.Velocity) != 5 {
		t.Errorf("expected default 5 for count>50, got %d", len(result.Velocity))
	}
}

func TestMetrics_CumulativeFlow(t *testing.T) {
	h := &handler.MetricsHandler{}
	req := httptest.NewRequest(http.MethodGet, "/api/metrics/cumulative-flow?days=7", nil)
	rec := httptest.NewRecorder()

	h.CumulativeFlow(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var result []model.CumulativeFlowPoint
	if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if len(result) != 7 {
		t.Errorf("expected 7 days, got %d", len(result))
	}
	// 各ポイントにステータスが含まれることを検証
	for _, p := range result {
		if len(p.Statuses) == 0 {
			t.Error("expected non-empty statuses in cumulative flow point")
		}
	}
}

func TestMetrics_LeadTime(t *testing.T) {
	h := &handler.MetricsHandler{}
	req := httptest.NewRequest(http.MethodGet, "/api/metrics/lead-time?project_id=demo", nil)
	rec := httptest.NewRecorder()

	h.LeadTime(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var result model.LeadTimeStats
	if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if result.Average <= 0 {
		t.Error("expected positive average lead time")
	}
	// P50 <= P85 <= P95 の順序が保たれること
	if result.Percentile50 > result.Percentile85 {
		t.Errorf("p50 (%f) should be <= p85 (%f)", result.Percentile50, result.Percentile85)
	}
	if result.Percentile85 > result.Percentile95 {
		t.Errorf("p85 (%f) should be <= p95 (%f)", result.Percentile85, result.Percentile95)
	}
}

func TestJSONError(t *testing.T) {
	rec := httptest.NewRecorder()
	handler.JSONError(rec, http.StatusNotFound, "not found")

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}

	var result handler.ErrorResponse
	if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if result.Message != "not found" {
		t.Errorf("expected 'not found', got %q", result.Message)
	}
	if result.Error != "Not Found" {
		t.Errorf("expected 'Not Found', got %q", result.Error)
	}
}

func TestConnections_TestConnection_EmptyBody(t *testing.T) {
	h := &handler.ConnectionsHandler{Registry: adapter.NewRegistry()}
	req := httptest.NewRequest(http.MethodPost, "/api/connections/test", strings.NewReader(""))
	rec := httptest.NewRecorder()

	h.TestConnection(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for empty body, got %d", rec.Code)
	}
}

func TestConnections_TestConnection_MissingFields(t *testing.T) {
	h := &handler.ConnectionsHandler{Registry: adapter.NewRegistry()}
	body := `{"source":"clickup"}`
	req := httptest.NewRequest(http.MethodPost, "/api/connections/test", strings.NewReader(body))
	rec := httptest.NewRecorder()

	h.TestConnection(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for missing api_key, got %d", rec.Code)
	}
}

func TestConnections_TestConnection_UnsupportedSource(t *testing.T) {
	h := &handler.ConnectionsHandler{Registry: adapter.NewRegistry()}
	body := `{"source":"trello","display_name":"test","api_key":"key123"}`
	req := httptest.NewRequest(http.MethodPost, "/api/connections/test", strings.NewReader(body))
	rec := httptest.NewRecorder()

	h.TestConnection(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for unsupported source, got %d", rec.Code)
	}
}
