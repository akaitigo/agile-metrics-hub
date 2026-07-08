package jira_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/akaitigo/agile-metrics-hub/internal/adapter"
	"github.com/akaitigo/agile-metrics-hub/internal/adapter/jira"
)

func TestNewAdapter_MissingConfig(t *testing.T) {
	_, err := jira.NewAdapter("token", nil)
	if err != adapter.ErrUnauthorized {
		t.Errorf("expected ErrUnauthorized, got %v", err)
	}
}

func TestNewAdapter_EmptyToken(t *testing.T) {
	_, err := jira.NewAdapter("", map[string]string{"base_url": "https://test.atlassian.net", "email": "test@example.com"})
	if err != adapter.ErrUnauthorized {
		t.Errorf("expected ErrUnauthorized, got %v", err)
	}
}

func TestNewAdapter_Valid(t *testing.T) {
	a, err := jira.NewAdapter("token", map[string]string{
		"base_url": "https://test.atlassian.net",
		"email":    "test@example.com",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a.Name() != "jira" {
		t.Errorf("expected name 'jira', got %q", a.Name())
	}
}

func TestNewAdapter_InvalidStoryPointsField(t *testing.T) {
	_, err := jira.NewAdapter("token", map[string]string{
		"base_url":           "https://test.atlassian.net",
		"email":              "test@example.com",
		"story_points_field": "field;DROP TABLE",
	})
	if err == nil {
		t.Fatal("expected error for invalid field name")
	}
}

func TestNewAdapter_ValidStoryPointsField(t *testing.T) {
	a, err := jira.NewAdapter("token", map[string]string{
		"base_url":           "https://test.atlassian.net",
		"email":              "test@example.com",
		"story_points_field": "customfield_10020",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a.Name() != "jira" {
		t.Errorf("expected name 'jira', got %q", a.Name())
	}
}

func TestNewAdapter_SSRF_HttpBlocked(t *testing.T) {
	_, err := jira.NewAdapter("token", map[string]string{
		"base_url": "http://test.atlassian.net",
		"email":    "test@example.com",
	})
	if err == nil {
		t.Fatal("expected error for http:// URL")
	}
}

func TestNewAdapter_SSRF_NonAtlassianBlocked(t *testing.T) {
	_, err := jira.NewAdapter("token", map[string]string{
		"base_url": "https://169.254.169.254",
		"email":    "test@example.com",
	})
	if err == nil {
		t.Fatal("expected error for non-atlassian domain")
	}
}

func TestNewAdapter_SSRF_LocalhostBlocked(t *testing.T) {
	_, err := jira.NewAdapter("token", map[string]string{
		"base_url": "https://localhost",
		"email":    "test@example.com",
	})
	if err == nil {
		t.Fatal("expected error for localhost")
	}
}

func TestAdapter_TestConnection_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/rest/api/3/myself" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		user, pass, ok := r.BasicAuth()
		if !ok || user != "test@example.com" || pass != "token" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]string{"displayName": "Test"})
	}))
	defer srv.Close()

	a := jira.NewAdapterWithBaseURL(srv.URL, "test@example.com", "token")
	if err := a.TestConnection(context.Background()); err != nil {
		t.Fatalf("TestConnection failed: %v", err)
	}
}

func TestAdapter_TestConnection_Unauthorized(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer srv.Close()

	a := jira.NewAdapterWithBaseURL(srv.URL, "bad@example.com", "bad")
	err := a.TestConnection(context.Background())
	if err != adapter.ErrUnauthorized {
		t.Errorf("expected ErrUnauthorized, got %v", err)
	}
}

func TestAdapter_FetchTasks_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		resp := map[string]interface{}{
			"issues": []map[string]interface{}{
				{
					"id":  "10001",
					"key": "PROJ-1",
					"fields": map[string]interface{}{
						"summary":        "Test Issue",
						"status":         map[string]string{"name": "To Do"},
						"labels":         []string{"bug"},
						"created":        "2026-04-01T10:00:00.000+0900",
						"updated":        "2026-04-02T10:00:00.000+0900",
						"resolutiondate": nil,
					},
				},
			},
			"total": 1,
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	a := jira.NewAdapterWithBaseURL(srv.URL, "test@example.com", "token")
	tasks, err := a.FetchTasks(context.Background(), "1")
	if err != nil {
		t.Fatalf("FetchTasks failed: %v", err)
	}
	if len(tasks) != 1 {
		t.Fatalf("expected 1 task, got %d", len(tasks))
	}
	if tasks[0].Title != "Test Issue" {
		t.Errorf("expected 'Test Issue', got %q", tasks[0].Title)
	}
	if tasks[0].ExternalID != "PROJ-1" {
		t.Errorf("expected 'PROJ-1', got %q", tasks[0].ExternalID)
	}
}
