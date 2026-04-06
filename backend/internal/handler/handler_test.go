package handler_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

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
}
